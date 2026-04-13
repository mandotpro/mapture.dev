---
id: 031
title: Explorer delivery modes from canonical export JSON
milestone: v0.3.0
status: done
prd: §18, §19, §29
depends_on: [030]
---

> Superseded by the JGF-first export model. The explorer is now delivered from the derived visualisation export, which is produced from the canonical JGF artifact.

## Goal
Make the explorer consume the canonical export JSON directly in every delivery mode:

- live `mapture serve`
- static/offline exported bundle
- standalone hosted web app
- direct “serve this export file” mode

## Why
The explorer is currently the strongest product surface, but it still reads like a special case. The intended model is:

- the backend builds one canonical export
- the explorer reads that export
- live mode is just “rebuild and resend the same export”

If the explorer keeps a private payload shape, every future consumer will repeat that mistake.

## Scope

### 1. Make the frontend adapter consume the canonical export
The Svelte app should normalize the shared export envelope directly, not a private server payload.

Any frontend-only derivation stays in the adapter/presentation layer, not the backend response shape.

### 2. Keep `mapture serve` as a live export server
`mapture serve [path]` should:

- rebuild the canonical export on source changes
- return that export from the API
- let the frontend consume it directly

### 3. Add file-driven serving
Add a mode such as:

```bash
mapture serve --from /tmp/ecommerce.json
```

This should:

- skip config discovery, scanning, and validation
- load the canonical export from disk
- serve the same explorer UI against that file
- disable watch/rebuild behavior

### 4. Implement static/offline explorer export
Add `mapture export-html` as a packaging command built on top of the canonical export:

- emit the explorer app
- emit `data.json` containing the canonical export
- make the explorer boot from that local file with no Go process required

### 5. Support standalone hosted web delivery
The hosted web app should be able to load:

- injected `window.__MAPTURE_DATA__`
- sibling `data.json`
- explicit `?data=<url>`
- user-selected local JSON file

All of those inputs are the same canonical export format.

## Acceptance

- `mapture serve examples/ecommerce` serves the explorer from the canonical export contract
- `mapture serve --from /tmp/ecommerce.json` serves the explorer without rescanning the repo
- `mapture export-html examples/ecommerce -o /tmp/ecommerce-view` creates a directory that works when opened from a simple static server
- a hosted standalone build can load a canonical export file via URL parameter or file drop
- there is no second backend-only explorer payload shape left in the codebase

## Implementation Notes

- keep the canonical export JSON generic; do not add Svelte-specific state to support this task
- live mode and offline mode should differ only in how the export is obtained
- if a bare `graph.Graph` file is still worth supporting for degraded viewing, wrap it at boot time instead of teaching the backend multiple schemas
