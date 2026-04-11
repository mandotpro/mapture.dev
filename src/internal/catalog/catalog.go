// Package catalog defines the canonical catalog model (teams and domains)
// and loads it from mapture.yaml and optional legacy YAML files.
//
// These types are the in-memory projection of the catalog files and the
// single source of truth consumed by scanner, validator, and exporter packages.
package catalog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mandotpro/mapture.dev/src/internal/config"
	"github.com/mandotpro/mapture.dev/src/internal/schema"
)

// Catalog is the in-memory union of inline config entries and optional
// legacy catalog YAML files.
type Catalog struct {
	Teams       []Team            `json:"-"`
	Domains     []Domain          `json:"-"`
	TeamsByID   map[string]Team   `json:"-"`
	DomainsByID map[string]Domain `json:"-"`
}

// Team mirrors an inline or file-backed team entry.
type Team = config.Team

// Domain mirrors an inline or file-backed domain entry.
type Domain = config.Domain

type teamsFile struct {
	Teams []Team `json:"teams"`
}

type domainsFile struct {
	Domains []Domain `json:"domains"`
}

// Load builds catalog indexes from inline mapture.yaml entries and, when
// configured, legacy teams.yaml / domains.yaml files under catalog.dir.
func Load(configPath string, cfg *config.Config) (*Catalog, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	c := &Catalog{
		Teams:       append([]Team(nil), cfg.Teams...),
		Domains:     append([]Domain(nil), cfg.Domains...),
		TeamsByID:   make(map[string]Team, len(cfg.Teams)),
		DomainsByID: make(map[string]Domain, len(cfg.Domains)),
	}

	dir, err := cfg.CatalogDir(configPath)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(dir) != "" {
		var teamDoc teamsFile
		if err := readYAML(filepath.Join(dir, "teams.yaml"), len(c.Teams) == 0, schema.TeamsDefinition, &teamDoc); err != nil {
			return nil, err
		}
		var domainDoc domainsFile
		if err := readYAML(filepath.Join(dir, "domains.yaml"), len(c.Domains) == 0, schema.DomainsDefinition, &domainDoc); err != nil {
			return nil, err
		}
		c.Teams = append(c.Teams, teamDoc.Teams...)
		c.Domains = append(c.Domains, domainDoc.Domains...)
	}

	c.TeamsByID = make(map[string]Team, len(c.Teams))
	c.DomainsByID = make(map[string]Domain, len(c.Domains))

	if len(c.Teams) == 0 {
		return nil, fmt.Errorf("%s: no teams configured; add inline teams to mapture.yaml or point catalog.dir at legacy files", configPath)
	}
	if len(c.Domains) == 0 {
		return nil, fmt.Errorf("%s: no domains configured; add inline domains to mapture.yaml or point catalog.dir at legacy files", configPath)
	}

	if err := c.buildIndexes(); err != nil {
		return nil, err
	}
	if err := c.validateReferences(); err != nil {
		return nil, err
	}

	return c, nil
}

func readYAML(path string, required bool, def schema.Definition, out any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) && !required {
			return nil
		}
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := schema.DecodeYAML(def, path, b, out); err != nil {
		return err
	}
	return nil
}

func (c *Catalog) buildIndexes() error {
	for _, team := range c.Teams {
		if _, exists := c.TeamsByID[team.ID]; exists {
			return fmt.Errorf("duplicate team id %q", team.ID)
		}
		c.TeamsByID[team.ID] = team
	}

	for _, domain := range c.Domains {
		if _, exists := c.DomainsByID[domain.ID]; exists {
			return fmt.Errorf("duplicate domain id %q", domain.ID)
		}
		c.DomainsByID[domain.ID] = domain
	}

	return nil
}

func (c *Catalog) validateReferences() error {
	for _, domain := range c.Domains {
		for _, owner := range domain.OwnerTeams {
			if _, exists := c.TeamsByID[owner]; !exists {
				return fmt.Errorf("domain %q references unknown team %q", domain.ID, owner)
			}
		}
		for _, target := range domain.AllowedOutboundDomains {
			if _, exists := c.DomainsByID[target]; !exists {
				return fmt.Errorf("domain %q references unknown outbound domain %q", domain.ID, target)
			}
		}
		for _, source := range domain.AllowedInboundDomains {
			if _, exists := c.DomainsByID[source]; !exists {
				return fmt.Errorf("domain %q references unknown inbound domain %q", domain.ID, source)
			}
		}
	}
	return nil
}
