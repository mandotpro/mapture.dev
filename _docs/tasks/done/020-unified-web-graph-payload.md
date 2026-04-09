---
id: 020
title: Unify web graph payload and remove frontend normalization tax
milestone: v0.3.0
status: done
prd: §10, §17, §18, §29
depends_on: [005, 010, 013]
---

## Why
The current web explorer pays an unnecessary complexity tax because the
backend exposes multiple overlapping response shapes:

- `/api/graph` returns a raw graph shape
- `/api/validate` returns a validator result shape
- `/api/catalog` returns catalog + UI metadata

The frontend then has to fetch multiple endpoints, guess which payload
shape it received, normalize them back into one model, and keep
recomputing mappings on every render. This creates:

- duplicate fetches for the same screen load
- repeated graph/catalog/data normalization in the browser
- extra runtime branching for live vs static vs imported payloads
- fragile code around `/api/graph` vs `/api/validate` differences
- unnecessary bundle complexity and wasted recomputation

The API should return one documented, UI-ready payload that the web app
can consume directly with minimal mapping.

## Scope
- Introduce one canonical web payload contract for explorer use, for example:
  ```json
  {
    "schemaVersion": 1,
    "graph": {...},
    "catalog": {...},
    "validation": {
      "diagnostics": [...],
      "summary": {...}
    },
    "ui": {
      "nodeColors": {...}
    },
    "meta": {
      "projectId": "...",
      "sourceLabel": "...",
      "mode": "live|offline"
    }
  }
  ```
- Add a dedicated internal type for this response under `src/internal/server`
  or a shared package if exporters also need it.
- Make the explorer load from one endpoint only, rather than stitching
  together `/api/graph`, `/api/validate`, and `/api/catalog`.
- Keep `/api/events` for live reload, but make the reload path refetch
  the same canonical payload endpoint.
- Remove frontend-side payload guessing and collapse the current
  `normalizePayload` / shape-reconciliation logic to a thin adapter.
- Ensure static/exported payload injection uses the exact same contract
  as the live server response.
- Reduce frontend recomputation:
  - do not rebuild team/domain/event lookup maps in multiple places
  - do not re-wrap graph payloads just to recover a stable model
  - prefer backend-prepared summaries/metadata when they are already known
- Decide the future of the old endpoints:
  - either deprecate `/api/graph` and `/api/catalog`
  - or keep them as compatibility/debug endpoints, but ensure the web UI
    no longer depends on them

## Implementation Notes
- The canonical payload should be versioned from day one.
- This task should align with the stable graph schema work in task 019,
  but it is broader than graph JSON alone because the explorer also
  needs catalog, diagnostics, UI config, and meta.
- The frontend should still be allowed to derive view-only state such as
  visible-node counts, active filters, and layout state, but not to
  repair backend response shape mismatches.
- If a summary block is introduced, it should include at least:
  - error count
  - warning count
  - total nodes
  - total edges

## Acceptance
- The explorer performs a single data fetch for initial load
- Live reload refetches the same canonical payload endpoint only
- Static/exported payload injection reuses the exact same JSON contract
- The frontend no longer contains branching logic to detect whether
  `/api/graph` returned a raw graph or wrapped validation payload
- Removing the old normalization glue reduces frontend data-mapping code
  materially and keeps rendering behavior unchanged
- Tests cover:
  - canonical payload serialization
  - live server response shape
  - static/offline payload compatibility
  - frontend boot using only the canonical payload
