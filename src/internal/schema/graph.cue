package schema

#GraphNodeType: "service" | "api" | "database" | "event"
#GraphEdgeType: "calls" | "depends_on" | "stores_in" | "reads_from" | "emits" | "consumes"
#RFC3339UTC:    =~"^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$"

#GraphMetadata: close({
	generatedAt:    #RFC3339UTC
	scannerVersion: string & != ""
	sourceRoot:     string & != ""
})

#GraphNode: close({
	id:      string & != ""
	type:    #GraphNodeType
	name:    string & != ""
	domain?: string
	owner?:  string
	file?:   string
	line?:   int & >=0
	symbol?: string
	summary?: string
	tags?: [...#KebabID]
	effectiveTags?: [...#KebabID]
	facets?: [#FacetID]: #KebabID
})

#GraphEdge: close({
	from: string & != ""
	to:   string & != ""
	type: #GraphEdgeType
})

#Graph: close({
	schemaVersion: 1
	metadata:      #GraphMetadata
	nodes:         [...#GraphNode]
	edges:         [...#GraphEdge]
})
