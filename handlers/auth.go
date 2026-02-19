// Package handlers
// Purpose: provides functions for handling authentication with the Helm registry.
package handlers

import (
	"github.com/rs/zerolog/log"
	"helm.sh/helm/v3/pkg/registry"
)

// Login to the Helm registry
func (h *HelmHandler) Login(url, username, password string) error {
	log.Info().Str("url", url).Msg("Attempting to log in to the Helm registry")

	if err := h.RegistryClient.Login(
		url,
		registry.LoginOptBasicAuth(username, password),
	); err != nil {
		log.Error().Err(err).Str("url", url).Msg("Failed to log in to the Helm registry")
		return err
	}

	log.Info().Str("url", url).Msg("Successfully logged in to the Helm registry")
	return nil
}
