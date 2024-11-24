package handlers

import (
    "fmt"
    "os"
    "io/ioutil"
    "path/filepath"

	"github.com/rs/zerolog/log"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/repo"
)

type HelmHandler struct {
	RegistryClient *registry.Client
	Settings       *cli.EnvSettings
}

// NewHelmHandler initializes and returns a HelmHandler
func NewHelmHandler() (*HelmHandler, error) {
	settings := cli.New()
	registryClient, err := registry.NewClient(
		registry.ClientOptCredentialsFile(settings.RegistryConfig),
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create registry client")
		return nil, err
	}

	log.Info().Msg("HelmHandler initialized successfully")
	return &HelmHandler{
		RegistryClient: registryClient,
		Settings:       settings,
	}, nil
}

// Helm login function
func (h *HelmHandler) Login(url, username, password string) error {
	log.Info().
		Str("url", url).
		Msg("Attempting to log in to the Helm registry")

	if err := h.RegistryClient.Login(
		url,
		registry.LoginOptBasicAuth(username, password),
	); err != nil {
		log.Error().
			Err(err).
			Str("url", url).
			Msg("Failed to log in to the Helm registry")
		return err
	}

	log.Info().
		Str("url", url).
		Msg("Successfully logged in to the Helm registry")
	return nil
}

// adding helm repo is needed since we are using a non OCI compliant registry
func (h *HelmHandler) AddRepo(name, url, username, password string) error {
	log.Info().
		Str("name", name).
		Str("url", url).
		Msg("Adding Helm repository")

    // Create a new repository entry
	entry := &repo.Entry{
		Name:     name,
		URL:      url,
		Username: username,
		Password: password,
	}

	// Load the current repositories file
	repoFile := h.Settings.RepositoryConfig
	r, err := repo.LoadFile(repoFile)
	if err != nil {
		log.Error().
			Err(err).
			Str("repoFile", repoFile).
			Msg("Failed to load repository configuration")
		return err
	}

	// Add or update the new repository
	r.Update(entry)

	// Write the updated repositories file to disk
	if err := r.WriteFile(repoFile, 0644); err != nil {
		log.Error().
			Err(err).
			Str("repoFile", repoFile).
			Msg("Failed to write repository configuration")
		return err
	}

	log.Info().
		Str("name", name).
		Str("url", url).
		Msg("Successfully added Helm repository")
	return nil
}

// PullChart pulls a Helm chart from an OCI registry and saves it to disk
func (h *HelmHandler) PullChart(repo, chart, version, outputPath string) error {
    // Construct the chart reference with the version
    chartRef := fmt.Sprintf("%s/%s:%s", repo, chart, version)

    log.Info().
        Str("chartRef", chartRef).
        Msg("Attempting to pull Helm chart")

    // Pull the chart using the registry client
    pullResult, err := h.RegistryClient.Pull(chartRef)
    if err != nil {
        log.Error().
            Err(err).
            Str("chartRef", chartRef).
            Msg("Failed to pull Helm chart")
        return err
    }

    log.Info().
        Str("chartRef", chartRef).
        Msg("Successfully pulled Helm chart")

    // Ensure the output directory exists
    if err := os.MkdirAll(outputPath, 0755); err != nil {
        log.Error().
            Err(err).
            Str("outputPath", outputPath).
            Msg("Failed to create output directory")
        return err
    }

    // Define the file path for the chart
    chartFile := filepath.Join(outputPath, fmt.Sprintf("%s-%s.tgz", chart, version))

    // Write the chart data to disk
    if err := ioutil.WriteFile(chartFile, pullResult.Chart.Data, 0644); err != nil {
        log.Error().
            Err(err).
            Str("chartFile", chartFile).
            Msg("Failed to save Helm chart to disk")
        return err
    }

    log.Info().
        Str("chartFile", chartFile).
        Msg("Successfully saved Helm chart to disk")
    return nil
}


