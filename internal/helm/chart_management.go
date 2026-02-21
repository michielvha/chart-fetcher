package helm

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	helmrepo "helm.sh/helm/v3/pkg/repo"
)

// PullLegacyChart pulls a Helm chart from a legacy HTTP repository.
func (h *HelmHandler) PullLegacyChart(repo, chart, version, outputPath, username, password string) error {
	log.Info().Str("chart", chart).Str("version", version).Msg("Attempting to pull legacy Helm chart")

	repoName, exists := h.RepoNames[repo]
	if !exists {
		log.Error().Str("repo", repo).Msg("Repository not found in RepoNames map")
		return fmt.Errorf("repository not found: %s", repo)
	}

	indexFile := filepath.Join(h.Settings.RepositoryCache, fmt.Sprintf("%s-index.yaml", repoName))
	index, err := helmrepo.LoadIndexFile(indexFile)
	if err != nil {
		log.Error().Err(err).Str("indexFile", indexFile).Msg("Failed to load repository index file")
		return err
	}

	chartVersion, err := index.Get(chart, version)
	if err != nil {
		log.Error().Err(err).Str("chart", chart).Str("version", version).Msg("Chart version not found in index")
		return err
	}

	chartURL := chartVersion.URLs[0]
	if !strings.HasPrefix(chartURL, "http") {
		chartURL = fmt.Sprintf("%s/%s", strings.TrimSuffix(repo, "/"), chartURL)
	}

	log.Info().Str("chartURL", chartURL).Msg("Resolved chart URL")

	parsedURL, err := url.ParseRequestURI(chartURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		log.Error().Str("chartURL", chartURL).Msg("Invalid or disallowed chart URL scheme")
		return fmt.Errorf("invalid chart URL: %s", chartURL)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, chartURL, nil)
	if err != nil {
		log.Error().Err(err).Str("chartURL", chartURL).Msg("Failed to create request for legacy chart")
		return err
	}

	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}

	resp, err := client.Do(req) // #nosec G704 -- URL scheme is validated above to be http or https only
	if err != nil {
		log.Error().Err(err).Str("chartURL", chartURL).Msg("Failed to fetch legacy chart")
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		log.Error().Int("statusCode", resp.StatusCode).Str("chartURL", chartURL).Msg("Unexpected status code while fetching legacy chart")
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	chartFile := filepath.Join(outputPath, filepath.Base(chartURL))
	if err := os.MkdirAll(outputPath, 0o750); err != nil {
		log.Error().Err(err).Str("outputPath", outputPath).Msg("Failed to create output directory")
		return err
	}
	file, err := os.Create(chartFile) // #nosec G304 -- chartFile is constructed from a validated URL base name and a caller-supplied output path
	if err != nil {
		log.Error().Err(err).Str("chartFile", chartFile).Msg("Failed to create chart file")
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close file")
		}
	}()

	if _, err := io.Copy(file, resp.Body); err != nil {
		log.Error().Err(err).Str("chartFile", chartFile).Msg("Failed to write legacy chart data to file")
		return err
	}

	log.Info().Str("chartFile", chartFile).Msg("Successfully pulled legacy Helm chart")
	return nil
}

// PullOCIChart pulls a Helm chart from an OCI-compliant registry.
func (h *HelmHandler) PullOCIChart(repo, chart, version, outputPath string) error {
	chartRef := fmt.Sprintf("%s/%s:%s", repo, chart, version)

	log.Info().Str("chartRef", chartRef).Msg("Attempting to pull OCI Helm chart")

	pullResult, err := h.RegistryClient.Pull(chartRef)
	if err != nil {
		log.Error().Err(err).Str("chartRef", chartRef).Msg("Failed to pull OCI Helm chart")
		return err
	}
	log.Info().Str("chartRef", chartRef).Msg("Successfully pulled OCI Helm chart")

	if err := os.MkdirAll(outputPath, 0o750); err != nil {
		log.Error().Err(err).Str("outputPath", outputPath).Msg("Failed to create output directory")
		return err
	}

	chartFile := filepath.Join(outputPath, fmt.Sprintf("%s-%s.tgz", chart, version))
	if err := os.WriteFile(chartFile, pullResult.Chart.Data, 0o600); err != nil {
		log.Error().Err(err).Str("chartFile", chartFile).Msg("Failed to write OCI chart to file")
		return err
	}

	log.Info().Str("chartFile", chartFile).Msg("Successfully saved OCI Helm chart to disk")
	return nil
}
