package schema

#ExportSource: close({
	projectRoot: string & != ""
	configPath:  string & != ""
	scopes?:     [...string]
})

#ExportValidationSummary: close({
	errors:   int & >=0
	warnings: int & >=0
	nodes:    int & >=0
	edges:    int & >=0
})

#ExportDiagnostic: close({
	severity: string & != ""
	layer:    int & >=0
	code:     string & != ""
	message:  string & != ""
	file?:    string
	line?:    int & >=0
})

#ExportValidation: close({
	summary:     #ExportValidationSummary
	diagnostics?: [...#ExportDiagnostic]
})

#ExportMeta: close({
	sourceLabel: string & != ""
	mode:        "live" | "offline" | "static"
})

#ExportCatalog: close({
	teams:   [...#Team]
	domains: [...#Domain]
})

#JGFNodeMetadata: close({
	id:      string & != ""
	type:    string & != ""
	domain?: string
	owner?:  string
	file?:   string
	line?:   int & >=0
	symbol?: string
	summary?: string
})

#JGFNode: close({
	label:    string & != ""
	metadata: #JGFNodeMetadata
})

#JGFEdgeMetadata: close({
	id: string & != ""
})

#JGFEdge: close({
	source:   string & != ""
	target:   string & != ""
	relation: string & != ""
	directed: bool
	metadata: #JGFEdgeMetadata
})

#MaptureMetadata: close({
	schemaVersion: 1
	generatedAt:   #RFC3339UTC
	toolVersion:   string & != ""
	source:        #ExportSource
	catalog:       #ExportCatalog
	validation:    #ExportValidation
	ui?: close({
		defaultLayout?: "freeform" | "clustered" | "elk-horizontal"
		nodeColors?: close({
			service?:  #HexColor
			api?:      #HexColor
			database?: #HexColor
			event?:    #HexColor
		})
	})
	meta: #ExportMeta
})

#JSONGraphExport: close({
	graph: close({
		id:       string & != ""
		type:     string & != ""
		label:    string & != ""
		directed: bool
		nodes: [string]: #JGFNode
		edges: [...#JGFEdge]
		metadata: close({
			mapture: #MaptureMetadata
		})
	})
})

#VisualizationExport: close({
	schemaVersion: 1
	generatedAt:   #RFC3339UTC
	toolVersion:   string & != ""
	source:        #ExportSource
	catalog:       #ExportCatalog
	validation:    #ExportValidation
	graph:         #Graph
	ui?: close({
		defaultLayout?: "freeform" | "clustered" | "elk-horizontal"
		nodeColors?: close({
			service?:  #HexColor
			api?:      #HexColor
			database?: #HexColor
			event?:    #HexColor
		})
	})
	meta: #ExportMeta
})
