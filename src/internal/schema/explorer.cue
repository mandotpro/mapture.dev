package schema

#ExplorerValidationSummary: close({
	errors:   int & >=0
	warnings: int & >=0
	nodes:    int & >=0
	edges:    int & >=0
})

#ExplorerDiagnostic: close({
	severity: string & != ""
	layer:    int & >=0
	code:     string & != ""
	message:  string & != ""
	file?:    string
	line?:    int & >=0
})

#ExplorerValidation: close({
	diagnostics?: [...#ExplorerDiagnostic]
	summary:      #ExplorerValidationSummary
})

#ExplorerMeta: close({
	projectId:   string & != ""
	sourceLabel: string & != ""
	mode:        "live" | "offline"
})

#ExplorerPayload: close({
	schemaVersion: 1
	graph:         #Graph
	catalog: close({
		teams:   [...#Team]
		domains: [...#Domain]
	})
	validation: #ExplorerValidation
	ui?: close({
		defaultLayout?: "freeform" | "clustered" | "elk-horizontal"
		nodeColors?: close({
			service?:  #HexColor
			api?:      #HexColor
			database?: #HexColor
			event?:    #HexColor
		})
	})
	meta: #ExplorerMeta
})
