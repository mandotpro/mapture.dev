---
id: 014
title: Svelte Full-Page Immersive UI Layout
milestone: v0.3.0
status: done
depends_on: [011]
---

## Implemented
- The explorer is now an edge-to-edge `100vw` / `100vh` Svelte Flow canvas with no permanent page grid around it.
- Status, telemetry, filters, diagnostics, and selected-node details are all rendered as floating overlays inside the canvas using Svelte Flow panels.
- Team, domain, and node-type filters are toggle-based and update the rendered graph immediately without a page reload.

## Rendering fix included
- The previous Svelte Flow integration loaded graph data but failed to render node wrappers on the canvas.
- The frontend now uses Svelte 5 runes with raw graph arrays for the Svelte Flow node/edge props, which fixes node rendering in the installed `@xyflow/svelte` version.
- Verified against `examples/ecommerce`: the live page now renders 26 graph nodes and the custom node cards are present in the DOM.

## Notes
- The top overlay defaults to a compact telemetry bar instead of a full dashboard shell.
- Filters default to collapsed so the canvas stays dominant.
- Selected-node details only appear when a node is clicked.
