// Package main is a temporary utility to assert schema layouts during testing.
package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/mandotpro/mapture.dev/src/internal/catalog"
	"github.com/mandotpro/mapture.dev/src/internal/config"
	"github.com/mandotpro/mapture.dev/src/internal/graph"
	"github.com/mandotpro/mapture.dev/src/internal/scanner"
	"github.com/mandotpro/mapture.dev/src/internal/validator"
)

func emit(rel string, sourceRoot string) error {
	root, err := filepath.Abs(rel)
	if err != nil {
		return err
	}
	cfg, err := config.Load(filepath.Join(root, "mapture.yaml"))
	if err != nil {
		return err
	}
	catalogDir, err := cfg.CatalogDir(filepath.Join(root, "mapture.yaml"))
	if err != nil {
		return err
	}
	cat, err := catalog.Load(catalogDir)
	if err != nil {
		return err
	}
	blocks, err := scanner.Scan(root, cfg)
	if err != nil {
		return err
	}
	result, err := validator.Build(cfg, cat, blocks, validator.BuildOptions{SourceRoot: sourceRoot, GeneratedAt: time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC), ScannerVersion: graph.DefaultScannerVersion})
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(result.Graph, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func main() {
	fmt.Println("===demo===")
	if err := emit("examples/demo", "/repo/examples/demo"); err != nil {
		panic(err)
	}
	fmt.Println("===ecommerce===")
	if err := emit("examples/ecommerce", "/repo/examples/ecommerce"); err != nil {
		panic(err)
	}
}
