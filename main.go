package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/rs/zerolog/log"
	"dev.azure.com/bnl-ms/AzureFoundation/charthost/handlers"
)

// Registry defines the structure for a registry in the configuration
type Registry struct {
	URL          string `json:"url"`
	UsernameEnv  string `json:"username_env,omitempty"`
	PasswordEnv  string `json:"password_env,omitempty"`
	Charts       []Chart `json:"charts"`
}

// Chart defines the structure for a Helm chart in the configuration
type Chart struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Config defines the structure for the configuration file
type Config struct {
	Registries []Registry `json:"registries"`
}

func main() {
	// Define a flag for the configuration file path, allows you to specify --config, if not mentioned it falls back to ./config.json which is the same dir as the binary
	configPath := flag.String("config", "./config.json", "Path to the configuration file")
    // Define output directory, use flag like in for configPath
    outputPath := flag.String("outputPath", "./charts", "Path to the output directory")
    flag.Parse()

	// Allow overriding the path with an environment variable
	if envPath := os.Getenv("CHARTHOST_CONFIG_PATH"); envPath != "" {
		*configPath = envPath
	}
    if envOutput := os.Getenv("OUTPUT_DIR"); envOutput != "" {
    *outputPath = envOutput
    }

	// Open the configuration file
	file, err := os.Open(*configPath)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to open configuration file: %s", *configPath)
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		log.Fatal().Err(err).Msg("Failed to parse configuration file")
	}

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
		}
		if registry.PasswordEnv != "" {
			password = os.Getenv(registry.PasswordEnv)
		}

		// Log in to the registry (if credentials are provided)
		if username != "" || password != "" {
			if err := handler.Login(registry.URL, username, password); err != nil {
				log.Error().Err(err).Str("url", registry.URL).Msg("Failed to log in to registry")
				continue
			}
		}

        // Pull each chart from the registry
        for _, chart := range registry.Charts {
            log.Info().
                Str("chart", chart.Name).
                Str("version", chart.Version).
                Msg("Pulling chart")

            // Call PullChart with the chart name, repository URL, and version
            if err := handler.PullChart(registry.URL, chart.Name, chart.Version, *outputPath); err != nil {
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
