import { forceCenter, forceCollide, forceLink, forceManyBody, forceSimulation, forceX, forceY } from 'd3-force';
import type { Edge, Node } from '@xyflow/svelte';

const NODE_WIDTH = 148;
const NODE_HEIGHT = 72;
const TICKS = 220;

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

export function layoutGraph(nodes: Node[], edges: Edge[], options: LayoutOptions): { nodes: Node[]; edges: Edge[] } {
  if (nodes.length === 0) {
    return { nodes, edges };
  }

  const clusterStrength = options.mode === 'clustered' ? 0.19 : 0.1;
  const linkStrength = options.mode === 'clustered' ? 0.22 : 0.12;
  const linkDistance = options.mode === 'clustered' ? 96 : 118;
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
    .force('collide', forceCollide<SimNode>().radius(56).strength(0.95))
    .force('link', forceLink<SimNode, { source: string; target: string }>(
      edges.map((edge) => ({ source: edge.source, target: edge.target })),
    ).id((node) => node.id).distance(linkDistance).strength(linkStrength))
    .force('center', forceCenter(0, 0))
    .force('cluster-x', forceX<SimNode>((node) => (domainCenters.get(node.domain) ?? { x: 0 }).x).strength(clusterStrength))
    .force('cluster-y', forceY<SimNode>((node) => (domainCenters.get(node.domain) ?? { y: 0 }).y).strength(clusterStrength))
    .stop();

  for (let tick = 0; tick < TICKS; tick += 1) {
    simulation.tick();
  }

  const byID = new Map(simNodes.map((node) => [node.id, node]));

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
