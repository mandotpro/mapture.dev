// Package catalog defines the canonical catalog model (teams, domains,
// events) and loads it from a repo-local directory of YAML files.
//
// These types are the in-memory projection of the catalog files and the
// single source of truth consumed by scanner, validator, and exporter packages.
package catalog

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/angelmanchev/mapture/src/internal/schema"
)

// Catalog is the in-memory union of all catalog YAML files under a
// repo's architecture/ directory.
type Catalog struct {
	Teams       []Team            `json:"-"`
	Domains     []Domain          `json:"-"`
	Events      []Event           `json:"-"`
	TeamsByID   map[string]Team   `json:"-"`
	DomainsByID map[string]Domain `json:"-"`
	EventsByID  map[string]Event  `json:"-"`
}

// Team mirrors an entry in teams.yaml.
type Team struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Contact string   `json:"contact,omitempty"`
	Slack   string   `json:"slack,omitempty"`
	Email   string   `json:"email,omitempty"`
	Tags    []string `json:"tags,omitempty"`
}

// Domain mirrors an entry in domains.yaml.
type Domain struct {
	ID                     string   `json:"id"`
	Name                   string   `json:"name"`
	Description            string   `json:"description,omitempty"`
	OwnerTeams             []string `json:"ownerTeams"`
	AllowedOutboundDomains []string `json:"allowedOutboundDomains,omitempty"`
	AllowedInboundDomains  []string `json:"allowedInboundDomains,omitempty"`
	Tags                   []string `json:"tags,omitempty"`
}

// Event mirrors an entry in events.yaml.
type Event struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Description          string   `json:"description,omitempty"`
	Domain               string   `json:"domain"`
	OwnerTeam            string   `json:"ownerTeam"`
	Kind                 string   `json:"kind"`
	Visibility           string   `json:"visibility"`
	Status               string   `json:"status"`
	Version              int      `json:"version,omitempty"`
	PayloadSchema        string   `json:"payloadSchema,omitempty"`
	AllowedTargetDomains []string `json:"allowedTargetDomains,omitempty"`
	AllowedProducers     []string `json:"allowedProducers,omitempty"`
	AllowedConsumers     []string `json:"allowedConsumers,omitempty"`
	Deprecated           bool     `json:"deprecated,omitempty"`
	ReplacedBy           string   `json:"replacedBy,omitempty"`
	Tags                 []string `json:"tags,omitempty"`
}

type teamsFile struct {
	Teams []Team `json:"teams"`
}

type domainsFile struct {
	Domains []Domain `json:"domains"`
}

type eventsFile struct {
	Events []Event `json:"events"`
}

// Load reads teams.yaml, domains.yaml, and events.yaml from dir, validates
// them, and builds fast lookup maps for downstream validation stages.
func Load(dir string) (*Catalog, error) {
	var teamDoc teamsFile
	if err := readYAML(filepath.Join(dir, "teams.yaml"), true, schema.TeamsDefinition, &teamDoc); err != nil {
		return nil, err
	}

	var domainDoc domainsFile
	if err := readYAML(filepath.Join(dir, "domains.yaml"), true, schema.DomainsDefinition, &domainDoc); err != nil {
		return nil, err
	}

	var eventDoc eventsFile
	if err := readYAML(filepath.Join(dir, "events.yaml"), false, schema.EventsDefinition, &eventDoc); err != nil {
		return nil, err
	}

	c := &Catalog{
		Teams:       teamDoc.Teams,
		Domains:     domainDoc.Domains,
		Events:      eventDoc.Events,
		TeamsByID:   make(map[string]Team, len(teamDoc.Teams)),
		DomainsByID: make(map[string]Domain, len(domainDoc.Domains)),
		EventsByID:  make(map[string]Event, len(eventDoc.Events)),
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

	for _, event := range c.Events {
		if _, exists := c.EventsByID[event.ID]; exists {
			return fmt.Errorf("duplicate event id %q", event.ID)
		}
		c.EventsByID[event.ID] = event
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

	for _, event := range c.Events {
		if _, exists := c.DomainsByID[event.Domain]; !exists {
			return fmt.Errorf("event %q references unknown domain %q", event.ID, event.Domain)
		}
		if _, exists := c.TeamsByID[event.OwnerTeam]; !exists {
			return fmt.Errorf("event %q references unknown team %q", event.ID, event.OwnerTeam)
		}
		for _, domain := range event.AllowedTargetDomains {
			if _, exists := c.DomainsByID[domain]; !exists {
				return fmt.Errorf("event %q references unknown target domain %q", event.ID, domain)
			}
		}
		if event.ReplacedBy != "" {
			if _, exists := c.EventsByID[event.ReplacedBy]; !exists {
				return fmt.Errorf("event %q references unknown replacement event %q", event.ID, event.ReplacedBy)
			}
		}
	}

	return nil
}
