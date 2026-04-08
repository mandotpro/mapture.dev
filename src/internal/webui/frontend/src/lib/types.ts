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

export interface UINodeColors {
  service?: string;
  api?: string;
  database?: string;
  event?: string;
}

export interface UIConfig {
  nodeColors?: UINodeColors;
}

export interface ExplorerMeta {
  projectId: string;
  sourceLabel: string;
  mode: 'live' | 'offline';
}

export interface CatalogPayload {
  teams: CatalogTeam[];
  domains: CatalogDomain[];
  events: CatalogEvent[];
}

export interface ValidationPayload {
  diagnostics: Diagnostic[];
  summary: ValidationSummary;
}

export interface ValidationSummary {
  errors: number;
  warnings: number;
  nodes: number;
  edges: number;
}

export interface ExplorerPayload {
  schemaVersion: number;
  graph: BackendGraph;
  catalog: CatalogPayload;
  validation: ValidationPayload;
  ui: UIConfig;
  meta: ExplorerMeta;
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

export type LayoutMode = 'freeform' | 'clustered' | 'elk-horizontal';

export type FilterPreset = 'service-map' | 'event-map' | 'producer-consumer' | 'api-dependencies';

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
  ui: Required<UIConfig>;
  projectId: string;
  sourceLabel: string;
  mode: 'live' | 'offline';
  summary: ValidationSummary;
}

export interface Filters {
  query: string;
  nodeTypes: string[];
  domains: string[];
  owners: string[];
  relationTypes: string[];
}

export interface WindowWithPayload extends Window {
  __MAPTURE_DATA__?: ExplorerPayload;
}
