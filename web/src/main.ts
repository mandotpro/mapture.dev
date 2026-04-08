// Mapture Explorer frontend.
//
// Shared by the HTML exporter and the `mapture serve` command. Loads a
// graph payload (either fetched from the server's /api/* endpoints or
// injected as window.__MAPTURE_DATA__ by the static exporter), renders
// it with Cytoscape.js, and wires search, legend toggles, node details,
// neighborhood isolation, and domain grouping.

declare const cytoscape: any;

interface GraphNode {
  id: string;
  type: string;
  name: string;
  domain?: string;
  owner?: string;
  file?: string;
  line?: number;
  summary?: string;
}

interface GraphEdge {
  from: string;
  to: string;
  type: string;
}

interface Graph {
  nodes: GraphNode[];
  edges: GraphEdge[];
}

interface Diagnostic {
  severity: string;
  layer: number;
  code: string;
  message: string;
  file?: string;
  line?: number;
}

interface Payload {
  graph: Graph;
  diagnostics?: Diagnostic[];
}

const NODE_COLORS: Record<string, string> = {
  service: "#2563eb",
  api: "#14b8a6",
  database: "#f59e0b",
  event: "#a855f7",
};

const EDGE_COLORS: Record<string, string> = {
  calls: "#2563eb",
  depends_on: "#64748b",
  stores_in: "#f59e0b",
  reads_from: "#0891b2",
  emits: "#a855f7",
  consumes: "#db2777",
};

interface State {
  cy: any;
  payload: Payload;
  hiddenNodeTypes: Set<string>;
  hiddenEdgeTypes: Set<string>;
  search: string;
  isolateRoot: string | null;
}

const state: State = {
  cy: null,
  payload: { graph: { nodes: [], edges: [] }, diagnostics: [] },
  hiddenNodeTypes: new Set(),
  hiddenEdgeTypes: new Set(),
  search: "",
  isolateRoot: null,
};

function $(id: string): HTMLElement {
  const el = document.getElementById(id);
  if (!el) throw new Error(`missing element #${id}`);
  return el;
}

function escapeHTML(value: unknown): string {
  return String(value ?? "").replace(/[&<>"']/g, (c) => ({
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    '"': "&quot;",
    "'": "&#39;",
  }[c] as string));
}

function matchesSearch(node: GraphNode, query: string): boolean {
  if (!query) return true;
  const q = query.toLowerCase();
  return [node.id, node.name, node.domain, node.owner]
    .filter((v): v is string => typeof v === "string")
    .some((v) => v.toLowerCase().includes(q));
}

function toElements(graph: Graph): any[] {
  const domains = new Map<string, string>();
  for (const n of graph.nodes) {
    if (n.domain && !domains.has(n.domain)) {
      domains.set(n.domain, `domain:${n.domain}`);
    }
  }

  const elements: any[] = [];
  for (const [domain, id] of domains) {
    elements.push({
      group: "nodes",
      data: { id, label: domain, isGroup: true },
      classes: "group",
    });
  }

  for (const n of graph.nodes) {
    elements.push({
      group: "nodes",
      data: {
        id: n.id,
        label: n.name || n.id,
        nodeType: n.type,
        domain: n.domain || "",
        owner: n.owner || "",
        file: n.file || "",
        line: n.line || 0,
        summary: n.summary || "",
        parent: n.domain ? `domain:${n.domain}` : undefined,
      },
    });
  }

  for (const e of graph.edges) {
    elements.push({
      group: "edges",
      data: {
        id: `${e.from}->${e.to}|${e.type}`,
        source: e.from,
        target: e.to,
        edgeType: e.type,
        label: e.type,
      },
    });
  }

  return elements;
}

function cyStyle(): any[] {
  return [
    {
      selector: "node",
      style: {
        "background-color": (ele: any) => NODE_COLORS[ele.data("nodeType")] || "#64748b",
        label: "data(label)",
        color: "#fff",
        "text-outline-color": "#1a1a1a",
        "text-outline-width": 2,
        "font-size": 11,
        width: 38,
        height: 38,
        "text-valign": "center",
        "text-halign": "center",
      },
    },
    {
      selector: "node.group",
      style: {
        "background-opacity": 0.05,
        "background-color": "#64748b",
        "border-color": "#64748b",
        "border-width": 1,
        "border-style": "dashed",
        label: "data(label)",
        color: "#64748b",
        "text-valign": "top",
        "text-halign": "center",
        "text-outline-width": 0,
        "font-size": 10,
        "text-transform": "uppercase",
        shape: "round-rectangle",
        padding: 14,
      },
    },
    {
      selector: "node:selected",
      style: { "border-width": 3, "border-color": "#2563eb" },
    },
    {
      selector: "edge",
      style: {
        "curve-style": "bezier",
        "target-arrow-shape": "triangle",
        width: 1.5,
        "line-color": (ele: any) => EDGE_COLORS[ele.data("edgeType")] || "#64748b",
        "target-arrow-color": (ele: any) => EDGE_COLORS[ele.data("edgeType")] || "#64748b",
        label: "data(label)",
        "font-size": 8,
        color: "#94a3b8",
        "text-background-color": "#0f1115",
        "text-background-opacity": 0.6,
        "text-background-padding": "1px",
      },
    },
    { selector: ".dimmed", style: { opacity: 0.08 } },
    { selector: ".hidden", style: { display: "none" } },
  ];
}

function uniqueTypes(graph: Graph): { nodes: string[]; edges: string[] } {
  const nt = new Set<string>();
  const et = new Set<string>();
  for (const n of graph.nodes) nt.add(n.type);
  for (const e of graph.edges) et.add(e.type);
  return {
    nodes: Array.from(nt).sort(),
    edges: Array.from(et).sort(),
  };
}

function renderLegend(): void {
  const { nodes, edges } = uniqueTypes(state.payload.graph);

  const nodeList = $("node-legend");
  nodeList.innerHTML = nodes
    .map((type) => {
      const off = state.hiddenNodeTypes.has(type) ? " off" : "";
      const color = NODE_COLORS[type] || "#64748b";
      return `<li data-node-type="${escapeHTML(type)}" class="${off.trim()}"><span class="swatch" style="background:${color}"></span>${escapeHTML(type)}</li>`;
    })
    .join("");
  nodeList.querySelectorAll("li").forEach((li) => {
    li.addEventListener("click", () => {
      const type = (li as HTMLElement).dataset.nodeType!;
      if (state.hiddenNodeTypes.has(type)) state.hiddenNodeTypes.delete(type);
      else state.hiddenNodeTypes.add(type);
      renderLegend();
      applyFilters();
    });
  });

  const edgeList = $("edge-legend");
  edgeList.innerHTML = edges
    .map((type) => {
      const off = state.hiddenEdgeTypes.has(type) ? " off" : "";
      const color = EDGE_COLORS[type] || "#64748b";
      return `<li data-edge-type="${escapeHTML(type)}" class="${off.trim()}"><span class="swatch" style="background:${color}"></span>${escapeHTML(type)}</li>`;
    })
    .join("");
  edgeList.querySelectorAll("li").forEach((li) => {
    li.addEventListener("click", () => {
      const type = (li as HTMLElement).dataset.edgeType!;
      if (state.hiddenEdgeTypes.has(type)) state.hiddenEdgeTypes.delete(type);
      else state.hiddenEdgeTypes.add(type);
      renderLegend();
      applyFilters();
    });
  });
}

function applyFilters(): void {
  if (!state.cy) return;

  const visibleNodeIds = new Set<string>();
  const searchQuery = state.search.trim();
  const searchMatches = new Set<string>();

  for (const n of state.payload.graph.nodes) {
    const typeHidden = state.hiddenNodeTypes.has(n.type);
    if (typeHidden) continue;
    if (matchesSearch(n, searchQuery)) {
      searchMatches.add(n.id);
    }
    visibleNodeIds.add(n.id);
  }

  let active = searchQuery ? new Set(searchMatches) : new Set(visibleNodeIds);

  if (state.isolateRoot && active.has(state.isolateRoot)) {
    const neighborhood = new Set<string>([state.isolateRoot]);
    for (const e of state.payload.graph.edges) {
      if (state.hiddenEdgeTypes.has(e.type)) continue;
      if (e.from === state.isolateRoot) neighborhood.add(e.to);
      if (e.to === state.isolateRoot) neighborhood.add(e.from);
    }
    active = new Set(Array.from(active).filter((id) => neighborhood.has(id)));
    active.add(state.isolateRoot);
  }

  state.cy.batch(() => {
    state.cy.nodes().forEach((el: any) => {
      if (el.data("isGroup")) return;
      const show = active.has(el.id());
      el.toggleClass("hidden", !show);
    });
    // Hide a domain group if all children are hidden.
    state.cy.nodes(".group").forEach((group: any) => {
      const children = group.children();
      const anyVisible = children.toArray().some((c: any) => !c.hasClass("hidden"));
      group.toggleClass("hidden", !anyVisible);
    });
    state.cy.edges().forEach((el: any) => {
      const type = el.data("edgeType");
      const src = el.source().id();
      const dst = el.target().id();
      const show =
        !state.hiddenEdgeTypes.has(type) && active.has(src) && active.has(dst);
      el.toggleClass("hidden", !show);
    });
  });
}

function showDetails(node: GraphNode | null): void {
  const details = $("details");
  if (!node) {
    details.innerHTML = `<h2>Details</h2><p class="muted">Click a node to inspect it.</p>`;
    return;
  }
  const location = node.file ? `${escapeHTML(node.file)}${node.line ? ":" + node.line : ""}` : "—";
  details.innerHTML = `
    <h2>Details</h2>
    <dl>
      <dt>id</dt><dd>${escapeHTML(node.id)}</dd>
      <dt>type</dt><dd>${escapeHTML(node.type)}</dd>
      <dt>name</dt><dd>${escapeHTML(node.name)}</dd>
      <dt>domain</dt><dd>${escapeHTML(node.domain || "—")}</dd>
      <dt>owner</dt><dd>${escapeHTML(node.owner || "—")}</dd>
      <dt>source</dt><dd>${location}</dd>
    </dl>
    ${node.summary ? `<p class="summary">${escapeHTML(node.summary)}</p>` : ""}
  `;
}

function renderDiagnostics(diagnostics: Diagnostic[]): void {
  const container = $("diagnostics");
  if (!diagnostics.length) {
    container.innerHTML = '<span class="muted">none</span>';
    return;
  }
  container.innerHTML = diagnostics
    .map((d) => {
      const loc = d.file ? ` ${escapeHTML(d.file)}${d.line ? ":" + d.line : ""}` : "";
      return `<div class="diag ${escapeHTML(d.severity)}">[${escapeHTML(d.severity)}] layer ${d.layer} ${escapeHTML(d.code)}${loc}: ${escapeHTML(d.message)}</div>`;
    })
    .join("");
}

function updateStatus(): void {
  const g = state.payload.graph;
  $("status").textContent = `${g.nodes.length} nodes · ${g.edges.length} edges`;
}

function buildGraph(): void {
  const container = $("cy");
  const start = performance.now();
  state.cy = cytoscape({
    container,
    elements: toElements(state.payload.graph),
    style: cyStyle(),
    layout: { name: "cose", animate: false, nodeDimensionsIncludeLabels: true },
    wheelSensitivity: 0.2,
  });
  const elapsed = Math.round(performance.now() - start);
  (window as any).__MAPTURE_LAST_RENDER_MS = elapsed;

  state.cy.on("tap", "node", (evt: any) => {
    const node = evt.target;
    if (node.data("isGroup")) return;
    const found = state.payload.graph.nodes.find((n) => n.id === node.id()) || null;
    showDetails(found);
    state.isolateRoot = found?.id || null;
  });
  state.cy.on("tap", (evt: any) => {
    if (evt.target === state.cy) {
      showDetails(null);
      state.isolateRoot = null;
      applyFilters();
    }
  });
}

export function renderPayload(payload: Payload): void {
  state.payload = {
    graph: payload.graph || { nodes: [], edges: [] },
    diagnostics: payload.diagnostics || [],
  };
  updateStatus();
  renderLegend();
  renderDiagnostics(state.payload.diagnostics || []);
  buildGraph();
  applyFilters();
}

async function refreshFromServer(): Promise<void> {
  try {
    const [graphRes, validateRes] = await Promise.all([
      fetch("/api/graph"),
      fetch("/api/validate"),
    ]);
    const graph = (await graphRes.json()) as Graph;
    const validate = (await validateRes.json()) as Payload;
    renderPayload({ graph, diagnostics: validate.diagnostics || [] });
  } catch (err) {
    $("status").textContent = "error: " + String(err);
  }
}

function wireTopbar(): void {
  $("search").addEventListener("input", (e) => {
    state.search = (e.target as HTMLInputElement).value;
    applyFilters();
  });
  $("isolate").addEventListener("click", () => {
    if (!state.cy) return;
    const selected = state.cy.$("node:selected").filter((n: any) => !n.data("isGroup"));
    if (selected.length === 0) return;
    state.isolateRoot = selected[0].id();
    applyFilters();
  });
  $("reset").addEventListener("click", () => {
    state.search = "";
    state.isolateRoot = null;
    state.hiddenNodeTypes.clear();
    state.hiddenEdgeTypes.clear();
    (document.getElementById("search") as HTMLInputElement).value = "";
    renderLegend();
    applyFilters();
    if (state.cy) state.cy.fit();
  });
}

function boot(): void {
  wireTopbar();

  const inline = (window as any).__MAPTURE_DATA__ as Payload | undefined;
  if (inline && inline.graph) {
    renderPayload(inline);
    return;
  }

  void refreshFromServer();

  if (typeof EventSource !== "undefined") {
    try {
      const es = new EventSource("/api/events");
      es.addEventListener("graph", () => {
        void refreshFromServer();
      });
    } catch {
      // SSE not available; fall back to static snapshot.
    }
  }
}

// Expose for the exporter-injected payload path and tests.
(window as any).mapture = { renderPayload };

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", boot);
} else {
  boot();
}
