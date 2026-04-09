# Task 032: Configurable Explorer UI Defaults & Visual Tuning

## Goal
Extend `mapture.yaml` so teams can optionally configure the explorer's default reading mode, density, viewport behavior, label strategy, and mode-specific visualization rules without making any of those settings required.

The current explorer already has stronger presentation behavior in the frontend, but almost all of it is hard-coded. This task turns the most valuable parts into repo-level defaults while preserving sensible built-in behavior for users who do not want extra configuration.

## Why this matters
The web explorer is now opinionated enough to be useful, but different repositories need different defaults:

- event-heavy repos want `Event Flow` up front
- domain-first teams want lane grouping by default
- large graphs need more aggressive initial zoom-out
- some teams want labels mostly hidden
- some teams want synthetic async summary edges on, others do not

If those choices stay hard-coded, users will keep fighting the UI instead of shaping it to their architecture.

The key product rule for this work:

**Every new explorer setting must be optional.**

If a repo omits all new keys, Mapture must behave exactly as it does today:

- existing `ui.defaultLayout` remains valid
- `System Map` remains the frontend default view
- `Standard` remains the default density
- current node colors remain the default palette
- current presentation heuristics remain the fallback behavior

## Current state

Today the Go config and schema expose only:

- `ui.defaultLayout`
- `ui.nodeColors.service`
- `ui.nodeColors.api`
- `ui.nodeColors.database`
- `ui.nodeColors.event`

The new frontend logic already has additional concepts that are not yet configurable:

- `ViewMode`
  - `system-map`
  - `event-flow`
  - `domain-lanes`
  - `workbench`
- `DensityMode`
  - `overview`
  - `standard`
  - `detailed`
- viewport refit behavior
- edge-label visibility rules
- synthetic async summary edges
- event collapsing/reveal rules
- domain-lane emphasis rules
- search suggestion behavior

That means the frontend and the config model are now out of sync. This task closes that gap intentionally rather than by leaking more hard-coded behavior into `App.svelte`.

## Design principles

### 1. Optional only
No new field should be required in `mapture.yaml`.

### 2. Repo defaults, not user session state
`mapture.yaml` should describe team-wide explorer defaults, not personal UI memory.

Good candidates for YAML:

- default view
- default density
- label policy
- viewport padding
- which modes are available
- mode-specific presentation behavior

Bad candidates for YAML:

- current search query
- currently selected node
- open popovers
- last zoom/pan
- per-user workbench node positions

### 3. Preserve backwards compatibility
Existing repos using only `ui.defaultLayout` and `ui.nodeColors` must continue to work unchanged.

### 4. Keep `src/cmd/` and server payload simple
The server should pass UI defaults through to the frontend in the explorer payload. The interpretation logic should stay mostly in the frontend presentation layer, not in command wiring.

### 5. Avoid config sprawl
Only expose knobs that materially affect readability or team workflow. Do not mirror every internal constant just because it exists.

## Proposed config shape

This is the intended direction, not a mandate to implement every field in one pass:

```yaml
ui:
  defaultLayout: elk-horizontal
  defaultView: system-map
  defaultDensity: standard

  views:
    enabled: [system-map, event-flow, domain-lanes, workbench]
    order: [system-map, event-flow, domain-lanes, workbench]

  viewport:
    fitPadding: 0.24
    minZoom: 0.18
    maxZoom: 1.2
    refitOnModeChange: true

  labels:
    edge: focused

  search:
    suggestionsLimit: 8

  systemMap:
    hideEventsByDefault: true
    showSyntheticAsyncEdges: true
    hideRelationsInOverview: [depends_on, reads_from]

  eventFlow:
    showCallsInDetailed: true
    revealDatabasesOnSearch: true

  domainLanes:
    showOwners: true
    emphasizeCrossDomainEdges: true
    showEventsInOverview: false

  nodeColors:
    service: "#1664d9"
    api: "#0f8f78"
    database: "#a56614"
    event: "#a73f7f"
```

## Field-by-field intent

### `ui.defaultView`
Maps directly to the frontend reading mode.

Allowed values:

- `system-map`
- `event-flow`
- `domain-lanes`
- `workbench`

Behavior:

- if present, it wins over the old `ui.defaultLayout` for frontend boot
- if absent, frontend derives the initial view from `ui.defaultLayout`

This lets us evolve the UX without breaking existing YAML.

### `ui.defaultDensity`
Controls initial clutter level.

Allowed values:

- `overview`
- `standard`
- `detailed`

This is the cleanest repo-level knob for large vs small graphs.

### `ui.views.enabled`
Allows a repo to remove modes that are misleading or low-value for that architecture.

Examples:

- event-sourced system disables `workbench`
- simple service map enables only `system-map` and `domain-lanes`

Rules:

- omit means all modes available
- if the configured default mode is disabled, fall back to the first enabled mode

### `ui.views.order`
Allows teams to choose which mode is first in the picker and how the choices are presented.

### `ui.viewport.fitPadding`
Controls how zoomed-out the initial render and refit behavior should be.

This is one of the highest-value settings for large graphs.

### `ui.viewport.minZoom` / `ui.viewport.maxZoom`
Useful guardrails for very dense or very small systems.

### `ui.viewport.refitOnModeChange`
Default should remain `true`. Some future power users may want to preserve their current camera on mode switches, but the default should optimize for readability.

### `ui.labels.edge`
Defines when edge labels should be visible.

Suggested enum:

- `hidden`
- `hover`
- `focused`
- `always`

Semantics:

- `hidden`: never show labels unless maybe future debug tools require them
- `hover`: show on hovered edge only
- `focused`: show on hovered edge and selected-node neighborhood
- `always`: show whenever the edge is visible

### `ui.search.suggestionsLimit`
Useful for very large graphs where suggestion strips could become noisy.

### `ui.systemMap.*`
Repo-level tuning for the default architecture overview mode.

- `hideEventsByDefault`
  - keep the current summary behavior configurable
- `showSyntheticAsyncEdges`
  - allow repos to disable collapsed producer-to-consumer summary links
- `hideRelationsInOverview`
  - likely restrict to known relation enums only

### `ui.eventFlow.*`
Controls secondary information in event-centric mode.

- `showCallsInDetailed`
- `revealDatabasesOnSearch`

### `ui.domainLanes.*`
Controls how strongly the domain-based grouping reads.

- `showOwners`
- `emphasizeCrossDomainEdges`
- `showEventsInOverview`

## Recommended implementation slices

This should be built in small vertical slices, not one giant schema dump.

### Slice 1: core frontend defaults

Implement:

- `ui.defaultView`
- `ui.defaultDensity`
- `ui.viewport.fitPadding`
- `ui.viewport.minZoom`
- `ui.viewport.maxZoom`
- `ui.viewport.refitOnModeChange`

Why first:

- directly visible impact
- low conceptual risk
- minimal schema expansion

### Slice 2: picker availability and label policy

Implement:

- `ui.views.enabled`
- `ui.views.order`
- `ui.labels.edge`

Why second:

- cleaner control over the new explorer experience
- removes hard-coded assumptions from `App.svelte`

### Slice 3: mode-specific presentation tuning

Implement:

- `ui.systemMap.*`
- `ui.eventFlow.*`
- `ui.domainLanes.*`

Why third:

- highest UX value, but also the most interaction complexity
- should land after the generic defaults are stable

### Slice 4: polish and docs

Implement:

- README section with example configs by repo type
- example fixture YAMLs that demonstrate different explorer defaults
- tests for defaulting and fallback behavior

## Backend and schema changes

### Go config model
Extend `src/internal/config/config.go` `UI` struct with optional nested sections for:

- view defaults
- viewport defaults
- label policy
- view availability/order
- mode-specific options
- search suggestion limit

The Go layer should apply defaults conservatively and only for fields that need stable payload values. The frontend can still own some fallback logic if that keeps the payload small.

### CUE schema
Update `src/internal/schema/config.cue` to validate:

- closed enums for `defaultView`, `defaultDensity`, `labels.edge`
- view lists containing only supported frontend view IDs
- safe numeric ranges for viewport settings
- relation lists containing only known edge types if we expose them directly

### Explorer payload
The server payload should expose enough UI config for the frontend to boot deterministically.

That likely means extending the frontend `UIConfig` payload contract, not inventing a second UI-specific payload shape.

## Frontend changes

### `types.ts`
Extend `UIConfig` to include the new optional fields.

### `adapter.ts`
Add normalization helpers that:

- resolve effective default view
- resolve effective density
- clamp viewport values
- resolve enabled/ordered modes safely
- apply label strategy
- resolve mode-specific fallback settings

The adapter should be the single place that turns partial payload config into stable runtime values.

### `App.svelte`
Replace remaining hard-coded UI defaults with adapter-provided effective config.

### presentation layer
Replace direct conditionals with config-driven behavior where appropriate:

- edge label visibility
- system map event collapsing
- synthetic async edge generation
- event-flow secondary edges
- domain-lane emphasis

## Tests

### Config/schema tests

- valid YAML with no new keys still loads
- valid YAML with every new key loads
- invalid enum values fail clearly
- invalid viewport numbers fail clearly
- disabled default view falls back predictably

### Frontend adapter tests

- `defaultView` wins over `defaultLayout`
- missing config falls back to current behavior
- label policy affects `showLabel`
- disabled modes are removed from the picker
- view ordering is deterministic
- `systemMap.showSyntheticAsyncEdges = false` suppresses synthetic async edges

### Smoke tests

- `make serve ecommerce` still boots without any new config
- adding only `ui.defaultDensity: overview` changes the initial presentation
- adding `ui.views.enabled: [system-map, domain-lanes]` removes the other modes from the UI

## Acceptance criteria

- all new explorer config keys are optional
- existing repos without new UI keys behave the same as before
- frontend boot behavior is deterministic from payload config
- invalid explorer config is rejected by schema validation
- users can set repo-wide explorer defaults without touching code

## Non-goals

- storing per-user workbench positions in YAML
- per-user themes or personalized UI memory
- exposing every internal layout constant
- backend-side graph transformations for presentation only

## Example repo profiles

### Service-oriented monolith

```yaml
ui:
  defaultView: system-map
  defaultDensity: overview
  viewport:
    fitPadding: 0.28
  systemMap:
    hideEventsByDefault: true
    showSyntheticAsyncEdges: true
```

### Event-driven platform

```yaml
ui:
  defaultView: event-flow
  defaultDensity: standard
  labels:
    edge: focused
  eventFlow:
    showCallsInDetailed: true
```

### Domain-boundary governance

```yaml
ui:
  defaultView: domain-lanes
  defaultDensity: standard
  domainLanes:
    emphasizeCrossDomainEdges: true
    showOwners: true
```

## Open questions

1. Should `ui.defaultLayout` eventually become legacy-only while `ui.defaultView` becomes canonical, or do we keep both long-term?
2. Do we want to expose per-mode numeric spacing controls later, or keep spacing internal to avoid fragile tuning?
3. Should `ui.views.enabled` be purely declarative, or can CLI flags temporarily override it for local usage?
4. Should `labels.edge` support a future `auto` mode that changes behavior by density?

## Suggested follow-up after this task

Once this configuration layer exists, the next logical improvement is a curated set of example `mapture.yaml` explorer presets in `examples/` and docs:

- service-map preset
- event-platform preset
- boundary-governance preset

That will help teams adopt the UI faster without forcing them to discover every option manually.
