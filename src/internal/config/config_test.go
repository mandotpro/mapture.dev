package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAppliesDefaults(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "mapture.yaml")
	content := `version: 1
scan:
  include:
    - ./src
languages:
  go: true
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Catalog.Dir != "./architecture" {
		t.Fatalf("expected default catalog dir, got %q", cfg.Catalog.Dir)
	}
	if cfg.Comments.Style != "tags" {
		t.Fatalf("expected default comment style, got %q", cfg.Comments.Style)
	}
	if !cfg.Validation.FailOnUnknownDomain || !cfg.Validation.FailOnUnknownTeam || !cfg.Validation.FailOnUnknownEvent || !cfg.Validation.FailOnUnknownNode {
		t.Fatalf("expected default failOnUnknown* values to be true: %+v", cfg.Validation)
	}
	if cfg.Validation.WarnOnOrphanedNodes {
		t.Fatalf("expected warnOnOrphanedNodes default false")
	}
	if !cfg.Validation.WarnOnDeprecatedEvents {
		t.Fatalf("expected warnOnDeprecatedEvents default true")
	}
}

func TestLoadRejectsInvalidRequireMetadataRole(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "mapture.yaml")
	content := `version: 1
scan:
  include:
    - ./src
languages:
  go: true
validation:
  requireMetadataOn:
    - random
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected invalid config to fail")
	}
	if !strings.Contains(err.Error(), "random") {
		t.Fatalf("expected error to mention invalid value, got %v", err)
	}
}

func TestDiscoverWalksUpToConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	deep := filepath.Join(root, "src", "nested")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(root, "mapture.yaml")
	if err := os.WriteFile(path, []byte("version: 1\nscan:\n  include:\n    - ./src\nlanguages:\n  go: true\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	found, err := Discover(deep)
	if err != nil {
		t.Fatalf("Discover returned error: %v", err)
	}
	if found != path {
		t.Fatalf("expected %s, got %s", path, found)
	}
}
