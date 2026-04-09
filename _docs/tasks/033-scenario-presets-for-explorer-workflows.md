# Task 033: Scenario Presets for Explorer Workflows

## Goal
Add first-class explorer presets that package multiple visualization controls into named, task-oriented workflows so users can answer common architecture questions with one click instead of manually combining view mode, density, filters, structure settings, and path/impact tools every time.

This task is intentionally about **presets as workflow shortcuts**, not just saved UI snapshots.

## Why this matters
The explorer is getting powerful:

- view modes
- density modes
- filters
- grouping/collapse controls
- cross-domain aggregation
- path tracing
- impact preview

That power is useful, but it shifts too much configuration work onto the user. Most people are not trying to "configure a graph"; they are trying to answer a question:

- what does the whole system look like?
- what crosses domain boundaries?
- how does this flow move through the system?
- what will this service affect?

Presets should turn those questions into immediate starting points.

## Product principle
Presets should be:

- intentional
- named by user goal
- fast to switch
- safe to override manually after activation

Presets should **not** lock the user into a mode or hide the underlying controls.

## What a preset should control

Each preset can optionally define:

- `viewMode`
- `densityMode`
- relation visibility defaults
- filter defaults
- grouping defaults
  - collapsed domains
  - collapsed teams
  - cross-domain aggregation
- trace behavior defaults
  - off by default
  - possibly preserve current trace when switching presets
- label policy overrides
- fit/refocus behavior

Presets should not own ephemeral user state like:

- currently selected node
- current popup
- current hover state
- current ad hoc query text unless the preset is explicitly query-based

## Preset philosophy

Presets are best when they are **opinionated starting states**. After the preset is applied, the user can still:

- change layout/view
- change density
- add filters
- collapse/expand groups
- trace a path

That means presets should feel like "jump to the right reading mode", not "load a different application".

## Recommended preset set for v1

### 1. System Overview

**Intent**
Understand the overall architecture shape quickly.

**Suggested behavior**

- `viewMode: system-map`
- `densityMode: overview`
- events hidden by default
- synthetic async edges on
- cross-domain aggregation off
- no collapsed groups

**What this helps**

- first-time repo exploration
- leadership and onboarding views
- "what exists here?" questions

### 2. Boundary Review

**Intent**
Find coupling and communication across domain lines.

**Suggested behavior**

- `viewMode: domain-lanes`
- `densityMode: standard`
- cross-domain aggregation on
- optional collapsed teams off by default
- edge labels on focused/hovered only

**What this helps**

- architecture governance
- domain ownership review
- boundary cleanup work

### 3. Event Journey

**Intent**
Read producers, events, and consumers as a flow.

**Suggested behavior**

- `viewMode: event-flow`
- `densityMode: standard`
- primary relations are `emits` and `consumes`
- databases mostly hidden
- calls only in detailed mode

**What this helps**

- event-driven systems
- debugging async chains
- onboarding into messaging-heavy domains

### 4. Storage Footprint

**Intent**
See where services store or read state.

**Suggested behavior**

- `viewMode: system-map`
- `densityMode: standard`
- focus relation visibility toward `stores_in` and `reads_from`
- optionally fade async links harder

**What this helps**

- data ownership review
- migration planning
- identifying hidden database coupling

### 5. Change Impact

**Intent**
Start from a selected service and understand what it can affect.

**Suggested behavior**

- does not set a specific selected node
- enables the strongest impact emphasis treatment
- keeps labels focused, not always-on
- possibly defaults to `system-map` or `domain-lanes`

**What this helps**

- refactoring planning
- change risk review
- debugging blast radius

### 6. Workbench

**Intent**
Manual editing and ad hoc arrangement.

**Suggested behavior**

- `viewMode: workbench`
- `densityMode: detailed`
- grouping off
- aggregation off

**What this helps**

- expert users
- screenshots
- manual storytelling

## Preset UX

### Location
Presets should be exposed as a compact top-right control near view and density, not buried in filters.

### Presentation
Each preset should show:

- name
- one-line purpose
- maybe a short glyph or icon

### Switching behavior
When a user switches preset:

- the preset applies its declared settings
- the canvas refits automatically
- active node selection may remain if still visible
- active trace may either:
  - be preserved by default, or
  - be cleared only if the preset explicitly says so

This should be a deliberate rule, not incidental behavior.

## Config design

Presets should eventually be configurable in `mapture.yaml`, but v1 can ship with built-ins first if needed.

Long-term shape:

```yaml
ui:
  presets:
    default: system-overview
    enabled:
      - system-overview
      - boundary-review
      - event-journey
      - storage-footprint
      - change-impact
      - workbench
```

Possible later extension for custom presets:

```yaml
ui:
  presets:
    default: boundary-review
    custom:
      - id: async-governance
        label: Async Governance
        summary: Events and cross-domain links
        view: event-flow
        density: standard
        aggregateCrossDomain: true
```

Do not start with arbitrary fully-user-defined preset schemas unless there is clear demand. Built-ins first is safer.

## Recommended implementation slices

### Slice 1: built-in preset layer

Implement:

- a frontend preset registry
- preset selector UI
- one-way preset application into current explorer state

No YAML yet.

### Slice 2: preset-aware state application rules

Define exactly which state a preset can override and which it preserves:

- view
- density
- structure controls
- label strategy
- relation-family visibility
- trace preservation rules

This is the hardest part conceptually.

### Slice 3: config exposure

Expose in `mapture.yaml`:

- default preset
- enabled presets

Avoid custom preset definitions in the same slice unless there is a strong reason.

### Slice 4: optional custom presets

Only after built-ins prove useful.

## Technical shape

Suggested frontend type:

```ts
type ExplorerPreset = {
  id: string;
  label: string;
  summary: string;
  viewMode?: ViewMode;
  densityMode?: DensityMode;
  aggregateCrossDomain?: boolean;
  collapsedDomains?: string[];
  collapsedOwners?: string[];
  relationFocus?: string[];
  clearTraceOnApply?: boolean;
}
```

Presets should be applied through one reducer/helper instead of scattered setters in `App.svelte`.

Example:

```ts
applyPreset(currentState, preset) => nextState
```

That helper should be deterministic and testable.

## Acceptance criteria

- users can switch between named presets in one click
- presets make the graph meaningfully different, not just cosmetically different
- switching presets is faster than manually configuring the equivalent controls
- presets remain compatible with manual overrides after activation
- preset application triggers a viewport refit

## Non-goals

- persisting arbitrary user-created presets in local storage
- multi-user shared preset editing
- server-side preset computation
- hidden "magic" presets that the user cannot inspect

## Open questions

1. Should presets live above or beside view/density controls?
2. Should a preset be allowed to override search text?
3. Should a preset clear collapsed groups when not explicitly specifying them?
4. Should presets be available in exported/static builds exactly the same way as in `serve`?

## Why this should be a dedicated task

Presets sound simple, but they sit at the intersection of:

- information architecture
- UI state ownership
- config design
- mental model clarity

If implemented casually, presets will become confusing or redundant. If implemented deliberately, they become the fastest path to making the explorer useful for non-expert users.
