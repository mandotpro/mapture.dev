# Task 031: Web Platform Separation & Multi-Source Data Loading

## Goal
Make the Mapture web explorer a standalone, independently deployable application that can load data from any source — live API, static file, user upload, or remote URL — while keeping the existing `mapture serve` experience intact. The result is one web UI that works in three deployment modes: embedded in the Go binary, hosted as a static site, and exported as a self-contained directory alongside a data file.

## Context

### What already works
The frontend (Svelte 5 + Svelte Flow, `src/internal/webui/frontend/`) already has the bones of dual-mode loading in `App.svelte:183-206`:

```
if window.__MAPTURE_DATA__ exists → use it (offline)
else → fetch /api/explorer + bind SSE (live)
```

The adapter layer (`lib/adapter.ts:normalizeGraph`) converts the canonical `ExplorerPayload` into the in-memory `GraphModel`. It already uses `??` fallbacks throughout, which means it tolerates missing optional fields. The `ExplorerPayload` contract lives in `server/payload.go` (Go) and `lib/types.ts` (TS).

### What's missing

1. **No file loading.** If neither `window.__MAPTURE_DATA__` nor `/api/explorer` is available, the UI shows an error. There is no file picker, drag-and-drop, or URL parameter to load arbitrary JSON. A user who has a valid `ExplorerPayload` JSON file (from `mapture export-html`, from CI, or from a third-party tool) cannot view it without running the Go binary.

2. **`export-html` is a stub.** The CLI command exists (`src/cmd/root.go:401`) but calls `todo("export-html")`. Task 008 describes the approach (eject SPA + `data.json`), but hasn't been implemented.

3. **No standalone deployment path.** `web/dist/` is embedded in the Go binary. There is no documented way to host the frontend on static infrastructure (GitHub Pages, S3, Netlify) without the Go backend.

4. **No partial payload support.** The frontend expects the full `ExplorerPayload` shape (graph + catalog + validation + ui + meta). A user who only has a `graph.Graph` JSON (nodes + edges) cannot load it — they'd need to wrap it in the full envelope manually. This blocks "any data provider" use cases.

### The data source question

The user asked whether we need a formal `DataSource` abstraction (API vs Files). After reviewing the code:

**No.** The current boot function is a one-time decision: "where does the initial payload come from?" Once `normalizeGraph()` runs, the entire UI operates on `GraphModel` regardless of origin. The only ongoing data flow is SSE reload, which only applies in live-server mode. All other modes are load-once (or reload-on-drop for file mode).

What we need is not a class hierarchy but a **priority chain** in the boot function:

```
1. URL has ?data=<url>         → fetch that URL, mode=offline
2. window.__MAPTURE_DATA__     → use it, mode=offline  (export-html)
3. /api/explorer responds      → use it, mode=live     (mapture serve)
4. None of the above           → show file picker, mode=offline
```

After boot, the only reload mechanism varies by mode:
- Live: SSE triggers re-fetch of `/api/explorer`
- File: user drops a new file → re-normalize
- Embedded/URL: no reload (static snapshot)

This is simple enough that a `reloadFn: (() => Promise<ExplorerPayload>) | null` on the app state is sufficient. No need for an abstract DataSource interface.

## Stories

### Story 1: Boot chain & file loading in the frontend

**What:** Extend `App.svelte:boot()` from a 2-branch if/else to the 4-step priority chain above. Add a file-picker/drag-drop landing screen that appears when no data source is detected (step 4). Add `?data=<url>` support (step 1).

**Why this is first:** Every other story depends on the frontend being able to load data without a running Go server. This unblocks standalone deployment, export-html testing, and third-party integration.

**Scope:**
- Refactor `boot()` in `App.svelte` into the 4-step chain. Each step is a plain async function, not an abstract class.
- New component: a minimal landing screen with a file drop zone and a URL input field. Shown only when steps 1-3 all fail (no URL param, no injected data, no server). Style it to match the existing explorer theme.
- File drop accepts `.json` files. On drop/select, parse as `ExplorerPayload`, run `normalizeGraph()`, render the explorer. If the JSON is a bare `graph.Graph` (has `nodes`/`edges` at top level but no `catalog` key), wrap it in a minimal `ExplorerPayload` envelope with empty catalog/validation/default UI before normalizing.
- `?data=<url>` fetches the URL, parses as above, renders the explorer. CORS is the user's problem; document this.
- After file load, show a "load another file" button in the header (where the SSE status indicator sits in live mode). Dropping a new file re-normalizes without page reload.
- No changes to `normalizeGraph()` or `lib/types.ts` — the adapter already handles optional fields.
- The `sourceLabel` in file/URL mode should show the filename or URL so the user knows what they're looking at.

**Acceptance:**
- Open `web/dist/index.html` directly in a browser (file:// protocol) → file picker appears.
- Drop a valid `ExplorerPayload` JSON → explorer renders with full filtering, search, side panel.
- Drop a bare `graph.Graph` JSON (just nodes + edges) → explorer renders; catalog-dependent features (domain names, team names, event details) degrade gracefully (show raw IDs instead of labels).
- Open `index.html?data=https://example.com/graph.json` → fetches and renders.

**Does not include:** Deploying the web anywhere, export-html, or the Go backend.

---

### Story 2: Implement `export-html` command

**What:** Implement the `mapture export-html` CLI command that produces a self-contained directory with the SPA bundle and a `data.json` file.

**Why:** This is the primary offline distribution path. A developer runs `mapture export-html examples/ecommerce -o dist/` and gets a directory they can zip, email, host on an internal wiki, or open locally — no Go binary needed by the viewer.

**Scope:**
- Implement `internal/exporter/html/exporter.go`. Accept a `graph.Graph`, catalog, validation result, and config. Produce an `ExplorerPayload` JSON file (`data.json`).
- The export command (`src/cmd/root.go:401`):
  1. Loads config, catalog, scans sources, runs validation (same pipeline as `validate`).
  2. Builds the `ExplorerPayload` with `meta.mode = "offline"` and `meta.sourceLabel = "static export"`.
  3. Copies the embedded `webui.FS()` contents into the output directory.
  4. Writes `data.json` alongside the SPA files.
  5. Writes a thin `index.html` wrapper that sets `window.__MAPTURE_DATA__` from a `<script>` tag loading `data.json`, then loads the SPA's `app.js`. (Or alternatively, the SPA detects a sibling `data.json` via relative fetch — simpler, no wrapper needed. Pick whichever is cleaner after Story 1 lands.)
- Flags: `-o <dir>` (required), `--domain`, `--team`, `--type` filters (same as `graph` command — apply before export so the payload only contains the filtered view).
- The output directory must be servable by any static file server (`python -m http.server`, `npx serve`, nginx, GitHub Pages) with zero configuration.

**Acceptance:**
- `mapture export-html examples/ecommerce -o /tmp/ecom-export` produces a directory.
- `cd /tmp/ecom-export && python3 -m http.server 9000` → browser at `localhost:9000` shows the full interactive explorer with all ecommerce nodes.
- `mapture export-html examples/ecommerce -o /tmp/billing --domain billing` → export contains only billing-domain nodes.
- The output directory weighs under 500KB (excluding data.json).

**Depends on:** Story 1 (the frontend must handle the `window.__MAPTURE_DATA__` or sibling-`data.json` path).

**Existing task:** This overlaps with `_docs/tasks/008-html-exporter.md`. Story 2 supersedes task 008 — when this ships, mark 008 as done.

---

### Story 3: Standalone web deployment

**What:** Make `web/dist/` independently hostable as a static site that shows the file-picker landing screen by default and can be pointed at data via URL parameters.

**Why:** Teams that want a persistent, always-available architecture viewer shouldn't need to re-run `mapture export-html` every time. They host the UI once and point it at data files produced by CI.

**Scope:**
- Ensure `web/dist/` works when served from a non-root path (e.g., `https://internal.example.com/mapture/`). This means all asset references in `index.html` must be relative, not absolute. Audit and fix.
- Add a `make deploy-web` target (or document the steps) that copies `web/dist/` to a publishable location.
- Document the three intended deployment modes in a `web/README.md`:
  1. **With `mapture serve`** — the Go binary embeds and serves `web/dist/`, hits `/api/*` endpoints, SSE reload. This is the developer's local workflow.
  2. **With `mapture export-html`** — a self-contained directory with embedded data. This is the "share with the team" workflow.
  3. **Hosted standalone** — `web/dist/` deployed to static hosting. Users load data via file drop or `?data=<url>`. This is the "org-wide architecture viewer" workflow. CI pushes `data.json` to a known URL; the hosted UI loads it.
- CI integration example in the README: a GitHub Actions snippet that runs `mapture validate --format json > data.json`, uploads `data.json` as an artifact, and the hosted UI loads it via `?data=`.

**Acceptance:**
- `web/dist/` served from a subdirectory (`/mapture/`) loads correctly with no broken asset paths.
- `web/README.md` documents all three deployment modes with concrete commands.

**Depends on:** Story 1 (file picker + URL param loading).

---

### Story 4: Graceful partial payload handling

**What:** Define the minimum viable payload the frontend will accept and make sure degradation is smooth, not broken.

**Why:** Third-party tools, CI pipelines, and quick prototypes may produce only a graph (nodes + edges) without catalog, validation, or UI config. The explorer should still render — just without catalog-enriched labels or diagnostics.

**Scope:**
- Define two payload tiers:
  1. **Minimal** — a bare `graph.Graph` object: `{ nodes: [...], edges: [...] }`. No catalog, no validation, no meta. This is what a third-party tool or a quick `jq` pipeline might produce.
  2. **Full** — the complete `ExplorerPayload` with all sections. This is what `mapture serve` and `export-html` produce.
- In the boot chain (Story 1), when a loaded JSON lacks the `catalog` key, wrap it:
  ```ts
  { schemaVersion: 1, graph: <the JSON>, catalog: { teams: [], domains: [], events: [] },
    validation: { diagnostics: [], summary: { errors: 0, warnings: 0, nodes: 0, edges: 0 } },
    ui: {}, meta: { projectId: '', sourceLabel: filename, mode: 'offline' } }
  ```
- Frontend degradation behavior when catalog is empty:
  - Domain filter still works (derives domains from `node.domain` strings) but shows raw IDs instead of display names.
  - Team filter still works (derives owners from `node.owner` strings) but shows raw IDs.
  - Event details panel shows "no catalog data" instead of event metadata.
  - Diagnostics panel shows "no validation data" instead of an empty table.
- Add 2-3 example minimal JSON files under `examples/` or `web/examples/` that demonstrate the minimal payload shape. These double as test fixtures for the file-loading path.

**Acceptance:**
- Drop a file containing `{ "nodes": [{"id":"service:a","type":"service","name":"A"}], "edges": [] }` into the explorer → renders one node, no errors.
- All catalog-dependent UI elements show graceful fallbacks, not broken states or JS errors.
- The minimal payload shape is documented (inline JSDoc on the type or in `web/README.md`).

**Depends on:** Story 1 (the file-loading path that detects and wraps bare graphs).

---

## Story dependency graph

```
Story 1: Boot chain + file loading
  ├── Story 2: export-html command
  ├── Story 3: Standalone web deployment
  └── Story 4: Partial payload handling
```

Story 1 is the foundation. Stories 2, 3, and 4 are independent of each other and can be worked in parallel after Story 1 lands.

## Relation to existing tasks

| Existing task | Relation |
|---|---|
| 008 (HTML exporter) | **Superseded by Story 2.** Story 2 is the concrete implementation plan; task 008 is the earlier sketch. Mark 008 done when Story 2 ships. |
| 013 (Stable graph JSON schema) | **Complementary.** Story 4 defines the minimal payload the frontend accepts; task 013 locks the full schema with CUE validation. Neither blocks the other, but they should share the same `schemaVersion` field. |
| 012 (Machine-readable reports) | **Independent.** The `--format json` output from validate is a diagnostics report, not an `ExplorerPayload`. They serve different consumers. |

## What this epic does NOT include

- **Server-side changes.** `mapture serve` and its `/api/*` endpoints stay exactly as they are. No new endpoints, no new flags.
- **New validation logic in the browser.** Validation runs in Go; the frontend only displays pre-computed results. Offline mode shows whatever diagnostics were in the payload at export time.
- **Authentication, multi-user, or persistence.** The web UI is stateless — it loads data, renders it, and forgets it on reload (except manual node positions, which are already stored in browser memory).
- **A formal DataSource abstraction layer.** The boot chain is a linear priority check, not a pluggable interface. If we ever need a plugin system (e.g., loading from S3, GCS, or a database), that's a separate task with its own justification.
