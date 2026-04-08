// Package validator turns scanned raw blocks plus catalog metadata into
// a normalized graph and enforces validation layers 4-6.
package validator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mandotpro/mapture.dev/src/internal/catalog"
	"github.com/mandotpro/mapture.dev/src/internal/config"
	"github.com/mandotpro/mapture.dev/src/internal/graph"
	"github.com/mandotpro/mapture.dev/src/internal/scanner"
)

const (
	severityError   = "error"
	severityWarning = "warning"
)

// Diagnostic is a machine-readable validation issue.
type Diagnostic struct {
	Severity string `json:"severity"`
	Layer    int    `json:"layer"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
}

// Result is the output of the validator pipeline.
type Result struct {
	Graph       graph.Graph  `json:"graph"`
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
}

// ValidationError reports one or more validation errors.
type ValidationError struct {
	Result *Result
}

func (e *ValidationError) Error() string {
	if e == nil || e.Result == nil {
		return "validation failed"
	}

	lines := make([]string, 0, len(e.Result.Diagnostics))
	for _, diagnostic := range e.Result.Diagnostics {
		if diagnostic.Severity != severityError {
			continue
		}

		var b strings.Builder
		_, _ = fmt.Fprintf(&b, "layer %d %s", diagnostic.Layer, diagnostic.Code)
		if diagnostic.File != "" {
			_, _ = fmt.Fprintf(&b, " %s", diagnostic.File)
			if diagnostic.Line > 0 {
				_, _ = fmt.Fprintf(&b, ":%d", diagnostic.Line)
			}
		}
		_, _ = fmt.Fprintf(&b, ": %s", diagnostic.Message)
		lines = append(lines, b.String())
	}

	if len(lines) == 0 {
		return "validation failed"
	}
	return strings.Join(lines, "\n")
}

// Build validates scanned blocks against catalog/config state and returns
// the normalized graph plus any warnings.
func Build(cfg *config.Config, cat *catalog.Catalog, blocks []scanner.RawBlock) (*Result, error) {
	result := &Result{}
	if cfg == nil {
		return result, fmt.Errorf("config is nil")
	}
	if cat == nil {
		return result, fmt.Errorf("catalog is nil")
	}

	builder := graph.NewBuilder()
	archNodes := make(map[string]graph.Node)
	fileNodes := make(map[string][]graph.Node)

	for _, event := range cat.Events {
		node := graph.Node{
			ID:      eventNodeID(event.ID),
			Type:    graph.NodeEvent,
			Name:    event.Name,
			Domain:  event.Domain,
			Owner:   event.OwnerTeam,
			Summary: event.Description,
		}
		if err := builder.AddNode(node); err != nil {
			addError(result, 6, "duplicate_node_id", err.Error(), "", 0)
		}
	}

	for _, block := range blocks {
		if block.Kind != "arch" {
			continue
		}
		node, err := buildArchNode(block)
		if err != nil {
			addError(result, 6, "invalid_node", err.Error(), block.File, block.Line)
			continue
		}

		checkDomain(result, cfg.Validation.FailOnUnknownDomain, block.File, block.Line, node.Domain, cat)
		checkOwner(result, cfg.Validation.FailOnUnknownTeam, block.File, block.Line, node.Owner, cat)

		if err := builder.AddNode(node); err != nil {
			addError(result, 6, "duplicate_node_id", err.Error(), block.File, block.Line)
			continue
		}
		archNodes[node.ID] = node
		fileNodes[node.File] = append(fileNodes[node.File], node)
	}

	for file := range fileNodes {
		sort.Slice(fileNodes[file], func(i, j int) bool {
			return fileNodes[file][i].Line < fileNodes[file][j].Line
		})
	}

	for _, block := range blocks {
		switch block.Kind {
		case "arch":
			sourceNode, err := buildArchNode(block)
			if err != nil {
				continue
			}
			for relation, targets := range block.Relations {
				for _, target := range targets {
					builder.AddEdge(graph.Edge{
						From: sourceNode.ID,
						To:   fmt.Sprintf("%s:%s", target.Type, target.ID),
						Type: relation,
					})
				}
			}
		case "event":
			validateEventBlock(result, cfg, cat, block)

			edgeType, ok := eventEdgeType(block.Fields["role"])
			if !ok {
				continue
			}

			source, found := nearestFileNode(fileNodes[block.File], block.Line)
			if !found {
				addWarning(result, 5, "unattached_event", "event block could not be attached to a nearby architecture node", block.File, block.Line)
				continue
			}

			builder.AddEdge(graph.Edge{
				From: source.ID,
				To:   eventNodeID(block.Fields["id"]),
				Type: edgeType,
			})
		}
	}

	result.Graph = builder.Build()

	if cfg.Validation.WarnOnOrphanedNodes {
		reportOrphanedNodes(result, result.Graph)
	}
	validateEdgeTargets(result, cfg, result.Graph)
	sortDiagnostics(result.Diagnostics)

	if hasErrors(result.Diagnostics) {
		return result, &ValidationError{Result: result}
	}

	return result, nil
}

func buildArchNode(block scanner.RawBlock) (graph.Node, error) {
	ref, err := parseNodeRef(block.Fields["node"])
	if err != nil {
		return graph.Node{}, err
	}

	return graph.Node{
		ID:      fmt.Sprintf("%s:%s", ref.Type, ref.ID),
		Type:    ref.Type,
		Name:    block.Fields["name"],
		Domain:  block.Fields["domain"],
		Owner:   block.Fields["owner"],
		File:    block.File,
		Line:    block.Line,
		Summary: block.Fields["description"],
	}, nil
}

type nodeRef struct {
	Type string
	ID   string
}

func parseNodeRef(value string) (nodeRef, error) {
	parts := strings.Fields(value)
	if len(parts) != 2 {
		return nodeRef{}, fmt.Errorf("expected <type> <id>, got %q", value)
	}
	return nodeRef{Type: parts[0], ID: parts[1]}, nil
}

func eventNodeID(eventID string) string {
	return graph.NodeEvent + ":" + eventID
}

func checkDomain(result *Result, fail bool, file string, line int, domainID string, cat *catalog.Catalog) {
	if _, ok := cat.DomainsByID[domainID]; ok {
		return
	}
	if fail {
		addError(result, 4, "unknown_domain", fmt.Sprintf("unknown domain %q", domainID), file, line)
		return
	}
	addWarning(result, 4, "unknown_domain", fmt.Sprintf("unknown domain %q", domainID), file, line)
}

func checkOwner(result *Result, fail bool, file string, line int, ownerID string, cat *catalog.Catalog) {
	if _, ok := cat.TeamsByID[ownerID]; ok {
		return
	}
	if fail {
		addError(result, 4, "unknown_team", fmt.Sprintf("unknown team %q", ownerID), file, line)
		return
	}
	addWarning(result, 4, "unknown_team", fmt.Sprintf("unknown team %q", ownerID), file, line)
}

func validateEventBlock(result *Result, cfg *config.Config, cat *catalog.Catalog, block scanner.RawBlock) {
	eventID := block.Fields["id"]
	event, ok := cat.EventsByID[eventID]
	if !ok {
		if cfg.Validation.FailOnUnknownEvent {
			addError(result, 4, "unknown_event", fmt.Sprintf("unknown event %q", eventID), block.File, block.Line)
		} else {
			addWarning(result, 4, "unknown_event", fmt.Sprintf("unknown event %q", eventID), block.File, block.Line)
		}
		return
	}

	if block.Fields["role"] == "definition" && block.Fields["domain"] != event.Domain {
		addError(result, 4, "event_domain_mismatch", fmt.Sprintf("event %q belongs to domain %q, not %q", eventID, event.Domain, block.Fields["domain"]), block.File, block.Line)
	}
	if block.Fields["role"] == "definition" {
		if owner := block.Fields["owner"]; owner != "" && owner != event.OwnerTeam {
			addError(result, 4, "event_owner_mismatch", fmt.Sprintf("event %q owner should be %q, not %q", eventID, event.OwnerTeam, owner), block.File, block.Line)
		}
	}
	if cfg.Validation.WarnOnDeprecatedEvents && (event.Deprecated || event.Status == "deprecated") {
		addWarning(result, 4, "deprecated_event", fmt.Sprintf("event %q is deprecated", eventID), block.File, block.Line)
	}
}

func eventEdgeType(role string) (string, bool) {
	switch role {
	case "trigger", "bridge-out", "publisher":
		return graph.EdgeEmits, true
	case "listener", "bridge-in", "subscriber":
		return graph.EdgeConsumes, true
	default:
		return "", false
	}
}

func nearestFileNode(nodes []graph.Node, line int) (graph.Node, bool) {
	var match graph.Node
	found := false
	for _, node := range nodes {
		if node.Line > line {
			break
		}
		match = node
		found = true
	}
	return match, found
}

func reportOrphanedNodes(result *Result, g graph.Graph) {
	linked := make(map[string]struct{}, len(g.Nodes))
	for _, edge := range g.Edges {
		linked[edge.From] = struct{}{}
		linked[edge.To] = struct{}{}
	}
	for _, node := range g.Nodes {
		if _, ok := linked[node.ID]; ok {
			continue
		}
		addWarning(result, 6, "orphaned_node", fmt.Sprintf("node %q has no edges", node.ID), node.File, node.Line)
	}
}

func validateEdgeTargets(result *Result, cfg *config.Config, g graph.Graph) {
	nodeIDs := make(map[string]struct{}, len(g.Nodes))
	for _, node := range g.Nodes {
		nodeIDs[node.ID] = struct{}{}
	}

	for _, edge := range g.Edges {
		if _, ok := nodeIDs[edge.To]; ok {
			continue
		}
		if cfg.Validation.FailOnUnknownNode {
			addError(result, 6, "unknown_node_target", fmt.Sprintf("edge target %q does not exist", edge.To), "", 0)
			continue
		}
		addWarning(result, 6, "unknown_node_target", fmt.Sprintf("edge target %q does not exist", edge.To), "", 0)
	}
}

func addError(result *Result, layer int, code string, message string, file string, line int) {
	result.Diagnostics = append(result.Diagnostics, Diagnostic{
		Severity: severityError,
		Layer:    layer,
		Code:     code,
		Message:  message,
		File:     file,
		Line:     line,
	})
}

func addWarning(result *Result, layer int, code string, message string, file string, line int) {
	result.Diagnostics = append(result.Diagnostics, Diagnostic{
		Severity: severityWarning,
		Layer:    layer,
		Code:     code,
		Message:  message,
		File:     file,
		Line:     line,
	})
}

func hasErrors(diagnostics []Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == severityError {
			return true
		}
	}
	return false
}

func sortDiagnostics(diagnostics []Diagnostic) {
	sort.Slice(diagnostics, func(i, j int) bool {
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
}
