// Package html writes static explorer bundles backed by canonical export JSON.
package html

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	exportercanonical "github.com/mandotpro/mapture.dev/src/internal/exporter/canonical"
	"github.com/mandotpro/mapture.dev/src/internal/webui"
)

// WriteBundle writes the static explorer bundle and sibling data.json into outputDir.
func WriteBundle(outputDir string, doc *exportercanonical.Document) error {
	if doc == nil {
		return fmt.Errorf("export html: document is required")
	}
	if outputDir == "" {
		return fmt.Errorf("export html: output directory is required")
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("export html: create output directory: %w", err)
	}
	if err := fs.WalkDir(webui.FS(), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := webui.ReadFile(path)
		if err != nil {
			return err
		}
		target := filepath.Join(outputDir, path)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	}); err != nil {
		return fmt.Errorf("export html: copy embedded bundle: %w", err)
	}

	payload, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("export html: marshal data.json: %w", err)
	}
	payload = append(payload, '\n')
	if err := os.WriteFile(filepath.Join(outputDir, "data.json"), payload, 0o644); err != nil {
		return fmt.Errorf("export html: write data.json: %w", err)
	}
	return nil
}
