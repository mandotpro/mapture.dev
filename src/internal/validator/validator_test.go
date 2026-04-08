package validator

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/mandotpro/mapture.dev/src/internal/catalog"
	"github.com/mandotpro/mapture.dev/src/internal/config"
	"github.com/mandotpro/mapture.dev/src/internal/scanner"
)

func TestBuildDemoFixture(t *testing.T) {
	t.Parallel()

	_, cfg, cat, blocks := loadFixture(t, "../../../examples/demo")

	result, err := Build(cfg, cat, blocks)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if len(result.Graph.Nodes) != 4 {
		t.Fatalf("expected 4 graph nodes, got %d", len(result.Graph.Nodes))
	}
	if len(result.Graph.Edges) != 4 {
		t.Fatalf("expected 4 graph edges, got %d", len(result.Graph.Edges))
	}
}

func TestBuildRejectsUnknownArchDomain(t *testing.T) {
	t.Parallel()

	cfg := strictConfig()
	cat := minimalCatalog()
	blocks := []scanner.RawBlock{
		{
			Kind: "arch",
			File: "src/app.go",
			Line: 3,
			Fields: map[string]string{
				"node":   "service checkout-service",
				"name":   "Checkout Service",
				"domain": "missing",
				"owner":  "team-commerce",
			},
		},
	}

	_, err := Build(cfg, cat, blocks)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "unknown domain") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildRejectsEventDomainMismatch(t *testing.T) {
	t.Parallel()

	cfg := strictConfig()
	cat := minimalCatalog()
	blocks := []scanner.RawBlock{
		{
			Kind: "arch",
			File: "src/app.go",
			Line: 1,
			Fields: map[string]string{
				"node":   "event order-placed-event",
				"name":   "Order Placed Event",
				"domain": "billing",
				"owner":  "team-commerce",
			},
		},
		{
			Kind: "event",
			File: "src/app.go",
			Line: 1,
			Fields: map[string]string{
				"id":     "order.placed",
				"role":   "definition",
				"domain": "billing",
				"owner":  "team-commerce",
			},
		},
	}

	_, err := Build(cfg, cat, blocks)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "event_domain_mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildRejectsUnknownNodeTarget(t *testing.T) {
	t.Parallel()

	cfg := strictConfig()
	cat := minimalCatalog()
	blocks := []scanner.RawBlock{
		{
			Kind: "arch",
			File: "src/app.go",
			Line: 1,
			Fields: map[string]string{
				"node":   "service checkout-service",
				"name":   "Checkout Service",
				"domain": "orders",
				"owner":  "team-commerce",
			},
			Relations: map[string][]scanner.TargetRef{
				"calls": {{Type: "api", ID: "missing-api"}},
			},
		},
	}

	_, err := Build(cfg, cat, blocks)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "unknown_node_target") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildWarnsOnDeprecatedEvent(t *testing.T) {
	t.Parallel()

	cfg := strictConfig()
	cfg.Validation.WarnOnDeprecatedEvents = true
	cat := minimalCatalog()
	event := cat.EventsByID["order.placed"]
	event.Status = "deprecated"
	cat.EventsByID["order.placed"] = event
	cat.Events[0] = event
	blocks := []scanner.RawBlock{
		{
			Kind: "arch",
			File: "src/app.go",
			Line: 1,
			Fields: map[string]string{
				"node":   "service checkout-service",
				"name":   "Checkout Service",
				"domain": "orders",
				"owner":  "team-commerce",
			},
		},
		{
			Kind: "event",
			File: "src/app.go",
			Line: 8,
			Fields: map[string]string{
				"id":       "order.placed",
				"role":     "trigger",
				"domain":   "orders",
				"producer": "CheckoutService::placeOrder",
			},
		},
	}

	result, err := Build(cfg, cat, blocks)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if len(result.Diagnostics) == 0 || result.Diagnostics[0].Severity != severityWarning {
		t.Fatalf("expected deprecation warning, got %#v", result.Diagnostics)
	}
}

func loadFixture(t *testing.T, rel string) (string, *config.Config, *catalog.Catalog, []scanner.RawBlock) {
	t.Helper()

	root, err := filepath.Abs(rel)
	if err != nil {
		t.Fatalf("Abs(%q): %v", rel, err)
	}

	cfg, err := config.Load(filepath.Join(root, "mapture.yaml"))
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	catalogDir, err := cfg.CatalogDir(filepath.Join(root, "mapture.yaml"))
	if err != nil {
		t.Fatalf("CatalogDir: %v", err)
	}
	cat, err := catalog.Load(catalogDir)
	if err != nil {
		t.Fatalf("catalog.Load: %v", err)
	}
	blocks, err := scanner.Scan(root, cfg)
	if err != nil {
		t.Fatalf("scanner.Scan: %v", err)
	}

	return root, cfg, cat, blocks
}

func strictConfig() *config.Config {
	return &config.Config{
		Validation: config.Validation{
			FailOnUnknownDomain: true,
			FailOnUnknownTeam:   true,
			FailOnUnknownEvent:  true,
			FailOnUnknownNode:   true,
		},
	}
}

func minimalCatalog() *catalog.Catalog {
	team := catalog.Team{ID: "team-commerce", Name: "Commerce"}
	domain := catalog.Domain{ID: "orders", Name: "Orders", OwnerTeams: []string{"team-commerce"}}
	event := catalog.Event{
		ID:        "order.placed",
		Name:      "Order Placed",
		Domain:    "orders",
		OwnerTeam: "team-commerce",
		Status:    "active",
	}
	return &catalog.Catalog{
		Teams:       []catalog.Team{team},
		Domains:     []catalog.Domain{domain},
		Events:      []catalog.Event{event},
		TeamsByID:   map[string]catalog.Team{team.ID: team},
		DomainsByID: map[string]catalog.Domain{domain.ID: domain},
		EventsByID:  map[string]catalog.Event{event.ID: event},
	}
}
