---
id: 013
title: Frontend UI — graph explorer
milestone: v0.3.0
status: todo
prd: §18 (UI goals), §29 v0.3.0
depends_on: [001]
---

## Why
Tasks 003 (HTML export) and 012 (server) both need the *same* frontend.
Picking it once and committing it under `web/` (built into Go's
`embed` FS) keeps the binary single-file and avoids the trap of
maintaining two UIs.

## Scope
- Decide between **Cytoscape.js** and **vis-network** (PRD §18 candidate
  list). Recommendation: Cytoscape.js — better filtering API, mature
  layouts, smaller surface area.
- Repo layout under `web/`:
  - `web/src/` — TypeScript sources
  - `web/dist/` — built output, committed and `//go:embed`-included by
    `internal/exporter/html` and `internal/server`
- Features (v0.3):
  - search box (filter by id/name/domain/owner)
  - node-type legend with toggle visibility
  - edge-type toggle (`calls`, `stores_in`, `emits`, `consumes`, …)
  - click → side panel (file/line/owner/domain/summary)
  - "isolate neighborhood" button — hides everything more than 1 hop away
  - group-by-domain layout
- Build script under `scripts/` so contributors don't need a separate Node
  workflow on top of the Go one
- Document the build in `AGENTS.md` (use the agent-docs skill)

## Acceptance
- `make web` produces `web/dist/` deterministically
- The same bundle works in `mapture export-html` and `mapture serve`
- Performance: 100-node graph renders in <100ms on M1
