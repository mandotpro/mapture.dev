import { MarkerType, Position, type Edge, type Node } from '@xyflow/svelte';
import type {
  BackendGraph,
  CanonicalExportDocument,
  DensityMode,
  Diagnostic,
  Filters,
  FlowPresentation,
  GraphEdge,
  GraphModel,
  GraphNode,
  ImpactDirection,
  LaneOverlay,
  LayoutMode,
  NodeTone,
  NodeStage,
  PresentedEdge,
  PresentedGraph,
  PresentedGroupKind,
  PresentedNode,
  PresenterFocus,
  StageBandOverlay,
  TypeSummary,
  UIConfig,
  ViewMode,
} from './types';
import { NODE_HEIGHT, NODE_WIDTH, layoutGraph } from './layout';

type BuildPresentationOptions = {
  viewMode: ViewMode;
  densityMode: DensityMode;
  focus: PresenterFocus;
  boundaryFocus: boolean;
  collapsedDomains: Set<string>;
  collapsedOwners: Set<string>;
  aggregateCrossDomain: boolean;
  manualPositions: Record<string, { x: number; y: number }>;
  reservedInsets: { top: number; left: number };
};

type WorkingNode = GraphNode & {
  kind: PresentedNode['kind'];
  groupKind: PresentedGroupKind;
  eyebrow: string;
  memberCount: number;
  typeSummary: TypeSummary;
  colorHint: string;
};

type ModeEdge = {
  id: string;
  from: string;
  to: string;
  type: string;
  label: string;
  synthetic: boolean;
  secondary: boolean;
  aggregated: boolean;
  weight: number;
};

type StructureOptions = {
  collapsedDomains: Set<string>;
  collapsedOwners: Set<string>;
  aggregateCrossDomain: boolean;
};

type AggregateBucket = {
  from: string;
  to: string;
  synthetic: boolean;
  secondary: boolean;
  aggregated: boolean;
  weight: number;
  typeCounts: Map<string, number>;
};

type FocusState = {
  active: boolean;
  nodeIDs: Set<string>;
  edgeIDs: Set<string>;
  anchorNodeId: string | null;
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
  aggregate: '#8c6d44',
};

export function normalizeGraph(payload: CanonicalExportDocument): GraphModel {
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
    ui: {
      defaultLayout: resolveDefaultLayout(payload.ui),
      nodeColors: resolveNodeColors(payload.ui),
    },
    projectId: payload.source.projectRoot,
    sourceLabel: payload.meta.sourceLabel,
    mode: payload.meta.mode === 'live' ? 'live' : 'offline',
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
    filteredNodes,
    filteredEdges,
    directNodeIDs,
    options.viewMode,
    options.densityMode,
    options.focus.selectedNodeId,
  );
  const structuredGraph = applyStructureTransform(model, modeGraph.nodes, modeGraph.edges, {
    collapsedDomains: options.collapsedDomains,
    collapsedOwners: options.collapsedOwners,
    aggregateCrossDomain: options.aggregateCrossDomain,
  });
  const graph = applyPresentation(
    model,
    structuredGraph.nodes,
    structuredGraph.edges,
    options.viewMode,
    options.densityMode,
    options.focus,
    options.boundaryFocus,
  );
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
  const stageBands = options.viewMode === 'event-flow'
    ? buildStageBandOverlays(graph.nodes, laidOut.nodes)
    : [];

  return {
    graph: {
      ...graph,
      lanes,
      stageBands,
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
    aggregate: 'links',
  };
  return labels[edgeType] ?? edgeType;
}

export function visibleNodesForFilters(model: GraphModel, filters: Filters): GraphNode[] {
  return model.nodes.filter((node) => matchesFilters(node, filters));
}

function deriveModeGraph(
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
        aggregated: count > 1,
        weight: count,
      };
    });
}

function applyStructureTransform(
  model: GraphModel,
  nodes: GraphNode[],
  edges: ModeEdge[],
  options: StructureOptions,
): { nodes: WorkingNode[]; edges: ModeEdge[] } {
  const workingNodes = nodes.map(toWorkingNode);
  const nodeMap = new Map(workingNodes.map((node) => [node.id, node]));
  const replacements = new Map<string, string>();
  const syntheticNodes = new Map<string, WorkingNode>();
  const groupMembers = new Map<string, WorkingNode[]>();

  for (const node of workingNodes) {
    const domainKey = node.domain || 'unassigned';
    let groupID = '';

    if (node.domain && options.collapsedDomains.has(node.domain)) {
      groupID = `group:domain:${domainKey}`;
    } else if (node.owner && options.collapsedOwners.has(node.owner)) {
      groupID = `group:team:${node.owner}:${domainKey}`;
    }

    if (!groupID) {
      continue;
    }

    replacements.set(node.id, groupID);
    const current = groupMembers.get(groupID) ?? [];
    current.push(node);
    groupMembers.set(groupID, current);
  }

  for (const [groupID, members] of groupMembers.entries()) {
    const first = members[0];
    if (!first) {
      continue;
    }

    if (groupID.startsWith('group:domain:')) {
      syntheticNodes.set(groupID, createDomainGroupNode(model, first.domain || 'unassigned', members));
      continue;
    }

    syntheticNodes.set(groupID, createTeamGroupNode(model, first.owner, first.domain || 'unassigned', members));
  }

  const edgeBuckets = new Map<string, AggregateBucket>();
  const boundaryDomains = new Set<string>();

  for (const edge of edges) {
    const source = nodeMap.get(edge.from);
    const target = nodeMap.get(edge.to);
    if (!source || !target) {
      continue;
    }

    const crossDomain = Boolean(
      source.domain &&
      target.domain &&
      source.domain !== target.domain,
    );

    let from = replacements.get(edge.from) ?? edge.from;
    let to = replacements.get(edge.to) ?? edge.to;

    if (options.aggregateCrossDomain && crossDomain) {
      from = remapCrossDomainEndpoint(from, source);
      to = remapCrossDomainEndpoint(to, target);
      if (from.startsWith('bridge:domain:')) {
        boundaryDomains.add(source.domain || 'unassigned');
      }
      if (to.startsWith('bridge:domain:')) {
        boundaryDomains.add(target.domain || 'unassigned');
      }
    }

    if (from === to) {
      continue;
    }

    const aggregated = from !== edge.from || to !== edge.to || (options.aggregateCrossDomain && crossDomain);
    const key = aggregated ? `${from}|${to}` : `${from}|${to}|${edge.type}`;
    const bucket = edgeBuckets.get(key) ?? {
      from,
      to,
      synthetic: edge.synthetic,
      secondary: edge.secondary,
      aggregated,
      weight: 0,
      typeCounts: new Map<string, number>(),
    };
    bucket.synthetic = bucket.synthetic || edge.synthetic;
    bucket.secondary = bucket.secondary && edge.secondary;
    bucket.aggregated = bucket.aggregated || aggregated;
    bucket.weight += edge.weight;
    bucket.typeCounts.set(edge.type, (bucket.typeCounts.get(edge.type) ?? 0) + edge.weight);
    edgeBuckets.set(key, bucket);
  }

  for (const domain of boundaryDomains) {
    const bridgeID = `bridge:domain:${domain}`;
    if (syntheticNodes.has(bridgeID)) {
      continue;
    }
    const members = workingNodes.filter((node) => (node.domain || 'unassigned') === domain);
    syntheticNodes.set(bridgeID, createBoundaryNode(model, domain, members));
  }

  const visibleNodes = workingNodes.filter((node) => !replacements.has(node.id));
  const structuredNodes = [
    ...visibleNodes,
    ...Array.from(syntheticNodes.values()),
  ].sort(compareWorkingNodes);
  const structuredEdges = Array.from(edgeBuckets.values())
    .map(finalizeAggregateBucket)
    .sort((left, right) => left.id.localeCompare(right.id));

  return {
    nodes: structuredNodes,
    edges: structuredEdges,
  };
}

function applyPresentation(
  model: GraphModel,
  nodes: WorkingNode[],
  edges: ModeEdge[],
  viewMode: ViewMode,
  densityMode: DensityMode,
  focus: PresenterFocus,
  boundaryFocus: boolean,
): PresentedGraph {
  const nodeStages = buildNodeStages(nodes, edges, viewMode);
  const nodeMap = new Map(nodes.map((node) => [node.id, node]));
  const boundaryState = buildBoundaryState(nodes, edges);
  const focusState = buildFocusState(edges, focus, new Set(nodes.map((node) => node.id)));
  const impactState = buildImpactState(edges, focusState.anchorNodeId);

  const presentedNodes: PresentedNode[] = nodes.map((node) => ({
    ...node,
    stage: nodeStages.get(node.id) ?? 'support',
    subtitle: resolveNodeSubtitle(model, node),
    tone: resolveNodeTone(node, viewMode, densityMode, focusState, boundaryState, boundaryFocus),
    kind: node.kind,
    groupKind: node.groupKind,
    eyebrow: node.eyebrow,
    memberCount: node.memberCount,
    typeSummary: node.typeSummary,
    colorHint: node.colorHint,
    impact: impactState.nodeDirections.get(node.id) ?? 'none',
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
      tone: resolveEdgeTone(edge, viewMode, crossDomain, focusState, boundaryState, boundaryFocus),
      showLabel: shouldShowEdgeLabel(edge, densityMode, focusState),
      crossDomain,
      aggregated: edge.aggregated,
      weight: edge.weight,
      impact: impactState.edgeDirections.get(edge.id) ?? 'none',
    };
  });

  return {
    nodes: presentedNodes,
    edges: presentedEdges,
    lanes: [],
    stageBands: [],
  };
}

function buildNodeStages(
  nodes: WorkingNode[],
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

function buildBoundaryState(
  nodes: WorkingNode[],
  edges: ModeEdge[],
): {
  nodeIDs: Set<string>;
  edgeIDs: Set<string>;
} {
  const nodeMap = new Map(nodes.map((node) => [node.id, node]));
  const nodeIDs = new Set<string>();
  const edgeIDs = new Set<string>();

  for (const edge of edges) {
    const source = nodeMap.get(edge.from);
    const target = nodeMap.get(edge.to);
    if (!(source?.domain && target?.domain && source.domain !== target.domain)) {
      continue;
    }
    edgeIDs.add(edge.id);
    nodeIDs.add(edge.from);
    nodeIDs.add(edge.to);
  }

  return {
    nodeIDs,
    edgeIDs,
  };
}

function buildFocusState(
  edges: ModeEdge[],
  focus: PresenterFocus,
  visibleNodeIDs: Set<string>,
): FocusState {
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

function buildImpactState(
  edges: ModeEdge[],
  anchorNodeId: string | null,
): {
  nodeDirections: Map<string, ImpactDirection>;
  edgeDirections: Map<string, ImpactDirection>;
} {
  const nodeDirections = new Map<string, ImpactDirection>();
  const edgeDirections = new Map<string, ImpactDirection>();

  if (!anchorNodeId) {
    return {
      nodeDirections,
      edgeDirections,
    };
  }

  nodeDirections.set(anchorNodeId, 'focus');
  for (const edge of edges) {
    if (edge.from === anchorNodeId && edge.to === anchorNodeId) {
      nodeDirections.set(anchorNodeId, 'mixed');
      edgeDirections.set(edge.id, 'mixed');
      continue;
    }
    if (edge.from === anchorNodeId) {
      edgeDirections.set(edge.id, 'outgoing');
      nodeDirections.set(edge.to, mergeImpactDirection(nodeDirections.get(edge.to), 'outgoing'));
    }
    if (edge.to === anchorNodeId) {
      edgeDirections.set(edge.id, 'incoming');
      nodeDirections.set(edge.from, mergeImpactDirection(nodeDirections.get(edge.from), 'incoming'));
    }
  }

  return {
    nodeDirections,
    edgeDirections,
  };
}

function resolveNodeTone(
  node: WorkingNode,
  viewMode: ViewMode,
  densityMode: DensityMode,
  focusState: FocusState,
  boundaryState: { nodeIDs: Set<string>; edgeIDs: Set<string> },
  boundaryFocus: boolean,
): NodeTone {
  const baseTone = baseNodeTone(node, viewMode, densityMode);

  if (!focusState.active) {
    if (boundaryFocus && !boundaryState.nodeIDs.has(node.id)) {
      return baseTone === 'primary' ? 'secondary' : 'muted';
    }
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
  focusState: FocusState,
  boundaryState: { nodeIDs: Set<string>; edgeIDs: Set<string> },
  boundaryFocus: boolean,
): NodeTone {
  const baseTone = baseEdgeTone(edge, viewMode, crossDomain);

  if (!focusState.active) {
    if (boundaryFocus && !boundaryState.edgeIDs.has(edge.id)) {
      return 'muted';
    }
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
  focusState: FocusState,
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

function baseNodeTone(node: WorkingNode, viewMode: ViewMode, densityMode: DensityMode): NodeTone {
  if (node.kind === 'group' || node.kind === 'bridge') {
    return node.kind === 'bridge' ? 'secondary' : 'primary';
  }

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

function baseEdgeTone(edge: ModeEdge, viewMode: ViewMode, crossDomain: boolean): NodeTone {
  if (edge.synthetic || edge.aggregated) {
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
    aggregated: false,
    weight: 1,
  };
}

function toFlowNode(model: GraphModel, node: PresentedNode, viewMode: ViewMode): Node {
  const componentType = node.kind === 'group'
    ? 'group'
    : node.kind === 'bridge'
      ? 'bridge'
      : node.type;
  return {
    id: node.id,
    type: componentType,
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
      color: node.colorHint || nodeColor(model, node.type),
      tone: node.tone,
      viewMode,
      stage: node.stage,
      kind: node.kind,
      groupKind: node.groupKind,
      eyebrow: node.eyebrow,
      memberCount: node.memberCount,
      typeSummary: node.typeSummary,
      impact: node.impact,
    },
    sourcePosition: Position.Right,
    targetPosition: Position.Left,
    selectable: true,
    draggable: viewMode === 'workbench',
    connectable: false,
  } satisfies Node;
}

function toFlowEdge(edge: PresentedEdge): Edge {
  const opacity = edge.tone === 'muted'
    ? 0.11
    : edge.tone === 'secondary'
      ? 0.34
      : 0.8;
  const strokeWidth = edge.aggregated
    ? 2.2
    : edge.synthetic
      ? 1.6
      : edge.crossDomain
        ? 1.9
        : edge.tone === 'primary'
          ? 1.7
          : 1.4;
  const dash = edge.synthetic
    ? 'stroke-dasharray:11 6;'
    : edge.type === 'depends_on'
      ? 'stroke-dasharray:9 5;'
      : edge.type === 'reads_from'
        ? 'stroke-dasharray:4 4;'
        : edge.type === 'consumes'
          ? 'stroke-dasharray:7 5;'
          : edge.aggregated
            ? 'stroke-dasharray:2 0;'
            : '';

  return {
    id: edge.id,
    source: edge.from,
    target: edge.to,
    type: 'smoothstep',
    label: edge.showLabel ? edge.label : '',
    animated: false,
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

function buildStageBandOverlays(
  nodes: PresentedNode[],
  laidOutNodes: Node[],
): StageBandOverlay[] {
  if (laidOutNodes.length === 0) {
    return [];
  }

  const nodeByID = new Map(nodes.map((node) => [node.id, node]));
  const stageOrder: Array<{ id: PresentedNode['stage']; label: string; summary: string; accent: string }> = [
    { id: 'support', label: 'Support', summary: 'Context and helpers', accent: '#8a744d' },
    { id: 'producer', label: 'Producers', summary: 'Emit messages', accent: '#1f6fe5' },
    { id: 'event', label: 'Events', summary: 'Contracts and signals', accent: '#cf2c7d' },
    { id: 'consumer', label: 'Consumers', summary: 'React downstream', accent: '#12806b' },
  ];
  const globalTop = Math.max(
    24,
    Math.min(...laidOutNodes.map((node) => node.position.y)) - 68,
  );
  const globalBottom = Math.max(
    ...laidOutNodes.map((node) => node.position.y + (typeof node.height === 'number' ? node.height : NODE_HEIGHT)),
  ) + 54;

  return stageOrder.flatMap((stage) => {
    const stageNodes = laidOutNodes.filter((node) => nodeByID.get(node.id)?.stage === stage.id);
    if (stageNodes.length === 0) {
      return [];
    }

    const minX = Math.min(...stageNodes.map((node) => node.position.x));
    const maxX = Math.max(
      ...stageNodes.map((node) => node.position.x + (typeof node.width === 'number' ? node.width : NODE_WIDTH)),
    );

    return [{
      id: stage.id,
      label: stage.label,
      summary: stage.summary,
      accent: stage.accent,
      x: minX - 36,
      width: maxX - minX + 72,
      top: globalTop,
      height: globalBottom - globalTop,
    }];
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

function toWorkingNode(node: GraphNode): WorkingNode {
  return {
    ...node,
    kind: 'node',
    groupKind: null,
    eyebrow: typeLabel(node.type),
    memberCount: 1,
    typeSummary: singleTypeSummary(node.type),
    colorHint: '',
  };
}

function createDomainGroupNode(model: GraphModel, domainID: string, members: WorkingNode[]): WorkingNode {
  const summary = buildTypeSummary(members);
  const label = domainID === 'unassigned' ? 'Unassigned Domain' : domainName(model, domainID);
  const owner = mostCommonValue(members.map((member) => member.owner)) ?? '';

  return {
    id: `group:domain:${domainID}`,
    type: 'service',
    name: label,
    domain: domainID === 'unassigned' ? '' : domainID,
    owner,
    file: '',
    line: 0,
    symbol: '',
    summary: `Collapsed ${summary.total} nodes across ${formatTypeSummary(summary)}.`,
    kind: 'group',
    groupKind: 'domain',
    eyebrow: 'Domain Group',
    memberCount: summary.total,
    typeSummary: summary,
    colorHint: domainAccent(domainID),
  };
}

function createTeamGroupNode(
  model: GraphModel,
  ownerID: string,
  domainID: string,
  members: WorkingNode[],
): WorkingNode {
  const summary = buildTypeSummary(members);
  const label = teamName(model, ownerID);
  const domainLabel = domainID === 'unassigned' ? 'Unassigned' : domainName(model, domainID);

  return {
    id: `group:team:${ownerID}:${domainID}`,
    type: 'service',
    name: label,
    domain: domainID === 'unassigned' ? '' : domainID,
    owner: ownerID,
    file: '',
    line: 0,
    symbol: '',
    summary: `Collapsed ${summary.total} nodes for ${label} in ${domainLabel}.`,
    kind: 'group',
    groupKind: 'team',
    eyebrow: 'Team Group',
    memberCount: summary.total,
    typeSummary: summary,
    colorHint: ownerAccent(ownerID),
  };
}

function createBoundaryNode(model: GraphModel, domainID: string, members: WorkingNode[]): WorkingNode {
  const summary = buildTypeSummary(members);
  const label = domainID === 'unassigned' ? 'Unassigned Boundary' : domainName(model, domainID);
  const owner = mostCommonValue(members.map((member) => member.owner)) ?? '';

  return {
    id: `bridge:domain:${domainID}`,
    type: 'api',
    name: label,
    domain: domainID === 'unassigned' ? '' : domainID,
    owner,
    file: '',
    line: 0,
    symbol: '',
    summary: `Aggregated cross-domain traffic touching ${summary.total} visible nodes in ${label}.`,
    kind: 'bridge',
    groupKind: 'boundary',
    eyebrow: 'Boundary',
    memberCount: summary.total,
    typeSummary: summary,
    colorHint: domainAccent(domainID),
  };
}

function remapCrossDomainEndpoint(currentID: string, node: WorkingNode): string {
  if (currentID.startsWith('group:')) {
    return currentID;
  }
  return `bridge:domain:${node.domain || 'unassigned'}`;
}

function finalizeAggregateBucket(bucket: AggregateBucket): ModeEdge {
  const topTypes = Array.from(bucket.typeCounts.entries()).sort((left, right) => {
    if (right[1] !== left[1]) {
      return right[1] - left[1];
    }
    return left[0].localeCompare(right[0]);
  });
  const primaryType = topTypes[0]?.[0] ?? 'aggregate';
  const mixed = topTypes.length > 1;

  return {
    id: bucket.aggregated
      ? `aggregate:${bucket.from}->${bucket.to}|${mixed ? 'mixed' : primaryType}`
      : `${bucket.from}->${bucket.to}|${primaryType}`,
    from: bucket.from,
    to: bucket.to,
    type: mixed ? 'aggregate' : primaryType,
    label: mixed
      ? `${bucket.weight} links`
      : bucket.weight === 1
        ? edgeLabel(primaryType)
        : `${bucket.weight} ${edgeLabel(primaryType)}`,
    synthetic: bucket.synthetic,
    secondary: bucket.secondary,
    aggregated: bucket.aggregated,
    weight: bucket.weight,
  };
}

function resolveNodeSubtitle(model: GraphModel, node: WorkingNode): string {
  if (node.kind === 'group') {
    if (node.groupKind === 'domain') {
      const ownerLabel = node.owner ? teamName(model, node.owner) : 'no owner';
      return `${node.memberCount} nodes · ${ownerLabel}`;
    }
    const domainLabel = node.domain ? domainName(model, node.domain) : 'Mixed';
    return `${domainLabel} · ${node.memberCount} nodes`;
  }

  if (node.kind === 'bridge') {
    const domainLabel = node.domain ? domainName(model, node.domain) : 'Unassigned';
    return `${domainLabel} boundary · ${node.memberCount} nodes`;
  }

  return node.domain || node.owner || '';
}

function mergeImpactDirection(
  current: ImpactDirection | undefined,
  next: ImpactDirection,
): ImpactDirection {
  if (!current || current === 'none') {
    return next;
  }
  if (current === next) {
    return current;
  }
  if (current === 'focus') {
    return 'focus';
  }
  return 'mixed';
}

function singleTypeSummary(type: string): TypeSummary {
  return {
    service: type === 'service' ? 1 : 0,
    api: type === 'api' ? 1 : 0,
    database: type === 'database' ? 1 : 0,
    event: type === 'event' ? 1 : 0,
    total: 1,
  };
}

function buildTypeSummary(nodes: Array<{ type: string }>): TypeSummary {
  return nodes.reduce<TypeSummary>(
    (summary, node) => {
      if (node.type === 'service') {
        summary.service += 1;
      } else if (node.type === 'api') {
        summary.api += 1;
      } else if (node.type === 'database') {
        summary.database += 1;
      } else if (node.type === 'event') {
        summary.event += 1;
      }
      summary.total += 1;
      return summary;
    },
    {
      service: 0,
      api: 0,
      database: 0,
      event: 0,
      total: 0,
    },
  );
}

function formatTypeSummary(summary: TypeSummary): string {
  const parts: string[] = [];
  if (summary.service > 0) {
    parts.push(`${summary.service} services`);
  }
  if (summary.api > 0) {
    parts.push(`${summary.api} apis`);
  }
  if (summary.database > 0) {
    parts.push(`${summary.database} databases`);
  }
  if (summary.event > 0) {
    parts.push(`${summary.event} events`);
  }
  return parts.join(', ') || '0 nodes';
}

function typeLabel(type: string): string {
  return type.replaceAll('_', ' ');
}

function ownerAccent(ownerID: string): string {
  const palette = [
    'hsl(168 66% 41%)',
    'hsl(212 78% 62%)',
    'hsl(33 74% 52%)',
    'hsl(334 61% 58%)',
    'hsl(194 63% 50%)',
  ];
  return palette[hashSeed(ownerID) % palette.length];
}

function mostCommonValue(values: string[]): string | null {
  const counts = new Map<string, number>();
  for (const value of values) {
    if (!value) {
      continue;
    }
    counts.set(value, (counts.get(value) ?? 0) + 1);
  }

  let best: string | null = null;
  let bestCount = 0;
  for (const [value, count] of counts.entries()) {
    if (count > bestCount) {
      best = value;
      bestCount = count;
    }
  }

  return best;
}

function compareWorkingNodes(left: WorkingNode, right: WorkingNode): number {
  const kindOrder = {
    bridge: 0,
    group: 1,
    node: 2,
  } satisfies Record<WorkingNode['kind'], number>;
  if (kindOrder[left.kind] !== kindOrder[right.kind]) {
    return kindOrder[left.kind] - kindOrder[right.kind];
  }
  return left.id.localeCompare(right.id);
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
