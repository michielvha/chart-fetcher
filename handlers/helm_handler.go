package handlers

import (
	"github.com/rs/zerolog/log"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
)

// HelmHandler manages Helm operations
type HelmHandler struct {
	RegistryClient *registry.Client
	Settings       *cli.EnvSettings
	RepoNames      map[string]string
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
