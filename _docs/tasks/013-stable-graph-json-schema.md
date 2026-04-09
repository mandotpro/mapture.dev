---
id: 019
title: Lock the graph JSON schema
milestone: v1.0.0
status: todo
prd: §17, §29 v1.0.0
depends_on: [001]
---

## Why
PRD §17 shows an illustrative graph JSON shape. PRD §32 lists "define
exact graph JSON schema" as an immediate next step. Once the AI bundle
(task 014), HTML exporter (task 003), and server API (task 012) all
consume this, every change becomes a breaking one — so the schema needs
to be locked, versioned, and validated like the catalog files are.

## Scope
- Add `internal/schema/graph.cue` mirroring the `graph.Graph` shape
  (`schemaVersion`, `nodes`, `edges`, `metadata` block with generated_at,
  scanner_version, source_root)
- Add a `schemaVersion` field to `graph.Graph` (start at `1`)
- Add `internal/graph/graph_test.go` snapshot covering the example
  graphs (works with task 016)
- Document the contract in the public contributor docs — "the graph JSON
  schema is stable; bump `schemaVersion` and write a migration note for
  any breaking change"

## Acceptance
- Graph JSON validates against the new CUE schema
- A test fails if a new field appears without bumping `schemaVersion`
