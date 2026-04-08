# Task 011: Svelte Flow Frontend Implementation (Graph Explorer)

## Goal
Implement the initial Svelte Flow frontend for Mapture. The goal is a pleasant architecture explorer—not a generic node editor—that successfully loads and visually explores normalized graph model data produced by the Go backend.

## Context
Mapture is a Go-based architecture mapping tool. The backend already has APIs implemented and produces mapping output files. Your job is to build the Svelte Flow UI by reading the existing backend/API contracts and wiring the frontend to real data instead of inventing a new backend shape.

## Requirements

### 1. Discover the Current Ecosystem
Inspect the existing codebase and identify:
- All currently implemented HTTP/API endpoints.
- Any existing graph/mapping output files and their schemas.
- Any existing serve/static file setup.
- Any current UI/frontend scaffolding, if present.

### 2. Backend Constraint
Do not redesign the backend unless absolutely necessary.
- Prefer adapting the Svelte Flow frontend to the existing API/output shape.
- Only propose backend changes if there is a clear blocker, and keep them minimal.
- Reuse current graph/mapping models whenever possible.

### 3. Build an Interactive Svelte Flow UI
Build a first usable interactive UI using Svelte Flow focusing on exploring architecture.

**Primary UI goals:**
- Load graph data from existing APIs or mapping output files.
- Display nodes and edges in Svelte Flow.
- Support basic node types such as `service`, `api`, `database`, and `event` (if they exist).
- Show labels clearly.
- Support zoom/pan/fit view.
- Show a details panel when a node is selected.
- Display source references if available (file, line, domain, owner, summary, tags, etc.).
- Support basic filtering by node type, domain, and owner if that data exists.
- Support search by node ID or label.
- Keep the implementation simple, maintainable, and production-lean.

**Important Data Handling Requirements:**
- Read the existing APIs first.
- Read the mapping output files first.
- Derive the frontend data adapter from the real backend data shape.
- Do not hardcode fake graph data except possibly as a temporary fallback during development.
- Keep the graph model renderer-agnostic where possible.
- Introduce a clear adapter layer:
  `backend/output format -> frontend normalized graph model -> Svelte Flow nodes/edges`

**Implementation Details:**
- Use Svelte Flow for rendering.
- Build reusable custom node renderers only where they add clear value.
- Keep styling clean and minimal.
- Prefer simple layout first; if there is no backend layout info, implement a basic automatic layout strategy.
- If useful, use `dagre` or `elkjs` for layout, but only if needed.
- Avoid overengineering.

## Expected Deliverables
1. **A short findings summary** describing:
   - which APIs exist
   - which output files/schemas exist
   - how the frontend consumes them
   - any gaps or blockers found
2. **A working Svelte Flow-based UI** integrated into the existing app.
3. **A normalized frontend graph adapter module**, for example:
   - `loadGraphFromApi(...)`
   - `loadGraphFromFile(...)`
   - `normalizeGraph(...)`
   - `toSvelteFlowNodes(...)`
   - `toSvelteFlowEdges(...)`
4. **Basic UI features:** graph canvas, filters, search, node details panel, loading and error states.
5. **A short implementation note** covering:
   - assumptions made
   - minimal backend changes needed, if any
   - next recommended improvements

## Engineering Constraints
- Prefer incremental changes.
- Respect existing project structure and conventions.
- Do not replace the existing serving model.
- Keep compatibility with the Go `serve` command and existing static asset flow.
- Make it easy to embed/build with the current binary-serving approach.

## Quality Bar
- The UI should feel like an architecture explorer.
- The code should be understandable and easy to extend.
- The implementation should be grounded in what already exists in the repository.

## Execution Plan
- First inspect backend APIs and output files.
- Then define the frontend adapter.
- Then implement the Svelte Flow UI.
- Then wire filters/search/details.
- Then document findings and any follow-up tasks.

*Note: If the codebase already contains partial UI code, extend it instead of replacing it. If the backend exposes more than one possible graph endpoint/output, choose the one closest to the canonical mapping output and explain why.*
