package graph_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/mandotpro/mapture.dev/src/internal/catalog"
	"github.com/mandotpro/mapture.dev/src/internal/config"
	"github.com/mandotpro/mapture.dev/src/internal/graph"
	"github.com/mandotpro/mapture.dev/src/internal/scanner"
	"github.com/mandotpro/mapture.dev/src/internal/schema"
	"github.com/mandotpro/mapture.dev/src/internal/validator"
)

func TestFixtureGraphsMatchSchemaAndGoldens(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fixtureRel string
		sourceRoot string
		golden     string
	}{
		{
			name:       "demo",
			fixtureRel: "../../../examples/demo",
			sourceRoot: "/repo/examples/demo",
			golden:     "testdata/demo.golden.json",
		},
		{
			name:       "ecommerce",
			fixtureRel: "../../../examples/ecommerce",
			sourceRoot: "/repo/examples/ecommerce",
			golden:     "testdata/ecommerce.golden.json",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := marshalFixtureGraph(t, tc.fixtureRel, tc.sourceRoot)
			if err := schema.ValidateJSON(schema.GraphDefinition, tc.golden, got); err != nil {
				t.Fatalf("graph schema validation failed: %v", err)
			}

			want, err := os.ReadFile(tc.golden)
			if err != nil {
				t.Fatalf("read golden: %v", err)
			}
			if string(got) != string(want) {
				t.Fatalf("unexpected graph JSON\nwant:\n%s\ngot:\n%s", want, got)
			}
		})
	}
}

func TestGraphSchemaVersionIsStable(t *testing.T) {
	t.Parallel()

	if graph.SchemaVersion != 1 {
		t.Fatalf("SchemaVersion = %d, want 1", graph.SchemaVersion)
	}
}

func marshalFixtureGraph(t *testing.T, rel string, sourceRoot string) []byte {
	t.Helper()

	root, err := fixtureAbs(rel)
	if err != nil {
		t.Fatalf("resolve fixture: %v", err)
	}

	cfg, err := config.Load(filepath.Join(root, "mapture.yaml"))
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	cat, err := catalog.Load(filepath.Join(root, "mapture.yaml"), cfg)
	if err != nil {
		t.Fatalf("catalog.Load: %v", err)
	}
	blocks, err := scanner.Scan(root, cfg)
	if err != nil {
		t.Fatalf("scanner.Scan: %v", err)
	}
	result, err := validator.Build(cfg, cat, blocks, validator.BuildOptions{
		SourceRoot:     sourceRoot,
		GeneratedAt:    time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC),
		ScannerVersion: graph.DefaultScannerVersion,
	})
	if err != nil {
		t.Fatalf("validator.Build: %v", err)
	}

	data, err := json.MarshalIndent(result.Graph, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent: %v", err)
	}
	return append(data, '\n')
}

func fixtureAbs(rel string) (string, error) {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Abs(filepath.Join(filepath.Dir(file), rel))
}
