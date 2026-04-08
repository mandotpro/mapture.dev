# Task 014: Svelte Full-Page Immersive UI Layout

## Goal
Design and implement the definitive Mapture frontend as a 100% full-page Svelte Flow canvas. Avoid dedicating permanent structural space to fixed filter blocks, traditional sidebars, or headers. Maximize viewport real estate purely for architectural exploration.

## Context
Graph exploration tools often suffer from "dashboard syndrome," where heavy toolbars and side-panels crush the actual graph into a tiny square. Mapture's focus is on massive architectural flows. The best UX mandates an edge-to-edge immersive canvas, where status badges and complex filtering tools "float" cleanly on top of the map just like interactive maps (e.g. Google Maps or Figma).

## Requirements

### 1. Edge-to-Edge Canvas
- Render the `SvelteFlow` component so it inherently consumes `100vw` and `100vh`. 
- Completely eliminate any surrounding `div` padding or HTML block layouts that push the canvas boundaries inward.

### 2. Floating Top Info Bar
- Embed an absolute-positioned floating bar at the top of the canvas.
- Utilize clean, pill-like badges to display operational data instantly:
  - An indicator confirming connection state (e.g., "Live API Connected" vs "Static Build" mode).
  - Quantitative summaries of the current payload (`Total Nodes`, counts per type like `Services` / `Events`, etc).

### 3. Internal Canvas Filtering & Overlays
- Shift all filtering interactions to overlay panels or canvas controls directly inside Svelte Flow.
- Provide actionable click/select options directly over the canvas to toggle visibility across:
  - Teams.
  - Domains.
  - Node Roles/Types.
- The filtering menus should act dynamically—toggling domains immediately recalculates and repaints the nodes without requiring a page refresh or shifting the canvas framing wrapper.

## Definition of Done
- Launching the UI presents a completely unobstructed edge-to-edge flow diagram.
- Users can clearly see their live connection status and aggregated telemetry in elegant, un-intrusive pill badges at the top boundary.
- Users can toggle domain/team filters via floating selectors, instantly reducing noise on the full-screen flow without sacrificing horizontal real estate.
