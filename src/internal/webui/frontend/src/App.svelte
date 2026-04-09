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
    buildFlowPresentation,
    domainName,
    findNode,
    nodeColor,
    normalizeGraph,
    severitySummary,
    teamName,
    viewModeFromLayout,
  } from './lib/adapter';
  import { resolvePositions } from './lib/layout';
  import DomainLanesBackdrop from './lib/DomainLanesBackdrop.svelte';
  import FlowViewportController from './lib/FlowViewportController.svelte';
  import FlowNode from './lib/FlowNode.svelte';
  import type {
    DensityMode,
    Filters,
    GraphModel,
    PresentedGraph,
    ViewMode,
    WindowWithPayload,
  } from './lib/types';

  type PopoverKind = 'search' | 'owners' | 'domains' | 'nodeTypes' | null;
  type NodePopup = {
    nodeId: string;
    x: number;
    y: number;
  } | null;
  type ManualPositions = Record<string, { x: number; y: number }>;
  type PersistedLayoutState = {
    version: 1;
    manualPositions: ManualPositions;
  };
  type ActiveFilterBadge = {
    kind: 'query' | 'owners' | 'domains' | 'nodeTypes';
    value: string;
    label: string;
    icon: string;
    tone: string;
  };

  const GITHUB_URL = 'https://github.com/mandotpro/mapture.dev';
  const STORAGE_PREFIX = 'mapture-layout';
  const FIT_VIEW_PADDING = 0.72;

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
      defaultLayout: 'elk-horizontal',
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

  const emptyGraph: PresentedGraph = {
    nodes: [],
    edges: [],
    lanes: [],
  };

  const nodeTypes = {
    architecture: FlowNode,
  } satisfies NodeTypes;

  const viewModeOptions: Array<{
    value: ViewMode;
    label: string;
    summary: string;
    glyph: string;
  }> = [
    { value: 'system-map', label: 'System Map', summary: 'Cleanest overview', glyph: 'SM' },
    { value: 'event-flow', label: 'Event Flow', summary: 'Producer to consumer', glyph: 'EF' },
    { value: 'domain-lanes', label: 'Domain Lanes', summary: 'Boundaries first', glyph: 'DL' },
    { value: 'workbench', label: 'Workbench', summary: 'Manual placement', glyph: 'WB' },
  ];

  const densityOptions: Array<{
    value: DensityMode;
    label: string;
    summary: string;
    glyph: string;
  }> = [
    { value: 'overview', label: 'Overview', summary: 'Low noise', glyph: 'OV' },
    { value: 'standard', label: 'Standard', summary: 'Balanced detail', glyph: 'ST' },
    { value: 'detailed', label: 'Detailed', summary: 'All labels', glyph: 'DT' },
  ];

  const railKinds = ['search', 'owners', 'domains', 'nodeTypes'] as const;

  let model = $state.raw<GraphModel>(emptyModel);
  let presentedGraph = $state.raw<PresentedGraph>(emptyGraph);
  let flowNodes = $state.raw<Node[]>([]);
  let flowEdges = $state.raw<Edge[]>([]);
  let loading = $state(true);
  let live = $state(false);
  let loadError = $state('');
  let sourceLabel = $state('api');
  let activePopover = $state<PopoverKind>(null);
  let nodePopup = $state<NodePopup>(null);
  let viewMode = $state<ViewMode>(viewModeFromLayout(emptyModel.ui.defaultLayout));
  let densityMode = $state<DensityMode>('standard');
  let modeMenuOpen = $state(false);
  let densityMenuOpen = $state(false);
  let hoveredNodeId = $state<string | null>(null);
  let hoveredEdgeId = $state<string | null>(null);
  let toolbarElement = $state<HTMLElement | null>(null);
  let toolbarSize = $state.raw({ width: 420, height: 52 });
  let manualPositions = $state.raw<ManualPositions>({});
  let lastStorageKey = '';
  let refreshVersion = 0;
  let refocusVersion = $state(0);
  let fitViewRequest = $state(0);
  let filters = $state.raw<Filters>({
    query: '',
    nodeTypes: [],
    domains: [],
    owners: [],
  });

  const popupNode = $derived(findNode(model, nodePopup?.nodeId ?? null));
  const activeViewOption = $derived(
    viewModeOptions.find((option) => option.value === viewMode) ?? viewModeOptions[0],
  );
  const activeDensityOption = $derived(
    densityOptions.find((option) => option.value === densityMode) ?? densityOptions[1],
  );
  const visible = $derived({
    nodes: presentedGraph.nodes.length,
    edges: presentedGraph.edges.length,
  });
  const summary = $derived(severitySummary(model.diagnostics));
  const activeFilterBadges = $derived(buildActiveFilterBadges(model, filters));
  const graphFingerprintKey = $derived(graphFingerprint(model));
  const projectIdentity = $derived(model.projectId || sourceLabel || 'default');
  const storageKey = $derived(
    viewMode === 'workbench' ? `${STORAGE_PREFIX}:${projectIdentity}:${graphFingerprintKey}:workbench` : '',
  );
  const paletteStyle = $derived(buildPaletteStyle(model));
  const visibleTypeCounts = $derived(countBy(presentedGraph.nodes, (node) => node.type));
  const visibleOwnerCounts = $derived(countBy(presentedGraph.nodes, (node) => node.owner));
  const visibleDomainCounts = $derived(countBy(presentedGraph.nodes, (node) => node.domain));
  const searchSuggestions = $derived(buildSearchSuggestions(model, filters.query));
  const filterCounts = $derived({
    query: filters.query ? 1 : 0,
    owners: filters.owners.length,
    domains: filters.domains.length,
    nodeTypes: filters.nodeTypes.length,
  });
  const reservedCanvasInsets = $derived({
    top: Math.ceil(toolbarSize.height + 72),
    left: Math.ceil(toolbarSize.width + 72),
  });

  $effect(() => {
    const nodeIDs = presentedGraph.nodes.map((node) => node.id);
    if (!storageKey) {
      lastStorageKey = '';
      if (Object.keys(manualPositions).length > 0) {
        manualPositions = {};
      }
      return;
    }

    if (storageKey !== lastStorageKey) {
      lastStorageKey = storageKey;
      manualPositions = pruneManualPositions(readLayoutState(storageKey), nodeIDs);
      return;
    }

    const pruned = pruneManualPositions(manualPositions, nodeIDs);
    if (!samePositions(manualPositions, pruned)) {
      manualPositions = pruned;
      persistLayoutState(storageKey, pruned);
    }
  });

  $effect(() => {
    const selectedNodeId = nodePopup?.nodeId ?? null;
    void refreshFlowGraph(
      model,
      filters,
      selectedNodeId,
      hoveredNodeId,
      hoveredEdgeId,
      viewMode,
      densityMode,
      manualPositions,
      reservedCanvasInsets,
    );
  });

  $effect(() => {
    const currentRefocusVersion = refocusVersion;
    if (currentRefocusVersion === 0 || flowNodes.length === 0) {
      return;
    }

    fitViewRequest = currentRefocusVersion;
    refocusVersion = 0;
  });

  async function boot(): Promise<void> {
    loading = true;
    loadError = '';

    try {
      const injected = (window as WindowWithPayload).__MAPTURE_DATA__;
      if (injected) {
        model = normalizeGraph(injected);
        viewMode = viewModeFromLayout(model.ui.defaultLayout);
        sourceLabel = injected.meta.sourceLabel;
      } else {
        const payload = await loadGraphFromApi();
        model = normalizeGraph(payload);
        viewMode = viewModeFromLayout(model.ui.defaultLayout);
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
    modeMenuOpen = false;
    densityMenuOpen = false;
  }

  async function refreshFlowGraph(
    currentModel: GraphModel,
    currentFilters: Filters,
    selectedNodeId: string | null,
    currentHoveredNodeId: string | null,
    currentHoveredEdgeId: string | null,
    currentViewMode: ViewMode,
    currentDensityMode: DensityMode,
    currentManualPositions: ManualPositions,
    currentReservedInsets: { top: number; left: number },
  ): Promise<void> {
    const revision = ++refreshVersion;
    const presentation = await buildFlowPresentation(currentModel, currentFilters, {
      viewMode: currentViewMode,
      densityMode: currentDensityMode,
      focus: {
        selectedNodeId,
        hoveredNodeId: currentHoveredNodeId,
        hoveredEdgeId: currentHoveredEdgeId,
      },
      manualPositions: currentManualPositions,
      reservedInsets: currentReservedInsets,
    });

    if (revision !== refreshVersion) {
      return;
    }

    presentedGraph = presentation.graph;
    flowNodes = presentation.flowNodes;
    flowEdges = presentation.flowEdges;

    if (nodePopup && !presentation.graph.nodes.some((node) => node.id === nodePopup.nodeId)) {
      nodePopup = null;
    }
    if (hoveredNodeId && !presentation.graph.nodes.some((node) => node.id === hoveredNodeId)) {
      hoveredNodeId = null;
    }
    if (hoveredEdgeId && !presentation.graph.edges.some((edge) => edge.id === hoveredEdgeId)) {
      hoveredEdgeId = null;
    }
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
    modeMenuOpen = false;
    densityMenuOpen = false;
  }

  function toggleModeMenu(): void {
    modeMenuOpen = !modeMenuOpen;
    densityMenuOpen = false;
    activePopover = null;
    nodePopup = null;
  }

  function toggleDensityMenu(): void {
    densityMenuOpen = !densityMenuOpen;
    modeMenuOpen = false;
    activePopover = null;
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

  function setViewMode(mode: ViewMode): void {
    if (viewMode === mode) {
      modeMenuOpen = false;
      return;
    }

    viewMode = mode;
    hoveredNodeId = null;
    hoveredEdgeId = null;
    nodePopup = null;
    activePopover = null;
    modeMenuOpen = false;
    densityMenuOpen = false;
    refocusVersion += 1;
  }

  function setDensityMode(mode: DensityMode): void {
    if (densityMode === mode) {
      densityMenuOpen = false;
      return;
    }

    densityMode = mode;
    densityMenuOpen = false;
  }

  function refitCanvas(): void {
    refocusVersion += 1;
    modeMenuOpen = false;
    densityMenuOpen = false;
  }

  function resetLayout(): void {
    manualPositions = {};
    if (storageKey) {
      clearLayoutState(storageKey);
    }
    nodePopup = null;
    activePopover = null;
    modeMenuOpen = false;
    densityMenuOpen = false;
    refocusVersion += 1;
  }

  function handleNodeClick({ node, event }: { node: Node; event: MouseEvent | TouchEvent }): void {
    activePopover = null;
    modeMenuOpen = false;
    densityMenuOpen = false;

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

  function handleNodePointerEnter({ node }: { node: Node }): void {
    hoveredNodeId = node.id;
  }

  function handleNodePointerLeave({ node }: { node: Node }): void {
    if (hoveredNodeId === node.id) {
      hoveredNodeId = null;
    }
  }

  function handleEdgePointerEnter({ edge }: { edge: Edge }): void {
    hoveredEdgeId = edge.id;
  }

  function handleEdgePointerLeave({ edge }: { edge: Edge }): void {
    if (hoveredEdgeId === edge.id) {
      hoveredEdgeId = null;
    }
  }

  function handleNodeDragStop({ nodes }: { nodes: Node[] }): void {
    if (viewMode !== 'workbench') {
      return;
    }

    const draggedPositions = new Map(nodes.map((node) => [node.id, node.position]));
    const nextManualPositions = {
      ...manualPositions,
      ...Object.fromEntries(
        nodes.map((node) => [
          node.id,
          {
            x: node.position.x,
            y: node.position.y,
          },
        ]),
      ),
    };
    const merged = flowNodes.map((node) => {
      const position = draggedPositions.get(node.id);
      return position ? { ...node, position } : node;
    });
    const resolved = resolvePositions(merged, viewMode, {
      lockedNodeIds: new Set(Object.keys(nextManualPositions)),
      priorityNodeIds: new Set(nodes.map((node) => node.id)),
      reservedInsets: reservedCanvasInsets,
    });
    const nextFlowNodes = merged.map((node) => ({
      ...node,
      position: resolved[node.id] ?? node.position,
    }));
    const next = {
      ...manualPositions,
      ...Object.fromEntries(
        nodes.map((node) => {
          const position = resolved[node.id] ?? node.position;
          return [
            node.id,
            {
              x: position.x,
              y: position.y,
            },
          ];
        }),
      ),
    };

    flowNodes = nextFlowNodes;
    manualPositions = next;
    if (storageKey) {
      persistLayoutState(storageKey, next);
    }
    nodePopup = null;
  }

  function closeTransientUI(): void {
    activePopover = null;
    nodePopup = null;
    modeMenuOpen = false;
    densityMenuOpen = false;
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
        icon: iconForKind('query'),
        tone: accentForKind(currentModel, 'query'),
      });
    }
    for (const owner of currentFilters.owners) {
      badges.push({
        kind: 'owners',
        value: owner,
        label: teamName(currentModel, owner),
        icon: iconForKind('owners'),
        tone: accentForKind(currentModel, 'owners'),
      });
    }
    for (const domain of currentFilters.domains) {
      badges.push({
        kind: 'domains',
        value: domain,
        label: domainName(currentModel, domain),
        icon: iconForKind('domains'),
        tone: accentForKind(currentModel, 'domains'),
      });
    }
    for (const nodeType of currentFilters.nodeTypes) {
      badges.push({
        kind: 'nodeTypes',
        value: nodeType,
        label: capitalize(nodeType),
        icon: iconForKind('nodeTypes', nodeType),
        tone: accentForKind(currentModel, 'nodeTypes', nodeType),
      });
    }
    return badges;
  }

  function railButtonLabel(kind: Exclude<PopoverKind, null>): string {
    const labels: Record<Exclude<PopoverKind, null>, string> = {
      search: 'Search',
      owners: 'Teams',
      domains: 'Domains',
      nodeTypes: 'Types',
    };
    return labels[kind];
  }

  function iconForKind(
    kind: ActiveFilterBadge['kind'] | Exclude<PopoverKind, null>,
    value?: string,
  ): string {
    if (kind === 'query' || kind === 'search') {
      return 'Q';
    }
    if (kind === 'owners') {
      return 'T';
    }
    if (kind === 'domains') {
      return 'D';
    }
    const nodeTypeIcons: Record<string, string> = {
      service: 'S',
      api: 'A',
      database: 'DB',
      event: 'E',
    };
    return nodeTypeIcons[value ?? ''] ?? 'N';
  }

  function accentForKind(
    currentModel: GraphModel,
    kind: ActiveFilterBadge['kind'] | Exclude<PopoverKind, null>,
    value?: string,
  ): string {
    if (kind === 'query' || kind === 'search') {
      return '#667076';
    }
    if (kind === 'owners') {
      return '#0d7661';
    }
    if (kind === 'domains') {
      return '#1664d9';
    }
    return nodeColor(currentModel, value ?? 'service');
  }

  function chipStyle(
    currentModel: GraphModel,
    kind: ActiveFilterBadge['kind'] | Exclude<PopoverKind, null>,
    value?: string,
  ): string {
    return `--chip-accent:${accentForKind(currentModel, kind, value)};`;
  }

  function capitalize(value: string): string {
    if (!value) {
      return value;
    }
    return `${value[0].toUpperCase()}${value.slice(1)}`;
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

  function readLayoutState(key: string): ManualPositions {
    try {
      const raw = window.localStorage.getItem(key);
      if (!raw) {
        return {};
      }
      const parsed = JSON.parse(raw) as PersistedLayoutState | ManualPositions;
      if (isManualPositions(parsed)) {
        return parsed;
      }
      if (
        parsed &&
        typeof parsed === 'object' &&
        'version' in parsed &&
        parsed.version === 1 &&
        'manualPositions' in parsed &&
        isManualPositions(parsed.manualPositions)
      ) {
        return parsed.manualPositions;
      }
      return {};
    } catch {
      return {};
    }
  }

  function persistLayoutState(key: string, positions: ManualPositions): void {
    try {
      const state: PersistedLayoutState = {
        version: 1,
        manualPositions: positions,
      };
      window.localStorage.setItem(key, JSON.stringify(state));
    } catch {
      return;
    }
  }

  function clearLayoutState(key: string): void {
    try {
      window.localStorage.removeItem(key);
    } catch {
      return;
    }
  }

  function pruneManualPositions(positions: ManualPositions, nodeIDs: string[]): ManualPositions {
    const allowed = new Set(nodeIDs);
    return Object.fromEntries(
      Object.entries(positions).filter(([nodeID]) => allowed.has(nodeID)),
    );
  }

  function samePositions(left: ManualPositions, right: ManualPositions): boolean {
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

  function isManualPositions(value: unknown): value is ManualPositions {
    if (!value || typeof value !== 'object' || Array.isArray(value)) {
      return false;
    }

    return Object.values(value).every((position) => {
      if (!position || typeof position !== 'object' || Array.isArray(position)) {
        return false;
      }

      const point = position as { x?: unknown; y?: unknown };
      return typeof point.x === 'number' && typeof point.y === 'number';
    });
  }

  function popupStyle(): string {
    if (!nodePopup) {
      return '';
    }
    const left = Math.max(20, Math.min(nodePopup.x + 14, window.innerWidth - 340));
    const top = Math.max(90, Math.min(nodePopup.y + 14, window.innerHeight - 260));
    return `left:${left}px;top:${top}px;`;
  }

  function applySearchSuggestion(value: string): void {
    filters = {
      ...filters,
      query: value,
    };
  }

  function buildSearchSuggestions(currentModel: GraphModel, query: string): string[] {
    const values = new Set<string>();
    for (const node of currentModel.nodes) {
      values.add(node.id);
      values.add(node.name);
      if (node.domain) {
        values.add(node.domain);
        values.add(domainName(currentModel, node.domain));
      }
    }

    const normalizedQuery = query.trim().toLowerCase();
    const suggestions = Array.from(values)
      .filter(Boolean)
      .sort((left, right) => left.localeCompare(right));

    if (!normalizedQuery) {
      return suggestions.slice(0, 10);
    }

    return suggestions
      .filter((value) => value.toLowerCase().includes(normalizedQuery))
      .slice(0, 8);
  }

  onMount(() => {
    void boot();

    function handleEscape(event: KeyboardEvent): void {
      if (event.key === 'Escape') {
        activePopover = null;
        nodePopup = null;
        modeMenuOpen = false;
        densityMenuOpen = false;
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
        target.closest('.svelte-flow__node') ||
        target.closest('.svelte-flow__edge')
      ) {
        return;
      }

      activePopover = null;
      nodePopup = null;
      modeMenuOpen = false;
      densityMenuOpen = false;
    }

    window.addEventListener('keydown', handleEscape);
    window.addEventListener('click', handleWindowClick);

    return () => {
      window.removeEventListener('keydown', handleEscape);
      window.removeEventListener('click', handleWindowClick);
    };
  });

  $effect(() => {
    const element = toolbarElement;
    if (!element || typeof ResizeObserver === 'undefined') {
      return;
    }

    const observer = new ResizeObserver((entries) => {
      const entry = entries[0];
      if (!entry) {
        return;
      }

      toolbarSize = {
        width: entry.contentRect.width,
        height: entry.contentRect.height,
      };
    });
    observer.observe(element);

    return () => {
      observer.disconnect();
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
      fitViewOptions={{ padding: FIT_VIEW_PADDING }}
      minZoom={0.18}
      maxZoom={2.2}
      nodesDraggable={viewMode === 'workbench'}
      nodesConnectable={false}
      elementsSelectable
      onnodeclick={handleNodeClick}
      onnodepointerenter={handleNodePointerEnter}
      onnodepointerleave={handleNodePointerLeave}
      onedgepointerenter={handleEdgePointerEnter}
      onedgepointerleave={handleEdgePointerLeave}
      onnodedragstop={handleNodeDragStop}
      onpaneclick={closeTransientUI}
      attributionPosition="bottom-left"
      class="immersive-flow"
    >
      <FlowViewportController request={fitViewRequest} padding={FIT_VIEW_PADDING} maxZoom={1.35} />
      <DomainLanesBackdrop lanes={presentedGraph.lanes} />
      <Background color="rgba(24, 34, 40, 0.07)" gap={26} />
      <MiniMap position="bottom-left" pannable zoomable />
      <Controls position="bottom-right" />

      <Panel position="top-left" class="canvas-toolbar-shell">
        <div class="canvas-toolbar" data-interactive-root bind:this={toolbarElement}>
          <div class="canvas-rail">
            {#each railKinds as kind}
              <button
                type="button"
                class={[
                  'rail-pill',
                  `rail-pill--${kind}`,
                  activePopover === kind ? 'active' : '',
                  (
                    (kind === 'search' && filterCounts.query > 0) ||
                    (kind === 'owners' && filterCounts.owners > 0) ||
                    (kind === 'domains' && filterCounts.domains > 0) ||
                    (kind === 'nodeTypes' && filterCounts.nodeTypes > 0)
                  ) ? 'has-value' : '',
                ].join(' ')}
                style={chipStyle(model, kind)}
                onclick={() => togglePopover(kind)}
              >
                <span class="chip-icon" aria-hidden="true">{iconForKind(kind)}</span>
                <span>{railButtonLabel(kind)}</span>
                {#if kind === 'search' && filterCounts.query > 0}
                  <small class="pill-count">{filterCounts.query}</small>
                {:else if kind === 'owners' && filterCounts.owners > 0}
                  <small class="pill-count">{filterCounts.owners}</small>
                {:else if kind === 'domains' && filterCounts.domains > 0}
                  <small class="pill-count">{filterCounts.domains}</small>
                {:else if kind === 'nodeTypes' && filterCounts.nodeTypes > 0}
                  <small class="pill-count">{filterCounts.nodeTypes}</small>
                {/if}
              </button>
            {/each}
          </div>

          {#if activePopover === 'search'}
            <div class="toolbar-popover search-popover" data-interactive-root>
              <input
                bind:value={filters.query}
                type="search"
                list="mapture-search-suggestions"
                autocomplete="off"
                placeholder="Search id, name, or domain"
              />
              <datalist id="mapture-search-suggestions">
                {#each searchSuggestions as suggestion}
                  <option value={suggestion}></option>
                {/each}
              </datalist>
              {#if searchSuggestions.length > 0}
                <div class="suggestion-strip">
                  {#each searchSuggestions as suggestion}
                    <button type="button" class="suggestion-chip" onclick={() => applySearchSuggestion(suggestion)}>
                      {suggestion}
                    </button>
                  {/each}
                </div>
              {/if}
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
                    class={['filter-chip', 'filter-chip--owner', filters.owners.includes(owner) ? 'active' : ''].join(' ')}
                    style={chipStyle(model, 'owners')}
                    onclick={() => toggleFilter('owners', owner)}
                  >
                    <span class="chip-icon" aria-hidden="true">{iconForKind('owners')}</span>
                    <span class="chip-label">{teamName(model, owner)}</span>
                    <small>{visibleOwnerCounts[owner] ?? 0}</small>
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
                    class={['filter-chip', 'filter-chip--domain', filters.domains.includes(domain) ? 'active' : ''].join(' ')}
                    style={chipStyle(model, 'domains')}
                    onclick={() => toggleFilter('domains', domain)}
                  >
                    <span class="chip-icon" aria-hidden="true">{iconForKind('domains')}</span>
                    <span class="chip-label">{domainName(model, domain)}</span>
                    <small>{visibleDomainCounts[domain] ?? 0}</small>
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
                    class={['filter-chip', 'filter-chip--node-type', 'kind-chip', filters.nodeTypes.includes(nodeType) ? 'active' : ''].join(' ')}
                    style={`${chipStyle(model, 'nodeTypes', nodeType)}--pill-color:${nodeColor(model, nodeType)};`}
                    onclick={() => toggleFilter('nodeTypes', nodeType)}
                  >
                    <span class="chip-icon" aria-hidden="true">{iconForKind('nodeTypes', nodeType)}</span>
                    <span class="chip-label">{capitalize(nodeType)}</span>
                    <small>{visibleTypeCounts[nodeType] ?? 0}</small>
                  </button>
                {/each}
              </div>
            </div>
          {/if}

          {#if activeFilterBadges.length > 0}
            <div class="active-strip">
              {#each activeFilterBadges as badge}
                <button
                  type="button"
                  class={['active-badge', `active-badge--${badge.kind}`].join(' ')}
                  style={`--chip-accent:${badge.tone};`}
                  onclick={() => removeBadge(badge)}
                >
                  <span class="active-badge__meta">
                    <span class="chip-icon" aria-hidden="true">{badge.icon}</span>
                  </span>
                  <span class="active-badge__label">{badge.label}</span>
                  <small aria-hidden="true">x</small>
                </button>
              {/each}
              <button type="button" class="active-reset" onclick={resetFilters}>Reset filters</button>
            </div>
          {/if}
        </div>
      </Panel>

      <Panel position="top-right" class="canvas-control-shell">
        <div class="control-stack" data-interactive-root>
          <div class="control-picker">
            <button
              type="button"
              class={['control-trigger', modeMenuOpen ? 'active' : ''].join(' ')}
              onclick={toggleModeMenu}
            >
              <span class="control-trigger__icon" aria-hidden="true">{activeViewOption.glyph}</span>
              <span class="control-trigger__copy">
                <strong>{activeViewOption.label}</strong>
                <small>{activeViewOption.summary}</small>
              </span>
              <span class="control-trigger__caret" aria-hidden="true">{modeMenuOpen ? 'x' : 'v'}</span>
            </button>

            {#if modeMenuOpen}
              <div class="control-menu">
                <div class="control-menu__head">
                  <strong>View</strong>
                  <button
                    type="button"
                    class="mini-action"
                    onclick={viewMode === 'workbench' ? resetLayout : refitCanvas}
                  >
                    {viewMode === 'workbench' ? 'Reset' : 'Refit'}
                  </button>
                </div>
                {#each viewModeOptions as option}
                  <button
                    type="button"
                    class={['control-option', viewMode === option.value ? 'active' : ''].join(' ')}
                    onclick={() => setViewMode(option.value)}
                  >
                    <span class="control-option__icon" aria-hidden="true">{option.glyph}</span>
                    <span class="control-option__copy">
                      <strong>{option.label}</strong>
                      <small>{option.summary}</small>
                    </span>
                  </button>
                {/each}
              </div>
            {/if}
          </div>

          <div class="control-picker control-picker--density">
            <button
              type="button"
              class={['control-trigger', 'control-trigger--density', densityMenuOpen ? 'active' : ''].join(' ')}
              onclick={toggleDensityMenu}
            >
              <span class="control-trigger__icon" aria-hidden="true">{activeDensityOption.glyph}</span>
              <span class="control-trigger__copy">
                <strong>{activeDensityOption.label}</strong>
                <small>{activeDensityOption.summary}</small>
              </span>
              <span class="control-trigger__caret" aria-hidden="true">{densityMenuOpen ? 'x' : 'v'}</span>
            </button>

            {#if densityMenuOpen}
              <div class="control-menu control-menu--density">
                <div class="control-menu__head">
                  <strong>Density</strong>
                </div>
                {#each densityOptions as option}
                  <button
                    type="button"
                    class={['control-option', densityMode === option.value ? 'active' : ''].join(' ')}
                    onclick={() => setDensityMode(option.value)}
                  >
                    <span class="control-option__icon" aria-hidden="true">{option.glyph}</span>
                    <span class="control-option__copy">
                      <strong>{option.label}</strong>
                      <small>{option.summary}</small>
                    </span>
                  </button>
                {/each}
              </div>
            {/if}
          </div>
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
