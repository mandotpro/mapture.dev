import { MarkerType, Position, type Edge, type Node } from '@xyflow/svelte';
import type {
  BackendGraph,
  CatalogEvent,
  DensityMode,
  Diagnostic,
  ExplorerPayload,
  Filters,
  FlowPresentation,
  GraphEdge,
  GraphModel,
  GraphNode,
  LaneOverlay,
  LayoutMode,
  NodeTone,
  NodeStage,
  PresentedEdge,
  PresentedGraph,
  PresentedNode,
  PresenterFocus,
  UIConfig,
  ViewMode,
} from './types';
import { NODE_HEIGHT, NODE_WIDTH, layoutGraph } from './layout';

type BuildPresentationOptions = {
  viewMode: ViewMode;
  densityMode: DensityMode;
  focus: PresenterFocus;
  manualPositions: Record<string, { x: number; y: number }>;
  reservedInsets: { top: number; left: number };
};

type ModeEdge = {
  id: string;
  from: string;
  to: string;
  type: string;
  label: string;
  synthetic: boolean;
  secondary: boolean;
};

const DEFAULT_NODE_COLORS: Required<NonNullable<UIConfig['nodeColors']>> = {
  service: '#1664d9',
  api: '#0f8f78',
  database: '#a56614',
  event: '#a73f7f',
};

const EDGE_COLORS: Record<string, string> = {
  calls: '#1664d9',
  depends_on: '#53657a',
  stores_in: '#a56614',
  reads_from: '#0f8f78',
  emits: '#a73f7f',
  consumes: '#cf6e26',
  async: '#9f52cc',
};

export function normalizeGraph(payload: ExplorerPayload): GraphModel {
  const rawGraph = normalizeBackendGraph(payload.graph ?? {});
  const diagnostics = payload.validation.diagnostics ?? [];
  const nodes = rawGraph.nodes.map((node) => ({
    id: node.id,
    type: node.type || inferNodeType(node.id),
    name: node.name || node.id,
    domain: node.domain ?? '',
    owner: node.owner ?? '',
    file: node.file ?? '',
    line: node.line ?? 0,
    symbol: node.symbol ?? '',
    summary: node.summary ?? '',
  }));
  const edges = rawGraph.edges.map((edge) => ({
    id: `${edge.from}->${edge.to}|${edge.type}`,
    from: edge.from,
    to: edge.to,
    type: edge.type,
  }));
  const teams = new Map((payload.catalog.teams ?? []).map((team) => [team.id, team.name]));
  const domainNames = new Map((payload.catalog.domains ?? []).map((domain) => [domain.id, domain.name]));
  const events = new Map<string, CatalogEvent>((payload.catalog.events ?? []).map((event) => [event.id, event]));

  return {
    nodes,
    edges,
    diagnostics,
    domains: unique(nodes.map((node) => node.domain).filter(Boolean)),
    owners: unique(nodes.map((node) => node.owner).filter(Boolean)),
    nodeTypes: unique(nodes.map((node) => node.type).filter(Boolean)),
    edgeTypes: unique(edges.map((edge) => edge.type).filter(Boolean)),
    teams,
    domainNames,
    events,
    ui: {
      defaultLayout: resolveDefaultLayout(payload.ui),
      nodeColors: resolveNodeColors(payload.ui),
    },
    projectId: payload.meta.projectId,
    sourceLabel: payload.meta.sourceLabel,
    mode: payload.meta.mode,
    summary: payload.validation.summary ?? {
      errors: diagnostics.filter((diagnostic) => diagnostic.severity === 'error').length,
      warnings: diagnostics.filter((diagnostic) => diagnostic.severity === 'warning').length,
      nodes: nodes.length,
      edges: edges.length,
    },
  };
}

export function resolveDefaultLayout(ui: UIConfig | undefined): LayoutMode {
  const defaultLayout = ui?.defaultLayout;
  if (defaultLayout === 'freeform' || defaultLayout === 'clustered' || defaultLayout === 'elk-horizontal') {
    return defaultLayout;
  }
  return 'elk-horizontal';
}

export function viewModeFromLayout(layout: LayoutMode): ViewMode {
  if (layout === 'freeform') {
    return 'workbench';
  }
  if (layout === 'clustered') {
    return 'domain-lanes';
  }
  return 'system-map';
}

export async function buildFlowPresentation(
  model: GraphModel,
  filters: Filters,
  options: BuildPresentationOptions,
): Promise<FlowPresentation> {
  const filteredNodes = visibleNodesForFilters(model, filters);
  const filteredNodeIDs = new Set(filteredNodes.map((node) => node.id));
  const filteredEdges = model.edges.filter((edge) => filteredNodeIDs.has(edge.from) && filteredNodeIDs.has(edge.to));
  const directNodeIDs = buildDirectNodeMatches(filteredNodes, filters, options.focus.selectedNodeId);
  const modeGraph = deriveModeGraph(
    model,
    filteredNodes,
    filteredEdges,
    directNodeIDs,
    options.viewMode,
    options.densityMode,
    options.focus.selectedNodeId,
  );
  const graph = applyPresentation(model, modeGraph.nodes, modeGraph.edges, options.viewMode, options.densityMode, options.focus);
  const flowNodesInput = graph.nodes.map((node) => toFlowNode(model, node, options.viewMode));
  const layoutEdges = graph.edges.map((edge) => ({
    id: edge.id,
    source: edge.from,
    target: edge.to,
  })) satisfies Edge[];
  const laidOut = await layoutGraph(flowNodesInput, layoutEdges, {
    viewMode: options.viewMode,
    manualPositions: options.viewMode === 'workbench' ? options.manualPositions : {},
    reservedInsets: options.reservedInsets,
  });
  const lanes = options.viewMode === 'domain-lanes'
    ? buildLaneOverlays(model, graph.nodes, laidOut.nodes)
    : [];

  return {
    graph: {
      ...graph,
      lanes,
    },
    flowNodes: laidOut.nodes,
    flowEdges: graph.edges.map((edge) => toFlowEdge(edge)),
  };
}

export function findNode(model: GraphModel, nodeID: string | null): GraphNode | null {
  if (!nodeID) {
    return null;
  }
  return model.nodes.find((node) => node.id === nodeID) ?? null;
}

export function severitySummary(diagnostics: Diagnostic[]): { errors: number; warnings: number } {
  return diagnostics.reduce(
    (summary, diagnostic) => {
      if (diagnostic.severity === 'error') {
        summary.errors += 1;
      } else if (diagnostic.severity === 'warning') {
        summary.warnings += 1;
      }
      return summary;
    },
    { errors: 0, warnings: 0 },
  );
}

export function teamName(model: GraphModel, ownerID: string): string {
  return model.teams.get(ownerID) ?? ownerID;
}

export function domainName(model: GraphModel, domainID: string): string {
  return model.domainNames.get(domainID) ?? domainID;
}

export function nodeColor(model: GraphModel, nodeType: string): string {
  return model.ui.nodeColors[nodeType as keyof typeof model.ui.nodeColors] ?? DEFAULT_NODE_COLORS.service;
}

export function edgeColor(edgeType: string): string {
  return EDGE_COLORS[edgeType] ?? '#53657a';
}

export function edgeLabel(edgeType: string): string {
  const labels: Record<string, string> = {
    calls: 'calls',
    depends_on: 'depends on',
    stores_in: 'stores in',
    reads_from: 'reads from',
    emits: 'emits',
    consumes: 'consumed by',
    async: 'async',
  };
  return labels[edgeType] ?? edgeType;
}

export function visibleNodesForFilters(model: GraphModel, filters: Filters): GraphNode[] {
  return model.nodes.filter((node) => matchesFilters(node, filters));
}

function deriveModeGraph(
  model: GraphModel,
  filteredNodes: GraphNode[],
  filteredEdges: GraphEdge[],
  directNodeIDs: Set<string>,
  viewMode: ViewMode,
  densityMode: DensityMode,
  selectedNodeId: string | null,
): { nodes: GraphNode[]; edges: ModeEdge[] } {
  if (viewMode === 'system-map') {
    return buildSystemMapGraph(filteredNodes, filteredEdges, directNodeIDs, densityMode, selectedNodeId);
  }
  if (viewMode === 'event-flow') {
    return buildEventFlowGraph(filteredNodes, filteredEdges, directNodeIDs, densityMode, selectedNodeId);
  }
  if (viewMode === 'domain-lanes') {
    return buildDomainLanesGraph(filteredNodes, filteredEdges, directNodeIDs, densityMode, selectedNodeId);
  }
  return buildWorkbenchGraph(filteredNodes, filteredEdges, densityMode);
}

function buildSystemMapGraph(
  filteredNodes: GraphNode[],
  filteredEdges: GraphEdge[],
  directNodeIDs: Set<string>,
  densityMode: DensityMode,
  selectedNodeId: string | null,
): { nodes: GraphNode[]; edges: ModeEdge[] } {
  const revealedEventIDs = buildRevealedEventIDs(filteredNodes, filteredEdges, directNodeIDs, selectedNodeId);
  const visibleNodes = filteredNodes.filter((node) => node.type !== 'event' || revealedEventIDs.has(node.id));
  const visibleNodeIDs = new Set(visibleNodes.map((node) => node.id));
  const hiddenEventIDs = new Set(
    filteredNodes
      .filter((node) => node.type === 'event' && !revealedEventIDs.has(node.id))
      .map((node) => node.id),
  );

  const visibleEdges = filteredEdges
    .filter((edge) => visibleNodeIDs.has(edge.from) && visibleNodeIDs.has(edge.to))
    .filter((edge) => densityMode !== 'overview' || (edge.type !== 'depends_on' && edge.type !== 'reads_from'))
    .map(toModeEdge);

  return {
    nodes: visibleNodes,
    edges: [
      ...visibleEdges,
      ...buildSyntheticAsyncEdges(filteredNodes, filteredEdges, visibleNodeIDs, hiddenEventIDs),
    ],
  };
}

function buildEventFlowGraph(
  filteredNodes: GraphNode[],
  filteredEdges: GraphEdge[],
  directNodeIDs: Set<string>,
  densityMode: DensityMode,
  selectedNodeId: string | null,
): { nodes: GraphNode[]; edges: ModeEdge[] } {
  const visibleNodes = filteredNodes.filter((node) => {
    if (node.type === 'database') {
      return directNodeIDs.has(node.id) || node.id === selectedNodeId;
    }
    return node.type === 'service' || node.type === 'api' || node.type === 'event';
  });
  const visibleNodeIDs = new Set(visibleNodes.map((node) => node.id));
  const allowedTypes = densityMode === 'detailed'
    ? new Set(['emits', 'consumes', 'calls'])
    : new Set(['emits', 'consumes']);

  return {
    nodes: visibleNodes,
    edges: filteredEdges
      .filter((edge) => visibleNodeIDs.has(edge.from) && visibleNodeIDs.has(edge.to))
      .filter((edge) => allowedTypes.has(edge.type))
      .map((edge) => toModeEdge(edge, edge.type === 'calls')),
  };
}

function buildDomainLanesGraph(
  filteredNodes: GraphNode[],
  filteredEdges: GraphEdge[],
  directNodeIDs: Set<string>,
  densityMode: DensityMode,
  selectedNodeId: string | null,
): { nodes: GraphNode[]; edges: ModeEdge[] } {
  const revealedEventIDs = densityMode === 'overview'
    ? buildRevealedEventIDs(filteredNodes, filteredEdges, directNodeIDs, selectedNodeId)
    : new Set(filteredNodes.filter((node) => node.type === 'event').map((node) => node.id));
  const visibleNodes = filteredNodes.filter((node) => node.type !== 'event' || revealedEventIDs.has(node.id));
  const visibleNodeIDs = new Set(visibleNodes.map((node) => node.id));

  return {
    nodes: visibleNodes,
    edges: filteredEdges
      .filter((edge) => visibleNodeIDs.has(edge.from) && visibleNodeIDs.has(edge.to))
      .filter((edge) => densityMode !== 'overview' || (edge.type !== 'depends_on' && edge.type !== 'reads_from'))
      .map(toModeEdge),
  };
}

function buildWorkbenchGraph(
  filteredNodes: GraphNode[],
  filteredEdges: GraphEdge[],
  densityMode: DensityMode,
): { nodes: GraphNode[]; edges: ModeEdge[] } {
  return {
    nodes: filteredNodes,
    edges: filteredEdges
      .filter((edge) => densityMode !== 'overview' || (edge.type !== 'depends_on' && edge.type !== 'reads_from'))
      .map(toModeEdge),
  };
}

function buildDirectNodeMatches(
  nodes: GraphNode[],
  filters: Filters,
  selectedNodeId: string | null,
): Set<string> {
  const matches = new Set<string>();

  for (const node of nodes) {
    if (selectedNodeId && node.id === selectedNodeId) {
      matches.add(node.id);
      continue;
    }

    if (filters.nodeTypes.includes(node.type)) {
      matches.add(node.id);
      continue;
    }

    if (matchesQuery(node, filters.query)) {
      matches.add(node.id);
    }
  }

  return matches;
}

function buildRevealedEventIDs(
  filteredNodes: GraphNode[],
  filteredEdges: GraphEdge[],
  directNodeIDs: Set<string>,
  selectedNodeId: string | null,
): Set<string> {
  const revealed = new Set<string>();

  for (const node of filteredNodes) {
    if (node.type === 'event' && directNodeIDs.has(node.id)) {
      revealed.add(node.id);
    }
  }

  if (!selectedNodeId) {
    return revealed;
  }

  for (const edge of filteredEdges) {
    if (edge.from === selectedNodeId) {
      const target = filteredNodes.find((node) => node.id === edge.to);
      if (target?.type === 'event') {
        revealed.add(target.id);
      }
    }

    if (edge.to === selectedNodeId) {
      const source = filteredNodes.find((node) => node.id === edge.from);
      if (source?.type === 'event') {
        revealed.add(source.id);
      }
    }
  }

  return revealed;
}

function buildSyntheticAsyncEdges(
  filteredNodes: GraphNode[],
  filteredEdges: GraphEdge[],
  visibleNodeIDs: Set<string>,
  hiddenEventIDs: Set<string>,
): ModeEdge[] {
  const aggregated = new Map<string, Set<string>>();

  for (const eventID of hiddenEventIDs) {
    const producers = unique(
      filteredEdges
        .filter((edge) => edge.type === 'emits' && edge.to === eventID && visibleNodeIDs.has(edge.from))
        .map((edge) => edge.from),
    );
    const consumers = unique(
      filteredEdges
        .filter((edge) => edge.type === 'consumes' && edge.from === eventID && visibleNodeIDs.has(edge.to))
        .map((edge) => edge.to),
    );

    for (const producer of producers) {
      for (const consumer of consumers) {
        if (producer === consumer) {
          continue;
        }
        const key = `${producer}|${consumer}`;
        const current = aggregated.get(key) ?? new Set<string>();
        current.add(eventID);
        aggregated.set(key, current);
      }
    }
  }

  return Array.from(aggregated.entries())
    .sort(([left], [right]) => left.localeCompare(right))
    .map(([key, eventIDs]) => {
      const [from, to] = key.split('|');
      const count = eventIDs.size;
      return {
        id: `synthetic:${from}->${to}|async`,
        from,
        to,
        type: 'async',
        label: count === 1 ? 'async' : `${count} async`,
        synthetic: true,
        secondary: true,
      };
    });
}

function applyPresentation(
  model: GraphModel,
  nodes: GraphNode[],
  edges: ModeEdge[],
  viewMode: ViewMode,
  densityMode: DensityMode,
  focus: PresenterFocus,
): PresentedGraph {
  const nodeStages = buildNodeStages(nodes, edges, viewMode);
  const nodeMap = new Map(nodes.map((node) => [node.id, node]));
  const focusState = buildFocusState(edges, focus, new Set(nodes.map((node) => node.id)));

  const presentedNodes: PresentedNode[] = nodes.map((node) => ({
    ...node,
    stage: nodeStages.get(node.id) ?? 'support',
    subtitle: node.domain || node.owner || '',
    tone: resolveNodeTone(node, viewMode, densityMode, focusState),
  }));

  const presentedEdges: PresentedEdge[] = edges.map((edge) => {
    const source = nodeMap.get(edge.from);
    const target = nodeMap.get(edge.to);
    const crossDomain = Boolean(
      source?.domain &&
      target?.domain &&
      source.domain !== target.domain,
    );
    return {
      ...edge,
      tone: resolveEdgeTone(edge, viewMode, crossDomain, focusState),
      showLabel: shouldShowEdgeLabel(edge, densityMode, focusState),
      crossDomain,
    };
  });

  return {
    nodes: presentedNodes,
    edges: presentedEdges,
    lanes: [],
  };
}

function buildNodeStages(
  nodes: GraphNode[],
  edges: ModeEdge[],
  viewMode: ViewMode,
): Map<string, NodeStage> {
  const stages = new Map<string, NodeStage>();
  const emitsFrom = new Set(edges.filter((edge) => edge.type === 'emits').map((edge) => edge.from));
  const consumesTo = new Set(edges.filter((edge) => edge.type === 'consumes').map((edge) => edge.to));

  for (const node of nodes) {
    if (node.type === 'event') {
      stages.set(node.id, 'event');
      continue;
    }

    if (viewMode === 'event-flow') {
      if (consumesTo.has(node.id)) {
        stages.set(node.id, 'consumer');
        continue;
      }
      if (emitsFrom.has(node.id)) {
        stages.set(node.id, 'producer');
        continue;
      }
      stages.set(node.id, 'support');
      continue;
    }

    stages.set(node.id, 'support');
  }

  return stages;
}

function buildFocusState(
  edges: ModeEdge[],
  focus: PresenterFocus,
  visibleNodeIDs: Set<string>,
): {
  active: boolean;
  nodeIDs: Set<string>;
  edgeIDs: Set<string>;
  anchorNodeId: string | null;
} {
  const hoveredEdge = focus.hoveredEdgeId
    ? edges.find((edge) => edge.id === focus.hoveredEdgeId) ?? null
    : null;

  if (hoveredEdge) {
    return {
      active: true,
      nodeIDs: new Set([hoveredEdge.from, hoveredEdge.to]),
      edgeIDs: new Set([hoveredEdge.id]),
      anchorNodeId: null,
    };
  }

  const anchorNodeId = focus.hoveredNodeId && visibleNodeIDs.has(focus.hoveredNodeId)
    ? focus.hoveredNodeId
    : focus.selectedNodeId && visibleNodeIDs.has(focus.selectedNodeId)
      ? focus.selectedNodeId
      : null;

  if (!anchorNodeId) {
    return {
      active: false,
      nodeIDs: new Set(),
      edgeIDs: new Set(),
      anchorNodeId: null,
    };
  }

  const nodeIDs = new Set([anchorNodeId]);
  const edgeIDs = new Set<string>();
  for (const edge of edges) {
    if (edge.from !== anchorNodeId && edge.to !== anchorNodeId) {
      continue;
    }
    edgeIDs.add(edge.id);
    nodeIDs.add(edge.from);
    nodeIDs.add(edge.to);
  }

  return {
    active: true,
    nodeIDs,
    edgeIDs,
    anchorNodeId,
  };
}

function resolveNodeTone(
  node: GraphNode,
  viewMode: ViewMode,
  densityMode: DensityMode,
  focusState: { active: boolean; nodeIDs: Set<string> },
): NodeTone {
  const baseTone = baseNodeTone(node, viewMode, densityMode);

  if (!focusState.active) {
    return baseTone;
  }

  if (!focusState.nodeIDs.has(node.id)) {
    return 'muted';
  }

  if (baseTone === 'muted') {
    return 'secondary';
  }

  return 'primary';
}

function resolveEdgeTone(
  edge: ModeEdge,
  viewMode: ViewMode,
  crossDomain: boolean,
  focusState: { active: boolean; nodeIDs: Set<string>; edgeIDs: Set<string> },
): 'primary' | 'secondary' | 'muted' {
  const baseTone = baseEdgeTone(edge, viewMode, crossDomain);

  if (!focusState.active) {
    return baseTone;
  }

  if (focusState.edgeIDs.has(edge.id)) {
    return 'primary';
  }

  if (focusState.nodeIDs.has(edge.from) && focusState.nodeIDs.has(edge.to)) {
    return edge.secondary ? 'secondary' : 'primary';
  }

  return 'muted';
}

function shouldShowEdgeLabel(
  edge: ModeEdge,
  densityMode: DensityMode,
  focusState: { active: boolean; edgeIDs: Set<string>; anchorNodeId: string | null },
): boolean {
  if (densityMode === 'detailed') {
    return true;
  }
  if (focusState.edgeIDs.has(edge.id)) {
    return true;
  }
  if (!focusState.anchorNodeId) {
    return false;
  }
  return edge.from === focusState.anchorNodeId || edge.to === focusState.anchorNodeId;
}

function baseNodeTone(node: GraphNode, viewMode: ViewMode, densityMode: DensityMode): NodeTone {
  if (viewMode === 'event-flow') {
    if (node.type === 'event') {
      return 'primary';
    }
    if (node.type === 'database') {
      return 'muted';
    }
    return 'secondary';
  }

  if (viewMode === 'system-map') {
    if (node.type === 'service') {
      return 'primary';
    }
    if (node.type === 'event') {
      return 'muted';
    }
    return 'secondary';
  }

  if (viewMode === 'domain-lanes') {
    if (node.type === 'service') {
      return 'primary';
    }
    if (node.type === 'event') {
      return densityMode === 'overview' ? 'muted' : 'secondary';
    }
    return 'secondary';
  }

  if (node.type === 'service') {
    return 'primary';
  }
  return 'secondary';
}

function baseEdgeTone(edge: ModeEdge, viewMode: ViewMode, crossDomain: boolean): 'primary' | 'secondary' | 'muted' {
  if (edge.synthetic) {
    return 'secondary';
  }

  if (viewMode === 'event-flow') {
    return edge.type === 'calls' ? 'secondary' : 'primary';
  }

  if (viewMode === 'domain-lanes') {
    return crossDomain ? 'primary' : 'secondary';
  }

  if (edge.type === 'depends_on' || edge.type === 'reads_from') {
    return 'secondary';
  }

  return 'primary';
}

function toModeEdge(edge: GraphEdge, secondary = false): ModeEdge {
  return {
    id: edge.id,
    from: edge.from,
    to: edge.to,
    type: edge.type,
    label: edgeLabel(edge.type),
    synthetic: false,
    secondary,
  };
}

function toFlowNode(model: GraphModel, node: PresentedNode, viewMode: ViewMode): Node {
  return {
    id: node.id,
    type: 'architecture',
    position: { x: 0, y: 0 },
    width: NODE_WIDTH,
    height: NODE_HEIGHT,
    data: {
      label: node.name,
      subtitle: node.subtitle,
      type: node.type,
      domain: node.domain,
      owner: node.owner,
      summary: node.summary,
      color: nodeColor(model, node.type),
      tone: node.tone,
      viewMode,
      stage: node.stage,
    },
    sourcePosition: Position.Right,
    targetPosition: Position.Left,
    selectable: true,
    draggable: viewMode === 'workbench',
    connectable: false,
  } satisfies Node;
}

function toFlowEdge(edge: PresentedEdge): Edge {
  const opacity = edge.tone === 'muted' ? 0.11 : edge.tone === 'secondary' ? 0.34 : 0.8;
  const strokeWidth = edge.synthetic ? 1.6 : edge.crossDomain ? 1.9 : edge.tone === 'primary' ? 1.7 : 1.4;
  const dash = edge.synthetic
    ? 'stroke-dasharray:11 6;'
    : edge.type === 'depends_on'
      ? 'stroke-dasharray:9 5;'
      : edge.type === 'reads_from'
        ? 'stroke-dasharray:4 4;'
        : edge.type === 'consumes'
          ? 'stroke-dasharray:7 5;'
          : '';

  return {
    id: edge.id,
    source: edge.from,
    target: edge.to,
    type: 'smoothstep',
    label: edge.showLabel ? edge.label : '',
    markerEnd: {
      type: MarkerType.ArrowClosed,
      color: edgeColor(edge.type),
    },
    style: `stroke:${edgeColor(edge.type)};stroke-width:${strokeWidth};opacity:${opacity};${dash}`,
    labelStyle: 'font-size:11px;font-weight:600;color:#4f5b66;background:rgba(255,252,246,0.96);border:1px solid rgba(23,32,39,0.08);border-radius:999px;padding:3px 8px;box-shadow:0 8px 20px rgba(58,39,14,0.08);',
  };
}

function buildLaneOverlays(
  model: GraphModel,
  nodes: PresentedNode[],
  laidOutNodes: Node[],
): LaneOverlay[] {
  if (laidOutNodes.length === 0) {
    return [];
  }

  const nodeByID = new Map(nodes.map((node) => [node.id, node]));
  const domains = unique(nodes.map((node) => node.domain || 'unassigned'));
  const globalTop = Math.max(
    32,
    Math.min(...laidOutNodes.map((node) => node.position.y)) - 52,
  );
  const globalBottom = Math.max(
    ...laidOutNodes.map((node) => node.position.y + (typeof node.height === 'number' ? node.height : NODE_HEIGHT)),
  ) + 44;

  return domains.map((domain) => {
    const nodesInDomain = laidOutNodes.filter((node) => {
      const source = nodeByID.get(node.id);
      return (source?.domain || 'unassigned') === domain;
    });
    const minX = Math.min(...nodesInDomain.map((node) => node.position.x));
    const maxX = Math.max(
      ...nodesInDomain.map((node) => node.position.x + (typeof node.width === 'number' ? node.width : NODE_WIDTH)),
    );
    const sample = nodeByID.get(nodesInDomain[0]?.id ?? '');

    return {
      id: domain,
      label: domain === 'unassigned' ? 'Unassigned' : domainName(model, domain),
      ownerLabel: sample?.owner ? teamName(model, sample.owner) : '',
      accent: domainAccent(domain),
      x: minX - 28,
      width: maxX - minX + 56,
      top: globalTop,
      height: globalBottom - globalTop,
    };
  });
}

function normalizeBackendGraph(graph: BackendGraph): { nodes: GraphNode[]; edges: GraphEdge[] } {
  return {
    nodes: (graph.nodes ?? []).map((node) => ({
      id: node.id,
      type: node.type,
      name: node.name,
      domain: node.domain ?? '',
      owner: node.owner ?? '',
      file: node.file ?? '',
      line: node.line ?? 0,
      symbol: node.symbol ?? '',
      summary: node.summary ?? '',
    })),
    edges: (graph.edges ?? []).map((edge) => ({
      id: `${edge.from}->${edge.to}|${edge.type}`,
      from: edge.from,
      to: edge.to,
      type: edge.type,
    })),
  };
}

function matchesFilters(node: GraphNode, filters: Filters): boolean {
  if (filters.nodeTypes.length > 0 && !filters.nodeTypes.includes(node.type)) {
    return false;
  }
  if (filters.domains.length > 0 && !filters.domains.includes(node.domain)) {
    return false;
  }
  if (filters.owners.length > 0 && !filters.owners.includes(node.owner)) {
    return false;
  }
  return matchesQuery(node, filters.query);
}

function matchesQuery(node: GraphNode, query: string): boolean {
  if (!query) {
    return true;
  }

  const normalizedQuery = query.toLowerCase();
  return [
    node.id,
    node.name,
    node.domain,
    node.owner,
    node.file,
    node.summary,
  ].join(' ').toLowerCase().includes(normalizedQuery);
}

function unique(values: string[]): string[] {
  return Array.from(new Set(values)).sort((left, right) => left.localeCompare(right));
}

function inferNodeType(nodeID: string): string {
  const [kind] = nodeID.split(':', 1);
  return kind || 'service';
}

function resolveNodeColors(ui: UIConfig | undefined): Required<NonNullable<UIConfig['nodeColors']>> {
  return {
    service: ui?.nodeColors?.service ?? DEFAULT_NODE_COLORS.service,
    api: ui?.nodeColors?.api ?? DEFAULT_NODE_COLORS.api,
    database: ui?.nodeColors?.database ?? DEFAULT_NODE_COLORS.database,
    event: ui?.nodeColors?.event ?? DEFAULT_NODE_COLORS.event,
  };
}

function domainAccent(domainID: string): string {
  const palette = [
    'hsl(212 78% 62%)',
    'hsl(168 66% 41%)',
    'hsl(334 61% 58%)',
    'hsl(33 74% 52%)',
    'hsl(262 54% 58%)',
    'hsl(194 63% 50%)',
  ];
  return palette[hashSeed(domainID) % palette.length];
}

function hashSeed(input: string): number {
  let hash = 2166136261;
  for (let index = 0; index < input.length; index += 1) {
    hash ^= input.charCodeAt(index);
    hash = Math.imul(hash, 16777619);
  }
  return hash >>> 0;
}
