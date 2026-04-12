// Package canonical builds the shared JSON export envelope for downstream consumers.
package canonical

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

// SchemaVersion is the stable public canonical export schema version.
const SchemaVersion = 1

// Mode values supported by the canonical export metadata.
const (
	ModeLive    = "live"
	ModeOffline = "offline"
	ModeStatic  = "static"
)

// Document is the canonical exported JSON contract for Mapture.
type Document struct {
	SchemaVersion int         `json:"schemaVersion"`
	GeneratedAt   string      `json:"generatedAt"`
	ToolVersion   string      `json:"toolVersion"`
	Source        Source      `json:"source"`
	Catalog       Catalog     `json:"catalog"`
	Validation    Validation  `json:"validation"`
	Graph         graph.Graph `json:"graph"`
	UI            config.UI   `json:"ui"`
	Meta          Meta        `json:"meta"`
}

// Source describes where and how the export was produced.
type Source struct {
	ProjectRoot string   `json:"projectRoot"`
	ConfigPath  string   `json:"configPath"`
	Scopes      []string `json:"scopes,omitempty"`
}

// Catalog contains the team/domain metadata needed by downstream tools.
type Catalog struct {
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

// BuildOptions configures a canonical export build from already-loaded project state.
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

// ProjectOptions configures a full canonical export build from a config path.
type ProjectOptions struct {
	Scopes      []string
	ToolVersion string
	GeneratedAt time.Time
	Mode        string
	SourceLabel string
}

// Build constructs the canonical export envelope from loaded config, catalog, and validator result.
func Build(opts BuildOptions) (*Document, error) {
	if opts.Config == nil {
		return nil, errors.New("canonical export requires config")
	}
	if opts.Catalog == nil {
		return nil, errors.New("canonical export requires catalog")
	}
	if opts.Result == nil {
		return nil, errors.New("canonical export requires validator result")
	}
	if opts.ConfigPath == "" {
		return nil, errors.New("canonical export requires config path")
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

	graphSnapshot := opts.Result.Graph
	graphSnapshot.Metadata = graph.NewMetadata(opts.ProjectRoot, generatedAt, opts.ToolVersion)

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

	sourceLabel := opts.SourceLabel
	if sourceLabel == "" {
		sourceLabel = projectscope.SourceLabel(opts.Mode, opts.Scopes)
	}

	return &Document{
		SchemaVersion: SchemaVersion,
		GeneratedAt:   generatedAt.Format(time.RFC3339),
		ToolVersion:   opts.ToolVersion,
		Source: Source{
			ProjectRoot: opts.ProjectRoot,
			ConfigPath:  opts.ConfigPath,
			Scopes:      append([]string(nil), opts.Scopes...),
		},
		Catalog: Catalog{
			Teams:   teams,
			Domains: domains,
		},
		Validation: Validation{
			Summary: ValidationSummary{
				Errors:   countDiagnostics(diagnostics, "error"),
				Warnings: countDiagnostics(diagnostics, "warning"),
				Nodes:    len(graphSnapshot.Nodes),
				Edges:    len(graphSnapshot.Edges),
			},
			Diagnostics: diagnostics,
		},
		Graph: graphSnapshot,
		UI:    opts.Config.UI,
		Meta: Meta{
			SourceLabel: sourceLabel,
			Mode:        opts.Mode,
		},
	}, nil
}

// BuildProject runs the config/catalog/scan/validate pipeline and returns a canonical export.
// If validation fails, the returned document will still contain the partial result and diagnostics.
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

// Result converts the canonical export back into the validator result shape.
func (d *Document) Result() validator.Result {
	if d == nil {
		return validator.Result{}
	}
	return validator.Result{
		Graph:       d.Graph,
		Diagnostics: append([]validator.Diagnostic(nil), d.Validation.Diagnostics...),
	}
}

// ExplorerCatalog converts the canonical export to the explorer catalog shape.
func (d *Document) ExplorerCatalog() Catalog {
	if d == nil {
		return Catalog{}
	}
	return Catalog{
		Teams:   append([]catalog.Team(nil), d.Catalog.Teams...),
		Domains: append([]catalog.Domain(nil), d.Catalog.Domains...),
	}
}

// ProjectID returns the stable project identifier used by the explorer.
func (d *Document) ProjectID() string {
	if d == nil {
		return ""
	}
	return d.Source.ProjectRoot
}

func (d *Document) String() string {
	if d == nil {
		return ""
	}
	return fmt.Sprintf("canonical export schema=%d mode=%s source=%s", d.SchemaVersion, d.Meta.Mode, d.Source.ProjectRoot)
}
