---
id: 014
title: AI bundle exporter (`mapture export-ai`)
milestone: v0.5.0
status: todo
prd: §19, §20
depends_on: [001, 004]
---

## Why
PRD §19 calls AI-first the *core differentiator*. `cmd/export-ai` is a
`todo()` stub today. The bundle is the artifact users feed into Claude
Code, ChatGPT, Cursor, MCP servers, etc.

## Scope
Create `internal/exporter/ai`. Output layout (PRD §19):

```
.mapture/ai/
  graph.json                      # full normalized graph
  entities/
    service-checkout-service.md   # one per node
    api-stripe-api.md
    event-order.placed.md
  views/
    system-overview.md            # full system, grouped by domain
    domain-orders.md              # one per domain
    team-commerce.md              # one per team
  prompts/
    explain-node.md
    blast-radius.md
    cross-domain-risk.md
    trace-flow.md
    find-undocumented.md
  glossary.md
```

Per-entity Markdown should follow the PRD §19 example shape: type,
domain, owner, depends-on / called-by lists, plus a generated natural
summary stitched from comment fields.

Wire `mapture export-ai [path]` to write the bundle into
`<repo>/.mapture/ai/`. Add `--out` flag for non-default locations.

## Acceptance
- `go run . export-ai examples/ecommerce` produces a bundle with one
  entity file per node and one view per domain/team
- `glossary.md` contains every catalog event with description and
  ownership
- File contents are stable across runs (golden snapshot test)

## Out of scope
- MCP server (PRD §19 future direction) — track separately when v0.5
  ships
