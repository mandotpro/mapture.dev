package server

import (
	"github.com/mandotpro/mapture.dev/src/internal/catalog"
	"github.com/mandotpro/mapture.dev/src/internal/config"
	exportercanonical "github.com/mandotpro/mapture.dev/src/internal/exporter/canonical"
	"github.com/mandotpro/mapture.dev/src/internal/graph"
	"github.com/mandotpro/mapture.dev/src/internal/validator"
)

const explorerPayloadSchemaVersion = 1

// ExplorerPayload is the canonical JSON contract consumed by the web UI.
type ExplorerPayload struct {
	SchemaVersion int                `json:"schemaVersion"`
	Graph         graph.Graph        `json:"graph"`
	Catalog       ExplorerCatalog    `json:"catalog"`
	Validation    ExplorerValidation `json:"validation"`
	UI            config.UI          `json:"ui"`
	Meta          ExplorerMeta       `json:"meta"`
}

// ExplorerCatalog carries the catalog entities needed by the explorer.
type ExplorerCatalog struct {
	Teams   []catalog.Team   `json:"teams"`
	Domains []catalog.Domain `json:"domains"`
}

// ExplorerValidation carries diagnostics plus summary metadata.
type ExplorerValidation struct {
	Diagnostics []validator.Diagnostic `json:"diagnostics,omitempty"`
	Summary     ValidationSummary      `json:"summary"`
}

// ValidationSummary is a lightweight aggregate for the explorer header.
type ValidationSummary struct {
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
	Nodes    int `json:"nodes"`
	Edges    int `json:"edges"`
}

// ExplorerMeta carries UI-boot metadata unrelated to the graph itself.
type ExplorerMeta struct {
	ProjectID   string `json:"projectId"`
	SourceLabel string `json:"sourceLabel"`
	Mode        string `json:"mode"`
}

func explorerPayloadFromCanonical(doc *exportercanonical.Document) *ExplorerPayload {
	if doc == nil {
		return nil
	}

	return &ExplorerPayload{
		SchemaVersion: explorerPayloadSchemaVersion,
		Graph:         doc.Graph,
		Catalog: ExplorerCatalog{
			Teams:   append([]catalog.Team(nil), doc.Catalog.Teams...),
			Domains: append([]catalog.Domain(nil), doc.Catalog.Domains...),
		},
		Validation: ExplorerValidation{
			Diagnostics: append([]validator.Diagnostic(nil), doc.Validation.Diagnostics...),
			Summary: ValidationSummary{
				Errors:   doc.Validation.Summary.Errors,
				Warnings: doc.Validation.Summary.Warnings,
				Nodes:    doc.Validation.Summary.Nodes,
				Edges:    doc.Validation.Summary.Edges,
			},
		},
		UI: doc.UI,
		Meta: ExplorerMeta{
			ProjectID:   doc.Source.ProjectRoot,
			SourceLabel: doc.Meta.SourceLabel,
			Mode:        doc.Meta.Mode,
		},
	}
}
