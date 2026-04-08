export interface BackendGraphNode {
  id: string;
  type: string;
  name: string;
  domain?: string;
  owner?: string;
  file?: string;
  line?: number;
  symbol?: string;
  summary?: string;
}

export interface BackendGraphEdge {
  from: string;
  to: string;
  type: string;
}

export interface BackendGraph {
  nodes?: BackendGraphNode[];
  edges?: BackendGraphEdge[];
}

export interface Diagnostic {
  severity: string;
  layer: number;
  code: string;
  message: string;
  file?: string;
  line?: number;
}

export interface CatalogTeam {
  id: string;
  name: string;
}

export interface CatalogDomain {
  id: string;
  name: string;
  owner_team: string;
}

export interface CatalogEvent {
  id: string;
  name: string;
  domain: string;
  owner_team: string;
  status?: string;
  description?: string;
}

export interface CatalogPayload {
  teams?: CatalogTeam[];
  domains?: CatalogDomain[];
  events?: CatalogEvent[];
}

export interface ValidationPayload {
  graph?: BackendGraph;
  diagnostics?: Diagnostic[];
}

export interface GraphNode {
  id: string;
  type: string;
  name: string;
  domain: string;
  owner: string;
  file: string;
  line: number;
  symbol: string;
  summary: string;
}

export interface GraphEdge {
  id: string;
  from: string;
  to: string;
  type: string;
}

export interface GraphModel {
  nodes: GraphNode[];
  edges: GraphEdge[];
  diagnostics: Diagnostic[];
  domains: string[];
  owners: string[];
  nodeTypes: string[];
  edgeTypes: string[];
  teams: Map<string, string>;
  domainNames: Map<string, string>;
  events: Map<string, CatalogEvent>;
}

export interface Filters {
  query: string;
  nodeTypes: string[];
  domains: string[];
  owners: string[];
}

export interface WindowWithPayload extends Window {
  __MAPTURE_DATA__?: ValidationPayload;
}
