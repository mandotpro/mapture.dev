package schema

#KebabID: =~"^[a-z0-9-]+$"
#EventID: =~"^[a-z0-9]+(?:\\.[a-z0-9]+)+$"
#Email:   =~"^[^@\\s]+@[^@\\s]+\\.[^@\\s]+$"

#Team: {
	id:      #KebabID
	name:    string & != ""
	contact?: string
	slack?:   string
	email?:   #Email
	tags?:    [...#KebabID]
}

#Domain: {
	id:                      #KebabID
	name:                    string & != ""
	description?:            string
	ownerTeams:              [#KebabID, ...#KebabID]
	allowedOutboundDomains?: [...#KebabID]
	allowedInboundDomains?:  [...#KebabID]
	tags?:                   [...#KebabID]
}

#EventKind:       "domain" | "integration" | "system" | "internal"
#EventVisibility: "internal" | "public" | "private" | "deprecated"
#EventStatus:     "active" | "deprecated" | "experimental"

#Event: {
	id:                   #EventID
	name:                 string & != ""
	description?:         string
	domain:               #KebabID
	ownerTeam:            #KebabID
	kind:                 #EventKind
	visibility:           #EventVisibility
	status:               #EventStatus
	version?:             *1 | (int & >0)
	payloadSchema?:       string
	allowedTargetDomains?: [...#KebabID]
	allowedProducers?:    [...string]
	allowedConsumers?:    [...string]
	deprecated?:          *false | bool
	replacedBy?:          #EventID
	tags?:                [...#KebabID]
}

#TeamsFile: {
	teams: [...#Team]
}

#DomainsFile: {
	domains: [...#Domain]
}

#EventsFile: {
	events: [...#Event]
}
