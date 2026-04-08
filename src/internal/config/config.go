// Package config loads and validates repository-level Mapture configuration.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mandotpro/mapture.dev/src/internal/schema"
)

const filename = "mapture.yaml"

// Config is the repository-level Mapture configuration.
type Config struct {
	Version    int        `json:"version"`
	Catalog    Catalog    `json:"catalog"`
	Scan       Scan       `json:"scan"`
	Languages  Languages  `json:"languages"`
	Comments   Comments   `json:"comments"`
	Validation Validation `json:"validation"`
}

// Catalog configures where catalog YAML files live.
type Catalog struct {
	Dir string `json:"dir"`
}

// Scan configures which repository paths should be scanned or skipped.
type Scan struct {
	Include []string `json:"include"`
	Exclude []string `json:"exclude"`
}

// Languages controls which source languages Mapture should inspect.
type Languages struct {
	PHP        bool `json:"php"`
	Go         bool `json:"go"`
	TypeScript bool `json:"typescript"`
	JavaScript bool `json:"javascript"`
}

// Comments defines the supported comment parsing style.
type Comments struct {
	Style string `json:"style"`
}

// Validation configures strictness for catalog and graph checks.
type Validation struct {
	FailOnUnknownDomain    bool     `json:"failOnUnknownDomain"`
	FailOnUnknownTeam      bool     `json:"failOnUnknownTeam"`
	FailOnUnknownEvent     bool     `json:"failOnUnknownEvent"`
	FailOnUnknownNode      bool     `json:"failOnUnknownNode"`
	RequireMetadataOn      []string `json:"requireMetadataOn"`
	WarnOnOrphanedNodes    bool     `json:"warnOnOrphanedNodes"`
	WarnOnDeprecatedEvents bool     `json:"warnOnDeprecatedEvents"`
}

// Discover walks up from start until it finds mapture.yaml.
func Discover(start string) (string, error) {
	absStart, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("resolve %s: %w", start, err)
	}

	current := absStart
	info, err := os.Stat(absStart)
	if err != nil {
		return "", fmt.Errorf("inspect %s: %w", absStart, err)
	}
	if !info.IsDir() {
		current = filepath.Dir(absStart)
	}

	for {
		candidate := filepath.Join(current, filename)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("%s not found from %s or any parent directory", filename, absStart)
		}
		current = parent
	}
}

// Load reads, validates, and decodes a config file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var cfg Config
	if err := schema.DecodeYAML(schema.ConfigDefinition, path, data, &cfg); err != nil {
		return nil, err
	}
	if !cfg.Languages.PHP && !cfg.Languages.Go && !cfg.Languages.TypeScript && !cfg.Languages.JavaScript {
		return nil, fmt.Errorf("%s: at least one language must be enabled", path)
	}

	return &cfg, nil
}

// CatalogDir returns the absolute path to the configured catalog directory.
func (c *Config) CatalogDir(configPath string) (string, error) {
	if c == nil {
		return "", errors.New("config is nil")
	}
	if filepath.IsAbs(c.Catalog.Dir) {
		return c.Catalog.Dir, nil
	}
	return filepath.Join(filepath.Dir(configPath), c.Catalog.Dir), nil
}
