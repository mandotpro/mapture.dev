---
id: 021
title: Fix web layout quality and guarantee no-overlap node placement
milestone: v0.3.0
status: done
prd: §10, §18, §29
depends_on: [010, 011, 014, 020]
---

## Why
The current explorer layout is not usable for real graph exploration.
Nodes overlap heavily, labels are obscured, edges stack into unreadable
clusters, and switching between layout modes does not reliably produce a
clear result. Even when the visual style is improving, the actual graph
placement is still failing the core UX requirement: users must be able
to understand the system from the first view.

This task should make layout quality a first-class behavior rather than
an incidental side effect of whatever force simulation happens to run.

## External Context
Svelte Flow’s docs explicitly separate layouting from rendering and point
to third-party strategies depending on the graph shape:

- `d3-force` is recommended for flexible, non-tree graph placement where
  a freeform map-like layout is desired
- the Svelte Flow `Node Collisions` example shows a concrete pattern for
  resolving overlap after interaction using `onnodedragstop` and a
  collision-resolution pass
- their example applies a collision resolver after dragging:
  `nodes = resolveCollisions(nodes, { maxIterations: Infinity, overlapThreshold: 0.5, margin: 15 })`

Relevant references:
- [Node Collisions](https://svelteflow.dev/examples/layout/node-collisions)
- [Layouting Libraries](https://svelteflow.dev/learn/layouting/layouting-libraries)

The goal here is not to copy the example blindly, but to adopt the same
principle: every layout mode must have an explicit anti-overlap pass and
dragging must not leave the graph in a broken state.

## Scope
- Define explicit layout quality rules for every supported layout mode:
  - no node card overlap at rest
  - no overlap after manual drag is released
  - enough spacing that node labels remain readable
  - graph should fit within a sensible viewport on first load
  - layout must remain stable across reloads for the same graph
- Add a dedicated collision-resolution utility in the frontend for
  Svelte Flow nodes, inspired by the Node Collisions example but adapted
  to Mapture node sizes and saved positions.
- Run collision resolution in these situations:
  - after initial layout computation
  - after switching layout modes
  - after drag stop
  - after restoring persisted positions from local storage
  - after filter changes reintroduce previously hidden nodes
- Define per-layout responsibilities clearly:
  - `Freeform`: flexible force-directed map with strong anti-overlap pass
  - `Clustered`: domain-aware grouped layout with stronger separation
    between clusters and the same anti-overlap pass
- Add layout-specific spacing controls instead of one shared magic set of
  constants. At minimum define configurable values for:
  - node margin
  - cluster spacing
  - edge/link distance
  - collision iterations / threshold
- Ensure node dimensions used for layout and collision resolution match
  the real rendered node card bounds closely enough to avoid visual
  collisions.
- Reduce visual pileups caused by restored positions:
  - if a saved position now collides with a newly visible or moved node,
    keep the user’s placement intent but nudge nodes apart automatically
  - if a saved layout is too dense to recover cleanly, fall back to a
    fresh layout pass before rendering the final positions
- Revisit initial `fitView` and viewport defaults so the first render
  shows a readable graph rather than a compressed pile of nodes.

## Implementation Notes
- Keep Svelte Flow as the renderer and interaction layer.
- Use `d3-force` as the base for freeform placement, but do not rely on
  force simulation alone to prevent overlap.
- Introduce a deterministic post-processing pass such as
  `resolveCollisions(nodes, options)` operating on final node positions.
- Wire collision handling through `onnodedragstop`, matching the Svelte
  Flow example’s interaction pattern.
- Prefer deterministic behavior: the same graph + same saved positions +
  same layout mode should converge to the same result.
- Persisted positions remain allowed, but they are subordinate to the
  no-overlap rule.
- If needed, split layout code into:
  - base layout generation
  - collision resolution
  - persisted-position reconciliation
  so each concern is testable independently.

## Acceptance
- `examples/ecommerce` renders with no overlapping node cards in both
  supported layout modes
- switching layout modes does not produce unreadable overlap
- dragging one or more nodes and releasing them results in a readable
  post-drag layout with no collisions
- restoring a saved layout from local storage does not reintroduce
  visible overlap
- filter toggles that bring nodes back into view do not cause collisions
- node labels remain readable and are not obscured by neighboring cards
- browser checks assert that rendered node bounding boxes do not overlap
  beyond a very small tolerance
- the layout system remains deterministic for the same graph and mode

## Suggested Tests
- unit tests for the collision-resolution helper using synthetic node
  rectangles with intentional overlap
- frontend/browser tests that:
  - load `examples/ecommerce`
  - collect rendered node bounding boxes
  - assert there are no meaningful intersections
  - switch layout modes and repeat the same assertion
  - drag a node into another node, release, and assert the overlap is
    resolved automatically
- regression test for saved positions:
  - persist overlapping positions in local storage
  - reload the page
  - assert the collision resolver separates them before final render
