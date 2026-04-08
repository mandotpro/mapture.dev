---
id: 022
title: Introduce semantic node visuals for the explorer
milestone: v0.3.0
status: done
prd: §10, §17, §18, §29
depends_on: [011, 014, 020, 021]
---

## Why
The current explorer still renders every node as the same rounded card
with only color and text to distinguish type. That is not enough for
fast system reading. Services, APIs, databases, and events should have
clearly different visual signatures so the graph is scannable even when
labels are partially obscured or the view is zoomed out.

## Scope
- Replace the single generic node card visual with per-type visuals for:
  - `service`
  - `api`
  - `database`
  - `event`
- Keep the current rounded soft design language, but add type-specific
  semantics through SVG or dedicated Svelte components.
- Define a stable visual system for each node type:
  - silhouette / icon treatment
  - accent placement
  - label hierarchy
  - selected state
  - hover state
- Ensure visuals still work in small-card mode and remain legible at the
  zoom levels used by the explorer.
- Keep node type colors configurable from `mapture.yaml` and make sure
  the new visual blocks respect those colors consistently.

## Implementation Notes
- Prefer a custom node component per semantic type or a shared component
  with explicit type render branches.
- Icons/SVGs must be embedded in the frontend bundle; no runtime asset
  loading.
- Visual differentiation should not rely on color alone.
- Preserve current node metadata needs:
  - name
  - domain
  - owner
  - selected state

## Acceptance
- A user can distinguish `service`, `api`, `database`, and `event`
  nodes from shape/visual treatment alone
- Selected and hovered nodes still feel consistent across all types
- The graph remains readable at normal zoom and zoomed-out overview
- Configured node colors still drive the semantic visuals cleanly
