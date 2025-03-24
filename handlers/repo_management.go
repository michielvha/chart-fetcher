// Contains functions related to repository management
package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"io"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	helmrepo "helm.sh/helm/v3/pkg/repo"
)

// EnsureRepoFileExists ensures the Helm repository file exists
func (h *HelmHandler) EnsureRepoFileExists() error {
    repoFile := h.Settings.RepositoryConfig

    // Check if the repository file exists
    if _, err := os.Stat(repoFile); os.IsNotExist(err) {
        // Create the directory if it doesn't exist
        if err := os.MkdirAll(filepath.Dir(repoFile), 0755); err != nil {
            log.Error().Err(err).Msg("Failed to create repository config directory")
            return err
        }

        // Initialize an empty repositories.yaml
        emptyRepoFile := helmrepo.NewFile()
        if err := emptyRepoFile.WriteFile(repoFile, 0644); err != nil {
            log.Error().Err(err).Msg("Failed to create empty repository config file")
            return err
        }

        log.Info().Str("repoFile", repoFile).Msg("Initialized new repositories.yaml")
    }

    return nil
}

// AddAndFetchRepo adds a repository and fetches its index
func (h *HelmHandler) AddAndFetchRepo(repoURL, username, password string) error {
    // Ensure the repositories file exists
    if err := h.EnsureRepoFileExists(); err != nil {
        return err
    }

    // Parse the URL to extract the repository name
    parsedURL, err := url.Parse(repoURL)
    if err != nil {
        log.Error().Err(err).Str("url", repoURL).Msg("Invalid repository URL")
        return err
    }
    // Get the last segment from the URL path
    segments := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
    name := segments[len(segments)-1]

    log.Info().Str("name", name).Str("url", repoURL).Msg("Adding Helm repository")

    // Create a new repository entry
    entry := &helmrepo.Entry{
        Name:     name,
        URL:      repoURL,
        Username: username,
        Password: password,
    }

    // Load the current repositories file
    repoFile := h.Settings.RepositoryConfig
    r, err := helmrepo.LoadFile(repoFile)
    if err != nil {
        log.Error().Err(err).Str("repoFile", repoFile).Msg("Failed to load repository configuration")
        return err
    }

    // Add or update the new repository
    r.Update(entry)

    // Write the updated repositories file to disk
    if err := r.WriteFile(repoFile, 0644); err != nil {
        log.Error().Err(err).Str("repoFile", repoFile).Msg("Failed to write repository configuration")
        return err
    }

    // Save the name to the RepoNames map
    if h.RepoNames == nil {
        h.RepoNames = make(map[string]string)
    }
    h.RepoNames[repoURL] = name
    log.Info().Str("name", name).Str("url", repoURL).Msg("Successfully added Helm repository")

    // Fetch the repository index
    log.Info().Str("url", repoURL).Str("name", name).Msg("Fetching repository index")

    // Define the index file path
    cacheDir := h.Settings.RepositoryCache
    indexFile := filepath.Join(cacheDir, fmt.Sprintf("%s-index.yaml", name))

    // Prepare the HTTP client
    client := &http.Client{}

    // Create the request with basic authentication
    req, err := http.NewRequest("GET", fmt.Sprintf("%s/index.yaml", repoURL), nil)
    if err != nil {
        log.Error().Err(err).Str("url", repoURL).Msg("Failed to create request for repository index")
        return err
    }

    // Add authentication if credentials exist
    if username != "" && password != "" {
        req.SetBasicAuth(username, password)
    }

    // Execute the request
    resp, err := client.Do(req)
    if err != nil {
        log.Error().Err(err).Str("url", repoURL).Msg("Failed to fetch repository index")
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        log.Error().Int("statusCode", resp.StatusCode).Str("url", repoURL).Msg("Unexpected status while fetching repository index")
        return fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }

    // Ensure the directory for the index file exists before writing
    if err := os.MkdirAll(filepath.Dir(indexFile), 0755); err != nil {
        log.Error().Err(err).Str("directory", filepath.Dir(indexFile)).Msg("Failed to create directory for repository index")
        return err
    }

    // Save the index file to the cache directory
    data, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Error().Err(err).Msg("Failed to read repository index")
        return err
    }
    if err := os.WriteFile(indexFile, data, 0644); err != nil {
        log.Error().Err(err).Str("indexFile", indexFile).Msg("Failed to write repository index")
        return err
    }

    log.Info().Str("indexFile", indexFile).Msg("Successfully fetched repository index")
    return nil
}
