---
id: 025
title: Make manual node dragging authoritative and support real reordering
milestone: v0.3.0
status: todo
prd: §10, §18, §29
depends_on: [014, 021]
---

## Why
Users still cannot actually arrange the graph. The current drag behavior
feels constrained by the layout engine and saved-position reconciliation,
so cards appear to snap back toward system-owned positions. That breaks
the core "workspace" feeling of the explorer.

## Scope
- Make manual dragging authoritative in all layout modes.
- Once a user drags a node, that node should remain where the user put
  it unless:
  - the user explicitly resets layout
  - the user switches to a different layout mode
  - collision recovery needs a minimal nudge to prevent overlap
- Preserve manual positions across:
  - filter changes
  - live reloads for unchanged nodes
  - page refresh for the same graph fingerprint and layout mode
- Support repeated user reordering of multiple nodes in one session
  without the graph feeling "magnetized" back to the base layout.
- Keep anti-overlap safeguards, but subordinate them to user intent as
  much as possible.

## Implementation Notes
- Separate:
  - system-computed layout positions
  - user-overridden positions
  - final resolved positions
- Track user-moved nodes explicitly rather than inferring everything
  from one saved-position map.
- Drag-stop collision resolution should preserve the dragged node and
  nudge surrounding nodes first where feasible.
- The explorer should feel like a canvas workspace, not a locked chart.

## Acceptance
- Users can drag any node and keep it there
- Filter changes do not snap manually placed nodes back
- Refreshing the page restores manual positions for the same graph/mode
- Collision handling nudges neighbors before it overrides the dragged
  node’s intent
- `Reset layout` is the explicit escape hatch to discard manual ordering

