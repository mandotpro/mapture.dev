// Package jgf builds the shareable JSON Graph Format export for downstream consumers.
package jgf

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/mandotpro/mapture.dev/src/internal/catalog"
	"github.com/mandotpro/mapture.dev/src/internal/config"
	"github.com/mandotpro/mapture.dev/src/internal/graph"
	"github.com/mandotpro/mapture.dev/src/internal/projectscope"
	"github.com/mandotpro/mapture.dev/src/internal/scanner"
	"github.com/mandotpro/mapture.dev/src/internal/validator"
)

// SchemaVersion is the stable public JGF export schema version.
const SchemaVersion = 1

// Mode values supported by the embedded Mapture export metadata.
const (
	ModeLive    = "live"
	ModeOffline = "offline"
	ModeStatic  = "static"
)

// Document is the shareable JGF export for a Mapture project.
type Document struct {
	Graph Graph `json:"graph"`
}

// Graph is the top-level JGF graph payload.
type Graph struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Label    string   `json:"label"`
	Directed bool     `json:"directed"`
	Nodes    NodeMap  `json:"nodes"`
	Edges    []Edge   `json:"edges"`
	Metadata Metadata `json:"metadata"`
}

// NodeMap stores JGF nodes by stable node ID.
type NodeMap map[string]Node

// Node is a single JGF node entry.
type Node struct {
	Label    string       `json:"label"`
	Metadata NodeMetadata `json:"metadata"`
}

// NodeMetadata carries Mapture-specific node details inside JGF.
type NodeMetadata struct {
	ID            string            `json:"id"`
	Type          string            `json:"type"`
	Domain        string            `json:"domain,omitempty"`
	Owner         string            `json:"owner,omitempty"`
	File          string            `json:"file,omitempty"`
	Line          int               `json:"line,omitempty"`
	Symbol        string            `json:"symbol,omitempty"`
	Summary       string            `json:"summary,omitempty"`
	Tags          []string          `json:"tags,omitempty"`
	EffectiveTags []string          `json:"effectiveTags,omitempty"`
	Facets        map[string]string `json:"facets,omitempty"`
}

// Edge is a single directed JGF edge entry.
type Edge struct {
	Source   string       `json:"source"`
	Target   string       `json:"target"`
	Relation string       `json:"relation"`
	Directed bool         `json:"directed"`
	Metadata EdgeMetadata `json:"metadata"`
}

// EdgeMetadata carries Mapture-specific edge details inside JGF.
type EdgeMetadata struct {
	ID string `json:"id"`
}

// Metadata is the JGF metadata envelope.
type Metadata struct {
	Mapture MaptureMetadata `json:"mapture"`
}

// MaptureMetadata carries all Mapture-specific export extensions under JGF metadata.
type MaptureMetadata struct {
	SchemaVersion int        `json:"schemaVersion"`
	GeneratedAt   string     `json:"generatedAt"`
	ToolVersion   string     `json:"toolVersion"`
	Source        Source     `json:"source"`
	Catalog       Catalog    `json:"catalog"`
	Validation    Validation `json:"validation"`
	UI            config.UI  `json:"ui"`
	Meta          Meta       `json:"meta"`
}

// Source describes where and how the export was produced.
type Source struct {
	ProjectRoot string   `json:"projectRoot"`
	ConfigPath  string   `json:"configPath"`
	Scopes      []string `json:"scopes,omitempty"`
}

// Catalog contains the team/domain metadata needed by downstream tools.
type Catalog struct {
	Tags    []string         `json:"tags,omitempty"`
	Facets  config.Facets    `json:"facets,omitempty"`
	Teams   []catalog.Team   `json:"teams"`
	Domains []catalog.Domain `json:"domains"`
}

// Validation carries diagnostics plus a summary snapshot.
type Validation struct {
	Summary     ValidationSummary      `json:"summary"`
	Diagnostics []validator.Diagnostic `json:"diagnostics,omitempty"`
}

// ValidationSummary is a small aggregate for downstream tool headers and gating.
type ValidationSummary struct {
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
	Nodes    int `json:"nodes"`
	Edges    int `json:"edges"`
}

// Meta carries generic consumption metadata unrelated to the normalized graph.
type Meta struct {
	SourceLabel string `json:"sourceLabel"`
	Mode        string `json:"mode"`
}

// BuildOptions configures a JGF export build from already-loaded project state.
type BuildOptions struct {
	ConfigPath  string
	ProjectRoot string
	Scopes      []string
	Config      *config.Config
	Catalog     *catalog.Catalog
	Result      *validator.Result
	ToolVersion string
	GeneratedAt time.Time
	Mode        string
	SourceLabel string
}

// ProjectOptions configures a full JGF export build from a config path.
type ProjectOptions struct {
	Scopes      []string
	ToolVersion string
	GeneratedAt time.Time
	Mode        string
	SourceLabel string
}

// Build constructs the JGF export from loaded config, catalog, and validator result.
func Build(opts BuildOptions) (*Document, error) {
	if opts.Config == nil {
		return nil, errors.New("json graph export requires config")
	}
	if opts.Catalog == nil {
		return nil, errors.New("json graph export requires catalog")
	}
	if opts.Result == nil {
		return nil, errors.New("json graph export requires validator result")
	}
	if opts.ConfigPath == "" {
		return nil, errors.New("json graph export requires config path")
	}
	if opts.ProjectRoot == "" {
		opts.ProjectRoot = filepath.Dir(opts.ConfigPath)
	}
	if opts.ToolVersion == "" {
		opts.ToolVersion = graph.DefaultScannerVersion
	}
	if opts.Mode == "" {
		opts.Mode = ModeStatic
	}

	generatedAt := opts.GeneratedAt.UTC()
	if generatedAt.IsZero() {
		generatedAt = time.Now().UTC()
	}

	sourceLabel := opts.SourceLabel
	if sourceLabel == "" {
		sourceLabel = projectscope.SourceLabel(opts.Mode, opts.Scopes)
	}

	graphNodes := append([]graph.Node(nil), opts.Result.Graph.Nodes...)
	sort.Slice(graphNodes, func(i, j int) bool { return graphNodes[i].ID < graphNodes[j].ID })
	graphEdges := append([]graph.Edge(nil), opts.Result.Graph.Edges...)
	sort.Slice(graphEdges, func(i, j int) bool {
		if graphEdges[i].From == graphEdges[j].From {
			if graphEdges[i].To == graphEdges[j].To {
				return graphEdges[i].Type < graphEdges[j].Type
			}
			return graphEdges[i].To < graphEdges[j].To
		}
		return graphEdges[i].From < graphEdges[j].From
	})

	nodes := make(NodeMap, len(graphNodes))
	for _, node := range graphNodes {
		nodes[node.ID] = Node{
			Label: node.Name,
			Metadata: NodeMetadata{
				ID:            node.ID,
				Type:          node.Type,
				Domain:        node.Domain,
				Owner:         node.Owner,
				File:          node.File,
				Line:          node.Line,
				Symbol:        node.Symbol,
				Summary:       node.Summary,
				Tags:          append([]string(nil), node.Tags...),
				EffectiveTags: append([]string(nil), node.EffectiveTags...),
				Facets:        cloneFacetAssignments(node.Facets),
			},
		}
	}

	edgeCounts := map[string]int{}
	edges := make([]Edge, 0, len(graphEdges))
	for _, edge := range graphEdges {
		baseID := fmt.Sprintf("%s:%s:%s", edge.Type, edge.From, edge.To)
		edgeCounts[baseID]++
		id := baseID
		if edgeCounts[baseID] > 1 {
			id = fmt.Sprintf("%s#%d", baseID, edgeCounts[baseID])
		}
		edges = append(edges, Edge{
			Source:   edge.From,
			Target:   edge.To,
			Relation: edge.Type,
			Directed: true,
			Metadata: EdgeMetadata{ID: id},
		})
	}

	teams := append([]catalog.Team(nil), opts.Catalog.Teams...)
	sort.Slice(teams, func(i, j int) bool { return teams[i].ID < teams[j].ID })
	domains := append([]catalog.Domain(nil), opts.Catalog.Domains...)
	sort.Slice(domains, func(i, j int) bool { return domains[i].ID < domains[j].ID })

	diagnostics := append([]validator.Diagnostic(nil), opts.Result.Diagnostics...)
	sort.SliceStable(diagnostics, func(i, j int) bool {
		if diagnostics[i].Severity == diagnostics[j].Severity {
			if diagnostics[i].Layer == diagnostics[j].Layer {
				if diagnostics[i].File == diagnostics[j].File {
					if diagnostics[i].Line == diagnostics[j].Line {
						return diagnostics[i].Code < diagnostics[j].Code
					}
					return diagnostics[i].Line < diagnostics[j].Line
				}
				return diagnostics[i].File < diagnostics[j].File
			}
			return diagnostics[i].Layer < diagnostics[j].Layer
		}
		return diagnostics[i].Severity < diagnostics[j].Severity
	})

	return &Document{
		Graph: Graph{
			ID:       filepath.Base(opts.ProjectRoot),
			Type:     "mapture.graph",
			Label:    filepath.Base(opts.ProjectRoot),
			Directed: true,
			Nodes:    nodes,
			Edges:    edges,
			Metadata: Metadata{
				Mapture: MaptureMetadata{
					SchemaVersion: SchemaVersion,
					GeneratedAt:   generatedAt.Format(time.RFC3339),
					ToolVersion:   opts.ToolVersion,
					Source: Source{
						ProjectRoot: opts.ProjectRoot,
						ConfigPath:  opts.ConfigPath,
						Scopes:      append([]string(nil), opts.Scopes...),
					},
					Catalog: Catalog{
						Tags:    append([]string(nil), opts.Config.Tags...),
						Facets:  cloneFacetDefinitions(opts.Config.Facets),
						Teams:   teams,
						Domains: domains,
					},
					Validation: Validation{
						Summary: ValidationSummary{
							Errors:   countDiagnostics(diagnostics, "error"),
							Warnings: countDiagnostics(diagnostics, "warning"),
							Nodes:    len(graphNodes),
							Edges:    len(graphEdges),
						},
						Diagnostics: diagnostics,
					},
					UI: opts.Config.UI,
					Meta: Meta{
						SourceLabel: sourceLabel,
						Mode:        opts.Mode,
					},
				},
			},
		},
	}, nil
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

// BuildProject runs the config/catalog/scan/validate pipeline and returns a JGF export.
func BuildProject(configPath string, opts ProjectOptions) (*Document, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}
	cat, err := catalog.Load(configPath, cfg)
	if err != nil {
		return nil, err
	}

	root := filepath.Dir(configPath)
	scoped, err := projectscope.Apply(root, cfg, opts.Scopes)
	if err != nil {
		return nil, err
	}

	blocks, err := scanner.Scan(root, scoped.Config)
	if err != nil {
		return nil, err
	}

	result, buildErr := validator.Build(cfg, cat, blocks, validator.BuildOptions{
		SourceRoot:     root,
		GeneratedAt:    opts.GeneratedAt,
		ScannerVersion: opts.ToolVersion,
		Scoped:         scoped.Scoped,
	})
	if result == nil {
		return nil, buildErr
	}

	doc, err := Build(BuildOptions{
		ConfigPath:  configPath,
		ProjectRoot: root,
		Scopes:      scoped.Scopes,
		Config:      cfg,
		Catalog:     cat,
		Result:      result,
		ToolVersion: opts.ToolVersion,
		GeneratedAt: opts.GeneratedAt,
		Mode:        opts.Mode,
		SourceLabel: sourceLabel(opts.SourceLabel, opts.Mode, scoped.Scopes),
	})
	if err != nil {
		return nil, err
	}
	return doc, buildErr
}

func sourceLabel(explicit string, mode string, scopes []string) string {
	if explicit != "" {
		return explicit
	}
	base := mode
	if base == "" {
		base = ModeStatic
	}
	return projectscope.SourceLabel(base, scopes)
}

func countDiagnostics(diagnostics []validator.Diagnostic, severity string) int {
	count := 0
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == severity {
			count++
		}
	}
	return count
}
