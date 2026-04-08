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
    normalizeGraph,
    severitySummary,
    teamName,
    toSvelteFlowEdges,
    toSvelteFlowNodes,
    visibleStats,
  } from './lib/adapter';
  import FlowNode from './lib/FlowNode.svelte';
  import type { Filters, GraphModel, WindowWithPayload } from './lib/types';

  type FilterPopover = 'owners' | 'domains' | 'nodeTypes' | null;
  type NodePopup = {
    nodeId: string;
    x: number;
    y: number;
  } | null;

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
  let loadError = $state('');
  let sourceLabel = $state('api');
  let fileInput = $state<HTMLInputElement | null>(null);
  let searchOpen = $state(false);
  let activePopover = $state<FilterPopover>(null);
  let nodePopup = $state<NodePopup>(null);
  let filters = $state.raw<Filters>({
    query: '',
    nodeTypes: [],
    domains: [],
    owners: [],
  });

  const popupNode = $derived(findNode(model, nodePopup?.nodeId ?? null));
  const visible = $derived(visibleStats(model, filters));
  const summary = $derived(severitySummary(model.diagnostics));
  const typeCounts = $derived(countBy(model.nodes, (node) => node.type));
  const ownerCounts = $derived(countBy(model.nodes, (node) => node.owner));
  const domainCounts = $derived(countBy(model.nodes, (node) => node.domain));
  const compactTypePills = $derived(model.nodeTypes.slice(0, 4));

  $effect(() => {
    flowNodes = toSvelteFlowNodes(model, filters, nodePopup?.nodeId ?? null);
    flowEdges = toSvelteFlowEdges(model, filters);
    if (nodePopup && !matchesFilters(nodePopup.nodeId)) {
      nodePopup = null;
    }
  });

  async function boot(): Promise<void> {
    loading = true;
    loadError = '';

    try {
      const injected = (window as WindowWithPayload).__MAPTURE_DATA__;
      if (injected?.graph) {
        model = normalizeGraph(injected, injected, { teams: [], domains: [], events: [] });
        sourceLabel = 'static build';
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

  function connectionLabel(): string {
    if (loadError) {
      return 'load failed';
    }
    if (loading) {
      return 'loading';
    }
    if (sourceLabel.startsWith('file:')) {
      return 'local file';
    }
    return sourceLabel;
  }

  function connectionTone(): string {
    if (loadError) {
      return 'error';
    }
    if (summary.warnings > 0) {
      return 'warning';
    }
    return 'ok';
  }

  function resetFilters(): void {
    filters = {
      query: '',
      nodeTypes: [],
      domains: [],
      owners: [],
    };
    activePopover = null;
    nodePopup = null;
  }

  function clearFilter(kind: 'owners' | 'domains' | 'nodeTypes'): void {
    filters = {
      ...filters,
      [kind]: [],
    };
  }

  function togglePopover(kind: FilterPopover): void {
    searchOpen = false;
    activePopover = activePopover === kind ? null : kind;
  }

  function toggleFilter(kind: 'owners' | 'domains' | 'nodeTypes', value: string): void {
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

  function handleNodeClick({ node, event }: { node: Node; event: MouseEvent | TouchEvent }): void {
    activePopover = null;
    searchOpen = false;

    let x = 220;
    let y = 160;
    if ('touches' in event && event.touches.length > 0) {
      x = event.touches[0].clientX;
      y = event.touches[0].clientY;
    } else if ('clientX' in event) {
      x = event.clientX;
      y = event.clientY;
    }

    nodePopup = {
      nodeId: node.id,
      x,
      y,
    };
  }

  function closeTransientUI(): void {
    activePopover = null;
    nodePopup = null;
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
      nodePopup = null;
      activePopover = null;
      loading = false;
    } catch (error) {
      loadError = error instanceof Error ? error.message : String(error);
    } finally {
      input.value = '';
    }
  }

  function matchesFilters(nodeID: string): boolean {
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

    function handleEscape(event: KeyboardEvent): void {
      if (event.key === 'Escape') {
        activePopover = null;
        nodePopup = null;
        searchOpen = false;
      }
    }

    function handleWindowClick(event: MouseEvent): void {
      const target = event.target as HTMLElement | null;
      if (!target) {
        return;
      }

      if (
        target.closest('[data-toolbar-root]') ||
        target.closest('[data-node-popup]') ||
        target.closest('.svelte-flow__node')
      ) {
        return;
      }

      activePopover = null;
      searchOpen = false;
      nodePopup = null;
    }

    window.addEventListener('keydown', handleEscape);
    window.addEventListener('click', handleWindowClick);

    return () => {
      window.removeEventListener('keydown', handleEscape);
      window.removeEventListener('click', handleWindowClick);
    };
  });
</script>

<main class="immersive-shell">
  <SvelteFlow
    nodes={flowNodes}
    edges={flowEdges}
    {nodeTypes}
    fitView
    fitViewOptions={{ padding: 0.08 }}
    minZoom={0.18}
    maxZoom={2.2}
    nodesDraggable
    nodesConnectable={false}
    elementsSelectable
    onnodeclick={handleNodeClick}
    onpaneclick={closeTransientUI}
    attributionPosition="bottom-left"
    class="immersive-flow"
  >
    <Background color="rgba(24, 34, 40, 0.07)" gap={26} />
    <MiniMap position="bottom-left" pannable zoomable />
    <Controls position="bottom-right" />

    <Panel position="top-left" class="top-toolbar-shell">
      <div class="top-toolbar" data-toolbar-root>
        <span class={['toolbar-pill', 'status-pill', connectionTone()].join(' ')}>
          <span class="status-dot"></span>
          {connectionLabel()}
        </span>
        <span class="toolbar-pill soft-pill">{visible.nodes} nodes</span>
        <span class="toolbar-pill soft-pill">{visible.edges} edges</span>
        {#each compactTypePills as nodeType}
          <span class={"toolbar-pill type-pill " + nodeType}>{nodeType} {typeCounts[nodeType] ?? 0}</span>
        {/each}

        <div class="toolbar-spacer"></div>

        <div class="toolbar-search">
          <button
            type="button"
            class={['toolbar-pill', 'toolbar-action', searchOpen ? 'active' : ''].join(' ')}
            onclick={() => {
              searchOpen = !searchOpen;
              activePopover = null;
            }}
          >
            Search
          </button>
          {#if searchOpen}
            <div class="toolbar-popover search-popover" data-toolbar-root>
              <input bind:value={filters.query} type="search" placeholder="Search id, name, domain, owner, file" />
              {#if filters.query}
                <button type="button" class="mini-action" onclick={() => (filters = { ...filters, query: '' })}>Clear</button>
              {/if}
            </div>
          {/if}
        </div>

        <div class="toolbar-filter">
          <button type="button" class={['toolbar-pill', 'toolbar-action', activePopover === 'owners' ? 'active' : ''].join(' ')} onclick={() => togglePopover('owners')}>
            Teams
          </button>
          {#if activePopover === 'owners'}
            <div class="toolbar-popover" data-toolbar-root>
              <div class="popover-head">
                <strong>Teams</strong>
                <button type="button" class="mini-action" onclick={() => clearFilter('owners')}>Reset</button>
              </div>
              <div class="chip-list">
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
            </div>
          {/if}
        </div>

        <div class="toolbar-filter">
          <button type="button" class={['toolbar-pill', 'toolbar-action', activePopover === 'domains' ? 'active' : ''].join(' ')} onclick={() => togglePopover('domains')}>
            Domains
          </button>
          {#if activePopover === 'domains'}
            <div class="toolbar-popover" data-toolbar-root>
              <div class="popover-head">
                <strong>Domains</strong>
                <button type="button" class="mini-action" onclick={() => clearFilter('domains')}>Reset</button>
              </div>
              <div class="chip-list">
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
            </div>
          {/if}
        </div>

        <div class="toolbar-filter">
          <button type="button" class={['toolbar-pill', 'toolbar-action', activePopover === 'nodeTypes' ? 'active' : ''].join(' ')} onclick={() => togglePopover('nodeTypes')}>
            Types
          </button>
          {#if activePopover === 'nodeTypes'}
            <div class="toolbar-popover" data-toolbar-root>
              <div class="popover-head">
                <strong>Types</strong>
                <button type="button" class="mini-action" onclick={() => clearFilter('nodeTypes')}>Reset</button>
              </div>
              <div class="chip-list">
                {#each model.nodeTypes as nodeType}
                  <button
                    type="button"
                    class={['filter-chip', 'kind-chip', nodeType, filters.nodeTypes.includes(nodeType) ? 'active' : ''].join(' ')}
                    onclick={() => toggleFilter('nodeTypes', nodeType)}
                  >
                    <span>{nodeType}</span>
                    <small>{typeCounts[nodeType] ?? 0}</small>
                  </button>
                {/each}
              </div>
            </div>
          {/if}
        </div>

        <button type="button" class="toolbar-pill toolbar-action" onclick={() => fileInput?.click()}>
          Load JSON
        </button>
        <input bind:this={fileInput} class="file-input" type="file" accept="application/json,.json" onchange={handleFileChange} />
      </div>
    </Panel>

    {#if nodePopup && popupNode}
      <Panel position="top-left" class="node-popup-shell">
        <article
          class="node-popup"
          data-node-popup
          style={`left:${Math.max(20, Math.min(nodePopup.x + 14, window.innerWidth - 304))}px;top:${Math.max(72, Math.min(nodePopup.y + 14, window.innerHeight - 240))}px;`}
        >
          <div class="node-popup__head">
            <strong>{popupNode.name}</strong>
            <span class={"toolbar-pill type-pill " + popupNode.type}>{popupNode.type}</span>
          </div>
          <div class="popup-meta">
            <span class="meta-label">Domain</span>
            <span>{popupNode.domain ? domainName(model, popupNode.domain) : 'n/a'}</span>
          </div>
          <div class="popup-meta">
            <span class="meta-label">Owner</span>
            <span>{popupNode.owner ? teamName(model, popupNode.owner) : 'n/a'}</span>
          </div>
          <div class="popup-meta">
            <span class="meta-label">Source</span>
            <span class="mono">
              {#if popupNode.file}
                {popupNode.file}{popupNode.line ? `:${popupNode.line}` : ''}
              {:else}
                n/a
              {/if}
            </span>
          </div>
          {#if popupNode.summary}
            <p class="popup-summary">{popupNode.summary}</p>
          {/if}
        </article>
      </Panel>
    {/if}
  </SvelteFlow>
</main>
