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
  import { loadGraphFromApi } from './lib/api';
  import {
    domainName,
    findNode,
    nodeColor,
    normalizeGraph,
    severitySummary,
    teamName,
    toSvelteFlowEdges,
    toSvelteFlowNodes,
    visibleStats,
  } from './lib/adapter';
  import { resolvePositions } from './lib/layout';
  import FlowNode from './lib/FlowNode.svelte';
  import type { Filters, GraphModel, WindowWithPayload } from './lib/types';

  type PopoverKind = 'search' | 'owners' | 'domains' | 'nodeTypes' | 'layout' | null;
  type LayoutMode = 'freeform' | 'clustered';
  type NodePopup = {
    nodeId: string;
    x: number;
    y: number;
  } | null;
  type SavedPositions = Record<string, { x: number; y: number }>;
  type ActiveFilterBadge = {
    kind: 'query' | 'owners' | 'domains' | 'nodeTypes';
    value: string;
    label: string;
  };

  const GITHUB_URL = 'https://github.com/mandotpro/mapture.dev';
  const STORAGE_PREFIX = 'mapture-layout';
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
    ui: {
      nodeColors: {
        service: '#1664d9',
        api: '#0f8f78',
        database: '#a56614',
        event: '#a73f7f',
      },
    },
    projectId: '',
    sourceLabel: 'offline',
    mode: 'offline',
    summary: {
      errors: 0,
      warnings: 0,
      nodes: 0,
      edges: 0,
    },
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
  let activePopover = $state<PopoverKind>(null);
  let nodePopup = $state<NodePopup>(null);
  let layoutMode = $state<LayoutMode>('freeform');
  let savedPositions = $state.raw<SavedPositions>({});
  let lastStorageKey = '';
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
  const activeFilterBadges = $derived(buildActiveFilterBadges(model, filters));
  const graphFingerprintKey = $derived(graphFingerprint(model));
  const projectIdentity = $derived(model.projectId || sourceLabel || 'default');
  const storageKey = $derived(`${STORAGE_PREFIX}:${projectIdentity}:${graphFingerprintKey}:${layoutMode}`);
  const paletteStyle = $derived(buildPaletteStyle(model));

  $effect(() => {
    const nodeIDs = model.nodes.map((node) => node.id);
    if (!storageKey) {
      return;
    }

    if (storageKey !== lastStorageKey) {
      lastStorageKey = storageKey;
      savedPositions = pruneSavedPositions(readSavedPositions(storageKey), nodeIDs);
      return;
    }

    const pruned = pruneSavedPositions(savedPositions, nodeIDs);
    if (!samePositions(savedPositions, pruned)) {
      savedPositions = pruned;
      persistSavedPositions(storageKey, pruned);
    }
  });

  $effect(() => {
    flowNodes = toSvelteFlowNodes(model, filters, nodePopup?.nodeId ?? null, layoutMode, savedPositions);
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
      if (injected) {
        model = normalizeGraph(injected);
        sourceLabel = injected.meta.sourceLabel;
      } else {
        const payload = await loadGraphFromApi();
        model = normalizeGraph(payload);
        sourceLabel = payload.meta.sourceLabel;
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
        model = normalizeGraph(payload);
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
      return 'Load failed';
    }
    if (loading) {
      return 'Loading';
    }
    if (sourceLabel.startsWith('file:') || sourceLabel === 'static build') {
      return 'Offline';
    }
    if (live) {
      return 'API connected';
    }
    return 'Offline';
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

  function togglePopover(kind: PopoverKind): void {
    activePopover = activePopover === kind ? null : kind;
    nodePopup = null;
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

  function removeBadge(badge: ActiveFilterBadge): void {
    if (badge.kind === 'query') {
      filters = {
        ...filters,
        query: '',
      };
      return;
    }

    if (badge.kind === 'owners') {
      filters = {
        ...filters,
        owners: filters.owners.filter((owner) => owner !== badge.value),
      };
      return;
    }

    if (badge.kind === 'domains') {
      filters = {
        ...filters,
        domains: filters.domains.filter((domain) => domain !== badge.value),
      };
      return;
    }

    filters = {
      ...filters,
      nodeTypes: filters.nodeTypes.filter((nodeType) => nodeType !== badge.value),
    };
  }

  function setLayoutMode(mode: LayoutMode): void {
    layoutMode = mode;
    nodePopup = null;
    activePopover = null;
  }

  function resetLayout(): void {
    savedPositions = {};
    clearSavedPositions(storageKey);
    nodePopup = null;
    activePopover = null;
  }

  function handleNodeClick({ node, event }: { node: Node; event: MouseEvent | TouchEvent }): void {
    activePopover = null;

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

  function handleNodeDragStop({ nodes }: { nodes: Node[] }): void {
    const locked = new Set(nodes.map((node) => node.id));
    const next = {
      ...savedPositions,
      ...resolvePositions(flowNodes, layoutMode, locked),
    };
    savedPositions = next;
    persistSavedPositions(storageKey, next);
  }

  function closeTransientUI(): void {
    activePopover = null;
    nodePopup = null;
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

  function buildActiveFilterBadges(currentModel: GraphModel, currentFilters: Filters): ActiveFilterBadge[] {
    const badges: ActiveFilterBadge[] = [];
    if (currentFilters.query) {
      badges.push({
        kind: 'query',
        value: currentFilters.query,
        label: `Search: ${currentFilters.query}`,
      });
    }
    for (const owner of currentFilters.owners) {
      badges.push({
        kind: 'owners',
        value: owner,
        label: teamName(currentModel, owner),
      });
    }
    for (const domain of currentFilters.domains) {
      badges.push({
        kind: 'domains',
        value: domain,
        label: domainName(currentModel, domain),
      });
    }
    for (const nodeType of currentFilters.nodeTypes) {
      badges.push({
        kind: 'nodeTypes',
        value: nodeType,
        label: nodeType,
      });
    }
    return badges;
  }

  function buildPaletteStyle(currentModel: GraphModel): string {
    return [
      `--service:${currentModel.ui.nodeColors.service}`,
      `--api:${currentModel.ui.nodeColors.api}`,
      `--database:${currentModel.ui.nodeColors.database}`,
      `--event:${currentModel.ui.nodeColors.event}`,
    ].join(';');
  }

  function graphFingerprint(currentModel: GraphModel): string {
    const nodeIDs = currentModel.nodes.map((node) => node.id).sort().join('|');
    const edgeIDs = currentModel.edges.map((edge) => edge.id).sort().join('|');
    return `${nodeIDs}::${edgeIDs}`;
  }

  function readSavedPositions(key: string): SavedPositions {
    try {
      const raw = window.localStorage.getItem(key);
      if (!raw) {
        return {};
      }
      return JSON.parse(raw) as SavedPositions;
    } catch {
      return {};
    }
  }

  function persistSavedPositions(key: string, positions: SavedPositions): void {
    try {
      window.localStorage.setItem(key, JSON.stringify(positions));
    } catch {
      return;
    }
  }

  function clearSavedPositions(key: string): void {
    try {
      window.localStorage.removeItem(key);
    } catch {
      return;
    }
  }

  function pruneSavedPositions(positions: SavedPositions, nodeIDs: string[]): SavedPositions {
    const allowed = new Set(nodeIDs);
    return Object.fromEntries(
      Object.entries(positions).filter(([nodeID]) => allowed.has(nodeID)),
    );
  }

  function samePositions(left: SavedPositions, right: SavedPositions): boolean {
    const leftKeys = Object.keys(left).sort();
    const rightKeys = Object.keys(right).sort();
    if (leftKeys.length !== rightKeys.length) {
      return false;
    }

    for (let index = 0; index < leftKeys.length; index += 1) {
      const key = leftKeys[index];
      if (key !== rightKeys[index]) {
        return false;
      }
      if (left[key].x !== right[key].x || left[key].y !== right[key].y) {
        return false;
      }
    }
    return true;
  }

  function popupStyle(): string {
    if (!nodePopup) {
      return '';
    }
    const left = Math.max(20, Math.min(nodePopup.x + 14, window.innerWidth - 340));
    const top = Math.max(90, Math.min(nodePopup.y + 14, window.innerHeight - 260));
    return `left:${left}px;top:${top}px;`;
  }

  onMount(() => {
    void boot();

    function handleEscape(event: KeyboardEvent): void {
      if (event.key === 'Escape') {
        activePopover = null;
        nodePopup = null;
      }
    }

    function handleWindowClick(event: MouseEvent): void {
      const target = event.target as HTMLElement | null;
      if (!target) {
        return;
      }

      if (
        target.closest('[data-interactive-root]') ||
        target.closest('[data-node-popup]') ||
        target.closest('.svelte-flow__node')
      ) {
        return;
      }

      activePopover = null;
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

<main class="app-shell" style={paletteStyle}>
  <header class="page-header">
    <div class="page-header__brand">
      <span class="wordmark">Mapture</span>
      <span class="header-pill soft-pill">{visible.nodes} nodes</span>
      <span class="header-pill soft-pill">{visible.edges} edges</span>
    </div>

    <div class="page-header__actions">
      <a class="header-pill header-link" href={GITHUB_URL} target="_blank" rel="noreferrer">GitHub</a>
      <span class={['header-pill', 'status-pill', connectionTone()].join(' ')}>
        <span class="status-dot"></span>
        {connectionLabel()}
      </span>
    </div>
  </header>

  <section class="canvas-shell">
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
      onnodedragstop={handleNodeDragStop}
      onpaneclick={closeTransientUI}
      attributionPosition="bottom-left"
      class="immersive-flow"
    >
      <Background color="rgba(24, 34, 40, 0.07)" gap={26} />
      <MiniMap position="bottom-left" pannable zoomable />
      <Controls position="bottom-right" />

      <Panel position="top-left" class="canvas-toolbar-shell">
        <div class="canvas-toolbar" data-interactive-root>
          <div class="canvas-rail">
            <button
              type="button"
              class={['rail-pill', activePopover === 'search' ? 'active' : ''].join(' ')}
              onclick={() => togglePopover('search')}
            >
              Search
            </button>
            <button
              type="button"
              class={['rail-pill', activePopover === 'owners' ? 'active' : ''].join(' ')}
              onclick={() => togglePopover('owners')}
            >
              Teams
            </button>
            <button
              type="button"
              class={['rail-pill', activePopover === 'domains' ? 'active' : ''].join(' ')}
              onclick={() => togglePopover('domains')}
            >
              Domains
            </button>
            <button
              type="button"
              class={['rail-pill', activePopover === 'nodeTypes' ? 'active' : ''].join(' ')}
              onclick={() => togglePopover('nodeTypes')}
            >
              Types
            </button>
            <button
              type="button"
              class={['rail-pill', activePopover === 'layout' ? 'active' : ''].join(' ')}
              onclick={() => togglePopover('layout')}
            >
              Layout
            </button>
          </div>

          {#if activePopover === 'search'}
            <div class="toolbar-popover search-popover" data-interactive-root>
              <input bind:value={filters.query} type="search" placeholder="Search id, name, domain, owner, file" />
              <button type="button" class="mini-action" onclick={() => (filters = { ...filters, query: '' })}>Clear</button>
              <button type="button" class="mini-action" onclick={resetFilters}>Reset filters</button>
            </div>
          {/if}

          {#if activePopover === 'owners'}
            <div class="toolbar-popover" data-interactive-root>
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

          {#if activePopover === 'domains'}
            <div class="toolbar-popover" data-interactive-root>
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

          {#if activePopover === 'nodeTypes'}
            <div class="toolbar-popover" data-interactive-root>
              <div class="popover-head">
                <strong>Types</strong>
                <button type="button" class="mini-action" onclick={() => clearFilter('nodeTypes')}>Reset</button>
              </div>
              <div class="chip-list">
                {#each model.nodeTypes as nodeType}
                  <button
                    type="button"
                    class={['filter-chip', 'kind-chip', filters.nodeTypes.includes(nodeType) ? 'active' : ''].join(' ')}
                    style={`--pill-color:${nodeColor(model, nodeType)};`}
                    onclick={() => toggleFilter('nodeTypes', nodeType)}
                  >
                    <span>{nodeType}</span>
                    <small>{typeCounts[nodeType] ?? 0}</small>
                  </button>
                {/each}
              </div>
            </div>
          {/if}

          {#if activePopover === 'layout'}
            <div class="toolbar-popover" data-interactive-root>
              <div class="popover-head">
                <strong>Layout</strong>
                <button type="button" class="mini-action" onclick={resetLayout}>Reset layout</button>
              </div>
              <div class="chip-list">
                <button
                  type="button"
                  class={['filter-chip', layoutMode === 'freeform' ? 'active' : ''].join(' ')}
                  onclick={() => setLayoutMode('freeform')}
                >
                  <span>Freeform</span>
                </button>
                <button
                  type="button"
                  class={['filter-chip', layoutMode === 'clustered' ? 'active' : ''].join(' ')}
                  onclick={() => setLayoutMode('clustered')}
                >
                  <span>Clustered</span>
                </button>
              </div>
            </div>
          {/if}

          {#if activeFilterBadges.length > 0}
            <div class="active-strip">
              {#each activeFilterBadges as badge}
                <button type="button" class="active-badge" onclick={() => removeBadge(badge)}>
                  <span>{badge.label}</span>
                  <small>x</small>
                </button>
              {/each}
              <button type="button" class="active-reset" onclick={resetFilters}>Reset filters</button>
            </div>
          {/if}
        </div>
      </Panel>

      {#if nodePopup && popupNode}
        <Panel position="top-left" class="node-popup-shell">
          <article class="node-popup" data-node-popup style={popupStyle()}>
            <div class="node-popup__head">
              <strong>{popupNode.name}</strong>
              <span class="type-badge" style={`--pill-color:${nodeColor(model, popupNode.type)};`}>{popupNode.type}</span>
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
              <span class="mono popup-source">
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
  </section>
</main>
