package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	exportercanonical "github.com/mandotpro/mapture.dev/src/internal/exporter/canonical"
	"github.com/mandotpro/mapture.dev/src/internal/schema"
)

func loadExportFromFile(path string) (*exportercanonical.Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read export file %s: %w", path, err)
	}
	if err := schema.ValidateJSON(schema.CanonicalDefinition, filepath.Base(path), data); err != nil {
		return nil, err
	}

	var doc exportercanonical.Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("decode export file %s: %w", path, err)
	}

	if doc.Meta.Mode == "" {
		doc.Meta.Mode = exportercanonical.ModeOffline
	}
	doc.Meta.Mode = exportercanonical.ModeOffline
	doc.Meta.SourceLabel = fmt.Sprintf("file: %s", filepath.Base(path))
	return &doc, nil
}

func cloneDocument(doc *exportercanonical.Document) *exportercanonical.Document {
	if doc == nil {
		return nil
	}

	cloned := *doc
	cloned.Source.Scopes = append([]string(nil), doc.Source.Scopes...)
	cloned.Catalog.Teams = append(cloned.Catalog.Teams[:0:0], doc.Catalog.Teams...)
	cloned.Catalog.Domains = append(cloned.Catalog.Domains[:0:0], doc.Catalog.Domains...)
	cloned.Validation.Diagnostics = append(cloned.Validation.Diagnostics[:0:0], doc.Validation.Diagnostics...)
	cloned.Graph.Nodes = append(cloned.Graph.Nodes[:0:0], doc.Graph.Nodes...)
	cloned.Graph.Edges = append(cloned.Graph.Edges[:0:0], doc.Graph.Edges...)
	return &cloned
}
