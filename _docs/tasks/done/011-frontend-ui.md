---
id: 011
title: Frontend UI
milestone: v0.3.0
status: done
depends_on: [006, 007, 012]
---

## Findings
- Existing backend APIs already exposed the needed explorer contract:
  - `GET /api/graph` returns the normalized graph payload used by exporters.
  - `GET /api/validate` returns graph plus diagnostics.
  - `GET /api/catalog` returns teams, domains, and events for filter metadata.
  - `GET /api/events` streams live reload notifications.
- The existing serving model was already correct for a single-binary app:
  - static frontend assets are embedded in Go and served from `mapture serve`
  - no backend redesign was required
- The existing frontend was a minimal custom Cytoscape bundle under the top-level `web/` directory. This task replaces that with a Svelte Flow explorer and moves the embedded frontend under `src/internal/webui/` so the repo structure matches the rest of the Go codebase.

## Implemented
- New frontend package layout:
  - `src/internal/webui/frontend/` — Svelte Flow source, adapter layer, Vite config, lockfile
  - `src/internal/webui/dist/` — committed production bundle embedded by Go
  - `src/internal/webui/webui.go` — `//go:embed` wrapper used by `src/internal/server`
- Frontend adapter modules:
  - `loadGraphFromApi(...)`
  - `loadGraphFromFile(...)`
  - `normalizeGraph(...)`
  - `toSvelteFlowNodes(...)`
  - `toSvelteFlowEdges(...)`
- UI features:
  - Svelte Flow graph canvas with zoom, pan, fit view, controls, and minimap
  - automatic Dagre layout derived from the normalized graph
  - search by id, name, domain, owner, file, and summary
  - filters for node type, domain, and owner
  - node details panel with source references and summaries
  - diagnostics panel fed from validator output
  - JSON file loader for local payload inspection
  - live reload via `/api/events`

## Notes
- Backend changes were intentionally minimal:
  - `src/internal/server` now imports `src/internal/webui` instead of the old top-level `web` package
  - existing API shapes were reused as-is
- `make web` now builds the Svelte Flow UI from `src/internal/webui/frontend/` into `src/internal/webui/dist/`.
- Next improvements:
  - HTML exporter can reuse the same embedded bundle and injected payload path.
  - Domain swimlanes and richer edge grouping can be layered on top without changing backend contracts.
