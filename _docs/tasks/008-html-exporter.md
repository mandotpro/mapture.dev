---
id: 003
title: Static HTML exporter
milestone: v0.1.0
status: todo
prd: §10, §18, §29 v0.1.0
depends_on: [001]
---

## Why
A self-contained HTML report is the artifact most teams will paste into
docs sites and PR comments. PRD §10 lists it as a v0.1 surface and the
"one binary, no runtime deps" invariant (CLAUDE.md) requires assets to be
embedded via `embed`.

## Scope
- Create `internal/exporter/html` with embedded assets via `//go:embed`
- Single self-contained `.html` file: HTML + CSS + JS + graph JSON inlined
- v0.1 UI is intentionally minimal: a Cytoscape.js (or vis-network) canvas
  with:
  - search box filtering nodes by id/name/domain
  - legend per node type
  - click → side panel with file/line/owner/domain/summary
- Wire `mapture export-html [path] -o file.html` (flag already declared in
  `cmd/root.go`).

## Acceptance
- `go run . export-html examples/ecommerce -o report.html` produces a single
  file (no external `<script src>`) that opens offline
- file size stays under ~500 KB minified
- works in Chrome + Safari + Firefox

## Notes
- Library choice should be the same one used by `mapture serve` (task 012)
  so we don't ship two frontends. Decide once.
