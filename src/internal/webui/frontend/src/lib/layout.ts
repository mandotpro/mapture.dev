import ELK from 'elkjs/lib/elk.bundled.js';
import { forceCenter, forceCollide, forceLink, forceManyBody, forceSimulation, forceX, forceY } from 'd3-force';
import type { Edge, Node } from '@xyflow/svelte';
import type { ViewMode } from './types';

export const NODE_WIDTH = 156;
export const NODE_HEIGHT = 82;

const VIEWPORT_MARGIN = 56;
const WORKBENCH_TICKS = 220;
const EVENT_FLOW_STAGE_GAP = 244;
const EVENT_FLOW_DOMAIN_GAP = 56;
const EVENT_FLOW_NODE_GAP = 26;
const LANE_WIDTH = 368;
const LANE_GAP = 72;
const LANE_COLUMN_GAP = 18;
const LANE_NODE_GAP = 24;

type SimNode = {
  id: string;
  domain: string;
  owner: string;
  x: number;
  y: number;
  vx: number;
  vy: number;
};

type LayoutOptions = {
  viewMode: ViewMode;
  manualPositions: Record<string, { x: number; y: number }>;
  reservedInsets: { top: number; left: number };
};

type ResolvePositionOptions = {
  lockedNodeIds?: Set<string>;
  priorityNodeIds?: Set<string>;
  reservedInsets?: { top: number; left: number };
};

type CollisionOptions = {
  margin: number;
  maxIterations: number;
  lockIDs?: Set<string>;
  priorityIDs?: Set<string>;
  reservedInsets?: { top: number; left: number };
};

type CollisionNode = {
  id: string;
  x: number;
  y: number;
  width: number;
  height: number;
};

const elk = new ELK();

const collisionTuning: Record<ViewMode, { margin: number; maxIterations: number }> = {
  'system-map': {
    margin: 34,
    maxIterations: 180,
  },
  'event-flow': {
    margin: 20,
    maxIterations: 80,
  },
  'domain-lanes': {
    margin: 16,
    maxIterations: 60,
  },
  workbench: {
    margin: 28,
    maxIterations: 160,
  },
};

export async function layoutGraph(
  nodes: Node[],
  edges: Edge[],
  options: LayoutOptions,
): Promise<{ nodes: Node[]; edges: Edge[] }> {
  if (nodes.length === 0) {
    return { nodes, edges };
  }

  if (options.viewMode === 'system-map') {
    return layoutWithSystemMap(nodes, edges, options);
  }

  if (options.viewMode === 'event-flow') {
    return layoutWithEventFlow(nodes, edges, options);
  }

  if (options.viewMode === 'domain-lanes') {
    return layoutWithDomainLanes(nodes, edges, options);
  }

  return layoutWithWorkbench(nodes, edges, options);
}

export function resolvePositions(
  nodes: Node[],
  viewMode: ViewMode,
  options: ResolvePositionOptions = {},
): Record<string, { x: number; y: number }> {
  if (viewMode !== 'workbench') {
    return Object.fromEntries(
      nodes.map((node) => [
        node.id,
        {
          x: node.position.x,
          y: node.position.y,
        },
      ]),
    );
  }

  const tuning = collisionTuning.workbench;
  const resolved = resolveNodeCollisions(
    nodes.map((node) => ({
      id: node.id,
      x: node.position.x,
      y: node.position.y,
      width: typeof node.width === 'number' ? node.width : NODE_WIDTH,
      height: typeof node.height === 'number' ? node.height : NODE_HEIGHT,
    })),
    {
      margin: tuning.margin,
      maxIterations: tuning.maxIterations,
      lockIDs: options.lockedNodeIds,
      priorityIDs: options.priorityNodeIds,
      reservedInsets: options.reservedInsets,
    },
  );

  return Object.fromEntries(
    resolved.map((node) => [
      node.id,
      {
        x: node.x,
        y: node.y,
      },
    ]),
  );
}

async function layoutWithSystemMap(
  nodes: Node[],
  edges: Edge[],
  options: LayoutOptions,
): Promise<{ nodes: Node[]; edges: Edge[] }> {
  const graph = await elk.layout({
    id: 'mapture-root',
    layoutOptions: {
      'elk.algorithm': 'layered',
      'elk.direction': 'RIGHT',
      'elk.layered.spacing.nodeNodeBetweenLayers': '132',
      'elk.spacing.nodeNode': '62',
      'elk.edgeRouting': 'ORTHOGONAL',
      'elk.layered.nodePlacement.strategy': 'NETWORK_SIMPLEX',
      'elk.layered.crossingMinimization.strategy': 'LAYER_SWEEP',
    },
    children: nodes.map((node) => ({
      id: node.id,
      width: NODE_WIDTH,
      height: NODE_HEIGHT,
    })),
    edges: edges.map((edge) => ({
      id: edge.id,
      sources: [edge.source],
      targets: [edge.target],
    })),
  });

  const byID = new Map(
    (graph.children ?? []).map((child) => [
      child.id,
      {
        x: (child.x ?? 0) + VIEWPORT_MARGIN + options.reservedInsets.left,
        y: (child.y ?? 0) + VIEWPORT_MARGIN + options.reservedInsets.top,
      },
    ]),
  );

  const laidOutNodes = nodes.map((node) => ({
    ...node,
    width: NODE_WIDTH,
    height: NODE_HEIGHT,
    position: byID.get(node.id) ?? { x: 180, y: 140 },
  }));

  const tuning = collisionTuning['system-map'];
  const resolved = resolveNodeCollisions(
    laidOutNodes.map((node) => ({
      id: node.id,
      x: node.position.x,
      y: node.position.y,
      width: NODE_WIDTH,
      height: NODE_HEIGHT,
    })),
    {
      margin: tuning.margin,
      maxIterations: tuning.maxIterations,
      reservedInsets: options.reservedInsets,
    },
  );
  const positions = new Map(resolved.map((node) => [node.id, node]));

  return {
    nodes: laidOutNodes.map((node) => ({
      ...node,
      position: {
        x: positions.get(node.id)?.x ?? node.position.x,
        y: positions.get(node.id)?.y ?? node.position.y,
      },
    })),
    edges,
  };
}

function layoutWithEventFlow(
  nodes: Node[],
  edges: Edge[],
  options: LayoutOptions,
): { nodes: Node[]; edges: Edge[] } {
  const stageOrder = ['support', 'producer', 'event', 'consumer'] as const;
  const domains = sortedDomains(nodes);
  const domainStartY = new Map<string, number>();
  let cursorY = options.reservedInsets.top + VIEWPORT_MARGIN;

  for (const domain of domains) {
    const stageCounts = stageOrder.map((stage) => (
      nodes.filter((node) => readString(node, 'domain') === domain && readString(node, 'stage') === stage).length
    ));
    const domainRows = Math.max(1, ...stageCounts);
    domainStartY.set(domain, cursorY);
    cursorY += domainRows * (NODE_HEIGHT + EVENT_FLOW_NODE_GAP) + EVENT_FLOW_DOMAIN_GAP;
  }

  const stageX = new Map<string, number>();
  stageOrder.forEach((stage, index) => {
    stageX.set(stage, options.reservedInsets.left + VIEWPORT_MARGIN + 48 + index * EVENT_FLOW_STAGE_GAP);
  });

  const nodesByStageDomain = new Map<string, Node[]>();
  for (const stage of stageOrder) {
    for (const domain of domains) {
      const key = `${stage}:${domain}`;
      nodesByStageDomain.set(
        key,
        nodes
          .filter((node) => readString(node, 'stage') === stage && readString(node, 'domain') === domain)
          .sort(compareNodes),
      );
    }
  }

  return {
    nodes: nodes.map((node) => {
      const stage = readString(node, 'stage') || 'support';
      const domain = readString(node, 'domain') || 'unassigned';
      const key = `${stage}:${domain}`;
      const siblings = nodesByStageDomain.get(key) ?? [];
      const index = siblings.findIndex((candidate) => candidate.id === node.id);
      return {
        ...node,
        width: NODE_WIDTH,
        height: NODE_HEIGHT,
        position: {
          x: stageX.get(stage) ?? stageX.get('support') ?? 180,
          y: (domainStartY.get(domain) ?? options.reservedInsets.top + VIEWPORT_MARGIN) + index * (NODE_HEIGHT + EVENT_FLOW_NODE_GAP),
        },
      };
    }),
    edges,
  };
}

function layoutWithDomainLanes(
  nodes: Node[],
  edges: Edge[],
  options: LayoutOptions,
): { nodes: Node[]; edges: Edge[] } {
  const domains = sortedDomains(nodes);
  const laneX = new Map<string, number>();
  domains.forEach((domain, index) => {
    laneX.set(
      domain,
      options.reservedInsets.left + VIEWPORT_MARGIN + 24 + index * (LANE_WIDTH + LANE_GAP),
    );
  });

  const grouped = new Map<string, { primary: Node[]; events: Node[]; databases: Node[] }>();
  for (const domain of domains) {
    const inDomain = nodes.filter((node) => readString(node, 'domain') === domain).sort(compareNodes);
    grouped.set(domain, {
      primary: inDomain.filter((node) => {
        const type = readString(node, 'type');
        return type === 'service' || type === 'api';
      }),
      events: inDomain.filter((node) => readString(node, 'type') === 'event'),
      databases: inDomain.filter((node) => readString(node, 'type') === 'database'),
    });
  }

  return {
    nodes: nodes.map((node) => {
      const domain = readString(node, 'domain') || 'unassigned';
      const laneStart = laneX.get(domain) ?? options.reservedInsets.left + VIEWPORT_MARGIN;
      const domainGroups = grouped.get(domain) ?? { primary: [], events: [], databases: [] };
      const type = readString(node, 'type');
      const top = options.reservedInsets.top + VIEWPORT_MARGIN + 46;
      let x = laneStart + 24;
      let y = top;

      if (type === 'event') {
        const index = domainGroups.events.findIndex((candidate) => candidate.id === node.id);
        x = laneStart + LANE_WIDTH - NODE_WIDTH - 24;
        y = top + index * (NODE_HEIGHT + LANE_NODE_GAP);
      } else if (type === 'database') {
        const primaryHeight = domainGroups.primary.length * (NODE_HEIGHT + LANE_NODE_GAP);
        const eventHeight = domainGroups.events.length * (NODE_HEIGHT + LANE_NODE_GAP);
        const databaseTop = top + Math.max(primaryHeight, eventHeight) + 60;
        const index = domainGroups.databases.findIndex((candidate) => candidate.id === node.id);
        x = laneStart + Math.round((LANE_WIDTH - NODE_WIDTH) / 2);
        y = databaseTop + index * (NODE_HEIGHT + LANE_NODE_GAP);
      } else {
        const index = domainGroups.primary.findIndex((candidate) => candidate.id === node.id);
        x = laneStart + 24;
        y = top + index * (NODE_HEIGHT + LANE_NODE_GAP);
      }

      return {
        ...node,
        width: NODE_WIDTH,
        height: NODE_HEIGHT,
        position: { x, y },
      };
    }),
    edges,
  };
}

function layoutWithWorkbench(
  nodes: Node[],
  edges: Edge[],
  options: LayoutOptions,
): { nodes: Node[]; edges: Edge[] } {
  const domainCenters = buildDomainCenters(nodes, options.reservedInsets);
  const random = mulberry32(hashSeed(nodes.map((node) => node.id).join('|')));

  const simNodes: SimNode[] = nodes.map((node) => {
    const domain = readString(node, 'domain');
    const owner = readString(node, 'owner');
    const center = domainCenters.get(domain) ?? {
      x: options.reservedInsets.left + 340,
      y: options.reservedInsets.top + 240,
    };
    const ownerOffset = hashSeed(owner || node.id) % 40;

    return {
      id: node.id,
      domain,
      owner,
      x: center.x + (random() - 0.5) * 140 + ownerOffset - 20,
      y: center.y + (random() - 0.5) * 140 - ownerOffset + 20,
      vx: 0,
      vy: 0,
    };
  });

  const simulation = forceSimulation(simNodes)
    .force('charge', forceManyBody<SimNode>().strength(-95))
    .force('collide', forceCollide<SimNode>().radius(62).strength(0.95))
    .force('link', forceLink<SimNode, { source: string; target: string }>(
      edges.map((edge) => ({ source: edge.source, target: edge.target })),
    ).id((node) => node.id).distance(118).strength(0.12))
    .force('center', forceCenter(0, 0))
    .force('cluster-x', forceX<SimNode>((node) => (domainCenters.get(node.domain) ?? { x: 0 }).x).strength(0.1))
    .force('cluster-y', forceY<SimNode>((node) => (domainCenters.get(node.domain) ?? { y: 0 }).y).strength(0.1))
    .stop();

  for (let tick = 0; tick < WORKBENCH_TICKS; tick += 1) {
    simulation.tick();
  }

  const resolved = resolveNodeCollisions(
    simNodes.map((node) => ({
      id: node.id,
      x: node.x,
      y: node.y,
      width: NODE_WIDTH,
      height: NODE_HEIGHT,
    })),
    {
      margin: collisionTuning.workbench.margin,
      maxIterations: collisionTuning.workbench.maxIterations,
      reservedInsets: options.reservedInsets,
    },
  );

  const byID = new Map(resolved.map((node) => [node.id, node]));
  const laidOutNodes = nodes.map((node) => {
    const position = byID.get(node.id);
    return {
      ...node,
      width: NODE_WIDTH,
      height: NODE_HEIGHT,
      position: {
        x: position?.x ?? 180,
        y: position?.y ?? 140,
      },
    };
  });

  const mergedNodes = applyManualPositions(laidOutNodes, options.manualPositions);
  const positions = resolvePositions(mergedNodes, 'workbench', {
    lockedNodeIds: new Set(Object.keys(options.manualPositions)),
    reservedInsets: options.reservedInsets,
  });

  return {
    nodes: mergedNodes.map((node) => ({
      ...node,
      position: positions[node.id] ?? node.position,
    })),
    edges,
  };
}

function resolveNodeCollisions(nodes: CollisionNode[], options: CollisionOptions): CollisionNode[] {
  const lockIDs = options.lockIDs ?? new Set<string>();
  const priorityIDs = options.priorityIDs ?? new Set<string>();
  const resolved = nodes
    .map((node) => ({ ...node }))
    .sort((left, right) => left.id.localeCompare(right.id));

  for (let iteration = 0; iteration < options.maxIterations; iteration += 1) {
    let moved = false;

    for (let leftIndex = 0; leftIndex < resolved.length; leftIndex += 1) {
      for (let rightIndex = leftIndex + 1; rightIndex < resolved.length; rightIndex += 1) {
        const left = resolved[leftIndex];
        const right = resolved[rightIndex];
        const overlap = intersection(left, right, options.margin);
        if (!overlap) {
          continue;
        }

        moved = true;
        const dx = centerX(right) - centerX(left);
        const dy = centerY(right) - centerY(left);
        const primary = Math.abs(dx) >= Math.abs(dy) ? 'x' : 'y';
        const directionX = dx === 0 ? (left.id < right.id ? -1 : 1) : Math.sign(dx);
        const directionY = dy === 0 ? (left.id < right.id ? -1 : 1) : Math.sign(dy);
        const shiftX = primary === 'x' ? overlap.x / 2 + 0.5 : overlap.x * 0.2;
        const shiftY = primary === 'y' ? overlap.y / 2 + 0.5 : overlap.y * 0.2;

        const leftLocked = lockIDs.has(left.id);
        const rightLocked = lockIDs.has(right.id);
        const leftPriority = priorityIDs.has(left.id);
        const rightPriority = priorityIDs.has(right.id);

        if (leftPriority && rightPriority) {
          nudge(left, -directionX * shiftX, -directionY * shiftY, primary);
          nudge(right, directionX * shiftX, directionY * shiftY, primary);
          continue;
        }

        if (leftPriority) {
          nudge(right, directionX * overlap.x, directionY * overlap.y, primary);
          continue;
        }

        if (rightPriority) {
          nudge(left, -directionX * overlap.x, -directionY * overlap.y, primary);
          continue;
        }

        if (leftLocked && rightLocked) {
          nudge(left, -directionX * shiftX, -directionY * shiftY, primary);
          nudge(right, directionX * shiftX, directionY * shiftY, primary);
          continue;
        }

        if (leftLocked) {
          nudge(right, directionX * overlap.x, directionY * overlap.y, primary);
          continue;
        }

        if (rightLocked) {
          nudge(left, -directionX * overlap.x, -directionY * overlap.y, primary);
          continue;
        }

        nudge(left, -directionX * shiftX, -directionY * shiftY, primary);
        nudge(right, directionX * shiftX, directionY * shiftY, primary);
      }
    }

    applyViewportBounds(resolved);
    applyReservedArea(resolved, options.reservedInsets);

    if (!moved) {
      break;
    }
  }

  applyViewportBounds(resolved);
  applyReservedArea(resolved, options.reservedInsets);
  return resolved;
}

function intersection(left: CollisionNode, right: CollisionNode, margin: number): { x: number; y: number } | null {
  const minGapX = (left.width + right.width) / 2 + margin;
  const minGapY = (left.height + right.height) / 2 + margin;
  const gapX = Math.abs(centerX(right) - centerX(left));
  const gapY = Math.abs(centerY(right) - centerY(left));
  const overlapX = minGapX - gapX;
  const overlapY = minGapY - gapY;

  if (overlapX <= 0 || overlapY <= 0) {
    return null;
  }

  return { x: overlapX, y: overlapY };
}

function nudge(node: CollisionNode, dx: number, dy: number, primary: 'x' | 'y'): void {
  if (primary === 'x') {
    node.x += dx;
    node.y += dy * 0.35;
    return;
  }

  node.x += dx * 0.35;
  node.y += dy;
}

function applyViewportBounds(nodes: CollisionNode[]): void {
  for (const node of nodes) {
    node.x = Math.max(node.x, VIEWPORT_MARGIN);
    node.y = Math.max(node.y, VIEWPORT_MARGIN);
  }
}

function applyReservedArea(nodes: CollisionNode[], reservedInsets?: { top: number; left: number }): void {
  if (!reservedInsets) {
    return;
  }

  const reservedRight = VIEWPORT_MARGIN + reservedInsets.left;
  const reservedBottom = VIEWPORT_MARGIN + reservedInsets.top;
  const buffer = 14;

  for (const node of nodes) {
    const overlapsReservedX = node.x < reservedRight;
    const overlapsReservedY = node.y < reservedBottom;
    if (!overlapsReservedX || !overlapsReservedY) {
      continue;
    }

    const shiftRight = reservedRight - node.x + buffer;
    const shiftDown = reservedBottom - node.y + buffer;
    if (shiftRight <= shiftDown) {
      node.x += shiftRight;
      continue;
    }

    node.y += shiftDown;
  }
}

function centerX(node: CollisionNode): number {
  return node.x + node.width / 2;
}

function centerY(node: CollisionNode): number {
  return node.y + node.height / 2;
}

function buildDomainCenters(nodes: Node[], reservedInsets: { top: number; left: number }): Map<string, { x: number; y: number }> {
  const domains = sortedDomains(nodes);
  const centers = new Map<string, { x: number; y: number }>();
  const centerX = reservedInsets.left + 420;
  const centerY = reservedInsets.top + 280;

  if (domains.length === 1) {
    centers.set(domains[0], { x: centerX, y: centerY });
    return centers;
  }

  const radius = Math.max(180, domains.length * 56);
  domains.forEach((domain, index) => {
    const angle = (index / domains.length) * Math.PI * 2;
    centers.set(domain, {
      x: centerX + Math.cos(angle) * radius,
      y: centerY + Math.sin(angle) * radius,
    });
  });

  return centers;
}

function sortedDomains(nodes: Node[]): string[] {
  return Array.from(new Set(nodes.map((node) => readString(node, 'domain') || 'unassigned'))).sort();
}

function compareNodes(left: Node, right: Node): number {
  const typeCompare = readString(left, 'type').localeCompare(readString(right, 'type'));
  if (typeCompare !== 0) {
    return typeCompare;
  }

  const labelCompare = readString(left, 'label').localeCompare(readString(right, 'label'));
  if (labelCompare !== 0) {
    return labelCompare;
  }

  return left.id.localeCompare(right.id);
}

function readString(node: Node, key: string): string {
  const value = (node.data as Record<string, unknown> | undefined)?.[key];
  return typeof value === 'string' ? value : '';
}

function applyManualPositions(
  nodes: Node[],
  manualPositions: Record<string, { x: number; y: number }>,
): Node[] {
  return nodes.map((node) => {
    const manualPosition = manualPositions[node.id];
    if (!manualPosition) {
      return node;
    }

    return {
      ...node,
      position: {
        x: manualPosition.x,
        y: manualPosition.y,
      },
    };
  });
}

function hashSeed(input: string): number {
  let hash = 2166136261;
  for (let index = 0; index < input.length; index += 1) {
    hash ^= input.charCodeAt(index);
    hash = Math.imul(hash, 16777619);
  }
  return hash >>> 0;
}

function mulberry32(seed: number): () => number {
  return () => {
    let t = seed += 0x6d2b79f5;
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}
