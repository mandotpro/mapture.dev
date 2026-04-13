package jgf_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/mandotpro/mapture.dev/src/internal/catalog"
	"github.com/mandotpro/mapture.dev/src/internal/config"
	"github.com/mandotpro/mapture.dev/src/internal/exporter/jgf"
	exportervis "github.com/mandotpro/mapture.dev/src/internal/exporter/visualization"
	"github.com/mandotpro/mapture.dev/src/internal/graph"
	"github.com/mandotpro/mapture.dev/src/internal/scanner"
	"github.com/mandotpro/mapture.dev/src/internal/schema"
	"github.com/mandotpro/mapture.dev/src/internal/validator"
)

func TestFixtureExportsMatchSchemaAndGoldens(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		fixtureRel  string
		configPath  string
		projectRoot string
		golden      string
	}{
		{
			name:        "demo",
			fixtureRel:  "../../../../examples/demo",
			configPath:  "/repo/examples/demo/mapture.yaml",
			projectRoot: "/repo/examples/demo",
			golden:      "testdata/demo.golden.json",
		},
		{
			name:        "ecommerce",
			fixtureRel:  "../../../../examples/ecommerce",
			configPath:  "/repo/examples/ecommerce/mapture.yaml",
			projectRoot: "/repo/examples/ecommerce",
			golden:      "testdata/ecommerce.golden.json",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := marshalFixtureExport(t, tc.fixtureRel, tc.configPath, tc.projectRoot)
			if err := schema.ValidateJSON(schema.JSONGraphDefinition, tc.golden, got); err != nil {
				t.Fatalf("json graph schema validation failed: %v", err)
			}

			want, err := os.ReadFile(tc.golden)
			if err != nil {
				t.Fatalf("read golden: %v", err)
			}
			if string(got) != string(want) {
				t.Fatalf("unexpected json graph export\nwant:\n%s\ngot:\n%s", want, got)
			}
		})
	}
}

func TestVisualizationConversionMatchesSchema(t *testing.T) {
	t.Parallel()

	doc := buildFixtureJGF(t, "../../../../examples/demo", "/repo/examples/demo/mapture.yaml", "/repo/examples/demo")
	vis, err := exportervis.FromJGF(doc)
	if err != nil {
		t.Fatalf("FromJGF: %v", err)
	}
	data, err := json.MarshalIndent(vis, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent: %v", err)
	}
	if err := schema.ValidateJSON(schema.VisualizationDefinition, "visualization.json", data); err != nil {
		t.Fatalf("visualization schema validation failed: %v", err)
	}
}

func marshalFixtureExport(t *testing.T, rel string, configPath string, sourceRoot string) []byte {
	t.Helper()

	doc := buildFixtureJGF(t, rel, configPath, sourceRoot)
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent: %v", err)
	}
	return append(data, '\n')
}

func buildFixtureJGF(t *testing.T, rel string, configPath string, sourceRoot string) *jgf.Document {
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

	doc, err := jgf.Build(jgf.BuildOptions{
		ConfigPath:  configPath,
		ProjectRoot: sourceRoot,
		Scopes:      nil,
		Config:      cfg,
		Catalog:     cat,
		Result:      result,
		ToolVersion: graph.DefaultScannerVersion,
		GeneratedAt: time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC),
		Mode:        jgf.ModeStatic,
		SourceLabel: "static",
	})
	if err != nil {
		t.Fatalf("jgf.Build: %v", err)
	}
	return doc
}

func fixtureAbs(rel string) (string, error) {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Abs(filepath.Join(filepath.Dir(file), rel))
}
