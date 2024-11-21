package handlers

import (
    "fmt"
	"github.com/rs/zerolog/log"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/getter"
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

// adding helmrepo should not be needed.
// func (h *HelmHandler) AddRepo(name, url, username, password string) error {
// 	log.Info().
// 		Str("name", name).
// 		Str("url", url).
// 		Msg("Adding Helm repository")
//
// 	entry := &repo.Entry{
// 		Name:     name,
// 		URL:      url,
// 		Username: username,
// 		Password: password,
// 	}
//
// 	// Load the current repositories file
// 	repoFile := h.Settings.RepositoryConfig
// 	r, err := repo.LoadFile(repoFile)
// 	if err != nil {
// 		log.Error().
// 			Err(err).
// 			Str("repoFile", repoFile).
// 			Msg("Failed to load repository configuration")
// 		return err
// 	}
//
// 	// Add the new repository
// 	if err := r.Update(entry); err != nil {
// 		log.Error().
// 			Err(err).
// 			Str("name", name).
// 			Str("url", url).
// 			Msg("Failed to add repository to configuration")
// 		return err
// 	}
//
// 	// Write the updated repositories file
// 	if err := r.WriteFile(repoFile, 0644); err != nil {
// 		log.Error().
// 			Err(err).
// 			Str("repoFile", repoFile).
// 			Msg("Failed to write repository configuration")
// 		return err
// 	}
//
// 	log.Info().
// 		Str("name", name).
// 		Str("url", url).
// 		Msg("Successfully added Helm repository")
// 	return nil
// }

// This function allows you to pull helm charts
func (h *HelmHandler) PullChart(repo, chart, version string) error {
	chartRef := fmt.Sprintf("%s/%s:%s", repo, chart, version)

	log.Info().
		Str("chartRef", chartRef).
		Msg("Attempting to pull Helm chart")

	if err := h.RegistryClient.Pull(chartRef, getter.All(h.Settings)); err != nil {
		log.Error().
			Err(err).
			Str("chartRef", chartRef).
			Msg("Failed to pull Helm chart")
		return err
	}

	log.Info().
		Str("chartRef", chartRef).
		Msg("Successfully pulled Helm chart")
	return nil
}