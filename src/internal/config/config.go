// Package config loads and validates repository-level Mapture configuration.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mandotpro/mapture.dev/src/internal/schema"
)

const filename = "mapture.yaml"

// Config is the repository-level Mapture configuration.
type Config struct {
	Version    int        `json:"version"`
	Catalog    Catalog    `json:"catalog"`
	Tags       []string   `json:"tags,omitempty"`
	Teams      []Team     `json:"teams,omitempty"`
	Domains    []Domain   `json:"domains,omitempty"`
	Scan       Scan       `json:"scan"`
	Languages  Languages  `json:"languages"`
	Comments   Comments   `json:"comments"`
	Validation Validation `json:"validation"`
	UI         UI         `json:"ui"`
}

// Catalog configures where catalog YAML files live.
type Catalog struct {
	Dir string `json:"dir"`
}

// Team is an inline team catalog entry defined in mapture.yaml.
type Team struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Contact string   `json:"contact,omitempty"`
	Slack   string   `json:"slack,omitempty"`
	Email   string   `json:"email,omitempty"`
	Tags    []string `json:"tags,omitempty"`
}

// Domain is an inline domain catalog entry defined in mapture.yaml.
type Domain struct {
	ID                     string   `json:"id"`
	Name                   string   `json:"name"`
	Description            string   `json:"description,omitempty"`
	OwnerTeams             []string `json:"ownerTeams"`
	AllowedOutboundDomains []string `json:"allowedOutboundDomains,omitempty"`
	AllowedInboundDomains  []string `json:"allowedInboundDomains,omitempty"`
	Tags                   []string `json:"tags,omitempty"`
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
	FailOnUnknownNode      bool     `json:"failOnUnknownNode"`
	RequireMetadataOn      []string `json:"requireMetadataOn"`
	WarnOnOrphanedNodes    bool     `json:"warnOnOrphanedNodes"`
	WarnOnDeprecatedEvents bool     `json:"warnOnDeprecatedEvents"`
}

// UI configures optional web explorer presentation settings.
type UI struct {
	DefaultLayout string     `json:"defaultLayout"`
	NodeColors    NodeColors `json:"nodeColors"`
}

// NodeColors controls node-type colors used by the web explorer.
type NodeColors struct {
	Service  string `json:"service"`
	API      string `json:"api"`
	Database string `json:"database"`
	Event    string `json:"event"`
}

var defaultNodeColors = NodeColors{
	Service:  "#1664d9",
	API:      "#0f8f78",
	Database: "#a56614",
	Event:    "#a73f7f",
}

// DefaultLayoutELKHorizontal is the default explorer layout written into UI config defaults.
const DefaultLayoutELKHorizontal = "elk-horizontal"

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
	if err := validateTagVocabulary(cfg.Tags); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if !cfg.Languages.PHP && !cfg.Languages.Go && !cfg.Languages.TypeScript && !cfg.Languages.JavaScript {
		return nil, fmt.Errorf("%s: at least one language must be enabled", path)
	}
	cfg.applyDefaults()

	return &cfg, nil
}

// CatalogDir returns the absolute path to the configured legacy catalog directory.
func (c *Config) CatalogDir(configPath string) (string, error) {
	if c == nil {
		return "", errors.New("config is nil")
	}
	if strings.TrimSpace(c.Catalog.Dir) == "" {
		return "", nil
	}
	if filepath.IsAbs(c.Catalog.Dir) {
		return c.Catalog.Dir, nil
	}
	return filepath.Join(filepath.Dir(configPath), c.Catalog.Dir), nil
}

func (c *Config) applyDefaults() {
	if c.UI.DefaultLayout == "" {
		c.UI.DefaultLayout = DefaultLayoutELKHorizontal
	}
	if c.UI.NodeColors.Service == "" {
		c.UI.NodeColors.Service = defaultNodeColors.Service
	}
	if c.UI.NodeColors.API == "" {
		c.UI.NodeColors.API = defaultNodeColors.API
	}
	if c.UI.NodeColors.Database == "" {
		c.UI.NodeColors.Database = defaultNodeColors.Database
	}
	if c.UI.NodeColors.Event == "" {
		c.UI.NodeColors.Event = defaultNodeColors.Event
	}
	c.Tags = normalizeTags(c.Tags)
	for index := range c.Teams {
		c.Teams[index].Tags = normalizeTags(c.Teams[index].Tags)
	}
	for index := range c.Domains {
		c.Domains[index].Tags = normalizeTags(c.Domains[index].Tags)
	}
}

func validateTagVocabulary(tags []string) error {
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		if _, exists := seen[tag]; exists {
			return fmt.Errorf("duplicate tag %q", tag)
		}
		seen[tag] = struct{}{}
	}
	return nil
}

func normalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(tags))
	normalized := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(strings.ToLower(tag))
		if tag == "" {
			continue
		}
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		normalized = append(normalized, tag)
	}
	sort.Strings(normalized)
	return normalized
}
