package schema

#EventRole: "definition" | "trigger" | "listener" | "bridge-out" | "bridge-in" | "publisher" | "subscriber"

#KebabID: =~"^[a-z0-9-]+$"
#EventID: =~"^[a-z0-9]+(?:\\.[a-z0-9]+)+$"
#Email:   =~"^[^@\\s]+@[^@\\s]+\\.[^@\\s]+$"
#HexColor: =~"^#[0-9a-fA-F]{6}$"

#Config: {
	version: 1

	catalog: {
		dir: *"./architecture" | string
	}

	scan: {
		include: [string, ...string]
		exclude: *[] | [...string]
	}

	languages: {
		php:        *false | bool
		go:         *false | bool
		typescript: *false | bool
		javascript: *false | bool
	}

	comments: {
		style: *"tags" | "tags"
	}

	validation: {
		failOnUnknownDomain:     *true | bool
		failOnUnknownTeam:       *true | bool
		failOnUnknownEvent:      *true | bool
		failOnUnknownNode:       *true | bool
		requireMetadataOn:       *[] | [...#EventRole]
		warnOnOrphanedNodes:     *false | bool
		warnOnDeprecatedEvents: *true | bool
	}

	ui?: {
		defaultLayout?: *"elk-horizontal" | "freeform" | "clustered"
		nodeColors?: {
			service?:  #HexColor
			api?:      #HexColor
			database?: #HexColor
			event?:    #HexColor
		}
	}
}
