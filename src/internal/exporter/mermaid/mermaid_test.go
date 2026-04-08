package mermaid

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mandotpro/mapture.dev/src/internal/catalog"
	"github.com/mandotpro/mapture.dev/src/internal/config"
	"github.com/mandotpro/mapture.dev/src/internal/graph"
	"github.com/mandotpro/mapture.dev/src/internal/scanner"
	"github.com/mandotpro/mapture.dev/src/internal/validator"
)

func TestRenderDemoGolden(t *testing.T) {
	t.Parallel()

	graph := loadFixtureGraph(t, "../../../../examples/demo")

	got, err := Render(&graph, Options{})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	want := readGolden(t, "testdata/demo.golden.mmd")
	if got != want {
		t.Fatalf("unexpected mermaid output\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestRenderEcommerceBillingFilterGolden(t *testing.T) {
	t.Parallel()

	graph := loadFixtureGraph(t, "../../../../examples/ecommerce")

	got, err := Render(&graph, Options{Domains: []string{"billing"}})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	want := readGolden(t, "testdata/ecommerce-billing.golden.mmd")
	if got != want {
		t.Fatalf("unexpected filtered mermaid output\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestRenderFiltersByNodeTypeAndTeam(t *testing.T) {
	t.Parallel()

	graph := loadFixtureGraph(t, "../../../../examples/ecommerce")

	got, err := Render(&graph, Options{
		NodeTypes: []string{"event"},
		Teams:     []string{"team-commerce"},
	})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	for _, want := range []string{"Order Placed", "Payment Captured", "Payment Failed", "Stripe Webhook Received"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in output, got:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{"Inventory Reserved", "Notification Sent Event", "Shipping Service"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("did not expect %q in output, got:\n%s", unwanted, got)
		}
	}
}

func readGolden(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", path, err)
	}
	return string(data)
}

func loadFixtureGraph(t *testing.T, rel string) graph.Graph {
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
	result, err := validator.Build(cfg, cat, blocks)
	if err != nil {
		t.Fatalf("validator.Build: %v", err)
	}
	return result.Graph
}
