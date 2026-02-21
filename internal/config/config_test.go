package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_YAML(t *testing.T) {
	content := `registries:
  - url: "oci://example.com"
    is_oci: true
    charts:
      - name: mychart
        version: "1.0.0"
`
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() unexpected error: %v", err)
	}
	if len(cfg.Registries) != 1 {
		t.Fatalf("expected 1 registry, got %d", len(cfg.Registries))
	}
	if cfg.Registries[0].URL != "oci://example.com" {
		t.Errorf("unexpected URL: %s", cfg.Registries[0].URL)
	}
	if !cfg.Registries[0].IsOCI {
		t.Error("expected IsOCI to be true")
	}
	if len(cfg.Registries[0].Charts) != 1 || cfg.Registries[0].Charts[0].Name != "mychart" {
		t.Errorf("unexpected charts: %+v", cfg.Registries[0].Charts)
	}
}

func TestLoadConfig_JSON(t *testing.T) {
	content := `{"registries":[{"url":"https://charts.example.com","charts":[{"name":"mychart","version":"1.0.0"}]}]}`
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() unexpected error: %v", err)
	}
	if len(cfg.Registries) != 1 || cfg.Registries[0].URL != "https://charts.example.com" {
		t.Errorf("unexpected config: %+v", cfg)
	}
}

func TestLoadConfig_UnsupportedFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("key = value"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig(path)
	if err == nil {
		t.Error("expected error for unsupported format, got nil")
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}
