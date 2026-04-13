---
id: 034
title: AI export from JGF graph export
milestone: v0.4.0
status: todo
prd: §19, §20
depends_on: [030, 033]
---

## Goal
Generate AI-friendly files from the JGF graph export so LLM workflows can understand the architecture without rescanning the repository.

## Why
AI export is only valuable if it is trustworthy and repeatable. That means:

- no second scanner
- no second validator
- no AI-specific graph builder

The AI bundle should be a view over the JGF export, not a second product.

## Scope

### 1. Add `mapture export-ai`
Support both modes:

```bash
mapture export-ai examples/ecommerce
mapture export-ai --from /tmp/ecommerce.json
```

The first command builds the JGF export first. The second reuses an existing one.

### 2. Bundle layout
Expected output:

```text
.mapture/ai/
  export.json
  entities/
  views/
  prompts/
  glossary.md
```

The bundle should include the raw JGF export so downstream tools can choose structured JSON or Markdown.

### 3. Generated materials

- per-node dossiers
- system overview
- per-domain and per-team summaries
- glossary of teams, domains, and events observed in the graph
- reusable prompt starter files for common architecture questions

### 4. Use export metadata, not ad hoc inference
Node summaries may be enriched for readability, but they must be grounded in:

- JGF node metadata
- validation diagnostics
- graph relationships
- repo-configured teams/domains/tags

## Acceptance

- AI bundle generation works from either a live repo path or an existing JGF export file
- generated files are deterministic across runs
- the bundle contains the raw JGF export plus derived Markdown views
- no scanner or validator logic is duplicated inside the AI exporter

## Out of scope

- interactive chat server
- MCP transport
- embeddings/vector search

Those belong in the MCP task once the AI bundle contract is stable.
