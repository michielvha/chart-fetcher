package helm

import (
	"os"
	"path/filepath"
	"testing"

	"helm.sh/helm/v3/pkg/cli"
)

func newTestHandler(t *testing.T) *HelmHandler {
	t.Helper()
	tmpDir := t.TempDir()
	s := cli.New()
	s.RepositoryConfig = filepath.Join(tmpDir, "repositories.yaml")
	s.RepositoryCache = filepath.Join(tmpDir, "cache")
	return &HelmHandler{
		Settings:  s,
		RepoNames: make(map[string]string),
	}
}

func TestEnsureRepoFileExists_CreatesFile(t *testing.T) {
	h := newTestHandler(t)

	if err := h.EnsureRepoFileExists(); err != nil {
		t.Fatalf("EnsureRepoFileExists() unexpected error: %v", err)
	}

	if _, err := os.Stat(h.Settings.RepositoryConfig); err != nil {
		t.Errorf("expected repositories.yaml to exist after call: %v", err)
	}
}

func TestEnsureRepoFileExists_Idempotent(t *testing.T) {
	h := newTestHandler(t)

	if err := h.EnsureRepoFileExists(); err != nil {
		t.Fatalf("first call error: %v", err)
	}
	if err := h.EnsureRepoFileExists(); err != nil {
		t.Fatalf("second call error (expected idempotent): %v", err)
	}
}

func TestNewHelmHandler_RepoNamesInitialized(t *testing.T) {
	// Verifies the constructor always returns a non-nil RepoNames map.
	h, err := NewHelmHandler()
	if err != nil {
		t.Fatalf("NewHelmHandler() unexpected error: %v", err)
	}
	if h.RepoNames == nil {
		t.Error("expected RepoNames to be initialized, got nil")
	}
	// Should be safe to write to immediately without a nil-map panic.
	h.RepoNames["test"] = "value"
}
