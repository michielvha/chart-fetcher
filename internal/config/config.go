// Package config handles loading and parsing the application configuration file.
package config

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// Registry defines the structure for a registry entry in the configuration.
type Registry struct {
	URL         string  `json:"url"                    yaml:"url"`
	UsernameEnv string  `json:"username_env,omitempty" yaml:"username_env,omitempty"`
	PasswordEnv string  `json:"password_env,omitempty" yaml:"password_env,omitempty"`
	Charts      []Chart `json:"charts"                 yaml:"charts"`
	IsOCI       bool    `json:"is_oci,omitempty"       yaml:"is_oci,omitempty"`
}

// Chart defines the name and version of a Helm chart to pull.
type Chart struct {
	Name    string `json:"name"    yaml:"name"`
	Version string `json:"version" yaml:"version"`
}

// Config is the top-level structure for the application configuration file.
type Config struct {
	Registries []Registry `json:"registries" yaml:"registries"`
}

// LoadConfig opens the config file at configPath, detects its format (JSON or YAML),
// and decodes it into a Config value.
func LoadConfig(configPath string) (Config, error) {
	file, err := os.Open(configPath) // #nosec G304 -- configPath is supplied by the operator via CLI flag or environment variable, not user input
	if err != nil {
		return Config{}, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close config file")
		}
	}()

	var cfg Config
	switch {
	case strings.HasSuffix(configPath, ".yaml") || strings.HasSuffix(configPath, ".yml"):
		if err := yaml.NewDecoder(file).Decode(&cfg); err != nil {
			return Config{}, err
		}
	case strings.HasSuffix(configPath, ".json"):
		if err := json.NewDecoder(file).Decode(&cfg); err != nil {
			return Config{}, err
		}
	default:
		return Config{}, errors.New("unsupported file format: " + configPath)
	}
	return cfg, nil
}
