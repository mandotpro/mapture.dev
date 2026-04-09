package graph

import (
	"fmt"
	"sort"
)

// Builder incrementally constructs a normalized graph while enforcing
// unique node identities.
type Builder struct {
	nodes map[string]Node
	edges []Edge
}

// NewBuilder creates an empty graph builder.
func NewBuilder() *Builder {
	return &Builder{
		nodes: make(map[string]Node),
	}
}

// AddNode inserts a node keyed by its stable ID.
func (b *Builder) AddNode(node Node) error {
	if node.ID == "" {
		return fmt.Errorf("graph node id is required")
	}
	if _, exists := b.nodes[node.ID]; exists {
		return fmt.Errorf("duplicate graph node id %q", node.ID)
	}
	b.nodes[node.ID] = node
	return nil
}

// AddEdge appends a graph edge.
func (b *Builder) AddEdge(edge Edge) {
	b.edges = append(b.edges, edge)
}

// HasNode reports whether the graph already contains nodeID.
func (b *Builder) HasNode(nodeID string) bool {
	_, ok := b.nodes[nodeID]
	return ok
}

// Build returns a stable graph snapshot sorted for deterministic output.
func (b *Builder) Build(metadata Metadata) Graph {
	nodes := make([]Node, 0, len(b.nodes))
	for _, node := range b.nodes {
		nodes = append(nodes, node)
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].ID < nodes[j].ID
	})

	edges := append([]Edge(nil), b.edges...)
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From == edges[j].From {
			if edges[i].To == edges[j].To {
				return edges[i].Type < edges[j].Type
			}
			return edges[i].To < edges[j].To
		}
		return edges[i].From < edges[j].From
	})

	return Graph{
		SchemaVersion: SchemaVersion,
		Metadata:      metadata,
		Nodes:         nodes,
		Edges:         edges,
	}
}
