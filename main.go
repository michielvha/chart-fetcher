package main

import (
	"os"

	"github.com/rs/zerolog/log"
	"dev.azure.com/bnl-ms/AzureFoundation/_git/charthost/handlers"
)

func main() {
	// Initialize the HelmHandler
	handler, err := handlers.NewHelmHandler()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize HelmHandler")
	}

	// Log in to a Helm registry
	err = handler.Login(
		"oci://registry.example.com",
		os.Getenv("REGISTRY_USERNAME"),
		os.Getenv("REGISTRY_PASSWORD"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to log in to registry")
	}

// 	// Add a Helm repository
// 	err = handler.AddRepo(
// 		"myrepo",
// 		"https://charts.example.com",
// 		"username",
// 		"password",
// 	)
// 	if err != nil {
// 		log.Fatal().Err(err).Msg("Failed to add repository")
// 	}

	// Pull a Helm chart
	err = handler.PullChart(
		"oci://registry.example.com",
		"mychart",
		"1.0.0",
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to pull chart")
	}
}
