// Package visualization builds the explorer-facing JSON derived from the JGF export.
package visualization

import (
	"fmt"
	"sort"

	"github.com/mandotpro/mapture.dev/src/internal/catalog"
	"github.com/mandotpro/mapture.dev/src/internal/config"
	jgfexport "github.com/mandotpro/mapture.dev/src/internal/exporter/jgf"
	"github.com/mandotpro/mapture.dev/src/internal/graph"
	"github.com/mandotpro/mapture.dev/src/internal/validator"
)

// SchemaVersion is the stable public visualisation export schema version.
const SchemaVersion = 1

// Document is the explorer-facing visualisation export derived from JGF.
type Document struct {
	SchemaVersion int         `json:"schemaVersion"`
	GeneratedAt   string      `json:"generatedAt"`
	ToolVersion   string      `json:"toolVersion"`
	Source        Source      `json:"source"`
	Graph         graph.Graph `json:"graph"`
	Catalog       Catalog     `json:"catalog"`
	Validation    Validation  `json:"validation"`
	UI            UIConfig    `json:"ui"`
	Meta          Meta        `json:"meta"`
}

// Source describes where and how the visualisation export was produced.
type Source = jgfexport.Source

// Catalog contains the team/domain metadata needed by the explorer.
type Catalog struct {
	Tags    []string         `json:"tags,omitempty"`
	Teams   []catalog.Team   `json:"teams"`
	Domains []catalog.Domain `json:"domains"`
}

// Validation carries diagnostics plus a summary snapshot for the explorer.
type Validation struct {
	Summary     ValidationSummary      `json:"summary"`
	Diagnostics []validator.Diagnostic `json:"diagnostics,omitempty"`
}

// ValidationSummary is a small aggregate for visualisation headers and gating.
type ValidationSummary = jgfexport.ValidationSummary

// Meta carries generic consumption metadata unrelated to the normalized graph.
type Meta = jgfexport.Meta

// UIConfig stores explorer defaults that downstream visualisations need.
type UIConfig = config.UI

// FromJGF converts a JGF document into the explorer-facing visualisation export.
func FromJGF(doc *jgfexport.Document) (*Document, error) {
	if doc == nil {
		return nil, fmt.Errorf("visualization export requires JGF document")
	}

	meta := doc.Graph.Metadata.Mapture
	nodeIDs := make([]string, 0, len(doc.Graph.Nodes))
	for id := range doc.Graph.Nodes {
		nodeIDs = append(nodeIDs, id)
	}
	sort.Strings(nodeIDs)

	nodes := make([]graph.Node, 0, len(nodeIDs))
	for _, id := range nodeIDs {
		node := doc.Graph.Nodes[id]
		nodes = append(nodes, graph.Node{
			ID:            id,
			Type:          node.Metadata.Type,
			Name:          node.Label,
			Domain:        node.Metadata.Domain,
			Owner:         node.Metadata.Owner,
			File:          node.Metadata.File,
			Line:          node.Metadata.Line,
			Symbol:        node.Metadata.Symbol,
			Summary:       node.Metadata.Summary,
			Tags:          append([]string(nil), node.Metadata.Tags...),
			EffectiveTags: append([]string(nil), node.Metadata.EffectiveTags...),
		})
	}

	edges := make([]graph.Edge, 0, len(doc.Graph.Edges))
	for _, edge := range doc.Graph.Edges {
		edges = append(edges, graph.Edge{
			From: edge.Source,
			To:   edge.Target,
			Type: edge.Relation,
		})
	}

	graphDoc := graph.Graph{
		SchemaVersion: graph.SchemaVersion,
		Metadata: graph.Metadata{
			GeneratedAt:    meta.GeneratedAt,
			ScannerVersion: meta.ToolVersion,
			SourceRoot:     meta.Source.ProjectRoot,
		},
		Nodes: nodes,
		Edges: edges,
	}

	return &Document{
		SchemaVersion: SchemaVersion,
		GeneratedAt:   meta.GeneratedAt,
		ToolVersion:   meta.ToolVersion,
		Source:        meta.Source,
		Graph:         graphDoc,
		Catalog: Catalog{
			Tags:    append([]string(nil), meta.Catalog.Tags...),
			Teams:   append([]catalog.Team(nil), meta.Catalog.Teams...),
			Domains: append([]catalog.Domain(nil), meta.Catalog.Domains...),
		},
		Validation: Validation{
			Summary:     meta.Validation.Summary,
			Diagnostics: append([]validator.Diagnostic(nil), meta.Validation.Diagnostics...),
		},
		UI:   meta.UI,
		Meta: meta.Meta,
	}, nil
}

// Result converts the visualisation export back into the validator result shape.
func (d *Document) Result() validator.Result {
	return validator.Result{
		Graph:       d.Graph,
		Diagnostics: append([]validator.Diagnostic(nil), d.Validation.Diagnostics...),
	}
}
