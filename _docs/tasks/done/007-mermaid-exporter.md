---
id: 002
title: Mermaid exporter
milestone: v0.1.0
status: todo
prd: §18, §29 v0.1.0
depends_on: [001]
---

## Why
The cheapest, highest-signal first exporter. Mermaid renders in GitHub
PRs and Markdown previews — perfect for demos and the website's
"before/after" section (PRD §27).

## Scope
Create `internal/exporter/mermaid` with:

- `Render(g *graph.Graph, opts Options) (string, error)`
- group nodes into `subgraph <Domain>` blocks (PRD §33 example)
- shape per node type:
  - `service` → `[Name]`
  - `api`     → `([Name])`
  - `database`→ `[(Name)]`
  - `event`   → `((Name))`
- edge labels = edge type (`calls`, `stores_in`, `emits`, …)
- deterministic ordering for stable diffs in CI
- options: filter by domain / team / node type (used later by `mapture graph`)

Wire `mapture graph` (currently `todo()`) to call this exporter and write
to stdout or `-o`.

## Acceptance
- `go run . graph examples/demo` prints a flowchart matching PRD §33
- `go run . graph examples/ecommerce -o ecommerce.mmd` writes a file that
  renders cleanly in mermaid.live
- golden snapshot test under `internal/exporter/mermaid`
