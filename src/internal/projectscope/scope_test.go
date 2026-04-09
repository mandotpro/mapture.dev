package projectscope

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mandotpro/mapture.dev/src/internal/config"
)

func TestApplyNarrowsIncludesDeterministically(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	for _, dir := range []string{
		filepath.Join(root, "src", "checkout"),
		filepath.Join(root, "src", "shared"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("MkdirAll(%q): %v", dir, err)
		}
	}

	cfg := &config.Config{
		Scan: config.Scan{
			Include: []string{"./src"},
		},
	}

	applied, err := Apply(root, cfg, []string{"./src/shared", "./src/checkout", "./src/shared"})
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	if !applied.Scoped {
		t.Fatal("expected scoped result")
	}
	if len(applied.Config.Scan.Include) != 2 {
		t.Fatalf("expected 2 effective includes, got %d", len(applied.Config.Scan.Include))
	}
	if applied.Config.Scan.Include[0] != "./src/checkout" || applied.Config.Scan.Include[1] != "./src/shared" {
		t.Fatalf("unexpected includes: %#v", applied.Config.Scan.Include)
	}
}

func TestApplyRejectsMissingScopePath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	cfg := &config.Config{
		Scan: config.Scan{
			Include: []string{"./src"},
		},
	}

	_, err := Apply(root, cfg, []string{"./src/missing"})
	if err == nil {
		t.Fatal("expected missing scope path error")
	}
}

func TestApplyRejectsOutOfIncludeScope(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "pkg"), 0o755); err != nil {
		t.Fatalf("MkdirAll(pkg): %v", err)
	}
	cfg := &config.Config{
		Scan: config.Scan{
			Include: []string{"./src"},
		},
	}

	_, err := Apply(root, cfg, []string{"./pkg"})
	if err == nil {
		t.Fatal("expected out-of-include scope error")
	}
}
