import { MarkerType, Position, type Edge, type Node } from '@xyflow/svelte';
import type {
  BackendGraph,
  CatalogPayload,
  CatalogEvent,
  Diagnostic,
  ExplorerPayload,
  Filters,
  GraphEdge,
  GraphModel,
  GraphNode,
  UIConfig,
} from './types';
import { layoutGraph } from './layout';

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
};

export function normalizeGraph(
  payload: ExplorerPayload,
): GraphModel {
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
      nodeColors: resolveNodeColors(payload.ui),
    },
    projectId: payload.meta.projectId,
    sourceLabel: payload.meta.sourceLabel,
    mode: payload.meta.mode,
    summary: payload.validation.summary,
  };
}

export function toSvelteFlowNodes(
  model: GraphModel,
  filters: Filters,
  selectedNodeId: string | null,
  layoutMode: 'freeform' | 'clustered',
  savedPositions: Record<string, { x: number; y: number }>,
): Node[] {
  const visibleNodes = model.nodes.filter((node) => matchesFilters(node, filters));
  const allowed = new Set(visibleNodes.map((node) => node.id));

  const nodes = visibleNodes.map((node) => ({
    id: node.id,
    type: 'architecture',
    position: { x: 0, y: 0 },
    data: {
      label: node.name,
      subtitle: node.domain || node.owner || '',
      type: node.type,
      domain: node.domain,
      owner: node.owner,
      summary: node.summary,
      color: nodeColor(model, node.type),
    },
    sourcePosition: Position.Right,
    targetPosition: Position.Left,
    selectable: true,
    draggable: true,
    connectable: false,
    class: selectedNodeId === node.id ? 'selected' : '',
  })) satisfies Node[];

  const edges = toSvelteFlowEdges(model, filters, allowed);
  return layoutGraph(nodes, edges, { mode: layoutMode, savedPositions }).nodes;
}

export function toSvelteFlowEdges(
  model: GraphModel,
  filters: Filters,
  allowedNodeIDs?: Set<string>,
): Edge[] {
  const visibleNodeIDs = allowedNodeIDs ?? new Set(
    model.nodes.filter((node) => matchesFilters(node, filters)).map((node) => node.id),
  );

  return model.edges
    .filter((edge) => visibleNodeIDs.has(edge.from) && visibleNodeIDs.has(edge.to))
    .map((edge) => ({
      id: edge.id,
      source: edge.from,
      target: edge.to,
      type: 'smoothstep',
      markerEnd: {
        type: MarkerType.ArrowClosed,
        color: edgeColor(edge.type),
      },
      style: {
        stroke: edgeColor(edge.type),
        strokeWidth: 1.5,
        opacity: 0.7,
      },
    }));
}

export function graphStats(model: GraphModel): Record<string, number> {
  return {
    nodes: model.nodes.length,
    edges: model.edges.length,
    domains: model.domains.length,
    owners: model.owners.length,
  };
}

export function visibleStats(model: GraphModel, filters: Filters): Record<string, number> {
  const visibleNodes = model.nodes.filter((node) => matchesFilters(node, filters));
  const visibleIDs = new Set(visibleNodes.map((node) => node.id));
  const visibleEdges = model.edges.filter((edge) => visibleIDs.has(edge.from) && visibleIDs.has(edge.to));
  return {
    nodes: visibleNodes.length,
    edges: visibleEdges.length,
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
  if (filters.query) {
    const query = filters.query.toLowerCase();
    const haystack = [
      node.id,
      node.name,
      node.domain,
      node.owner,
      node.file,
      node.summary,
    ].join(' ').toLowerCase();
    if (!haystack.includes(query)) {
      return false;
    }
  }
  return true;
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

function edgeColor(edgeType: string): string {
  return EDGE_COLORS[edgeType] ?? '#53657a';
}
