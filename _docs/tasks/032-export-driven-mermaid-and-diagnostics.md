---
id: 032
title: Export-driven Mermaid and diagnostics outputs
milestone: v0.3.0
status: todo
prd: §15, §17, §29
depends_on: [030]
---

## Goal
Make Mermaid rendering and machine-readable diagnostics derive from the same JGF export and validation structures instead of bespoke command-specific formats.

## Why
The JGF export is only useful if follow-on outputs stop bypassing it.

Two places still matter immediately:

1. Mermaid output for docs, PRs, and static diagrams
2. CI-friendly diagnostics for validation failures

Both should reuse the same data model that powers the explorer instead of rebuilding their own interpretation layer.

## Scope

### 1. Mermaid from JGF export
Refactor Mermaid generation so it can consume:

- a live build produced in-process
- a saved JGF export JSON file

Add either:

- `mapture graph --from export.json`

or

- `mapture export-mermaid --from export.json`

The exact command shape is less important than the rule:

**Mermaid should be renderable from the saved JGF export without rescanning the repo.**

### 2. Structured diagnostics model
Define one stable diagnostics shape that is also used inside the JGF export under `graph.metadata.mapture.validation`.

Then wire `mapture validate --format json` to emit that same diagnostics object rather than inventing a parallel report type.

Expected behavior:

- `validate --format text` remains the human CLI
- `validate --format json` emits the structured diagnostics section
- the JGF export embeds the same summary/diagnostics block

### 3. Exit code policy
Lock a clean exit code contract:

- `0` no validation errors
- `1` validation errors
- `2` tool/runtime errors

### 4. Converter parity tests
Add tests proving:

- `mapture graph examples/ecommerce`
- `mapture export-json-graph examples/ecommerce`
- `mapture graph --from ecommerce.json`

all describe the same underlying graph relationships

## Acceptance

- Mermaid output can be generated from a JGF export file without rescanning the repository
- `mapture validate --format json` emits the same diagnostics structure embedded in the JGF export
- exit codes are stable and documented
- no command-specific diagnostics schema exists outside the shared validation model

## Out of scope

- AI bundle generation
- MCP serving
- frontend rendering

Those tasks should build on the same export/diagnostics contract later.
