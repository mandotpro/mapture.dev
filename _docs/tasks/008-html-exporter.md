# Task 008: Static HTML Exporter & SPA Setup

## Goal
Establish the static web tooling pipeline for Mapture using SvelteKit (with `adapter-static`). Output a clean, decoupled Single Page Application (SPA) into a directory like `output/web/`, effectively retiring the brittle "one-file HTML string" approach in favor of robust, embedded, cleanly-routed static web artifacts.

## Context
A major requirement for architecture tools is providing an artifact that teams can easily share or host on internal doc sites without running the original CLI or Go binaries. By adopting SvelteKit + Svelte Flow (per **Task 011**), the UI can cleanly build into strict static CSS/HTML/JS assets. Go will embed these using `//go:embed`. Instead of trying to aggressively bundle everything into a single literal `.html` file natively, `export-html` will simply eject the bundled SPA directory alongside a static data payload.

## Requirements

### 1. Leverage SvelteKit + `adapter-static`
- Ensure the frontend web directory relies on `@sveltejs/adapter-static` configured strict SPA mode (no active node servers / SSR allowed).
- The compilation goal is to generate static bundles cleanly mapped (typically into a `build/` or `dist/` root).

### 2. Provide the Data Adapter Bridge
- Since the exported output must run "offline" statically without dynamically hitting `mapture serve` APIs, the UI must have an adapter pattern implemented. 
- If running exported, the Svelte application fetches a locally emitted `graph-data.json` relative configuration instead of a dynamic port.

### 3. Implement the `mapture export-html` Logic (`internal/exporter/html`)
- Update the command flags interface so `-o` correctly targets a directory (e.g., `-o output/web`).
- The exporter should:
  1. Parse the architecture payload via the standard internal graph generators.
  2. Unpack the embedded UI static assets from the embedded `src/internal/webui/dist/` Go file system onto the local disk target at `output/web/`.
  3. Emit the final `data.json` mapping output directly inside that `output/web/` folder for the isolated SPA to consume entirely self-contained.

## Definition of Done
- `mapture export-html examples/demo -o output/web` completely outputs standard SvelteKit SPA static assets.
- Local engineers can spin up a basic static file watcher over `output/web/` or drop it into GitHub Pages seamlessly.
- Svelte Flow behaves natively utilizing the static graph JSON file.
- We have completely dropped attempts to over-optimize code into a single, fragile monolithic HTML file payload for v0.1.
