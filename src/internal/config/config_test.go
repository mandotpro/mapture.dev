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

	if cfg.Comments.Style != "tags" {
		t.Fatalf("expected default comment style, got %q", cfg.Comments.Style)
	}
	if !cfg.Validation.FailOnUnknownDomain || !cfg.Validation.FailOnUnknownTeam || !cfg.Validation.FailOnUnknownNode {
		t.Fatalf("expected default failOnUnknown* values to be true: %+v", cfg.Validation)
	}
	if cfg.Validation.WarnOnOrphanedNodes {
		t.Fatalf("expected warnOnOrphanedNodes default false")
	}
	if !cfg.Validation.WarnOnDeprecatedEvents {
		t.Fatalf("expected warnOnDeprecatedEvents default true")
	}
	if cfg.UI.NodeColors.Service != "#1664d9" || cfg.UI.NodeColors.API != "#0f8f78" || cfg.UI.NodeColors.Database != "#a56614" || cfg.UI.NodeColors.Event != "#a73f7f" {
		t.Fatalf("expected default UI node colors, got %+v", cfg.UI.NodeColors)
	}
	if cfg.UI.DefaultLayout != DefaultLayoutELKHorizontal {
		t.Fatalf("expected default UI layout %q, got %q", DefaultLayoutELKHorizontal, cfg.UI.DefaultLayout)
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

func TestLoadAcceptsCustomNodeColors(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "mapture.yaml")
	content := `version: 1
scan:
  include:
    - ./src
languages:
  go: true
ui:
  nodeColors:
    service: "#112233"
    api: "#223344"
    database: "#334455"
    event: "#445566"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.UI.NodeColors.Service != "#112233" || cfg.UI.NodeColors.API != "#223344" || cfg.UI.NodeColors.Database != "#334455" || cfg.UI.NodeColors.Event != "#445566" {
		t.Fatalf("expected custom UI node colors, got %+v", cfg.UI.NodeColors)
	}
}

func TestLoadAcceptsCustomDefaultLayout(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "mapture.yaml")
	content := `version: 1
scan:
  include:
    - ./src
languages:
  go: true
ui:
  defaultLayout: clustered
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.UI.DefaultLayout != "clustered" {
		t.Fatalf("expected custom UI defaultLayout, got %q", cfg.UI.DefaultLayout)
	}
}

func TestLoadRejectsInvalidNodeColor(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "mapture.yaml")
	content := `version: 1
scan:
  include:
    - ./src
languages:
  go: true
ui:
  nodeColors:
    service: "blue"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected invalid color to fail")
	}
	if !strings.Contains(err.Error(), "blue") {
		t.Fatalf("expected error to mention invalid color, got %v", err)
	}
}

func TestLoadRejectsDuplicateTopLevelTags(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "mapture.yaml")
	content := `version: 1
tags:
  - critical-path
  - critical-path
scan:
  include:
    - ./src
languages:
  go: true
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected duplicate tag config to fail")
	}
	if !strings.Contains(err.Error(), `duplicate tag "critical-path"`) {
		t.Fatalf("expected duplicate tag error, got %v", err)
	}
}

func TestLoadRejectsMalformedTopLevelTags(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "mapture.yaml")
	content := `version: 1
tags:
  - CriticalPath
scan:
  include:
    - ./src
languages:
  go: true
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected malformed tag config to fail")
	}
	if !strings.Contains(err.Error(), "CriticalPath") {
		t.Fatalf("expected malformed tag error to mention invalid value, got %v", err)
	}
}

func TestLoadAcceptsFacetDefinitions(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "mapture.yaml")
	content := `version: 1
facets:
  event.type:
    label: Event Type
    values:
      - sync
      - async
  db.type:
    label: Database Type
    values:
      - tenant
      - shared
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

	if cfg.Facets["event.type"].Label != "Event Type" {
		t.Fatalf("unexpected facet label: %+v", cfg.Facets)
	}
	if got := strings.Join(cfg.Facets["db.type"].Values, ","); got != "tenant,shared" {
		t.Fatalf("unexpected facet values: %q", got)
	}
}

func TestLoadRejectsDuplicateFacetValues(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "mapture.yaml")
	content := `version: 1
facets:
  event.type:
    label: Event Type
    values:
      - async
      - async
scan:
  include:
    - ./src
languages:
  go: true
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected duplicate facet values to fail")
	}
	if !strings.Contains(err.Error(), `facet "event.type" has duplicate value "async"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadRejectsMalformedFacetKey(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "mapture.yaml")
	content := `version: 1
facets:
  EventType:
    label: Event Type
    values:
      - async
scan:
  include:
    - ./src
languages:
  go: true
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected malformed facet key to fail")
	}
	if !strings.Contains(err.Error(), "EventType") {
		t.Fatalf("expected malformed facet key error to mention invalid value, got %v", err)
	}
}
