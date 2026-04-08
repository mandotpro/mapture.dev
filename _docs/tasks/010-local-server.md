---
id: 012
title: Local explorer server (`mapture serve`)
milestone: v0.3.0
status: todo
prd: §10, §18, §29 v0.3.0
depends_on: [001, 003]
---

## Why
PRD §10 lists `mapture serve` as a primary surface. v0.3 success criteria
(PRD §29) say users should be able to interactively explore the graph,
filter, and drill down. `cmd/serve` is a `todo()` stub today.

## Scope
- New `internal/server` package
- `Serve(ctx, cfg, addr)` boots an `http.Server` that:
  - serves embedded UI assets (`//go:embed`)
  - exposes `GET /api/graph` returning the same JSON as `mapture scan`
  - exposes `GET /api/catalog` returning teams/domains/events
  - exposes `GET /api/validate` returning the validator report (task 008)
  - watches files under `cfg.Scan.Include` and re-scans on change (use
    `fsnotify`); broadcast over Server-Sent Events for live reload
- Wire `cmd/serve` to call `internal/server.Serve` with `--addr`,
  `--no-watch`, `--open` flags
- Reuse the same UI library as the HTML exporter (task 003) — do not
  ship two frontends

## Acceptance
- `go run . serve examples/ecommerce` opens an explorer on
  `http://127.0.0.1:8765` showing the full graph
- Editing a comment in a fixture file refreshes the open browser tab
  within ~1s
- `Ctrl-C` shuts down cleanly with no goroutine leak
