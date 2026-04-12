package schema

#CanonicalSource: close({
	projectRoot: string & != ""
	configPath:  string & != ""
	scopes?:     [...string]
})

#CanonicalValidationSummary: close({
	errors:   int & >=0
	warnings: int & >=0
	nodes:    int & >=0
	edges:    int & >=0
})

#CanonicalDiagnostic: close({
	severity: string & != ""
	layer:    int & >=0
	code:     string & != ""
	message:  string & != ""
	file?:    string
	line?:    int & >=0
})

#CanonicalValidation: close({
	summary:     #CanonicalValidationSummary
	diagnostics?: [...#CanonicalDiagnostic]
})

#CanonicalMeta: close({
	sourceLabel: string & != ""
	mode:        "live" | "offline" | "static"
})

#CanonicalCatalog: close({
	teams:   [...#Team]
	domains: [...#Domain]
})

#CanonicalExport: close({
	schemaVersion: 1
	generatedAt:   #RFC3339UTC
	toolVersion:   string & != ""
	source:        #CanonicalSource
	catalog:       #CanonicalCatalog
	validation:    #CanonicalValidation
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
	meta: #CanonicalMeta
})
