package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadBuildsIndexes(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeCatalogFile(t, root, "teams.yaml", "teams:\n  - id: team-commerce\n    name: Commerce Team\n    email: commerce@example.com\n")
	writeCatalogFile(t, root, "domains.yaml", "domains:\n  - id: orders\n    name: Orders\n    ownerTeams: [team-commerce]\n")
	writeCatalogFile(t, root, "events.yaml", "events:\n  - id: order.placed\n    name: Order Placed\n    domain: orders\n    ownerTeam: team-commerce\n    kind: domain\n    visibility: internal\n    status: active\n")

	c, err := Load(root)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if _, ok := c.TeamsByID["team-commerce"]; !ok {
		t.Fatalf("expected team index to be populated")
	}
	if _, ok := c.DomainsByID["orders"]; !ok {
		t.Fatalf("expected domain index to be populated")
	}
	if _, ok := c.EventsByID["order.placed"]; !ok {
		t.Fatalf("expected event index to be populated")
	}
}

func TestLoadRejectsDuplicateTeamIDs(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeCatalogFile(t, root, "teams.yaml", "teams:\n  - id: team-commerce\n    name: One\n  - id: team-commerce\n    name: Two\n")
	writeCatalogFile(t, root, "domains.yaml", "domains:\n  - id: orders\n    name: Orders\n    ownerTeams: [team-commerce]\n")
	writeCatalogFile(t, root, "events.yaml", "events: []\n")

	_, err := Load(root)
	if err == nil {
		t.Fatalf("expected duplicate team ids to fail")
	}
	if !strings.Contains(err.Error(), "duplicate team id") {
		t.Fatalf("expected duplicate id error, got %v", err)
	}
}

func TestLoadRejectsUnknownDomainOwnerTeam(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeCatalogFile(t, root, "teams.yaml", "teams:\n  - id: team-commerce\n    name: Commerce Team\n")
	writeCatalogFile(t, root, "domains.yaml", "domains:\n  - id: orders\n    name: Orders\n    ownerTeams: [team-missing]\n")
	writeCatalogFile(t, root, "events.yaml", "events: []\n")

	_, err := Load(root)
	if err == nil {
		t.Fatalf("expected missing team reference to fail")
	}
	if !strings.Contains(err.Error(), "unknown team") {
		t.Fatalf("expected unknown team error, got %v", err)
	}
}

func TestLoadRejectsInvalidEventStatusViaSchema(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeCatalogFile(t, root, "teams.yaml", "teams:\n  - id: team-commerce\n    name: Commerce Team\n")
	writeCatalogFile(t, root, "domains.yaml", "domains:\n  - id: orders\n    name: Orders\n    ownerTeams: [team-commerce]\n")
	writeCatalogFile(t, root, "events.yaml", "events:\n  - id: order.placed\n    name: Order Placed\n    domain: orders\n    ownerTeam: team-commerce\n    kind: domain\n    visibility: internal\n    status: random\n")

	_, err := Load(root)
	if err == nil {
		t.Fatalf("expected invalid event status to fail")
	}
	if !strings.Contains(err.Error(), "status") || !strings.Contains(err.Error(), "random") {
		t.Fatalf("expected schema error mentioning status/random, got %v", err)
	}
}

func writeCatalogFile(t *testing.T, root string, name string, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}
