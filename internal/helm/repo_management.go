package helm

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	helmrepo "helm.sh/helm/v3/pkg/repo"
)

// EnsureRepoFileExists ensures the Helm repository config file exists, creating it if necessary.
func (h *HelmHandler) EnsureRepoFileExists() error {
	repoFile := h.Settings.RepositoryConfig

	if _, err := os.Stat(repoFile); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(repoFile), 0o750); err != nil {
			log.Error().Err(err).Msg("Failed to create repository config directory")
			return err
		}

		emptyRepoFile := helmrepo.NewFile()
		if err := emptyRepoFile.WriteFile(repoFile, 0o600); err != nil {
			log.Error().Err(err).Msg("Failed to create empty repository config file")
			return err
		}

		log.Info().Str("repoFile", repoFile).Msg("Initialized new repositories.yaml")
	}

	return nil
}

// AddAndFetchRepo adds a Helm repository entry and downloads its index file.
func (h *HelmHandler) AddAndFetchRepo(repoURL, username, password string) error {
	if err := h.EnsureRepoFileExists(); err != nil {
		return err
	}

	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		log.Error().Err(err).Str("url", repoURL).Msg("Invalid repository URL")
		return err
	}
	segments := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	name := segments[len(segments)-1]

	log.Info().Str("name", name).Str("url", repoURL).Msg("Adding Helm repository")

	entry := &helmrepo.Entry{
		Name:     name,
		URL:      repoURL,
		Username: username,
		Password: password,
	}

	repoFile := h.Settings.RepositoryConfig
	r, err := helmrepo.LoadFile(repoFile)
	if err != nil {
		log.Error().Err(err).Str("repoFile", repoFile).Msg("Failed to load repository configuration")
		return err
	}

	r.Update(entry)

	if err := r.WriteFile(repoFile, 0o600); err != nil {
		log.Error().Err(err).Str("repoFile", repoFile).Msg("Failed to write repository configuration")
		return err
	}

	h.RepoNames[repoURL] = name
	log.Info().Str("name", name).Str("url", repoURL).Msg("Successfully added Helm repository")

	log.Info().Str("url", repoURL).Str("name", name).Msg("Fetching repository index")

	cacheDir := h.Settings.RepositoryCache
	indexFile := filepath.Join(cacheDir, fmt.Sprintf("%s-index.yaml", name))

	parsedIndexURL, err := url.ParseRequestURI(fmt.Sprintf("%s/index.yaml", repoURL))
	if err != nil || (parsedIndexURL.Scheme != "http" && parsedIndexURL.Scheme != "https") {
		log.Error().Str("url", repoURL).Msg("Invalid or disallowed repository URL scheme")
		return fmt.Errorf("invalid repository URL: %s", repoURL)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, parsedIndexURL.String(), nil)
	if err != nil {
		log.Error().Err(err).Str("url", repoURL).Msg("Failed to create request for repository index")
		return err
	}

	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}

	resp, err := client.Do(req) // #nosec G704 -- URL scheme is validated above to be http or https only
	if err != nil {
		log.Error().Err(err).Str("url", repoURL).Msg("Failed to fetch repository index")
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		log.Error().Int("statusCode", resp.StatusCode).Str("url", repoURL).Msg("Unexpected status while fetching repository index")
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(indexFile), 0o750); err != nil { // #nosec G703 -- indexFile is constructed from the Helm settings cache dir and a sanitised repo name
		log.Error().Err(err).Str("directory", filepath.Dir(indexFile)).Msg("Failed to create directory for repository index")
		return err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read repository index")
		return err
	}
	if err := os.WriteFile(indexFile, data, 0o600); err != nil { // #nosec G703 -- indexFile is constructed from the Helm settings cache dir and a sanitised repo name
		log.Error().Err(err).Str("indexFile", indexFile).Msg("Failed to write repository index")
		return err
	}

	log.Info().Str("indexFile", indexFile).Msg("Successfully fetched repository index")
	return nil
}
