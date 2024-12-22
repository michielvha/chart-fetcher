package handlers

import (
    "fmt"
    "strings"
    "net/url"
    "os"
    "net/http"
    "io/ioutil"
    "path/filepath"
    "io"

	"github.com/rs/zerolog/log"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
	helmrepo "helm.sh/helm/v3/pkg/repo"
)

type HelmHandler struct {
	RegistryClient *registry.Client
	Settings       *cli.EnvSettings
	RepoNames      map[string]string // Map to store URL -> RepoName
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

// Helm login function
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

// new pull functions, seperation between OCI complaint and non OCI compliant.repository

// PullLegacyChart pulls a Helm chart from a non-OCI-compliant repository
func (h *HelmHandler) PullLegacyChart(repo, chart, version, outputPath, username, password string) error {

    log.Info().Str("chart", chart).Str("version", version).Msg("Attempting to pull legacy Helm chart")

    // Resolve the repository name
    repoName, exists := h.RepoNames[repo]
    if !exists {
        log.Error().Str("repo", repo).Msg("Repository not found in RepoNames map")
        return fmt.Errorf("repository not found: %s", repo)
    }

    // Load the index file, in AddAndFetchRepo When ensure it already exists
    indexFile := filepath.Join(h.Settings.RepositoryCache, fmt.Sprintf("%s-index.yaml", repoName))
    index, err := helmrepo.LoadIndexFile(indexFile)
    if err != nil {
        log.Error().Err(err).Str("indexFile", indexFile).Msg("Failed to load repository index file")
        return err
    }

    // Find the chart version
    chartVersion, err := index.Get(chart, version)
    if err != nil {
        log.Error().Err(err).Str("chart", chart).Str("version", version).Msg("Chart version not found in index")
        return err
    }

    // Resolve the chart URL
    chartURL := chartVersion.URLs[0]
    if !strings.HasPrefix(chartURL, "http") {
        chartURL = fmt.Sprintf("%s/%s", strings.TrimSuffix(repo, "/"), chartURL)
    }

    log.Info().Str("chartURL", chartURL).Msg("Resolved chart URL")

    // Prepare the HTTP client and request
    client := &http.Client{}
    req, err := http.NewRequest("GET", chartURL, nil)
    if err != nil {
        log.Error().Err(err).Str("chartURL", chartURL).Msg("Failed to create request for legacy chart")
        return err
    }

    // Add authentication if credentials are provided
    if username != "" && password != "" {
        req.SetBasicAuth(username, password)
    }

    // Execute the request
    resp, err := client.Do(req)
    if err != nil {
        log.Error().Err(err).Str("chartURL", chartURL).Msg("Failed to fetch legacy chart")
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        log.Error().Int("statusCode", resp.StatusCode).Str("chartURL", chartURL).Msg("Unexpected status code while fetching legacy chart")
        return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }

    // Save the chart to disk
    chartFile := filepath.Join(outputPath, filepath.Base(chartURL))
    if err := os.MkdirAll(outputPath, 0755); err != nil {
        log.Error().Err(err).Str("outputPath", outputPath).Msg("Failed to create output directory")
        return err
    }
    file, err := os.Create(chartFile)
    if err != nil {
        log.Error().Err(err).Str("chartFile", chartFile).Msg("Failed to create chart file")
        return err
    }
    defer file.Close()

    if _, err := io.Copy(file, resp.Body); err != nil {
        log.Error().Err(err).Str("chartFile", chartFile).Msg("Failed to write legacy chart data to file")
        return err
    }

    log.Info().Str("chartFile", chartFile).Msg("Successfully pulled legacy Helm chart")
    return nil
}

func (h *HelmHandler) PullOCIChart(repo, chart, version, outputPath string) error {
    // Construct the OCI reference
    chartRef := fmt.Sprintf("%s/%s:%s", repo, chart, version)

    log.Info().Str("chartRef", chartRef).Msg("Attempting to pull OCI Helm chart")

    // Pull the chart using the Helm Registry Client
    pullResult, err := h.RegistryClient.Pull(chartRef)
    if err != nil {
        log.Error().Err(err).Str("chartRef", chartRef).Msg("Failed to pull OCI Helm chart")
        return err
    }
    log.Info().Str("chartRef", chartRef).Msg("Successfully pulled OCI Helm chart")

    // Save the chart to disk
    if err := os.MkdirAll(outputPath, 0755); err != nil {
        log.Error().Err(err).Str("outputPath", outputPath).Msg("Failed to create output directory")
        return err
    }

    chartFile := filepath.Join(outputPath, fmt.Sprintf("%s-%s.tgz", chart, version))
    if err := ioutil.WriteFile(chartFile, pullResult.Chart.Data, 0644); err != nil {
        log.Error().Err(err).Str("chartFile", chartFile).Msg("Failed to write OCI chart to file")
        return err
    }

    log.Info().Str("chartFile", chartFile).Msg("Successfully saved OCI Helm chart to disk")
    return nil
}

// combined fetchrepoindex and addrepo functions
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
    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Error().Err(err).Msg("Failed to read repository index")
        return err
    }
    if err := ioutil.WriteFile(indexFile, data, 0644); err != nil {
        log.Error().Err(err).Str("indexFile", indexFile).Msg("Failed to write repository index")
        return err
    }

    log.Info().Str("indexFile", indexFile).Msg("Successfully fetched repository index")
    return nil
}
