// Contains functions related to chart pulling
package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	helmrepo "helm.sh/helm/v3/pkg/repo"
)

// PullLegacyChart pulls a Helm chart from a legacy repository
func (h *HelmHandler) PullLegacyChart(repo, chart, version, outputPath, username, password string) error {
	log.Info().Str("chart", chart).Str("version", version).Msg("Attempting to pull legacy Helm chart")

	// Resolve the repository name
	repoName, exists := h.RepoNames[repo]
	if !exists {
		log.Error().Str("repo", repo).Msg("Repository not found in RepoNames map")
		return fmt.Errorf("repository not found: %s", repo)
	}

	// Load the index file, in AddAndFetchRepo When ensure it already exists
	indexFile := filepath.Join(h.Settings.RepositoryCache, fmt.Sprintf("%s-index.yaml", repoName))
	index, err := helmrepo.LoadIndexFile(indexFile)
	if err != nil {
		log.Error().Err(err).Str("indexFile", indexFile).Msg("Failed to load repository index file")
		return err
	}

	// Find the chart version
	chartVersion, err := index.Get(chart, version)
	if err != nil {
		log.Error().Err(err).Str("chart", chart).Str("version", version).Msg("Chart version not found in index")
		return err
	}

	// Resolve the chart URL
	chartURL := chartVersion.URLs[0]
	if !strings.HasPrefix(chartURL, "http") {
		chartURL = fmt.Sprintf("%s/%s", strings.TrimSuffix(repo, "/"), chartURL)
	}

	log.Info().Str("chartURL", chartURL).Msg("Resolved chart URL")

	// Prepare the HTTP client and request
	client := &http.Client{}
	req, err := http.NewRequest("GET", chartURL, nil)
	if err != nil {
		log.Error().Err(err).Str("chartURL", chartURL).Msg("Failed to create request for legacy chart")
		return err
	}

	// Add authentication if credentials are provided
	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Str("chartURL", chartURL).Msg("Failed to fetch legacy chart")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error().Int("statusCode", resp.StatusCode).Str("chartURL", chartURL).Msg("Unexpected status code while fetching legacy chart")
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Save the chart to disk
	chartFile := filepath.Join(outputPath, filepath.Base(chartURL))
	if err := os.MkdirAll(outputPath, 0o755); err != nil {
		log.Error().Err(err).Str("outputPath", outputPath).Msg("Failed to create output directory")
		return err
	}
	file, err := os.Create(chartFile)
	if err != nil {
		log.Error().Err(err).Str("chartFile", chartFile).Msg("Failed to create chart file")
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		log.Error().Err(err).Str("chartFile", chartFile).Msg("Failed to write legacy chart data to file")
		return err
	}

	log.Info().Str("chartFile", chartFile).Msg("Successfully pulled legacy Helm chart")
	return nil
}

// PullOCIChart pulls a Helm chart from an OCI-compliant repository
func (h *HelmHandler) PullOCIChart(repo, chart, version, outputPath string) error {
	// Construct the OCI reference
	chartRef := fmt.Sprintf("%s/%s:%s", repo, chart, version)

	log.Info().Str("chartRef", chartRef).Msg("Attempting to pull OCI Helm chart")

	// Pull the chart using the Helm Registry Client
	pullResult, err := h.RegistryClient.Pull(chartRef)
	if err != nil {
		log.Error().Err(err).Str("chartRef", chartRef).Msg("Failed to pull OCI Helm chart")
		return err
	}
	log.Info().Str("chartRef", chartRef).Msg("Successfully pulled OCI Helm chart")

	// Save the chart to disk
	if err := os.MkdirAll(outputPath, 0o755); err != nil {
		log.Error().Err(err).Str("outputPath", outputPath).Msg("Failed to create output directory")
		return err
	}

	chartFile := filepath.Join(outputPath, fmt.Sprintf("%s-%s.tgz", chart, version))
	if err := os.WriteFile(chartFile, pullResult.Chart.Data, 0o644); err != nil {
		log.Error().Err(err).Str("chartFile", chartFile).Msg("Failed to write OCI chart to file")
		return err
	}

	log.Info().Str("chartFile", chartFile).Msg("Successfully saved OCI Helm chart to disk")
	return nil
}
