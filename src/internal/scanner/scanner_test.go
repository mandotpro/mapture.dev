package scanner

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/angelmanchev/mapture/src/internal/config"
)

func TestScanDemoFixture(t *testing.T) {
	t.Parallel()

	root, cfg := loadFixtureConfig(t, "../../../examples/demo")

	blocks, err := Scan(root, cfg)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(blocks) != 5 {
		t.Fatalf("expected 5 blocks, got %d", len(blocks))
	}

	var hasGoArch bool
	var hasPHPEvent bool
	var hasTSEvent bool
	for _, block := range blocks {
		switch {
		case block.Kind == "arch" && strings.HasSuffix(block.File, "src/go/ordersdb/ordersdb.go"):
			hasGoArch = true
		case block.Kind == "event" && strings.HasSuffix(block.File, "src/php/CheckoutService.php"):
			hasPHPEvent = true
		case block.Kind == "event" && strings.HasSuffix(block.File, "src/ts/PaymentApiClient.ts"):
			hasTSEvent = true
		}
	}

	if !hasGoArch || !hasPHPEvent || !hasTSEvent {
		t.Fatalf("expected demo scan to cover Go/PHP/TS blocks, got %#v", blocks)
	}
}

func TestScanEcommerceFixtureCoversRolesLanguagesAndRelations(t *testing.T) {
	t.Parallel()

	root, cfg := loadFixtureConfig(t, "../../../examples/ecommerce")

	blocks, err := Scan(root, cfg)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(blocks) != 39 {
		t.Fatalf("expected 39 blocks, got %d", len(blocks))
	}

	roles := make(map[string]struct{})
	relations := make(map[string]struct{})
	extensions := make(map[string]struct{})
	for _, block := range blocks {
		extensions[filepath.Ext(block.File)] = struct{}{}
		if role := block.Fields["role"]; role != "" {
			roles[role] = struct{}{}
		}
		for key := range block.Relations {
			relations[key] = struct{}{}
		}
	}

	for _, role := range []string{"definition", "trigger", "listener", "bridge-in", "bridge-out", "publisher", "subscriber"} {
		if _, ok := roles[role]; !ok {
			t.Fatalf("expected role %q in ecommerce fixture, got %#v", role, roles)
		}
	}
	for _, relation := range []string{"calls", "depends_on", "stores_in", "reads_from"} {
		if _, ok := relations[relation]; !ok {
			t.Fatalf("expected relation %q in ecommerce fixture, got %#v", relation, relations)
		}
	}
	for _, ext := range []string{".go", ".php", ".ts"} {
		if _, ok := extensions[ext]; !ok {
			t.Fatalf("expected extension %q in ecommerce fixture, got %#v", ext, extensions)
		}
	}
}

func TestScanRejectsMalformedCommentFixtures(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		fixture string
		wantErr string
	}{
		{name: "missing owner", fixture: "../../../examples/invalid/comment-missing-owner", wantErr: "arch.owner"},
		{name: "bad event role", fixture: "../../../examples/invalid/comment-bad-event-role", wantErr: "event.role"},
		{name: "unknown key", fixture: "../../../examples/invalid/comment-unknown-key", wantErr: "arch.foobar"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root, cfg := loadFixtureConfig(t, tc.fixture)
			_, err := Scan(root, cfg)
			if err == nil {
				t.Fatalf("expected scanner error for %s", tc.fixture)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected %q in error, got %v", tc.wantErr, err)
			}
		})
	}
}

func loadFixtureConfig(t *testing.T, rel string) (string, *config.Config) {
	t.Helper()

	root, err := filepath.Abs(rel)
	if err != nil {
		t.Fatalf("Abs(%q): %v", rel, err)
	}

	cfg, err := config.Load(filepath.Join(root, "mapture.yaml"))
	if err != nil {
		t.Fatalf("config.Load(%q): %v", rel, err)
	}

	return root, cfg
}
