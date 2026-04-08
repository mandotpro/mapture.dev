<script lang="ts">
  import { onMount } from 'svelte';
  import { Background, Controls, MiniMap, SvelteFlow } from '@xyflow/svelte';
  import type { Edge, Node } from '@xyflow/svelte';
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

  let model: GraphModel = emptyModel;
  let nodes: Node[] = [];
  let edges: Edge[] = [];
  let loading = true;
  let live = false;
  let selectedNodeId: string | null = null;
  let loadError = '';
  let sourceLabel = 'api';
  let fileInput: HTMLInputElement | null = null;
  let filters: Filters = {
    query: '',
    nodeType: '',
    domain: '',
    owner: '',
  };

  $: selectedNode = findNode(model, selectedNodeId);
  $: counts = graphStats(model);
  $: visible = visibleStats(model, filters);
  $: summary = severitySummary(model.diagnostics);
  $: applyGraph();

  function applyGraph(): void {
    const nextNodes = toSvelteFlowNodes(model, filters, selectedNodeId);
    const nextEdges = toSvelteFlowEdges(model, filters, new Set(nextNodes.map((node) => node.id)));
    nodes = nextNodes;
    edges = nextEdges;
    if (selectedNodeId && !nextNodes.some((node) => node.id === selectedNodeId)) {
      selectedNodeId = null;
    }
  }

  async function boot(): Promise<void> {
    loading = true;
    loadError = '';

    try {
      const injected = (window as WindowWithPayload).__MAPTURE_DATA__;
      if (injected?.graph) {
        model = normalizeGraph(injected, injected, { teams: [], domains: [], events: [] });
        sourceLabel = 'embedded payload';
      } else {
        const payload = await loadGraphFromApi();
        model = normalizeGraph(payload.graph, payload.validation, payload.catalog);
        sourceLabel = 'live server';
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
      } catch (error) {
        loadError = error instanceof Error ? error.message : String(error);
      }
    });
    stream.addEventListener('error', () => {
      live = false;
      stream.close();
    });
  }

  function handleNodeClick(event: CustomEvent<{ node: Node }>): void {
    selectedNodeId = event.detail.node.id;
  }

  function resetFilters(): void {
    filters = {
      query: '',
      nodeType: '',
      domain: '',
      owner: '',
    };
    selectedNodeId = null;
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
      loading = false;
      selectedNodeId = null;
    } catch (error) {
      loadError = error instanceof Error ? error.message : String(error);
    } finally {
      input.value = '';
    }
  }

  onMount(() => {
    void boot();
  });
</script>

<div class="shell">
  <section class="hero">
    <div class="panel hero-card">
      <div class="hero-title">
        <div>
          <p class="eyebrow">Architecture Explorer</p>
          <h1>Explore the validated graph, not raw comments.</h1>
        </div>
        <div class="status-pill" class:error={!!loadError} class:warning={!loadError && summary.warnings > 0}>
          {#if loadError}
            load failed
          {:else if loading}
            loading
          {:else if summary.errors > 0}
            validation issues
          {:else}
            ready
          {/if}
        </div>
      </div>
      <p class="hero-copy">
        This UI reads the existing Go endpoints, adapts the normalized graph model into Svelte Flow,
        and keeps details, diagnostics, search, and filtering centered around the current backend shape.
      </p>
      <div class="hero-stats">
        <div class="stat">
          <span class="detail-label">Total nodes</span>
          <strong>{counts.nodes}</strong>
        </div>
        <div class="stat">
          <span class="detail-label">Total edges</span>
          <strong>{counts.edges}</strong>
        </div>
        <div class="stat">
          <span class="detail-label">Visible nodes</span>
          <strong>{visible.nodes}</strong>
        </div>
        <div class="stat">
          <span class="detail-label">Visible edges</span>
          <strong>{visible.edges}</strong>
        </div>
      </div>
    </div>
    <div class="status-stack">
      <div class="panel status-card">
        <div class="status-row">
          <h2>Backend contract</h2>
          <span class="status-pill">{sourceLabel}</span>
        </div>
        <div class="status-meta">
          <span>`GET /api/graph` nodes and edges</span>
          <span>`GET /api/validate` diagnostics</span>
          <span>`GET /api/catalog` owners and domains</span>
          <span>`GET /api/events` live reload</span>
        </div>
      </div>
      <div class="panel status-card">
        <div class="status-row">
          <h2>Diagnostics</h2>
          <span>{summary.errors} errors · {summary.warnings} warnings</span>
        </div>
        {#if loadError}
          <p class="empty">{loadError}</p>
        {:else if loading}
          <p class="empty">Loading graph and catalog data…</p>
        {:else}
          <p class="empty">
            The explorer stays compatible with current backend output and can also load exported JSON payloads from disk.
          </p>
        {/if}
      </div>
    </div>
  </section>

  <aside class="panel left-rail">
    <div class="toolbar">
      <div>
        <h2>Filters</h2>
        <p class="empty">Search is matched against id, name, domain, owner, file, and summary.</p>
      </div>

      <input bind:value={filters.query} type="search" placeholder="Search graph" />

      <select bind:value={filters.nodeType}>
        <option value="">All node types</option>
        {#each model.nodeTypes as nodeType}
          <option value={nodeType}>{nodeType}</option>
        {/each}
      </select>

      <select bind:value={filters.domain}>
        <option value="">All domains</option>
        {#each model.domains as domain}
          <option value={domain}>{domainName(model, domain)}</option>
        {/each}
      </select>

      <select bind:value={filters.owner}>
        <option value="">All owners</option>
        {#each model.owners as owner}
          <option value={owner}>{teamName(model, owner)}</option>
        {/each}
      </select>

      <div class="actions">
        <button type="button" on:click={resetFilters}>Reset filters</button>
        <button class="secondary" type="button" on:click={() => fileInput?.click()}>Load JSON file</button>
        <input bind:this={fileInput} class="file-input" type="file" accept="application/json,.json" on:change={handleFileChange} />
      </div>
    </div>

    <div class="legend">
      <h3>Node types</h3>
      {#each model.nodeTypes as nodeType}
        <button
          type="button"
          class:off={filters.nodeType !== '' && filters.nodeType !== nodeType}
          on:click={() => {
            filters.nodeType = filters.nodeType === nodeType ? '' : nodeType;
            filters = filters;
          }}
        >
          <span class="legend-label">
            <span class={"dot " + nodeType}></span>
            <span>{nodeType}</span>
          </span>
          <span class="edge-chip">
            {model.nodes.filter((node) => node.type === nodeType).length}
          </span>
        </button>
      {/each}
    </div>

    <div class="legend">
      <h3>Edge types</h3>
      {#each model.edgeTypes as edgeType}
        <div class="detail-card">
          <span class="detail-label">{edgeType}</span>
          <div class="edge-chip">{model.edges.filter((edge) => edge.type === edgeType).length} edges</div>
        </div>
      {/each}
    </div>
  </aside>

  <section class="panel canvas">
    <div class="flow-host">
      <SvelteFlow
        class="flow-shell"
        {nodes}
        {edges}
        fitView
        minZoom={0.1}
        maxZoom={1.8}
        on:nodeclick={handleNodeClick}
      >
        <Background color="rgba(24, 34, 40, 0.09)" gap={22} />
        <Controls />
        <MiniMap pannable zoomable />
      </SvelteFlow>
    </div>
  </section>

  <aside class="panel right-rail">
    <div class="details-grid">
      <div>
        <h2>Node details</h2>
        {#if selectedNode}
          <div class="detail-card">
            <span class="detail-label">Node</span>
            <strong>{selectedNode.name}</strong>
            <div class="diag-meta mono">{selectedNode.id}</div>
          </div>
          <div class="detail-card">
            <span class="detail-label">Type</span>
            <div>{selectedNode.type}</div>
          </div>
          <div class="detail-card">
            <span class="detail-label">Domain</span>
            <div>{selectedNode.domain ? domainName(model, selectedNode.domain) : 'n/a'}</div>
          </div>
          <div class="detail-card">
            <span class="detail-label">Owner</span>
            <div>{selectedNode.owner ? teamName(model, selectedNode.owner) : 'n/a'}</div>
          </div>
          <div class="detail-card">
            <span class="detail-label">Source</span>
            <div class="mono">
              {#if selectedNode.file}
                {selectedNode.file}{selectedNode.line ? `:${selectedNode.line}` : ''}
              {:else}
                n/a
              {/if}
            </div>
          </div>
          {#if selectedNode.symbol}
            <div class="detail-card">
              <span class="detail-label">Symbol</span>
              <div class="mono">{selectedNode.symbol}</div>
            </div>
          {/if}
          {#if selectedNode.summary}
            <div class="detail-card">
              <span class="detail-label">Summary</span>
              <div>{selectedNode.summary}</div>
            </div>
          {/if}
        {:else}
          <p class="empty">Select a node on the canvas to inspect its owner, source attachment, and summary.</p>
        {/if}
      </div>

      <div>
        <h2>Validation feed</h2>
        {#if model.diagnostics.length === 0}
          <p class="empty">No diagnostics for the current graph.</p>
        {:else}
          <div class="diagnostics">
            {#each model.diagnostics as diagnostic}
              <div class={"diag-item " + diagnostic.severity}>
                <strong>{diagnostic.message}</strong>
                <div class="diag-meta">
                  {diagnostic.severity} · layer {diagnostic.layer} · {diagnostic.code}
                  {#if diagnostic.file}
                    · {diagnostic.file}{diagnostic.line ? `:${diagnostic.line}` : ''}
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  </aside>
</div>
