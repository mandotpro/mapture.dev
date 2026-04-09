package server

import (
	"github.com/mandotpro/mapture.dev/src/internal/catalog"
	"github.com/mandotpro/mapture.dev/src/internal/config"
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
