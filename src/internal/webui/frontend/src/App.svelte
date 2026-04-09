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
    nodeColor,
    normalizeGraph,
    severitySummary,
    teamName,
    visibleNodesForFilters,
    viewModeFromLayout,
  } from './lib/adapter';
  import { resolvePositions } from './lib/layout';
  import DomainLanesBackdrop from './lib/DomainLanesBackdrop.svelte';
  import FlowViewportController from './lib/FlowViewportController.svelte';
  import ApiNode from './lib/nodes/ApiNode.svelte';
  import BridgeNode from './lib/nodes/BridgeNode.svelte';
  import DatabaseNode from './lib/nodes/DatabaseNode.svelte';
  import EventNode from './lib/nodes/EventNode.svelte';
  import GroupNode from './lib/nodes/GroupNode.svelte';
  import ServiceNode from './lib/nodes/ServiceNode.svelte';
  import CanvasModal from './lib/ui/CanvasModal.svelte';
  import ImpactPreviewPanel from './lib/ui/ImpactPreviewPanel.svelte';
  import NodeInspector from './lib/ui/NodeInspector.svelte';
  import SettingsField from './lib/ui/SettingsField.svelte';
  import SettingsSection from './lib/ui/SettingsSection.svelte';
  import TokenBadge from './lib/ui/TokenBadge.svelte';
  import type {
    ExplorerSettings,
    DensityMode,
    Filters,
    GraphModel,
    ImpactPreview,
    NodeInspectorAction,
    PresentedGraph,
    PresentedNode,
    ResolvedTheme,
    SettingsSectionConfig,
    ThemePreference,
    TraceSelection,
    ViewMode,
    WindowWithPayload,
  } from './lib/types';

  type PopoverKind = 'search' | 'trace' | 'structure' | 'owners' | 'domains' | 'nodeTypes' | null;
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
  type TraceDraft = {
    sourceQuery: string;
    targetQuery: string;
  };

  const GITHUB_URL = 'https://github.com/mandotpro/mapture.dev';
  const STORAGE_PREFIX = 'mapture-layout';
  const SETTINGS_STORAGE_KEY = 'mapture-explorer-settings';
  const FIT_VIEW_PADDING = 0.72;
  const defaultExplorerSettings: ExplorerSettings = {
    version: 2,
    appearance: {
      themePreference: 'system',
    },
    experimental: {
      traceTools: false,
      structureTools: false,
      impactPreview: false,
    },
  };

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
    trace: {
      active: false,
      sourceId: null,
      targetId: null,
      found: false,
      directed: true,
      nodeIDs: [],
      edgeIDs: [],
    },
  };

  const nodeTypes = {
    service: ServiceNode,
    api: ApiNode,
    database: DatabaseNode,
    event: EventNode,
    group: GroupNode,
    bridge: BridgeNode,
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

  let model = $state.raw<GraphModel>(emptyModel);
  let presentedGraph = $state.raw<PresentedGraph>(emptyGraph);
  let flowNodes = $state.raw<Node[]>([]);
  let flowEdges = $state.raw<Edge[]>([]);
  let loading = $state(true);
  let live = $state(false);
  let loadError = $state('');
  let sourceLabel = $state('api');
  let activePopover = $state<PopoverKind>(null);
  let selectedNodeId = $state<string | null>(null);
  let viewMode = $state<ViewMode>(viewModeFromLayout(emptyModel.ui.defaultLayout));
  let densityMode = $state<DensityMode>('standard');
  let modeMenuOpen = $state(false);
  let densityMenuOpen = $state(false);
  let settingsOpen = $state(false);
  let hoveredNodeId = $state<string | null>(null);
  let hoveredEdgeId = $state<string | null>(null);
  let toolbarElement = $state<HTMLElement | null>(null);
  let toolbarSize = $state.raw({ width: 420, height: 52 });
  let manualPositions = $state.raw<ManualPositions>({});
  let traceSelection = $state.raw<TraceSelection>({
    sourceNodeId: null,
    targetNodeId: null,
  });
  let traceDraft = $state.raw<TraceDraft>({
    sourceQuery: '',
    targetQuery: '',
  });
  let collapsedDomains = $state.raw<string[]>([]);
  let collapsedOwners = $state.raw<string[]>([]);
  let aggregateCrossDomain = $state(false);
  let explorerSettings = $state.raw<ExplorerSettings>(defaultExplorerSettings);
  let systemPrefersDark = $state(false);
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

  const popupNode = $derived(
    presentedGraph.nodes.find((node) => node.id === (selectedNodeId ?? '')) ?? null,
  );
  const baseVisibleNodes = $derived(visibleNodesForFilters(model, filters));
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
  const traceSourceSuggestions = $derived(buildNodeSuggestions(presentedGraph.nodes, traceDraft.sourceQuery, traceSelection.targetNodeId));
  const traceTargetSuggestions = $derived(buildNodeSuggestions(presentedGraph.nodes, traceDraft.targetQuery, traceSelection.sourceNodeId));
  const structureDomainCounts = $derived(countBy(baseVisibleNodes, (node) => node.domain));
  const structureOwnerCounts = $derived(countBy(baseVisibleNodes, (node) => node.owner));
  const traceStatus = $derived(buildTraceStatus(presentedGraph, traceSelection));
  const popupImpact = $derived(buildImpactPreview(presentedGraph, popupNode?.id ?? null));
  const resolvedTheme = $derived<ResolvedTheme>(
    explorerSettings.appearance.themePreference === 'system'
      ? (systemPrefersDark ? 'dark' : 'light')
      : explorerSettings.appearance.themePreference,
  );
  const settingsSections = $derived(buildSettingsSections(explorerSettings));
  const nodeInspectorActions = $derived(buildNodeInspectorActions());
  const filterCounts = $derived({
    query: filters.query ? 1 : 0,
    trace: explorerSettings.experimental.traceTools && (traceSelection.sourceNodeId || traceSelection.targetNodeId)
      ? (traceSelection.sourceNodeId ? 1 : 0) + (traceSelection.targetNodeId ? 1 : 0)
      : 0,
    structure: explorerSettings.experimental.structureTools
      ? collapsedDomains.length + collapsedOwners.length + (aggregateCrossDomain ? 1 : 0)
      : 0,
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
    void refreshFlowGraph(
      model,
      filters,
      selectedNodeId,
      hoveredNodeId,
      hoveredEdgeId,
      viewMode,
      densityMode,
      traceSelection,
      collapsedDomains,
      collapsedOwners,
      aggregateCrossDomain,
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

  $effect(() => {
    if (explorerSettings.experimental.traceTools) {
      return;
    }

    if (activePopover === 'trace') {
      activePopover = null;
    }
    if (traceSelection.sourceNodeId || traceSelection.targetNodeId) {
      traceSelection = {
        sourceNodeId: null,
        targetNodeId: null,
      };
    }
    if (traceDraft.sourceQuery || traceDraft.targetQuery) {
      traceDraft = {
        sourceQuery: '',
        targetQuery: '',
      };
    }
  });

  $effect(() => {
    if (explorerSettings.experimental.structureTools) {
      return;
    }

    if (activePopover === 'structure') {
      activePopover = null;
    }
    if (collapsedDomains.length > 0) {
      collapsedDomains = [];
    }
    if (collapsedOwners.length > 0) {
      collapsedOwners = [];
    }
    if (aggregateCrossDomain) {
      aggregateCrossDomain = false;
    }
  });

  $effect(() => {
    if (typeof document === 'undefined') {
      return;
    }

    document.documentElement.dataset.theme = resolvedTheme;
    document.documentElement.style.colorScheme = resolvedTheme;
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

  function readExplorerSettings(): ExplorerSettings {
    if (typeof window === 'undefined') {
      return defaultExplorerSettings;
    }

    try {
      const raw = window.localStorage.getItem(SETTINGS_STORAGE_KEY);
      if (!raw) {
        return defaultExplorerSettings;
      }
      const parsed = JSON.parse(raw) as Partial<ExplorerSettings> & {
        appearance?: { themePreference?: ThemePreference };
      };
      return {
        version: 2,
        appearance: {
          themePreference: isThemePreference(parsed?.appearance?.themePreference)
            ? parsed.appearance.themePreference
            : 'system',
        },
        experimental: {
          traceTools: parsed?.experimental?.traceTools === true,
          structureTools: parsed?.experimental?.structureTools === true,
          impactPreview: parsed?.experimental?.impactPreview === true,
        },
      };
    } catch {
      return defaultExplorerSettings;
    }
  }

  function persistExplorerSettings(next: ExplorerSettings): void {
    if (typeof window === 'undefined') {
      return;
    }
    try {
      window.localStorage.setItem(SETTINGS_STORAGE_KEY, JSON.stringify(next));
    } catch {
      return;
    }
  }

  function updateExplorerSettings(next: ExplorerSettings): void {
    explorerSettings = next;
    persistExplorerSettings(next);
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
    selectedNodeId = null;
    modeMenuOpen = false;
    densityMenuOpen = false;
    settingsOpen = false;
  }

  async function refreshFlowGraph(
    currentModel: GraphModel,
    currentFilters: Filters,
    selectedNodeId: string | null,
    currentHoveredNodeId: string | null,
    currentHoveredEdgeId: string | null,
    currentViewMode: ViewMode,
    currentDensityMode: DensityMode,
    currentTraceSelection: TraceSelection,
    currentCollapsedDomains: string[],
    currentCollapsedOwners: string[],
    currentAggregateCrossDomain: boolean,
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
      trace: currentTraceSelection,
      collapsedDomains: new Set(currentCollapsedDomains),
      collapsedOwners: new Set(currentCollapsedOwners),
      aggregateCrossDomain: currentAggregateCrossDomain,
      manualPositions: currentManualPositions,
      reservedInsets: currentReservedInsets,
    });

    if (revision !== refreshVersion) {
      return;
    }

    presentedGraph = presentation.graph;
    flowNodes = presentation.flowNodes;
    flowEdges = presentation.flowEdges;

    if (selectedNodeId && !presentation.graph.nodes.some((node) => node.id === selectedNodeId)) {
      selectedNodeId = null;
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

  function visibleRailKinds(): Array<Exclude<PopoverKind, null>> {
    const kinds: Array<Exclude<PopoverKind, null>> = ['search'];
    if (explorerSettings.experimental.traceTools) {
      kinds.push('trace');
    }
    if (explorerSettings.experimental.structureTools) {
      kinds.push('structure');
    }
    kinds.push('owners', 'domains', 'nodeTypes');
    return kinds;
  }

  function togglePopover(kind: PopoverKind): void {
    activePopover = activePopover === kind ? null : kind;
    selectedNodeId = null;
    modeMenuOpen = false;
    densityMenuOpen = false;
    settingsOpen = false;
  }

  function toggleSettingsPanel(): void {
    settingsOpen = !settingsOpen;
    activePopover = null;
    selectedNodeId = null;
    modeMenuOpen = false;
    densityMenuOpen = false;
  }

  function handleSettingsFieldChange(id: string, value: boolean | string): void {
    if (id === 'themePreference' && typeof value === 'string' && isThemePreference(value)) {
      updateExplorerSettings({
        ...explorerSettings,
        appearance: {
          ...explorerSettings.appearance,
          themePreference: value,
        },
      });
      return;
    }

    if (typeof value !== 'boolean') {
      return;
    }

    if (id === 'traceTools' || id === 'structureTools' || id === 'impactPreview') {
      updateExplorerSettings({
        ...explorerSettings,
        experimental: {
          ...explorerSettings.experimental,
          [id]: value,
        },
      });
    }
  }

  function buildSettingsSections(settings: ExplorerSettings): SettingsSectionConfig[] {
    return [
      {
        id: 'appearance',
        title: 'Appearance',
        description: 'Control the explorer theme. System follows the OS preference by default.',
        fields: [
          {
            id: 'themePreference',
            kind: 'choice',
            label: 'Theme',
            description: 'Switch between system, light, and dark appearance.',
            value: settings.appearance.themePreference,
            options: [
              { value: 'system', label: 'System', glyph: 'OS' },
              { value: 'light', label: 'Light', glyph: 'LT' },
              { value: 'dark', label: 'Dark', glyph: 'DK' },
            ],
          },
        ],
      },
      {
        id: 'experimental',
        title: 'Experimental',
        description: 'Hidden tools that need more iteration before they become default.',
        fields: [
          {
            id: 'traceTools',
            kind: 'toggle',
            label: 'Trace tools',
            description: 'Path tracing workflows and quick trace actions.',
            value: settings.experimental.traceTools,
            badge: 'FT',
          },
          {
            id: 'structureTools',
            kind: 'toggle',
            label: 'Structure tools',
            description: 'Collapse domains and teams, plus cross-domain aggregation.',
            value: settings.experimental.structureTools,
            badge: 'FT',
          },
          {
            id: 'impactPreview',
            kind: 'toggle',
            label: 'Impact preview',
            description: 'Shows upstream and downstream reach for the selected node.',
            value: settings.experimental.impactPreview,
            badge: 'FT',
          },
        ],
      },
    ];
  }

  function buildNodeInspectorActions(): NodeInspectorAction[] {
    if (!popupNode) {
      return [];
    }

    const actions: NodeInspectorAction[] = [];
    if (explorerSettings.experimental.traceTools) {
      actions.push(
        { id: 'trace-source', label: 'Trace from here', tone: 'accent', badge: 'FT' },
        { id: 'trace-target', label: 'Trace to here', tone: 'accent', badge: 'FT' },
      );
    }
    if (explorerSettings.experimental.structureTools && popupNode.domain) {
      actions.push({
        id: 'toggle-domain',
        label: collapsedDomains.includes(popupNode.domain) ? 'Expand domain' : 'Collapse domain',
        badge: 'FT',
      });
    }
    if (explorerSettings.experimental.structureTools && popupNode.owner) {
      actions.push({
        id: 'toggle-owner',
        label: collapsedOwners.includes(popupNode.owner) ? 'Expand team' : 'Collapse team',
        badge: 'FT',
      });
    }
    return actions;
  }

  function handleNodeInspectorAction(actionId: string): void {
    if (actionId === 'trace-source') {
      usePopupNodeAsTrace('source');
      return;
    }
    if (actionId === 'trace-target') {
      usePopupNodeAsTrace('target');
      return;
    }
    if (actionId === 'toggle-domain') {
      togglePopupDomainCollapse();
      return;
    }
    if (actionId === 'toggle-owner') {
      togglePopupOwnerCollapse();
    }
  }

  function resetStructure(): void {
    collapsedDomains = [];
    collapsedOwners = [];
    aggregateCrossDomain = false;
  }

  function toggleCollapsedDomain(domain: string): void {
    collapsedDomains = toggleValue(collapsedDomains, domain);
  }

  function toggleCollapsedOwner(owner: string): void {
    collapsedOwners = toggleValue(collapsedOwners, owner);
  }

  function toggleCrossDomainAggregation(): void {
    aggregateCrossDomain = !aggregateCrossDomain;
  }

  function assignTraceEndpoint(slot: 'source' | 'target', node: PresentedNode): void {
    if (slot === 'source') {
      traceSelection = {
        ...traceSelection,
        sourceNodeId: node.id,
      };
      traceDraft = {
        ...traceDraft,
        sourceQuery: node.id,
      };
      return;
    }

    traceSelection = {
      ...traceSelection,
      targetNodeId: node.id,
    };
    traceDraft = {
      ...traceDraft,
      targetQuery: node.id,
    };
  }

  function clearTrace(): void {
    traceSelection = {
      sourceNodeId: null,
      targetNodeId: null,
    };
    traceDraft = {
      sourceQuery: '',
      targetQuery: '',
    };
  }

  function clearTraceEndpoint(slot: 'source' | 'target'): void {
    if (slot === 'source') {
      traceSelection = {
        ...traceSelection,
        sourceNodeId: null,
      };
      traceDraft = {
        ...traceDraft,
        sourceQuery: '',
      };
      return;
    }

    traceSelection = {
      ...traceSelection,
      targetNodeId: null,
    };
    traceDraft = {
      ...traceDraft,
      targetQuery: '',
    };
  }

  function swapTrace(): void {
    traceSelection = {
      sourceNodeId: traceSelection.targetNodeId,
      targetNodeId: traceSelection.sourceNodeId,
    };
    traceDraft = {
      sourceQuery: traceDraft.targetQuery,
      targetQuery: traceDraft.sourceQuery,
    };
  }

  function updateTraceDraft(slot: 'source' | 'target', value: string): void {
    if (slot === 'source') {
      traceDraft = {
        ...traceDraft,
        sourceQuery: value,
      };
      traceSelection = {
        ...traceSelection,
        sourceNodeId: traceSelection.sourceNodeId && traceSelection.sourceNodeId === value ? traceSelection.sourceNodeId : null,
      };
      return;
    }

    traceDraft = {
      ...traceDraft,
      targetQuery: value,
    };
    traceSelection = {
      ...traceSelection,
      targetNodeId: traceSelection.targetNodeId && traceSelection.targetNodeId === value ? traceSelection.targetNodeId : null,
    };
  }

  function applyTraceSelection(): void {
    const source = traceSelection.sourceNodeId
      ? presentedGraph.nodes.find((node) => node.id === traceSelection.sourceNodeId) ?? null
      : resolveTraceReference(traceDraft.sourceQuery);
    const target = traceSelection.targetNodeId
      ? presentedGraph.nodes.find((node) => node.id === traceSelection.targetNodeId) ?? null
      : resolveTraceReference(traceDraft.targetQuery);

    traceSelection = {
      sourceNodeId: source?.id ?? null,
      targetNodeId: target?.id ?? null,
    };
    traceDraft = {
      sourceQuery: source?.id ?? traceDraft.sourceQuery,
      targetQuery: target?.id ?? traceDraft.targetQuery,
    };
  }

  function toggleModeMenu(): void {
    modeMenuOpen = !modeMenuOpen;
    densityMenuOpen = false;
    settingsOpen = false;
    activePopover = null;
    selectedNodeId = null;
  }

  function toggleDensityMenu(): void {
    densityMenuOpen = !densityMenuOpen;
    modeMenuOpen = false;
    settingsOpen = false;
    activePopover = null;
    selectedNodeId = null;
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
    selectedNodeId = null;
    activePopover = null;
    modeMenuOpen = false;
    densityMenuOpen = false;
    settingsOpen = false;
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
    settingsOpen = false;
  }

  function resetLayout(): void {
    manualPositions = {};
    if (storageKey) {
      clearLayoutState(storageKey);
    }
    selectedNodeId = null;
    activePopover = null;
    modeMenuOpen = false;
    densityMenuOpen = false;
    settingsOpen = false;
    refocusVersion += 1;
  }

  function handleNodeClick({ node }: { node: Node; event: MouseEvent | TouchEvent }): void {
    activePopover = null;
    modeMenuOpen = false;
    densityMenuOpen = false;
    settingsOpen = false;
    selectedNodeId = node.id;
  }

  function usePopupNodeAsTrace(slot: 'source' | 'target'): void {
    if (!popupNode) {
      return;
    }
    assignTraceEndpoint(slot, popupNode);
  }

  function togglePopupDomainCollapse(): void {
    if (!popupNode?.domain) {
      return;
    }
    toggleCollapsedDomain(popupNode.domain);
  }

  function togglePopupOwnerCollapse(): void {
    if (!popupNode?.owner) {
      return;
    }
    toggleCollapsedOwner(popupNode.owner);
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
    selectedNodeId = null;
  }

  function closeTransientUI(): void {
    activePopover = null;
    selectedNodeId = null;
    modeMenuOpen = false;
    densityMenuOpen = false;
    settingsOpen = false;
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
      trace: 'Trace',
      structure: 'Structure',
      owners: 'Teams',
      domains: 'Domains',
      nodeTypes: 'Types',
    };
    return labels[kind];
  }

  function popoverCount(kind: Exclude<PopoverKind, null>): number {
    if (kind === 'search') {
      return filterCounts.query;
    }
    if (kind === 'trace') {
      return filterCounts.trace;
    }
    if (kind === 'structure') {
      return filterCounts.structure;
    }
    if (kind === 'owners') {
      return filterCounts.owners;
    }
    if (kind === 'domains') {
      return filterCounts.domains;
    }
    return filterCounts.nodeTypes;
  }

  function popoverHasValue(kind: Exclude<PopoverKind, null>): boolean {
    return popoverCount(kind) > 0;
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
    if (kind === 'trace') {
      return 'P';
    }
    if (kind === 'structure') {
      return 'GR';
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
    if (kind === 'trace') {
      return '#8c6d44';
    }
    if (kind === 'structure') {
      return '#8f4a18';
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

  function isThemePreference(value: unknown): value is ThemePreference {
    return value === 'system' || value === 'light' || value === 'dark';
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

  function buildNodeSuggestions(
    nodes: PresentedNode[],
    query: string,
    excludeNodeId: string | null,
  ): PresentedNode[] {
    const normalizedQuery = query.trim().toLowerCase();
    const visible = nodes.filter((node) => node.id !== excludeNodeId);
    const sorted = [...visible].sort((left, right) => {
      const leftLabel = `${left.name} ${left.id}`.toLowerCase();
      const rightLabel = `${right.name} ${right.id}`.toLowerCase();
      return leftLabel.localeCompare(rightLabel);
    });

    if (!normalizedQuery) {
      return sorted.slice(0, 6);
    }

    return sorted
      .filter((node) => (
        `${node.id} ${node.name} ${node.domain} ${node.owner}`.toLowerCase().includes(normalizedQuery)
      ))
      .slice(0, 6);
  }

  function resolveTraceReference(query: string): PresentedNode | null {
    const trimmed = query.trim();
    if (!trimmed) {
      return null;
    }

    const byID = presentedGraph.nodes.find((node) => node.id === trimmed);
    if (byID) {
      return byID;
    }

    const lowered = trimmed.toLowerCase();
    const exactName = presentedGraph.nodes.find((node) => node.name.toLowerCase() === lowered);
    if (exactName) {
      return exactName;
    }

    const fuzzy = presentedGraph.nodes.filter((node) => (
      `${node.id} ${node.name}`.toLowerCase().includes(lowered)
    ));
    return fuzzy.length === 1 ? fuzzy[0] : null;
  }

  function buildTraceStatus(
    currentGraph: PresentedGraph,
    currentTraceSelection: TraceSelection,
  ): { label: string; tone: 'muted' | 'ok' | 'warning'; hops: number } {
    if (!currentTraceSelection.sourceNodeId && !currentTraceSelection.targetNodeId) {
      return {
        label: 'Pick two visible nodes to trace a path.',
        tone: 'muted',
        hops: 0,
      };
    }

    if (!(currentTraceSelection.sourceNodeId && currentTraceSelection.targetNodeId)) {
      return {
        label: 'Choose both a start and an end node.',
        tone: 'muted',
        hops: 0,
      };
    }

    if (!currentGraph.trace.found) {
      return {
        label: 'No visible path in the current view.',
        tone: 'warning',
        hops: 0,
      };
    }

    const hops = Math.max(0, currentGraph.trace.edgeIDs.length);
    return {
      label: currentGraph.trace.directed
        ? `Directed path across ${hops} hop${hops === 1 ? '' : 's'}.`
        : `Related path across ${hops} hop${hops === 1 ? '' : 's'} in the current view.`,
      tone: 'ok',
      hops,
    };
  }

  function buildImpactPreview(currentGraph: PresentedGraph, nodeId: string | null): ImpactPreview {
    if (!nodeId) {
      return {
        directUpstream: [],
        directDownstream: [],
        upstreamReach: 0,
        downstreamReach: 0,
        crossBoundaryTouches: 0,
      };
    }

    const nodeMap = new Map(currentGraph.nodes.map((node) => [node.id, node]));
    const directUpstreamIDs = uniquePreservingOrder(
      currentGraph.edges.filter((edge) => edge.to === nodeId).map((edge) => edge.from),
    );
    const directDownstreamIDs = uniquePreservingOrder(
      currentGraph.edges.filter((edge) => edge.from === nodeId).map((edge) => edge.to),
    );
    const crossBoundaryTouches = currentGraph.edges.filter((edge) => {
      if (edge.from !== nodeId && edge.to !== nodeId) {
        return false;
      }
      const source = nodeMap.get(edge.from);
      const target = nodeMap.get(edge.to);
      return Boolean(source?.domain && target?.domain && source.domain !== target.domain);
    }).length;

    return {
      directUpstream: directUpstreamIDs.map((id) => nodeMap.get(id)).filter(Boolean) as PresentedNode[],
      directDownstream: directDownstreamIDs.map((id) => nodeMap.get(id)).filter(Boolean) as PresentedNode[],
      upstreamReach: countReachable(currentGraph.edges, nodeId, 'upstream'),
      downstreamReach: countReachable(currentGraph.edges, nodeId, 'downstream'),
      crossBoundaryTouches,
    };
  }

  function countReachable(
    edges: PresentedGraph['edges'],
    originId: string,
    direction: 'upstream' | 'downstream',
  ): number {
    const seen = new Set<string>();
    const queue = [originId];

    while (queue.length > 0) {
      const current = queue.shift();
      if (!current) {
        break;
      }

      for (const edge of edges) {
        const next = direction === 'downstream'
          ? edge.from === current
            ? edge.to
            : null
          : edge.to === current
            ? edge.from
            : null;
        if (!next || next === originId || seen.has(next)) {
          continue;
        }
        seen.add(next);
        queue.push(next);
      }
    }

    return seen.size;
  }

  function popupBadgeLabel(node: PresentedNode): string {
    if (node.groupKind === 'domain') {
      return 'domain group';
    }
    if (node.groupKind === 'team') {
      return 'team group';
    }
    if (node.groupKind === 'boundary') {
      return 'boundary';
    }
    return node.type;
  }

  function popupBadgeColor(node: PresentedNode): string {
    return node.colorHint || nodeColor(model, node.type);
  }

  function popupTypeSummary(node: PresentedNode): string {
    return [
      node.typeSummary.service > 0 ? `${node.typeSummary.service}S` : '',
      node.typeSummary.api > 0 ? `${node.typeSummary.api}A` : '',
      node.typeSummary.database > 0 ? `${node.typeSummary.database}DB` : '',
      node.typeSummary.event > 0 ? `${node.typeSummary.event}E` : '',
    ].filter(Boolean).join(' · ');
  }

  function popupSourceLabel(node: PresentedNode): string {
    if (!node.file) {
      return 'n/a';
    }
    return `${node.file}${node.line ? `:${node.line}` : ''}`;
  }

  function popupCompositionLabel(node: PresentedNode): string {
    const summary = popupTypeSummary(node);
    return summary ? `${node.memberCount} nodes · ${summary}` : `${node.memberCount} nodes`;
  }

  function traceNodeName(nodeId: string | null): string {
    if (!nodeId) {
      return 'None';
    }
    const node = presentedGraph.nodes.find((candidate) => candidate.id === nodeId);
    return node ? node.name : nodeId;
  }

  function toggleValue(values: string[], value: string): string[] {
    const next = new Set(values);
    if (next.has(value)) {
      next.delete(value);
    } else {
      next.add(value);
    }
    return Array.from(next).sort((left, right) => left.localeCompare(right));
  }

  function uniquePreservingOrder(values: string[]): string[] {
    const seen = new Set<string>();
    const result: string[] = [];
    for (const value of values) {
      if (!value || seen.has(value)) {
        continue;
      }
      seen.add(value);
      result.push(value);
    }
    return result;
  }

  onMount(() => {
    explorerSettings = readExplorerSettings();
    const mediaQuery = typeof window !== 'undefined'
      ? window.matchMedia('(prefers-color-scheme: dark)')
      : null;
    systemPrefersDark = mediaQuery?.matches ?? false;
    void boot();

    function handleEscape(event: KeyboardEvent): void {
      if (event.key === 'Escape') {
        activePopover = null;
        selectedNodeId = null;
        modeMenuOpen = false;
        densityMenuOpen = false;
        settingsOpen = false;
      }
    }

    function handleWindowClick(event: MouseEvent): void {
      const target = event.target as HTMLElement | null;
      if (!target) {
        return;
      }

      if (
        target.closest('[data-interactive-root]') ||
        target.closest('.svelte-flow__node') ||
        target.closest('.svelte-flow__edge')
      ) {
        return;
      }

      activePopover = null;
      selectedNodeId = null;
      modeMenuOpen = false;
      densityMenuOpen = false;
      settingsOpen = false;
    }

    function handleThemeChange(event: MediaQueryListEvent): void {
      systemPrefersDark = event.matches;
    }

    window.addEventListener('keydown', handleEscape);
    window.addEventListener('click', handleWindowClick);
    if (mediaQuery) {
      if ('addEventListener' in mediaQuery) {
        mediaQuery.addEventListener('change', handleThemeChange);
      } else {
        mediaQuery.addListener(handleThemeChange);
      }
    }

    return () => {
      window.removeEventListener('keydown', handleEscape);
      window.removeEventListener('click', handleWindowClick);
      if (mediaQuery) {
        if ('removeEventListener' in mediaQuery) {
          mediaQuery.removeEventListener('change', handleThemeChange);
        } else {
          mediaQuery.removeListener(handleThemeChange);
        }
      }
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

<main class="app-shell" style={paletteStyle} data-theme={resolvedTheme}>
  <header class="page-header">
    <div class="page-header__brand">
      <span class="wordmark">Mapture</span>
      <TokenBadge label="Nodes" count={visible.nodes} interactive={false} quiet compact className="header-token" />
      <TokenBadge label="Edges" count={visible.edges} interactive={false} quiet compact className="header-token" />
    </div>

    <div class="page-header__actions">
      <span class={['header-pill', 'status-pill', connectionTone()].join(' ')}>
        <span class="status-dot"></span>
        {connectionLabel()}
      </span>
      <a
        class="header-pill header-link icon-pill"
        href={GITHUB_URL}
        target="_blank"
        rel="noreferrer"
        aria-label="Open GitHub repository"
        title="GitHub"
      >
        <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
          <path d="M9 19c-4 1.2-4-2.1-5.6-2.6M14.6 21v-3.1c0-1 .1-1.5-.4-2.1 2.3-.3 4.7-1.1 4.7-5a3.9 3.9 0 0 0-1-2.7 3.6 3.6 0 0 0-.1-2.7s-.9-.3-2.9 1a10.1 10.1 0 0 0-5.8 0c-2-1.3-2.9-1-2.9-1a3.6 3.6 0 0 0-.1 2.7 3.9 3.9 0 0 0-1 2.7c0 3.9 2.4 4.7 4.7 5-.5.6-.5 1.2-.4 2.1V21"></path>
        </svg>
        <span class="sr-only">GitHub</span>
      </a>
      <div class="header-control" data-interactive-root>
        <button
          type="button"
          class={['header-pill', 'header-button', 'icon-pill', settingsOpen ? 'active' : ''].join(' ')}
          onclick={toggleSettingsPanel}
          aria-label="Open explorer settings"
          title="Settings"
        >
          <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
            <path d="M10.3 3.3h3.4l.4 2a6.7 6.7 0 0 1 1.7.7l1.7-1 2.4 2.4-1 1.7c.3.5.5 1.1.7 1.7l2 .4v3.4l-2 .4a6.7 6.7 0 0 1-.7 1.7l1 1.7-2.4 2.4-1.7-1a6.7 6.7 0 0 1-1.7.7l-.4 2h-3.4l-.4-2a6.7 6.7 0 0 1-1.7-.7l-1.7 1-2.4-2.4 1-1.7a6.7 6.7 0 0 1-.7-1.7l-2-.4v-3.4l2-.4c.2-.6.4-1.2.7-1.7l-1-1.7L8 4.9l1.7 1c.5-.3 1.1-.5 1.7-.7l.4-2z"></path>
            <circle cx="12" cy="12" r="3.1"></circle>
          </svg>
          <span class="sr-only">Settings</span>
        </button>
      </div>
    </div>
  </header>

  <CanvasModal
    open={settingsOpen}
    title="Explorer Settings"
    description="Local preferences and feature toggles for the current browser."
    width="min(760px, calc(100vw - 2rem))"
    onclose={() => (settingsOpen = false)}
  >
    <div class="settings-modal-grid">
      {#each settingsSections as section}
        <SettingsSection title={section.title} description={section.description}>
          {#each section.fields as field}
            <SettingsField field={field} onchange={handleSettingsFieldChange} />
          {/each}
        </SettingsSection>
      {/each}
    </div>
  </CanvasModal>

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
      <Background color="var(--canvas-grid)" gap={26} />
      <MiniMap position="bottom-left" pannable zoomable />
      <Controls position="bottom-right" />

      <Panel position="top-left" class="canvas-toolbar-shell">
        <div class="canvas-toolbar" data-interactive-root bind:this={toolbarElement}>
          <div class="canvas-rail">
            {#each visibleRailKinds() as kind}
              <TokenBadge
                label={railButtonLabel(kind)}
                icon={iconForKind(kind)}
                count={popoverHasValue(kind) ? popoverCount(kind) : null}
                accent={accentForKind(model, kind)}
                active={activePopover === kind}
                compact
                className={[
                  'rail-pill',
                  `rail-pill--${kind}`,
                  popoverHasValue(kind) ? 'has-value' : '',
                ].join(' ')}
                onclick={() => togglePopover(kind)}
              />
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

          {#if explorerSettings.experimental.traceTools && activePopover === 'trace'}
            <div class="toolbar-popover trace-popover" data-interactive-root>
              <div class="popover-head">
                <strong>Path Trace</strong>
                <button type="button" class="mini-action" onclick={clearTrace}>Clear</button>
              </div>
              <div class="trace-fields">
                <label class="trace-field">
                  <span>From</span>
                  <input
                    type="search"
                    value={traceDraft.sourceQuery}
                    placeholder="Start node id or name"
                    oninput={(event) => updateTraceDraft('source', (event.currentTarget as HTMLInputElement).value)}
                    onblur={applyTraceSelection}
                  />
                  {#if traceSourceSuggestions.length > 0}
                    <div class="trace-suggestions">
                      {#each traceSourceSuggestions as node}
                        <button type="button" class="trace-suggestion" onclick={() => assignTraceEndpoint('source', node)}>
                          <strong>{node.name}</strong>
                          <small>{node.id}</small>
                        </button>
                      {/each}
                    </div>
                  {/if}
                </label>
                <label class="trace-field">
                  <span>To</span>
                  <input
                    type="search"
                    value={traceDraft.targetQuery}
                    placeholder="Target node id or name"
                    oninput={(event) => updateTraceDraft('target', (event.currentTarget as HTMLInputElement).value)}
                    onblur={applyTraceSelection}
                  />
                  {#if traceTargetSuggestions.length > 0}
                    <div class="trace-suggestions">
                      {#each traceTargetSuggestions as node}
                        <button type="button" class="trace-suggestion" onclick={() => assignTraceEndpoint('target', node)}>
                          <strong>{node.name}</strong>
                          <small>{node.id}</small>
                        </button>
                      {/each}
                    </div>
                  {/if}
                </label>
              </div>
              <div class="trace-actions">
                <button type="button" class="mini-action" onclick={applyTraceSelection}>Trace</button>
                <button type="button" class="mini-action" onclick={swapTrace}>Swap</button>
                <button type="button" class="mini-action" onclick={() => clearTraceEndpoint('source')}>Clear start</button>
                <button type="button" class="mini-action" onclick={() => clearTraceEndpoint('target')}>Clear end</button>
              </div>
              <article class={['trace-status', `trace-status--${traceStatus.tone}`].join(' ')}>
                <strong>{traceStatus.label}</strong>
                <small>{traceNodeName(traceSelection.sourceNodeId)} -> {traceNodeName(traceSelection.targetNodeId)}</small>
              </article>
            </div>
          {/if}

          {#if explorerSettings.experimental.structureTools && activePopover === 'structure'}
            <div class="toolbar-popover structure-popover" data-interactive-root>
              <div class="popover-head">
                <strong>Structure</strong>
                <button type="button" class="mini-action" onclick={resetStructure}>Reset</button>
              </div>
              <TokenBadge
                label="Summarize cross-domain links"
                icon={iconForKind('structure')}
                accent={accentForKind(model, 'structure')}
                active={aggregateCrossDomain}
                trailingText={aggregateCrossDomain ? 'On' : 'Off'}
                className="filter-chip structure-toggle"
                onclick={toggleCrossDomainAggregation}
              />

              <section class="structure-section">
                <div class="structure-section__head">
                  <strong>Collapse Domains</strong>
                  <button type="button" class="mini-action" onclick={() => (collapsedDomains = [])}>Clear</button>
                </div>
                <div class="structure-list">
                  {#each model.domains as domain}
                    <button
                      type="button"
                      class={['structure-row', collapsedDomains.includes(domain) ? 'active' : ''].join(' ')}
                      style={chipStyle(model, 'domains')}
                      onclick={() => toggleCollapsedDomain(domain)}
                    >
                      <span class="structure-row__copy">
                        <strong>{domainName(model, domain)}</strong>
                        <small>{structureDomainCounts[domain] ?? 0} visible</small>
                      </span>
                      <span class="structure-row__state">{collapsedDomains.includes(domain) ? 'Collapsed' : 'Open'}</span>
                    </button>
                  {/each}
                </div>
              </section>

              <section class="structure-section">
                <div class="structure-section__head">
                  <strong>Collapse Teams</strong>
                  <button type="button" class="mini-action" onclick={() => (collapsedOwners = [])}>Clear</button>
                </div>
                <div class="structure-list">
                  {#each model.owners as owner}
                    <button
                      type="button"
                      class={['structure-row', collapsedOwners.includes(owner) ? 'active' : ''].join(' ')}
                      style={chipStyle(model, 'owners')}
                      onclick={() => toggleCollapsedOwner(owner)}
                    >
                      <span class="structure-row__copy">
                        <strong>{teamName(model, owner)}</strong>
                        <small>{structureOwnerCounts[owner] ?? 0} visible</small>
                      </span>
                      <span class="structure-row__state">{collapsedOwners.includes(owner) ? 'Collapsed' : 'Open'}</span>
                    </button>
                  {/each}
                </div>
              </section>
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
                  <TokenBadge
                    label={teamName(model, owner)}
                    icon={iconForKind('owners')}
                    count={visibleOwnerCounts[owner] ?? 0}
                    accent={accentForKind(model, 'owners')}
                    active={filters.owners.includes(owner)}
                    className="filter-chip filter-chip--owner"
                    onclick={() => toggleFilter('owners', owner)}
                  />
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
                  <TokenBadge
                    label={domainName(model, domain)}
                    icon={iconForKind('domains')}
                    count={visibleDomainCounts[domain] ?? 0}
                    accent={accentForKind(model, 'domains')}
                    active={filters.domains.includes(domain)}
                    className="filter-chip filter-chip--domain"
                    onclick={() => toggleFilter('domains', domain)}
                  />
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
                  <TokenBadge
                    label={capitalize(nodeType)}
                    icon={iconForKind('nodeTypes', nodeType)}
                    count={visibleTypeCounts[nodeType] ?? 0}
                    accent={nodeColor(model, nodeType)}
                    active={filters.nodeTypes.includes(nodeType)}
                    className="filter-chip filter-chip--node-type kind-chip"
                    onclick={() => toggleFilter('nodeTypes', nodeType)}
                  />
                {/each}
              </div>
            </div>
          {/if}

          {#if activeFilterBadges.length > 0}
            <div class="active-strip">
              {#each activeFilterBadges as badge}
                <TokenBadge
                  label={badge.label}
                  icon={badge.icon}
                  accent={badge.tone}
                  trailingText="x"
                  className={['active-badge', `active-badge--${badge.kind}`].join(' ')}
                  onclick={() => removeBadge(badge)}
                />
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

      {#if popupNode}
        <Panel position="bottom-right" class="node-inspector-shell">
          <div class="node-inspector-stack">
            <NodeInspector
              node={popupNode}
              badgeLabel={popupBadgeLabel(popupNode)}
              badgeAccent={popupBadgeColor(popupNode)}
              domainLabel={popupNode.domain ? domainName(model, popupNode.domain) : 'n/a'}
              ownerLabel={popupNode.owner ? teamName(model, popupNode.owner) : 'n/a'}
              sourceLabel={popupSourceLabel(popupNode)}
              compositionLabel={popupCompositionLabel(popupNode)}
              summary={popupNode.summary}
              actions={nodeInspectorActions}
              onaction={handleNodeInspectorAction}
              onclose={() => (selectedNodeId = null)}
            />

            {#if explorerSettings.experimental.impactPreview}
              <ImpactPreviewPanel
                preview={popupImpact}
                ontrace={explorerSettings.experimental.traceTools ? assignTraceEndpoint : undefined}
              />
            {/if}
          </div>
        </Panel>
      {/if}
    </SvelteFlow>
  </section>
</main>
