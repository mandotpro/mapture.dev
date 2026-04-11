package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mandotpro/mapture.dev/src/internal/config"
)

func TestLoadBuildsIndexesFromInlineConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	configPath := filepath.Join(root, "mapture.yaml")
	cfg := &config.Config{
		Teams: []config.Team{
			{ID: "team-commerce", Name: "Commerce Team", Email: "commerce@example.com"},
		},
		Domains: []config.Domain{
			{ID: "orders", Name: "Orders", OwnerTeams: []string{"team-commerce"}},
		},
	}

	c, err := Load(configPath, cfg)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if _, ok := c.TeamsByID["team-commerce"]; !ok {
		t.Fatalf("expected team index to be populated")
	}
	if _, ok := c.DomainsByID["orders"]; !ok {
		t.Fatalf("expected domain index to be populated")
	}
}

func TestLoadSupportsLegacyCatalogDirForTeamsAndDomains(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	configPath := filepath.Join(root, "mapture.yaml")
	if err := os.MkdirAll(filepath.Join(root, "architecture"), 0o755); err != nil {
		t.Fatalf("mkdir architecture: %v", err)
	}
	writeCatalogFile(t, filepath.Join(root, "architecture"), "teams.yaml", "teams:\n  - id: team-commerce\n    name: Commerce Team\n    email: commerce@example.com\n")
	writeCatalogFile(t, filepath.Join(root, "architecture"), "domains.yaml", "domains:\n  - id: orders\n    name: Orders\n    ownerTeams: [team-commerce]\n")

	cfg := &config.Config{
		Catalog: config.Catalog{Dir: "./architecture"},
	}

	c, err := Load(configPath, cfg)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(c.Teams) != 1 || len(c.Domains) != 1 {
		t.Fatalf("unexpected catalog sizes: %+v", c)
	}
}

func TestLoadRejectsDuplicateTeamIDs(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	configPath := filepath.Join(root, "mapture.yaml")
	cfg := &config.Config{
		Teams: []config.Team{
			{ID: "team-commerce", Name: "One"},
			{ID: "team-commerce", Name: "Two"},
		},
		Domains: []config.Domain{
			{ID: "orders", Name: "Orders", OwnerTeams: []string{"team-commerce"}},
		},
	}

	_, err := Load(configPath, cfg)
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
	configPath := filepath.Join(root, "mapture.yaml")
	cfg := &config.Config{
		Teams: []config.Team{
			{ID: "team-commerce", Name: "Commerce Team"},
		},
		Domains: []config.Domain{
			{ID: "orders", Name: "Orders", OwnerTeams: []string{"team-missing"}},
		},
	}

	_, err := Load(configPath, cfg)
	if err == nil {
		t.Fatalf("expected missing team reference to fail")
	}
	if !strings.Contains(err.Error(), "unknown team") {
		t.Fatalf("expected unknown team error, got %v", err)
	}
}

func TestLoadRequiresTeamsAndDomainsFromEitherSource(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	configPath := filepath.Join(root, "mapture.yaml")

	_, err := Load(configPath, &config.Config{})
	if err == nil {
		t.Fatalf("expected missing catalog data to fail")
	}
	if !strings.Contains(err.Error(), "no teams configured") {
		t.Fatalf("expected missing teams error, got %v", err)
	}
}

func writeCatalogFile(t *testing.T, root string, name string, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}
