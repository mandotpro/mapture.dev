package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mandotpro/mapture.dev/src/internal/config"
	exportervis "github.com/mandotpro/mapture.dev/src/internal/exporter/visualization"
	"github.com/mandotpro/mapture.dev/src/internal/schema"
)

func loadVisualizationFromFile(path string) (*exportervis.Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read export file %s: %w", path, err)
	}
	if err := schema.ValidateJSON(schema.VisualizationDefinition, filepath.Base(path), data); err != nil {
		return nil, err
	}

	var doc exportervis.Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("decode export file %s: %w", path, err)
	}

	if doc.Meta.Mode == "" {
		doc.Meta.Mode = "offline"
	}
	doc.Meta.Mode = "offline"
	doc.Meta.SourceLabel = fmt.Sprintf("file: %s", filepath.Base(path))
	return &doc, nil
}

func cloneVisualizationDocument(doc *exportervis.Document) *exportervis.Document {
	if doc == nil {
		return nil
	}

	cloned := *doc
	cloned.Source.Scopes = append([]string(nil), doc.Source.Scopes...)
	cloned.Catalog.Tags = append([]string(nil), doc.Catalog.Tags...)
	cloned.Catalog.Facets = cloneFacetDefinitions(doc.Catalog.Facets)
	cloned.Catalog.Teams = append(cloned.Catalog.Teams[:0:0], doc.Catalog.Teams...)
	cloned.Catalog.Domains = append(cloned.Catalog.Domains[:0:0], doc.Catalog.Domains...)
	cloned.Validation.Diagnostics = append(cloned.Validation.Diagnostics[:0:0], doc.Validation.Diagnostics...)
	cloned.Graph.Nodes = append(cloned.Graph.Nodes[:0:0], doc.Graph.Nodes...)
	for index := range cloned.Graph.Nodes {
		cloned.Graph.Nodes[index].Tags = append([]string(nil), doc.Graph.Nodes[index].Tags...)
		cloned.Graph.Nodes[index].EffectiveTags = append([]string(nil), doc.Graph.Nodes[index].EffectiveTags...)
		cloned.Graph.Nodes[index].Facets = cloneFacetAssignments(doc.Graph.Nodes[index].Facets)
	}
	cloned.Graph.Edges = append(cloned.Graph.Edges[:0:0], doc.Graph.Edges...)
	return &cloned
}

func cloneFacetAssignments(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}

	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func cloneFacetDefinitions(facets config.Facets) config.Facets {
	if len(facets) == 0 {
		return nil
	}

	cloned := make(config.Facets, len(facets))
	for id, definition := range facets {
		cloned[id] = config.FacetDefinition{
			Label:  definition.Label,
			Values: append([]string(nil), definition.Values...),
		}
	}
	return cloned
}
