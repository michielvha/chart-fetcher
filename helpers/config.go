// helpers/config.go
// Purpose: help with handling the configuration file.
package helpers

import (
    "encoding/json"
    "errors"
    "os"
    "strings"

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

// LoadConfig handles opening the config file, detecting the format, and decoding it.
func LoadConfig(configPath string) (Config, error) {
    file, err := os.Open(configPath)
    if err != nil {
        return Config{}, err
    }
    defer file.Close()

    var config Config
    if strings.HasSuffix(configPath, ".yaml") || strings.HasSuffix(configPath, ".yml") {
        if err := yaml.NewDecoder(file).Decode(&config); err != nil {
            return Config{}, err
        }
    } else if strings.HasSuffix(configPath, ".json") {
        if err := json.NewDecoder(file).Decode(&config); err != nil {
            return Config{}, err
        }
    } else {
        return Config{}, errors.New("unsupported file format: " + configPath)
    }
    return config, nil
}
