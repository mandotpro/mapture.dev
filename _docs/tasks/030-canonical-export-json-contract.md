---
id: 030
title: Canonical export JSON contract and shared build pipeline
milestone: v0.3.0
status: todo
prd: §17, §19, §29
depends_on: [004, 005, 020]
---

## Goal
Define one stable exported JSON envelope for Mapture and make the scanner/validator pipeline produce it as the single source of truth for downstream consumers.

## Why
The product direction is no longer “CLI output plus a special web payload plus future AI files.” It is:

- one architecture build pipeline
- one exported JSON artifact
- many consumers on top of that artifact

Today those concerns are still split:

- `validate` owns diagnostics
- `graph` owns Mermaid conversion
- `serve` owns a web-specific payload
- future AI and MCP work still read like separate products

That drift will keep increasing unless there is one canonical export contract.

## Scope

### 1. Introduce one public export model
Add a dedicated export package and schema that is not “the web payload” and not “the raw graph” alone.

Suggested shape:

```json
{
  "schemaVersion": 1,
  "generatedAt": "2026-04-11T10:00:00Z",
  "toolVersion": "v0.3.0",
  "source": {
    "projectRoot": "...",
    "configPath": "...",
    "scopes": ["./src/orders"]
  },
  "catalog": {
    "teams": [...],
    "domains": [...]
  },
  "validation": {
    "summary": {...},
    "diagnostics": [...]
  },
  "graph": {
    "nodes": [...],
    "edges": [...]
  },
  "ui": {
    "defaultLayout": "...",
    "nodeColors": {...}
  },
  "meta": {
    "sourceLabel": "...",
    "mode": "live|offline|static"
  }
}
```

The exact field names can change, but the contract must satisfy these rules:

- it is generic enough for CLI, web, Mermaid, AI, and MCP consumers
- it is not tied to Svelte or any frontend-only concept
- it contains enough metadata for downstream consumers to avoid rescanning

### 2. Add a first-class `export-json` path
Add `mapture export-json [path]` with:

- `-o, --out <file>`
- the same scope/filter inputs that are safe to apply before export
- deterministic output ordering

The command should:

1. discover config
2. load config and inlined catalog data
3. scan sources
4. validate and build the graph
5. build the canonical export envelope
6. write JSON

### 3. Make the export builder reusable
Introduce a shared internal builder so downstream commands do not each assemble their own view of the world.

Good shape:

- `internal/export/model`
- `internal/export/builder`

Bad shape:

- one export shape in the server package
- another shape in future AI code
- another shape in CLI diagnostics

### 4. Lock the contract
Add:

- CUE schema for the canonical export
- golden snapshot tests for `examples/demo` and `examples/ecommerce`
- explicit `schemaVersion` bump rules for breaking changes

## Out of scope

- loading the export in the explorer
- HTML/static export packaging
- Mermaid conversion
- AI bundle generation
- MCP server

Those all depend on this task and should not re-specify the export shape themselves.

## Acceptance

- `mapture export-json examples/ecommerce -o /tmp/ecommerce.json` produces a stable JSON file
- the export validates against a CUE schema
- snapshot tests fail if a field changes without an intentional schema update
- the export includes graph, diagnostics, team/domain metadata, and enough meta information for downstream consumers
- the builder lives outside `src/cmd/` and outside the web server package

## Implementation Notes

- keep frontend-only state out of the canonical export
- keep the export generic enough that a static site, an MCP server, and a CLI converter can all consume it unchanged
- if `validate --format json` later reuses the `validation` subsection, that is a feature, not duplication
