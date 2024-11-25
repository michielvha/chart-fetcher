package main

import (
	"encoding/json"
	"flag"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"dev.azure.com/bnl-ms/AzureFoundation/charthost/handlers"
	"gopkg.in/yaml.v3"
)

// Registry defines the structure for a registry in the configuration
type Registry struct {
    URL          string  `json:"url" yaml:"url"`
    UsernameEnv  string  `json:"username_env,omitempty" yaml:"username_env,omitempty"`
    PasswordEnv  string  `json:"password_env,omitempty" yaml:"password_env,omitempty"`
    Charts       []Chart `json:"charts" yaml:"charts"`
    IsOCI        bool    `json:"is_oci,omitempty" yaml:"is_oci,omitempty"`
}

// Chart defines the structure for a Helm chart in the configuration
type Chart struct {
    Name    string `json:"name" yaml:"name"`
    Version string `json:"version" yaml:"version"`
}

// Config defines the structure for the configuration file
type Config struct {
	Registries []Registry `json:"registries" yaml:"registries"`
}

func main() {
	// Define a flag for the configuration file path, allows you to specify --config, if not mentioned it falls back to ./config.json which is the same dir as the binary
	configPath := flag.String("config", "./config.json", "Path to the configuration file (JSON OR YAML)")
    // Define output directory, use flag like in for configPath
    outputPath := flag.String("outputPath", "./charts", "Path to the output directory")
    flag.Parse()

// debug statement to be removed
//     for _, env := range os.Environ() {
//         log.Debug().Msgf("Env: %s", env)
//     }

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
    // Detect file type based on extension and decode this enables support for both JSON and YAML
    if strings.HasSuffix(*configPath, ".yaml") || strings.HasSuffix(*configPath, ".yml") {
        if err := yaml.NewDecoder(file).Decode(&config); err != nil {
            log.Fatal().Err(err).Msg("Failed to parse YAML configuration file")
        }
    } else if strings.HasSuffix(*configPath, ".json") {
        if err := json.NewDecoder(file).Decode(&config); err != nil {
            log.Fatal().Err(err).Msg("Failed to parse JSON configuration file")
        }
    } else {
        log.Fatal().Msgf("Unsupported configuration file format: %s", *configPath)
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
			log.Info().Str("username", username).Msg("Fetched username from environment variable")
		}
		if registry.PasswordEnv != "" {
			password = os.Getenv(registry.PasswordEnv)
			log.Info().Msg("Fetched password from environment variable")
		}

        // TODO: this needs to be seperated based on OCI compliant or not. Only if OCI complaint login else just use repo add.
		// Log in to the registry & add repository (if credentials are provided)
		if username != "" || password != "" {
			if err := handler.Login(registry.URL, username, password); err != nil {
				log.Error().Err(err).Str("url", registry.URL).Msg("Failed to log in to registry")
				continue
			} else {
			    log.Info().Str("url", registry.URL).Msg("Successfully logged in to registry")
            }

            if err := handler.AddRepo(registry.URL, username, password); err != nil {
				log.Error().Err(err).Str("url", registry.URL).Msg("Failed to add repository")
				continue
			} else {
			    log.Info().Str("url", registry.URL).Msg("Successfully added repository")
            }
            // Fetch the repository index
            if err := handler.FetchRepoIndex(registry.URL, username, password); err != nil {
                log.Error().Err(err).Str("url", registry.URL).Msg("Failed to fetch repository index")
                continue
            } else {
                log.Info().Str("url", registry.URL).Msg("Successfully fetched repository index")
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
