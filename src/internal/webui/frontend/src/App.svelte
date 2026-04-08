<svelte:options runes={true} />

<script lang="ts">
  import { onMount } from 'svelte';
  import {
    Background,
    Controls,
    MiniMap,
    Panel,
    SvelteFlow,
    type Edge,
    type Node,
    type NodeTypes,
  } from '@xyflow/svelte';
  import { loadGraphFromApi, loadGraphFromFile } from './lib/api';
  import {
    domainName,
    findNode,
    graphStats,
    normalizeGraph,
    severitySummary,
    teamName,
    toSvelteFlowEdges,
    toSvelteFlowNodes,
    visibleStats,
  } from './lib/adapter';
  import FlowNode from './lib/FlowNode.svelte';
  import type { Filters, GraphModel, WindowWithPayload } from './lib/types';

  const emptyModel: GraphModel = {
    nodes: [],
    edges: [],
    diagnostics: [],
    domains: [],
    owners: [],
    nodeTypes: [],
    edgeTypes: [],
    teams: new Map(),
    domainNames: new Map(),
    events: new Map(),
  };

  const nodeTypes = {
    architecture: FlowNode,
  } satisfies NodeTypes;

  let model = $state.raw<GraphModel>(emptyModel);
  let flowNodes = $state.raw<Node[]>([]);
  let flowEdges = $state.raw<Edge[]>([]);
  let loading = $state(true);
  let live = $state(false);
  let selectedNodeId = $state<string | null>(null);
  let loadError = $state('');
  let sourceLabel = $state('api');
  let fileInput = $state<HTMLInputElement | null>(null);
  let showFilters = $state(false);
  let showDiagnostics = $state(false);
  let filters = $state.raw<Filters>({
    query: '',
    nodeTypes: [],
    domains: [],
    owners: [],
  });

  const selectedNode = $derived(findNode(model, selectedNodeId));
  const counts = $derived(graphStats(model));
  const visible = $derived(visibleStats(model, filters));
  const summary = $derived(severitySummary(model.diagnostics));
  const typeCounts = $derived(countBy(model.nodes, (node) => node.type));
  const domainCounts = $derived(countBy(model.nodes, (node) => node.domain));
  const ownerCounts = $derived(countBy(model.nodes, (node) => node.owner));

  $effect(() => {
    flowNodes = toSvelteFlowNodes(model, filters, selectedNodeId);
    flowEdges = toSvelteFlowEdges(model, filters);
    if (selectedNodeId && !matchesSelected(selectedNodeId)) {
      selectedNodeId = null;
    }
  });

  async function boot(): Promise<void> {
    loading = true;
    loadError = '';

    try {
      const injected = (window as WindowWithPayload).__MAPTURE_DATA__;
      if (injected?.graph) {
        model = normalizeGraph(injected, injected, { teams: [], domains: [], events: [] });
        sourceLabel = 'static payload';
      } else {
        const payload = await loadGraphFromApi();
        model = normalizeGraph(payload.graph, payload.validation, payload.catalog);
        sourceLabel = 'live api';
        bindLiveReload();
      }
    } catch (error) {
      loadError = error instanceof Error ? error.message : String(error);
      model = emptyModel;
    } finally {
      loading = false;
    }
  }

  function bindLiveReload(): void {
    if (live || typeof EventSource === 'undefined') {
      return;
    }
    live = true;
    const stream = new EventSource('/api/events');
    stream.addEventListener('graph', async () => {
      try {
        const payload = await loadGraphFromApi();
        model = normalizeGraph(payload.graph, payload.validation, payload.catalog);
        loadError = '';
      } catch (error) {
        loadError = error instanceof Error ? error.message : String(error);
      }
    });
    stream.addEventListener('error', () => {
      live = false;
      stream.close();
    });
  }

  function handleNodeClick({ node }: { node: Node }): void {
    selectedNodeId = node.id;
  }

  function resetFilters(): void {
    filters = {
      query: '',
      nodeTypes: [],
      domains: [],
      owners: [],
    };
    selectedNodeId = null;
  }

  function toggleFilter(kind: 'nodeTypes' | 'domains' | 'owners', value: string): void {
    const next = new Set(filters[kind]);
    if (next.has(value)) {
      next.delete(value);
    } else {
      next.add(value);
    }
    filters = {
      ...filters,
      [kind]: Array.from(next).sort((left, right) => left.localeCompare(right)),
    };
  }

  function matchesSelected(nodeID: string): boolean {
    const node = model.nodes.find((candidate) => candidate.id === nodeID);
    if (!node) {
      return false;
    }
    return (
      (filters.nodeTypes.length === 0 || filters.nodeTypes.includes(node.type)) &&
      (filters.domains.length === 0 || filters.domains.includes(node.domain)) &&
      (filters.owners.length === 0 || filters.owners.includes(node.owner)) &&
      (!filters.query ||
        [node.id, node.name, node.domain, node.owner, node.file, node.summary]
          .join(' ')
          .toLowerCase()
          .includes(filters.query.toLowerCase()))
    );
  }

  async function handleFileChange(event: Event): Promise<void> {
    const input = event.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) {
      return;
    }

    try {
      const payload = await loadGraphFromFile(file);
      model = normalizeGraph(payload, payload, { teams: [], domains: [], events: [] });
      sourceLabel = `file: ${file.name}`;
      loadError = '';
      selectedNodeId = null;
      loading = false;
    } catch (error) {
      loadError = error instanceof Error ? error.message : String(error);
    } finally {
      input.value = '';
    }
  }

  function connectionPill(): string {
    if (loadError) {
      return 'load failed';
    }
    if (loading) {
      return 'loading';
    }
    if (sourceLabel.startsWith('file:')) {
      return 'local file';
    }
    if (sourceLabel === 'static payload') {
      return 'static build';
    }
    return 'live api connected';
  }

  function countBy<T>(items: T[], pick: (item: T) => string): Record<string, number> {
    return items.reduce<Record<string, number>>((result, item) => {
      const key = pick(item);
      if (!key) {
        return result;
      }
      result[key] = (result[key] ?? 0) + 1;
      return result;
    }, {});
  }

  onMount(() => {
    void boot();
  });
</script>

<main class="immersive-shell">
  <SvelteFlow
    nodes={flowNodes}
    edges={flowEdges}
    {nodeTypes}
    fitView
    fitViewOptions={{ padding: 0.22 }}
    minZoom={0.12}
    maxZoom={1.8}
    nodesDraggable={false}
    nodesConnectable={false}
    elementsSelectable
    onnodeclick={handleNodeClick}
    attributionPosition="bottom-left"
    class="immersive-flow"
  >
    <Background color="rgba(24, 34, 40, 0.08)" gap={24} />
    <MiniMap position="bottom-left" pannable zoomable />
    <Controls position="bottom-right" />

    <Panel position="top-left" class="overlay-stack overlay-top">
      <section class="overlay-card overlay-hero">
        <div class="pill-row">
          <span class={['pill', 'status', loadError ? 'error' : summary.warnings > 0 ? 'warning' : 'ok'].join(' ')}>
            {connectionPill()}
          </span>
          <span class="pill soft">{visible.nodes} visible nodes</span>
          <span class="pill soft">{visible.edges} visible edges</span>
        </div>
        <div class="overlay-hero__title">
          <div>
            <p class="eyebrow">Mapture Explorer</p>
            <h1>Architecture graph on the full canvas.</h1>
          </div>
          <div class="pill-grid">
            <span class="metric-pill"><strong>{counts.nodes}</strong><small>nodes</small></span>
            <span class="metric-pill"><strong>{counts.edges}</strong><small>edges</small></span>
            <span class="metric-pill"><strong>{counts.domains}</strong><small>domains</small></span>
            <span class="metric-pill"><strong>{counts.owners}</strong><small>teams</small></span>
          </div>
        </div>
        <div class="pill-row pill-row--wrap">
          {#each model.nodeTypes as nodeType}
            <span class={"pill type-pill " + nodeType}>
              {nodeType} {typeCounts[nodeType] ?? 0}
            </span>
          {/each}
        </div>
      </section>
    </Panel>

    <Panel position="top-right" class="overlay-stack overlay-right">
      <section class="overlay-card overlay-controls">
        <div class="control-row">
          <input bind:value={filters.query} type="search" placeholder="Search id, name, domain, owner, file" />
          <button type="button" class="secondary" onclick={() => (showFilters = !showFilters)}>
            {showFilters ? 'Hide filters' : 'Show filters'}
          </button>
        </div>
        <div class="control-row">
          <button type="button" onclick={resetFilters}>Reset</button>
          <button type="button" class="secondary" onclick={() => fileInput?.click()}>Load JSON</button>
          <button type="button" class="secondary" onclick={() => (showDiagnostics = !showDiagnostics)}>
            {showDiagnostics ? 'Hide issues' : 'Show issues'}
          </button>
          <input bind:this={fileInput} class="file-input" type="file" accept="application/json,.json" onchange={handleFileChange} />
        </div>

        {#if showFilters}
          <div class="filter-groups">
            <section>
              <div class="filter-heading">Teams</div>
              <div class="chip-grid">
                {#each model.owners as owner}
                  <button
                    type="button"
                    class={['filter-chip', filters.owners.includes(owner) ? 'active' : ''].join(' ')}
                    onclick={() => toggleFilter('owners', owner)}
                  >
                    <span>{teamName(model, owner)}</span>
                    <small>{ownerCounts[owner] ?? 0}</small>
                  </button>
                {/each}
              </div>
            </section>

            <section>
              <div class="filter-heading">Domains</div>
              <div class="chip-grid">
                {#each model.domains as domain}
                  <button
                    type="button"
                    class={['filter-chip', filters.domains.includes(domain) ? 'active' : ''].join(' ')}
                    onclick={() => toggleFilter('domains', domain)}
                  >
                    <span>{domainName(model, domain)}</span>
                    <small>{domainCounts[domain] ?? 0}</small>
                  </button>
                {/each}
              </div>
            </section>

            <section>
              <div class="filter-heading">Node types</div>
              <div class="chip-grid">
                {#each model.nodeTypes as nodeType}
                  <button
                    type="button"
                    class={['filter-chip', 'kind', nodeType, filters.nodeTypes.includes(nodeType) ? 'active' : ''].join(' ')}
                    onclick={() => toggleFilter('nodeTypes', nodeType)}
                  >
                    <span>{nodeType}</span>
                    <small>{typeCounts[nodeType] ?? 0}</small>
                  </button>
                {/each}
              </div>
            </section>
          </div>
        {/if}
      </section>
    </Panel>

    {#if selectedNode}
      <Panel position="bottom-right" class="overlay-stack overlay-bottom-right">
        <section class="overlay-card detail-panel">
          <div class="detail-heading">
            <h2>{selectedNode.name}</h2>
            <span class="pill soft">{selectedNode.type}</span>
          </div>

          <div class="detail-grid">
            <div>
              <span class="detail-label">Node id</span>
              <div class="mono">{selectedNode.id}</div>
            </div>
            <div>
              <span class="detail-label">Domain</span>
              <div>{selectedNode.domain ? domainName(model, selectedNode.domain) : 'n/a'}</div>
            </div>
            <div>
              <span class="detail-label">Owner</span>
              <div>{selectedNode.owner ? teamName(model, selectedNode.owner) : 'n/a'}</div>
            </div>
            <div>
              <span class="detail-label">Source</span>
              <div class="mono">
                {#if selectedNode.file}
                  {selectedNode.file}{selectedNode.line ? `:${selectedNode.line}` : ''}
                {:else}
                  n/a
                {/if}
              </div>
            </div>
            {#if selectedNode.summary}
              <div class="detail-summary">
                <span class="detail-label">Summary</span>
                <p>{selectedNode.summary}</p>
              </div>
            {/if}
          </div>
        </section>
      </Panel>
    {/if}

    {#if showDiagnostics}
      <Panel position="bottom-left" class="overlay-stack overlay-bottom-left">
        <section class="overlay-card diagnostics-panel">
          <div class="detail-heading">
            <h2>Validation feed</h2>
            <span class="pill soft">{summary.errors} errors · {summary.warnings} warnings</span>
          </div>

          {#if loadError}
            <p class="empty">{loadError}</p>
          {:else if loading}
            <p class="empty">Loading graph data…</p>
          {:else if model.diagnostics.length === 0}
            <p class="empty">No diagnostics for the current graph.</p>
          {:else}
            <div class="diagnostic-list">
              {#each model.diagnostics as diagnostic}
                <article class={['diagnostic-item', diagnostic.severity].join(' ')}>
                  <strong>{diagnostic.message}</strong>
                  <div class="diag-meta">
                    {diagnostic.severity} · layer {diagnostic.layer} · {diagnostic.code}
                    {#if diagnostic.file}
                      · {diagnostic.file}{diagnostic.line ? `:${diagnostic.line}` : ''}
                    {/if}
                  </div>
                </article>
              {/each}
            </div>
          {/if}
        </section>
      </Panel>
    {/if}
  </SvelteFlow>
</main>
