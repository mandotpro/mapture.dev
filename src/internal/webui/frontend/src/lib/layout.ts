import { forceCenter, forceCollide, forceLink, forceManyBody, forceSimulation, forceX, forceY } from 'd3-force';
import type { Edge, Node } from '@xyflow/svelte';

const NODE_WIDTH = 156;
const NODE_HEIGHT = 82;
const TICKS = 220;
const VIEWPORT_MARGIN = 56;

type SimNode = {
  id: string;
  domain: string;
  owner: string;
  x: number;
  y: number;
  vx: number;
  vy: number;
};

type LayoutMode = 'freeform' | 'clustered';

type LayoutOptions = {
  mode: LayoutMode;
  savedPositions: Record<string, { x: number; y: number }>;
};

type CollisionOptions = {
  margin: number;
  maxIterations: number;
  lockIDs?: Set<string>;
};

type CollisionNode = {
  id: string;
  x: number;
  y: number;
  width: number;
  height: number;
};

const layoutTuning: Record<LayoutMode, {
  clusterStrength: number;
  linkStrength: number;
  linkDistance: number;
  collisionMargin: number;
  collisionIterations: number;
}> = {
  freeform: {
    clusterStrength: 0.1,
    linkStrength: 0.12,
    linkDistance: 118,
    collisionMargin: 28,
    collisionIterations: 160,
  },
  clustered: {
    clusterStrength: 0.19,
    linkStrength: 0.22,
    linkDistance: 96,
    collisionMargin: 34,
    collisionIterations: 220,
  },
};

export function layoutGraph(nodes: Node[], edges: Edge[], options: LayoutOptions): { nodes: Node[]; edges: Edge[] } {
  if (nodes.length === 0) {
    return { nodes, edges };
  }

  const tuning = layoutTuning[options.mode];
  const domainCenters = buildDomainCenters(nodes);
  const random = mulberry32(hashSeed(nodes.map((node) => node.id).join('|')));

  const simNodes: SimNode[] = nodes.map((node) => {
    const saved = options.savedPositions[node.id];
    const domain = readString(node, 'domain');
    const owner = readString(node, 'owner');
    const center = domainCenters.get(domain) ?? { x: 0, y: 0 };
    const ownerOffset = hashSeed(owner || node.id) % 40;

    return {
      id: node.id,
      domain,
      owner,
      x: saved?.x ?? center.x + (random() - 0.5) * 140 + ownerOffset - 20,
      y: saved?.y ?? center.y + (random() - 0.5) * 140 - ownerOffset + 20,
      vx: 0,
      vy: 0,
    };
  });

  const simulation = forceSimulation(simNodes)
    .force('charge', forceManyBody<SimNode>().strength(-95))
    .force('collide', forceCollide<SimNode>().radius(62).strength(0.95))
    .force('link', forceLink<SimNode, { source: string; target: string }>(
      edges.map((edge) => ({ source: edge.source, target: edge.target })),
    ).id((node) => node.id).distance(tuning.linkDistance).strength(tuning.linkStrength))
    .force('center', forceCenter(0, 0))
    .force('cluster-x', forceX<SimNode>((node) => (domainCenters.get(node.domain) ?? { x: 0 }).x).strength(tuning.clusterStrength))
    .force('cluster-y', forceY<SimNode>((node) => (domainCenters.get(node.domain) ?? { y: 0 }).y).strength(tuning.clusterStrength))
    .stop();

  for (let tick = 0; tick < TICKS; tick += 1) {
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
      margin: tuning.collisionMargin,
      maxIterations: tuning.collisionIterations,
    },
  );
  const byID = new Map(resolved.map((node) => [node.id, node]));

  return {
    nodes: nodes.map((node) => {
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
    }),
    edges,
  };
}

export function resolvePositions(
  nodes: Node[],
  mode: LayoutMode,
  lockIDs?: Set<string>,
): Record<string, { x: number; y: number }> {
  const tuning = layoutTuning[mode];
  const resolved = resolveNodeCollisions(
    nodes.map((node) => ({
      id: node.id,
      x: node.position.x,
      y: node.position.y,
      width: typeof node.width === 'number' ? node.width : NODE_WIDTH,
      height: typeof node.height === 'number' ? node.height : NODE_HEIGHT,
    })),
    {
      margin: tuning.collisionMargin,
      maxIterations: tuning.collisionIterations,
      lockIDs,
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

function resolveNodeCollisions(nodes: CollisionNode[], options: CollisionOptions): CollisionNode[] {
  const lockIDs = options.lockIDs ?? new Set<string>();
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

        if (leftLocked && rightLocked) {
          nudge(right, directionX * overlap.x, directionY * overlap.y, primary);
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

    if (!moved) {
      break;
    }
  }

  normalizePositions(resolved);
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

function normalizePositions(nodes: CollisionNode[]): void {
  let minX = Number.POSITIVE_INFINITY;
  let minY = Number.POSITIVE_INFINITY;

  for (const node of nodes) {
    minX = Math.min(minX, node.x);
    minY = Math.min(minY, node.y);
  }

  const offsetX = minX < VIEWPORT_MARGIN ? VIEWPORT_MARGIN - minX : 0;
  const offsetY = minY < VIEWPORT_MARGIN ? VIEWPORT_MARGIN - minY : 0;

  for (const node of nodes) {
    node.x += offsetX;
    node.y += offsetY;
  }
}

function centerX(node: CollisionNode): number {
  return node.x + node.width / 2;
}

function centerY(node: CollisionNode): number {
  return node.y + node.height / 2;
}

function buildDomainCenters(nodes: Node[]): Map<string, { x: number; y: number }> {
  const domains = Array.from(new Set(nodes.map((node) => readString(node, 'domain') || 'unassigned'))).sort();
  const centers = new Map<string, { x: number; y: number }>();
  const centerX = 420;
  const centerY = 280;

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

function readString(node: Node, key: string): string {
  const value = (node.data as Record<string, unknown> | undefined)?.[key];
  return typeof value === 'string' ? value : '';
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
