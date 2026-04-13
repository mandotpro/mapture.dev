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
  tags?: string[];
  effectiveTags?: string[];
}

export interface BackendGraphEdge {
  from: string;
  to: string;
  type: string;
}

export interface BackendGraphMetadata {
  generatedAt: string;
  scannerVersion: string;
  sourceRoot: string;
}

export interface BackendGraph {
  schemaVersion: number;
  metadata: BackendGraphMetadata;
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

export interface UINodeColors {
  service?: string;
  api?: string;
  database?: string;
  event?: string;
}

export interface UIConfig {
  defaultLayout?: LayoutMode;
  nodeColors?: UINodeColors;
}

export interface VisualizationSource {
  projectRoot: string;
  configPath: string;
  scopes?: string[];
}

export interface VisualizationMeta {
  sourceLabel: string;
  mode: 'live' | 'offline' | 'static';
}

export interface CatalogPayload {
  tags?: string[];
  teams: CatalogTeam[];
  domains: CatalogDomain[];
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

export interface VisualizationExportDocument {
  schemaVersion: number;
  generatedAt: string;
  toolVersion: string;
  source: VisualizationSource;
  graph: BackendGraph;
  catalog: CatalogPayload;
  validation: ValidationPayload;
  ui: UIConfig;
  meta: VisualizationMeta;
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
  tags: string[];
  effectiveTags: string[];
}

export interface GraphEdge {
  id: string;
  from: string;
  to: string;
  type: string;
}

export type LayoutMode = 'freeform' | 'clustered' | 'elk-horizontal';
export type ViewMode = 'system-map' | 'event-flow' | 'domain-lanes' | 'workbench';
export type DensityMode = 'overview' | 'standard' | 'detailed';
export type ThemePreference = 'system' | 'light' | 'dark';
export type ResolvedTheme = 'light' | 'dark';
export type NodeTone = 'primary' | 'secondary' | 'muted';
export type EdgeTone = 'primary' | 'secondary' | 'muted';
export type NodeStage = 'support' | 'producer' | 'event' | 'consumer';
export type PresentedNodeKind = 'node' | 'group' | 'bridge';
export type PresentedGroupKind = 'domain' | 'team' | 'boundary' | null;
export type ImpactDirection = 'none' | 'focus' | 'incoming' | 'outgoing' | 'mixed';

export interface GraphModel {
  nodes: GraphNode[];
  edges: GraphEdge[];
  diagnostics: Diagnostic[];
  tags: string[];
  domains: string[];
  owners: string[];
  nodeTypes: string[];
  edgeTypes: string[];
  teams: Map<string, string>;
  domainNames: Map<string, string>;
  ui: {
    defaultLayout: LayoutMode;
    nodeColors: Required<NonNullable<UIConfig['nodeColors']>>;
  };
  projectId: string;
  sourceLabel: string;
  mode: 'live' | 'offline';
  summary: ValidationSummary;
}

export interface Filters {
  query: string;
  tags: string[];
  nodeTypes: string[];
  domains: string[];
  owners: string[];
}

export interface PresenterFocus {
  selectedNodeId: string | null;
  hoveredNodeId: string | null;
  hoveredEdgeId: string | null;
}

export interface TypeSummary {
  service: number;
  api: number;
  database: number;
  event: number;
  total: number;
}

export interface PresentedNode extends GraphNode {
  stage: NodeStage;
  subtitle: string;
  tone: NodeTone;
  kind: PresentedNodeKind;
  groupKind: PresentedGroupKind;
  eyebrow: string;
  memberCount: number;
  typeSummary: TypeSummary;
  colorHint: string;
  impact: ImpactDirection;
}

export interface PresentedEdge {
  id: string;
  from: string;
  to: string;
  type: string;
  label: string;
  tone: EdgeTone;
  showLabel: boolean;
  synthetic: boolean;
  crossDomain: boolean;
  secondary: boolean;
  aggregated: boolean;
  weight: number;
  impact: ImpactDirection;
}

export interface LaneOverlay {
  id: string;
  label: string;
  ownerLabel: string;
  accent: string;
  x: number;
  width: number;
  top: number;
  height: number;
}

export interface StageBandOverlay {
  id: string;
  label: string;
  summary: string;
  accent: string;
  x: number;
  width: number;
  top: number;
  height: number;
}

export interface PresentedGraph {
  nodes: PresentedNode[];
  edges: PresentedEdge[];
  lanes: LaneOverlay[];
  stageBands: StageBandOverlay[];
}

export interface FlowPresentation {
  graph: PresentedGraph;
  flowNodes: import('@xyflow/svelte').Node[];
  flowEdges: import('@xyflow/svelte').Edge[];
}

export interface ArchitectureNodeData {
  label: string;
  subtitle: string;
  type: string;
  domain: string;
  owner: string;
  summary: string;
  color: string;
  tone: NodeTone;
  viewMode: ViewMode;
  stage: NodeStage;
  kind: PresentedNodeKind;
  groupKind: PresentedGroupKind;
  eyebrow: string;
  memberCount: number;
  typeSummary: TypeSummary;
  impact: ImpactDirection;
}

export interface ImpactPreview {
  directUpstream: PresentedNode[];
  directDownstream: PresentedNode[];
  upstreamReach: number;
  downstreamReach: number;
  crossBoundaryTouches: number;
}

export interface ExplorerSettings {
  version: 3;
  appearance: {
    themePreference: ThemePreference;
  };
  inspector: {
    impactPreviewEnabled: boolean;
    impactPreviewDefaultExpanded: boolean;
  };
  experimental: {
    structureTools: boolean;
  };
}

export interface SettingsChoiceOption {
  value: string;
  label: string;
  description?: string;
  glyph?: string;
}

interface SettingsFieldBase {
  id: string;
  label: string;
  description: string;
  badge?: string;
  disabled?: boolean;
}

export interface SettingsToggleField extends SettingsFieldBase {
  kind: 'toggle' | 'checkbox';
  value: boolean;
}

export interface SettingsChoiceField extends SettingsFieldBase {
  kind: 'choice';
  value: string;
  options: SettingsChoiceOption[];
}

export interface SettingsInputField extends SettingsFieldBase {
  kind: 'input';
  value: string;
  placeholder?: string;
  inputType?: string;
}

export type SettingsFieldConfig =
  | SettingsToggleField
  | SettingsChoiceField
  | SettingsInputField;

export interface SettingsSectionConfig {
  id: string;
  title: string;
  description: string;
  fields: SettingsFieldConfig[];
}

export interface NodeInspectorAction {
  id: string;
  label: string;
  tone?: 'default' | 'accent';
  badge?: string;
}

export interface WindowWithPayload extends Window {
  __MAPTURE_DATA__?: VisualizationExportDocument;
}
