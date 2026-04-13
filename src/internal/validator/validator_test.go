package validator

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/mandotpro/mapture.dev/src/internal/catalog"
	"github.com/mandotpro/mapture.dev/src/internal/config"
	"github.com/mandotpro/mapture.dev/src/internal/graph"
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

func TestBuildUsesProducerToEventToConsumerFlow(t *testing.T) {
	t.Parallel()

	_, cfg, cat, blocks := loadFixture(t, "../../../examples/demo")

	result, err := Build(cfg, cat, blocks)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	if !hasEdge(result.Graph, graph.Edge{
		From: "service:checkout-service",
		To:   "event:order.placed",
		Type: graph.EdgeEmits,
	}) {
		t.Fatalf("expected emit edge from service to event, got %#v", result.Graph.Edges)
	}

	if !hasEdge(result.Graph, graph.Edge{
		From: "event:order.placed",
		To:   "api:payment-api",
		Type: graph.EdgeConsumes,
	}) {
		t.Fatalf("expected consume edge from event to consumer, got %#v", result.Graph.Edges)
	}
}

func TestBuildCanonicalizesPairedEventDefinitionNodes(t *testing.T) {
	t.Parallel()

	_, cfg, cat, blocks := loadFixture(t, "../../../examples/ecommerce")

	result, err := Build(cfg, cat, blocks)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	for _, node := range result.Graph.Nodes {
		if node.ID == "event:order-placed-event" {
			t.Fatalf("expected aliased event node id to be removed, got %#v", result.Graph.Nodes)
		}
	}
	if !hasNode(result.Graph, "event:order.placed") {
		t.Fatalf("expected canonical event node event:order.placed, got %#v", result.Graph.Nodes)
	}
	if !hasEdge(result.Graph, graph.Edge{
		From: "service:checkout-service",
		To:   "event:order.placed",
		Type: graph.EdgeDependsOn,
	}) {
		t.Fatalf("expected relation to use canonical event id, got %#v", result.Graph.Edges)
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
				"domain": "orders",
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

func TestBuildScopedSynthesizesBoundaryNodeForMissingTarget(t *testing.T) {
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
				"calls": {{Type: "api", ID: "payment-api"}},
			},
		},
	}

	result, err := Build(cfg, cat, blocks, BuildOptions{Scoped: true})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if !hasEdge(result.Graph, graph.Edge{
		From: "service:checkout-service",
		To:   "api:payment-api",
		Type: graph.EdgeCalls,
	}) {
		t.Fatalf("expected scoped edge to survive, got %#v", result.Graph.Edges)
	}

	for _, node := range result.Graph.Nodes {
		if node.ID != "api:payment-api" {
			continue
		}
		if node.File != "" {
			t.Fatalf("expected synthesized node to have no file, got %#v", node)
		}
		if !strings.Contains(node.Summary, "out-of-scope") {
			t.Fatalf("expected synthesized node summary marker, got %#v", node)
		}
		return
	}

	t.Fatalf("expected synthesized api:payment-api node, got %#v", result.Graph.Nodes)
}

func TestBuildScopedKeepsOnlyReferencedEvents(t *testing.T) {
	t.Parallel()

	root, cfg, cat, _ := loadFixture(t, "../../../examples/ecommerce")
	scopedCfg := *cfg
	scopedCfg.Scan.Include = []string{"./src/php/orders"}

	blocks, err := scanner.Scan(root, &scopedCfg)
	if err != nil {
		t.Fatalf("scanner.Scan: %v", err)
	}

	result, err := Build(cfg, cat, blocks, BuildOptions{Scoped: true})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	if !hasNode(result.Graph, "event:order.placed") {
		t.Fatalf("expected referenced event node in scoped graph, got %#v", result.Graph.Nodes)
	}
	if hasNode(result.Graph, "event:inventory.reserved") {
		t.Fatalf("did not expect unrelated event in scoped graph, got %#v", result.Graph.Nodes)
	}
}

func TestBuildWarnsOnDeprecatedEventDefinition(t *testing.T) {
	t.Parallel()

	cfg := strictConfig()
	cfg.Validation.WarnOnDeprecatedEvents = true
	cat := minimalCatalog()
	blocks := []scanner.RawBlock{
		{
			Kind: "arch",
			File: "src/contracts.go",
			Line: 1,
			Fields: map[string]string{
				"node":        "event order-placed-event",
				"name":        "Order Placed Event",
				"domain":      "orders",
				"owner":       "team-commerce",
				"status":      "deprecated",
				"description": "Legacy order event definition.",
			},
		},
		{
			Kind: "event",
			File: "src/contracts.go",
			Line: 1,
			Fields: map[string]string{
				"id":     "order.placed",
				"role":   "definition",
				"domain": "orders",
				"owner":  "team-commerce",
			},
		},
	}

	result, err := Build(cfg, cat, blocks)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if !hasDiagnostic(result.Diagnostics, severityWarning, "deprecated_event") {
		t.Fatalf("expected deprecated_event warning, got %#v", result.Diagnostics)
	}
}

func TestBuildMigrationFixtureWarnsOnlyOnDeprecatedLegacyEvent(t *testing.T) {
	t.Parallel()

	_, cfg, cat, blocks := loadFixture(t, "../../../examples/migration")

	result, err := Build(cfg, cat, blocks)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if len(result.Diagnostics) == 0 {
		t.Fatal("expected deprecated event warnings for migration fixture")
	}

	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Severity != severityWarning {
			t.Fatalf("expected warnings only, got %#v", result.Diagnostics)
		}
		if diagnostic.Code != "deprecated_event" {
			t.Fatalf("expected deprecated_event warnings only, got %#v", result.Diagnostics)
		}
		if !strings.Contains(diagnostic.Message, "legacy.order.created") {
			t.Fatalf("expected warning message to mention legacy.order.created, got %#v", result.Diagnostics)
		}
	}
}

func TestBuildRejectsUnknownConfiguredTags(t *testing.T) {
	t.Parallel()

	cfg := strictConfig()
	cfg.Tags = []string{"critical-path"}

	team := catalog.Team{ID: "team-commerce", Name: "Commerce", Tags: []string{"critical-path", "unknown-tag"}}
	domain := catalog.Domain{
		ID:         "orders",
		Name:       "Orders",
		OwnerTeams: []string{"team-commerce"},
		Tags:       []string{"critical-path"},
	}
	cat := &catalog.Catalog{
		Teams:       []catalog.Team{team},
		Domains:     []catalog.Domain{domain},
		TeamsByID:   map[string]catalog.Team{team.ID: team},
		DomainsByID: map[string]catalog.Domain{domain.ID: domain},
	}

	result, err := Build(cfg, cat, nil)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !hasDiagnostic(result.Diagnostics, severityError, "unknown_tag") {
		t.Fatalf("expected unknown_tag diagnostic, got %#v", result.Diagnostics)
	}
}

func TestBuildParsesDirectTagsAndComputesEffectiveTags(t *testing.T) {
	t.Parallel()

	cfg := strictConfig()
	cfg.Tags = []string{"critical-path", "customer-facing", "pci"}
	team := catalog.Team{ID: "team-commerce", Name: "Commerce", Tags: []string{"customer-facing"}}
	domain := catalog.Domain{
		ID:         "orders",
		Name:       "Orders",
		OwnerTeams: []string{"team-commerce"},
		Tags:       []string{"critical-path"},
	}
	cat := &catalog.Catalog{
		Teams:       []catalog.Team{team},
		Domains:     []catalog.Domain{domain},
		TeamsByID:   map[string]catalog.Team{team.ID: team},
		DomainsByID: map[string]catalog.Domain{domain.ID: domain},
	}
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
				"tags":   "pci, critical-path, pci",
			},
		},
		{
			Kind: "event",
			File: "src/app.go",
			Line: 10,
			Fields: map[string]string{
				"id":       "order.placed",
				"role":     "trigger",
				"domain":   "orders",
				"owner":    "team-commerce",
				"producer": "CheckoutService::placeOrder",
				"tags":     "pci, customer-facing, pci",
			},
		},
	}

	result, err := Build(cfg, cat, blocks)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	checkout := findGraphNode(t, result.Graph, "service:checkout-service")
	if got, want := strings.Join(checkout.Tags, ","), "critical-path,pci"; got != want {
		t.Fatalf("unexpected direct tags: got %q want %q", got, want)
	}
	if got, want := strings.Join(checkout.EffectiveTags, ","), "critical-path,customer-facing,pci"; got != want {
		t.Fatalf("unexpected effective tags: got %q want %q", got, want)
	}

	event := findGraphNode(t, result.Graph, "event:order.placed")
	if got, want := strings.Join(event.Tags, ","), "customer-facing,pci"; got != want {
		t.Fatalf("unexpected event direct tags: got %q want %q", got, want)
	}
	if got, want := strings.Join(event.EffectiveTags, ","), "critical-path,customer-facing,pci"; got != want {
		t.Fatalf("unexpected event effective tags: got %q want %q", got, want)
	}
}

func TestBuildRejectsUnknownDirectTag(t *testing.T) {
	t.Parallel()

	cfg := strictConfig()
	cfg.Tags = []string{"critical-path"}
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
				"tags":   "unknown-tag",
			},
		},
	}

	result, err := Build(cfg, cat, blocks)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !hasDiagnostic(result.Diagnostics, severityError, "unknown_tag") {
		t.Fatalf("expected unknown_tag diagnostic, got %#v", result.Diagnostics)
	}
}

func TestBuildParsesDirectFacets(t *testing.T) {
	t.Parallel()

	cfg := strictConfig()
	cfg.Facets = config.Facets{
		"event.type": {Label: "Event Type", Values: []string{"async", "queue"}},
		"db.type":    {Label: "Database Type", Values: []string{"tenant", "shared"}},
	}
	cat := minimalCatalog()
	blocks := []scanner.RawBlock{
		{
			Kind: "arch",
			File: "src/db.go",
			Line: 1,
			Fields: map[string]string{
				"node":    "database orders-db",
				"name":    "Orders DB",
				"domain":  "orders",
				"owner":   "team-commerce",
				"db.type": "tenant",
			},
		},
		{
			Kind: "event",
			File: "src/app.go",
			Line: 10,
			Fields: map[string]string{
				"id":         "order.placed",
				"role":       "trigger",
				"domain":     "orders",
				"owner":      "team-commerce",
				"producer":   "CheckoutService::placeOrder",
				"event.type": "async",
			},
		},
	}

	result, err := Build(cfg, cat, blocks)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	database := findGraphNode(t, result.Graph, "database:orders-db")
	if got := database.Facets["db.type"]; got != "tenant" {
		t.Fatalf("unexpected database facet value: %q", got)
	}

	event := findGraphNode(t, result.Graph, "event:order.placed")
	if got := event.Facets["event.type"]; got != "async" {
		t.Fatalf("unexpected event facet value: %q", got)
	}
}

func TestBuildRejectsUnknownFacetKeyAndValue(t *testing.T) {
	t.Parallel()

	cfg := strictConfig()
	cfg.Facets = config.Facets{
		"event.type": {Label: "Event Type", Values: []string{"async"}},
	}
	cat := minimalCatalog()
	blocks := []scanner.RawBlock{
		{
			Kind: "arch",
			File: "src/app.go",
			Line: 1,
			Fields: map[string]string{
				"node":         "service checkout-service",
				"name":         "Checkout Service",
				"domain":       "orders",
				"owner":        "team-commerce",
				"message.type": "queue",
				"event.type":   "event-bus",
			},
		},
	}

	result, err := Build(cfg, cat, blocks)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !hasDiagnostic(result.Diagnostics, severityError, "unknown_facet_key") {
		t.Fatalf("expected unknown_facet_key diagnostic, got %#v", result.Diagnostics)
	}
	if !hasDiagnostic(result.Diagnostics, severityError, "unknown_facet_value") {
		t.Fatalf("expected unknown_facet_value diagnostic, got %#v", result.Diagnostics)
	}
}

func TestBuildRejectsConflictingFacetValuesAcrossEventBlocks(t *testing.T) {
	t.Parallel()

	cfg := strictConfig()
	cfg.Facets = config.Facets{
		"event.type": {Label: "Event Type", Values: []string{"async", "queue"}},
	}
	cat := minimalCatalog()
	blocks := []scanner.RawBlock{
		{
			Kind: "event",
			File: "src/app.go",
			Line: 10,
			Fields: map[string]string{
				"id":         "order.placed",
				"role":       "trigger",
				"domain":     "orders",
				"owner":      "team-commerce",
				"producer":   "CheckoutService::placeOrder",
				"event.type": "async",
			},
		},
		{
			Kind: "event",
			File: "src/app.go",
			Line: 20,
			Fields: map[string]string{
				"id":         "order.placed",
				"role":       "listener",
				"domain":     "orders",
				"consumer":   "capture_payment",
				"event.type": "queue",
			},
		},
	}

	result, err := Build(cfg, cat, blocks)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !hasDiagnostic(result.Diagnostics, severityError, "conflicting_facet_value") {
		t.Fatalf("expected conflicting_facet_value diagnostic, got %#v", result.Diagnostics)
	}
}

func loadFixture(t *testing.T, rel string) (string, *config.Config, *catalog.Catalog, []scanner.RawBlock) {
	t.Helper()

	root, err := filepath.Abs(rel)
	if err != nil {
		t.Fatalf("Abs(%q): %v", rel, err)
	}

	configPath := filepath.Join(root, "mapture.yaml")
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	cat, err := catalog.Load(configPath, cfg)
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
			FailOnUnknownNode:   true,
		},
	}
}

func minimalCatalog() *catalog.Catalog {
	team := catalog.Team{ID: "team-commerce", Name: "Commerce"}
	domain := catalog.Domain{ID: "orders", Name: "Orders", OwnerTeams: []string{"team-commerce"}}
	return &catalog.Catalog{
		Teams:       []catalog.Team{team},
		Domains:     []catalog.Domain{domain},
		TeamsByID:   map[string]catalog.Team{team.ID: team},
		DomainsByID: map[string]catalog.Domain{domain.ID: domain},
	}
}

func hasEdge(g graph.Graph, want graph.Edge) bool {
	for _, edge := range g.Edges {
		if edge == want {
			return true
		}
	}
	return false
}

func hasNode(g graph.Graph, nodeID string) bool {
	for _, node := range g.Nodes {
		if node.ID == nodeID {
			return true
		}
	}
	return false
}

func hasDiagnostic(diagnostics []Diagnostic, severity string, code string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == severity && diagnostic.Code == code {
			return true
		}
	}
	return false
}

func findGraphNode(t *testing.T, g graph.Graph, nodeID string) graph.Node {
	t.Helper()

	for _, node := range g.Nodes {
		if node.ID == nodeID {
			return node
		}
	}

	t.Fatalf("node %q not found in %#v", nodeID, g.Nodes)
	return graph.Node{}
}
