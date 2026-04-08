// Package graph defines the normalized architecture graph model that
// the scanner produces and that exporters consume.
//
// NodeType and EdgeType are intentionally open string types so
// contributors can add new kinds without touching core plumbing.
// Validation enforces the allowed set in the validator package.
package graph

// NodeType values supported in v1.
const (
	NodeService  = "service"
	NodeAPI      = "api"
	NodeDatabase = "database"
	NodeEvent    = "event"
)

// EdgeType values supported in v1.
//
// Direction semantics:
//   - calls: source calls target
//   - depends_on: source depends on target
//   - stores_in: source stores state in target
//   - reads_from: source reads from target
//   - emits: source emits event target
//   - consumes: event source is consumed by target
const (
	EdgeCalls     = "calls"
	EdgeDependsOn = "depends_on"
	EdgeStoresIn  = "stores_in"
	EdgeReadsFrom = "reads_from"
	EdgeEmits     = "emits"
	EdgeConsumes  = "consumes"
)

// Node is a single architecture entity. ID is the stable "type:name"
// form (e.g. "service:checkout-service") used as the identity across
// the graph. File/Line/Symbol are best-effort source attachment — they
// may be empty if the comment could not be tied to a concrete location.
type Node struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Domain  string `json:"domain,omitempty"`
	Owner   string `json:"owner,omitempty"`
	File    string `json:"file,omitempty"`
	Line    int    `json:"line,omitempty"`
	Symbol  string `json:"symbol,omitempty"`
	Summary string `json:"summary,omitempty"`
}

// Edge is a typed directed relation between two node IDs.
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}

// Graph is the normalized scan result. It is the shared payload between
// scanner output, validator input, and every exporter.
type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}
