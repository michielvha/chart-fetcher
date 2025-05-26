package main

import (
	"flag"
	"os"

	"dev.azure.com/bnl-ms/AzureFoundation/charthost/handlers"
	"dev.azure.com/bnl-ms/AzureFoundation/charthost/helpers"
	"github.com/rs/zerolog/log"
)

func main() {
	// Define a flag for the configuration file path, allows you to specify --config, if not mentioned it falls back to ./config.json which is the same dir as the binary
	configPath := flag.String("config", "./config.yaml", "Path to the configuration file (JSON OR YAML)")
	// Define output directory, use flag like in for configPath
	outputPath := flag.String("outputPath", "./charts", "Path to the output directory")
	flag.Parse()

	// Override using environment variables with helper function
	helpers.OverrideFromEnv(configPath, "CONFIG_PATH")
	helpers.OverrideFromEnv(outputPath, "OUTPUT_PATH")

	// Load the configuration
	config, err := helpers.LoadConfig(*configPath)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to load configuration file: %s", *configPath)
	}

	// debug statement to be removed
	//     for _, env := range os.Environ() {
	//         log.Debug().Msgf("Env: %s", env)
	//     }

	// Initialize the HelmHandler
	handler, err := handlers.NewHelmHandler()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize HelmHandler")
	}

	// Iterate over registries and process charts
	for _, registry := range config.Registries {
		log.Info().Str("url", registry.URL).Msg("Processing registry")

		// Fetch credentials from environment variables if specified
		var username, password string
		if registry.UsernameEnv != "" {
			username = os.Getenv(registry.UsernameEnv)
			log.Debug().Str("username", username).Msg("Fetched username from environment variable")
		}
		if registry.PasswordEnv != "" {
			password = os.Getenv(registry.PasswordEnv)
			log.Debug().Msg("Fetched password from environment variable")
		}

		// separate based on OCI compliant or not. Only if OCI complaint login else just use repo add. v1.0.1
		// Log in to the registry & add repository (if credentials are provided)
		if username != "" || password != "" {
			if registry.IsOCI {
				if err := handler.Login(registry.URL, username, password); err != nil {
					log.Error().Err(err).Str("url", registry.URL).Msg("Failed to log in to registry")
					continue
				} else {
					log.Info().Str("url", registry.URL).Msg("Successfully logged in to registry")
				}
			} else {
				if err := handler.AddAndFetchRepo(registry.URL, username, password); err != nil {
					log.Error().Err(err).Str("url", registry.URL).Msg("Failed to add repository or fetch repository index")
					continue
				} else {
					log.Info().Str("url", registry.URL).Msg("Successfully added repository and fetched repository index")
				}
			}
		}

		// Pull each chart from the registry
		for _, chart := range registry.Charts {
			log.Info().
				Str("chart", chart.Name).
				Str("version", chart.Version).
				Msg("Pulling chart")

			var pullErr error
			if registry.IsOCI {
				pullErr = handler.PullOCIChart(registry.URL, chart.Name, chart.Version, *outputPath)
			} else {
				pullErr = handler.PullLegacyChart(registry.URL, chart.Name, chart.Version, *outputPath, username, password)
			}

			if pullErr != nil {
				log.Error().
					Err(err).
					Str("chart", chart.Name).
					Str("version", chart.Version).
					Msg("Failed to pull chart")
			} else {
				log.Info().
					Str("chart", chart.Name).
					Str("version", chart.Version).
					Str("outputPath", *outputPath).
					Msgf("Successfully pulled and saved chart %s:%s to disk at location: %s", chart.Name, chart.Version, *outputPath)
			}
		}

	}
	log.Info().Msg("Processing completed")
}
