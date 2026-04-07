// Package catalog defines the canonical catalog model (teams, domains,
// events) and loads it from a repo-local directory of YAML files.
//
// See PRD §13 for the authoritative schema. The types here are the
// in-memory projection of those YAML files and are the single source of
// truth consumed by scanner, validator, and exporter packages.
package catalog

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Catalog is the in-memory union of all catalog YAML files under a
// repo's architecture/ directory.
type Catalog struct {
	Teams   []Team   `yaml:"-"`
	Domains []Domain `yaml:"-"`
	Events  []Event  `yaml:"-"`
}

// Team mirrors an entry in teams.yaml. See PRD §13.1.
type Team struct {
	ID      string   `yaml:"id"`
	Name    string   `yaml:"name"`
	Contact string   `yaml:"contact,omitempty"`
	Slack   string   `yaml:"slack,omitempty"`
	Email   string   `yaml:"email,omitempty"`
	Tags    []string `yaml:"tags,omitempty"`
}

// Domain mirrors an entry in domains.yaml. See PRD §13.2.
type Domain struct {
	ID                     string   `yaml:"id"`
	Name                   string   `yaml:"name"`
	Description            string   `yaml:"description,omitempty"`
	OwnerTeams             []string `yaml:"ownerTeams"`
	AllowedOutboundDomains []string `yaml:"allowedOutboundDomains,omitempty"`
	AllowedInboundDomains  []string `yaml:"allowedInboundDomains,omitempty"`
	Tags                   []string `yaml:"tags,omitempty"`
}

// Event mirrors an entry in events.yaml. See PRD §13.3.
type Event struct {
	ID                   string   `yaml:"id"`
	Name                 string   `yaml:"name"`
	Description          string   `yaml:"description,omitempty"`
	Domain               string   `yaml:"domain"`
	OwnerTeam            string   `yaml:"ownerTeam"`
	Kind                 string   `yaml:"kind"`       // domain|integration|system|internal
	Visibility           string   `yaml:"visibility"` // internal|public|private|deprecated
	Status               string   `yaml:"status"`     // active|deprecated|experimental
	Version              int      `yaml:"version,omitempty"`
	PayloadSchema        string   `yaml:"payloadSchema,omitempty"`
	AllowedTargetDomains []string `yaml:"allowedTargetDomains,omitempty"`
	AllowedProducers     []string `yaml:"allowedProducers,omitempty"`
	AllowedConsumers     []string `yaml:"allowedConsumers,omitempty"`
	Deprecated           bool     `yaml:"deprecated,omitempty"`
	ReplacedBy           string   `yaml:"replacedBy,omitempty"`
	Tags                 []string `yaml:"tags,omitempty"`
}

// Load reads teams.yaml, domains.yaml, and events.yaml from dir.
// Missing files are tolerated (they produce empty slices); malformed
// YAML or IO errors are returned verbatim so callers can surface them.
func Load(dir string) (*Catalog, error) {
	c := &Catalog{}

	if err := readYAML(filepath.Join(dir, "teams.yaml"), &struct {
		Teams *[]Team `yaml:"teams"`
	}{Teams: &c.Teams}); err != nil {
		return nil, err
	}
	if err := readYAML(filepath.Join(dir, "domains.yaml"), &struct {
		Domains *[]Domain `yaml:"domains"`
	}{Domains: &c.Domains}); err != nil {
		return nil, err
	}
	if err := readYAML(filepath.Join(dir, "events.yaml"), &struct {
		Events *[]Event `yaml:"events"`
	}{Events: &c.Events}); err != nil {
		return nil, err
	}

	return c, nil
}

func readYAML(path string, out any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := yaml.Unmarshal(b, out); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}
