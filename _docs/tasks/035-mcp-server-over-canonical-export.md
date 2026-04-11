---
id: 035
title: MCP server over canonical export JSON
milestone: v0.5.0
status: todo
prd: §19 future direction
depends_on: [030, 034]
---

## Goal
Expose Mapture as an MCP server so users can ask architecture questions through chat tools while the server answers from the canonical export JSON.

## Why
“Chat with your infrastructure” is compelling only if the architecture model is:

- deterministic
- inspectable
- identical to what the CLI and explorer see

The MCP server should not build a parallel architecture world. It should sit on top of the same export artifact.

## Scope

### 1. Add an MCP server command
Expected command surface:

```bash
mapture mcp serve examples/ecommerce
mapture mcp serve --from /tmp/ecommerce.json
```

The live mode builds the canonical export once at startup and can optionally watch/reload later. The `--from` mode serves a static export.

### 2. Expose useful resources/tools
Initial MCP surface should stay narrow and trustworthy:

- fetch full export
- fetch node by id
- list teams/domains/tags
- list diagnostics
- search nodes by type/domain/tag
- explain neighborhood of a node

If richer “trace” or “impact” tools appear later, they should still operate over the export, not direct repo scanning.

### 3. Keep answers grounded
Every MCP result should be explainable from:

- canonical export fields
- graph relationships
- validation diagnostics

The server may synthesize response text, but it must not invent hidden architecture state.

## Acceptance

- an MCP client can connect to a local `mapture mcp serve`
- the client can retrieve nodes, diagnostics, domains, and graph neighborhoods from the canonical export
- `--from export.json` works without repo access
- no scanner or validator code is embedded in MCP request handlers

## Notes

- start with read-only architecture introspection
- mutation, remediation suggestions, or PR-writing flows are separate future work
