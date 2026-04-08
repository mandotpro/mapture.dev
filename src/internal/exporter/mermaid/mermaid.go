// Package mermaid renders normalized graphs as Mermaid flowcharts.
package mermaid

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mandotpro/mapture.dev/src/internal/graph"
)

// Options controls exporter filtering.
type Options struct {
	Domains   []string
	Teams     []string
	NodeTypes []string
}

// Render renders a graph as deterministic Mermaid flowchart output.
func Render(g *graph.Graph, opts Options) (string, error) {
	if g == nil {
		return "", fmt.Errorf("graph is nil")
	}

	selectedNodes := filterNodes(g.Nodes, opts)
	nodeIDs := make(map[string]graph.Node, len(selectedNodes))
	for _, node := range selectedNodes {
		nodeIDs[node.ID] = node
	}

	selectedEdges := make([]graph.Edge, 0, len(g.Edges))
	seenEdges := make(map[string]struct{}, len(g.Edges))
	for _, edge := range g.Edges {
		if _, ok := nodeIDs[edge.From]; !ok {
			continue
		}
		if _, ok := nodeIDs[edge.To]; !ok {
			continue
		}
		key := edge.From + "|" + edge.Type + "|" + edge.To
		if _, ok := seenEdges[key]; ok {
			continue
		}
		seenEdges[key] = struct{}{}
		selectedEdges = append(selectedEdges, edge)
	}

	sort.Slice(selectedNodes, func(i, j int) bool {
		if selectedNodes[i].Domain == selectedNodes[j].Domain {
			return selectedNodes[i].ID < selectedNodes[j].ID
		}
		return selectedNodes[i].Domain < selectedNodes[j].Domain
	})
	sort.Slice(selectedEdges, func(i, j int) bool {
		if selectedEdges[i].From == selectedEdges[j].From {
			if selectedEdges[i].To == selectedEdges[j].To {
				return selectedEdges[i].Type < selectedEdges[j].Type
			}
			return selectedEdges[i].To < selectedEdges[j].To
		}
		return selectedEdges[i].From < selectedEdges[j].From
	})

	domainOrder := make([]string, 0)
	domainNodes := make(map[string][]graph.Node)
	for _, node := range selectedNodes {
		domain := node.Domain
		if _, ok := domainNodes[domain]; !ok {
			domainOrder = append(domainOrder, domain)
		}
		domainNodes[domain] = append(domainNodes[domain], node)
	}

	aliases := make(map[string]string, len(selectedNodes))
	for _, node := range selectedNodes {
		aliases[node.ID] = mermaidID(node.ID)
	}

	var b strings.Builder
	b.WriteString("flowchart LR\n")
	for _, domain := range domainOrder {
		_, _ = fmt.Fprintf(&b, "  subgraph %s\n", mermaidLabel(domainTitle(domain)))
		for _, node := range domainNodes[domain] {
			_, _ = fmt.Fprintf(&b, "    %s%s\n", aliases[node.ID], nodeShape(node))
		}
		b.WriteString("  end\n")
	}

	if len(selectedEdges) > 0 && len(selectedNodes) > 0 {
		b.WriteString("\n")
	}
	for _, edge := range selectedEdges {
		_, _ = fmt.Fprintf(&b, "  %s -->|%s| %s\n", aliases[edge.From], edgeLabel(edge.Type), aliases[edge.To])
	}

	return b.String(), nil
}

func filterNodes(nodes []graph.Node, opts Options) []graph.Node {
	domainFilter := sliceSet(opts.Domains)
	teamFilter := sliceSet(opts.Teams)
	typeFilter := sliceSet(opts.NodeTypes)

	selected := make([]graph.Node, 0, len(nodes))
	for _, node := range nodes {
		if len(domainFilter) > 0 {
			if _, ok := domainFilter[node.Domain]; !ok {
				continue
			}
		}
		if len(teamFilter) > 0 {
			if _, ok := teamFilter[node.Owner]; !ok {
				continue
			}
		}
		if len(typeFilter) > 0 {
			if _, ok := typeFilter[node.Type]; !ok {
				continue
			}
		}
		selected = append(selected, node)
	}
	return selected
}

func sliceSet(values []string) map[string]struct{} {
	if len(values) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		set[trimmed] = struct{}{}
	}
	return set
}

func domainTitle(domain string) string {
	if domain == "" {
		return "Ungrouped"
	}
	parts := strings.FieldsFunc(domain, func(r rune) bool {
		return r == '-' || r == '_' || r == '.'
	})
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func mermaidID(nodeID string) string {
	var b strings.Builder
	b.WriteString("n_")
	for _, r := range nodeID {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	return b.String()
}

func nodeShape(node graph.Node) string {
	label := mermaidLabel(node.Name)
	switch node.Type {
	case graph.NodeService:
		return "[" + label + "]"
	case graph.NodeAPI:
		return "([" + label + "])"
	case graph.NodeDatabase:
		return "[(" + label + ")]"
	case graph.NodeEvent:
		return "((" + label + "))"
	default:
		return "[" + label + "]"
	}
}

func mermaidLabel(value string) string {
	return strings.ReplaceAll(value, "\"", "\\\"")
}

func edgeLabel(edgeType string) string {
	if edgeType == graph.EdgeConsumes {
		return "consumed by"
	}
	return edgeType
}
