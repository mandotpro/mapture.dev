# AGENTS.md

This file is the single source of truth for AI coding agents working in this repository — Claude Code, Codex, Cursor, Copilot, Aider, Jules, Windsurf, and any future tool that reads [agents.md](https://agents.md). Tool-specific files like `CLAUDE.md` are symlinks to this file; do not edit them directly.

Mapture is an MIT-licensed, single-binary, repo-native architecture graph tool written in Go. It turns a small YAML catalog plus `@arch.*` / `@event.*` comment tags in source files into a normalized architecture graph, an interactive UI, static HTML, and an AI-ready bundle.

The canonical product spec lives in [`_docs/mapture-dev-prd-v1.md`](./_docs/mapture-dev-prd-v1.md). Consult it before changing behavior, but do not add spec-section traceability comments to the code.

## Commands

```bash
go mod tidy                  # after editing go.mod / adding deps
go build ./src/...          # compile the application packages
go vet ./src/...            # static checks
go run src/main.go --help   # smoke-test the CLI
go run src/main.go validate examples/demo   # validate config, scan comments, and build the graph for the bundled example
go run src/main.go scan examples/ecommerce  # extract raw comment blocks from the polyglot fixture
go run src/main.go graph examples/demo      # render Mermaid from the built graph

./scripts/test-go.sh        # run Go tests via gotestsum with AI-friendly output
./scripts/lint-go.sh        # run golangci-lint against src/
./scripts/check-fmt.sh      # enforce gofmt on src/ and fail if files were changed
go test ./src/internal/catalog -run TestLoad   # single test, once tests exist

make help                   # discover the repo's day-to-day commands
make test                   # run the full verification suite
make lint                   # run the Go lint suite
make web                    # rebuild the frontend bundle under src/internal/webui/dist/
make testing-help           # show the testing-first wrapper commands and fixture paths
make testing-build          # build the current source into testing/bin/mapture
make testing-ecommerce-web  # run the web UI for the polyglot fixture from testing/
make testing-demo-scan      # write testing/outputs/demo.scan.json
make validate-demo          # validate the canonical demo fixture

./scripts/build.sh          # build build/mapture for local development
./scripts/test.sh           # run tests, vet, and CLI smoke checks
./scripts/test-go.sh --install-only   # install gotestsum into testing/tools/bin
./scripts/init-hooks.sh     # configure the repo-managed git hooks once per clone
./scripts/go.sh help        # show the testing wrapper, fixtures, and output paths
./scripts/go.sh init        # build into testing/ and run the playground wrapper
./scripts/go.sh web ecommerce  # open the web UI against examples/ecommerce
```

The `examples/demo/` tree is the canonical fixture: the minimal end-to-end example with catalog YAMLs and annotated PHP/Go/TS sources. Use it as the test fixture rather than inventing new ones.

## Architecture

Three layers, all normalized through one graph model:

1. **Config + Schema** (`src/internal/config`, `src/internal/schema`) — `mapture.yaml` and catalog YAML files are validated and decoded through embedded CUE schemas before the rest of the pipeline runs.
2. **Catalog** (`src/internal/catalog`) — YAML files under `architecture/` (`teams.yaml`, `domains.yaml`, `events.yaml`) are the source of truth for teams, domains, and events. Comments reference catalog IDs; they do not redefine them.
3. **Scanner** (`src/internal/scanner`) — walks include paths, parses flat `@arch.*` / `@event.*` tag comments from Go, PHP, TS, and JS comment forms, and emits raw blocks with file/line attachment. Comments-only in v0.1 — no AST or Tree-sitter.
4. **Validator** (`src/internal/validator`) — enforces catalog cross-references, builds the normalized graph, and emits diagnostics for layers 4-6.
5. **UI** (`src/internal/ui`) — owns shared CLI presentation rules so commands report stages, warnings, errors, and summaries consistently in TTY and plain-text environments.
6. **Exporter** (`src/internal/exporter/mermaid`) — renders the normalized graph as deterministic Mermaid flowcharts with optional domain/team/type filters.
7. **Graph** (`src/internal/graph`) — the normalized `Node`/`Edge`/`Graph` model is the shared payload between scanner output, validator input, and every exporter. Node identity is `type:name` (e.g. `service:checkout-service`) across the entire pipeline.
8. **Frontend bundle** (`src/internal/webui`) — the explorer UI lives under `src/internal/webui/frontend/` as a Svelte Flow app built by Vite. `make web` installs the frontend dependencies from the committed lockfile, writes the bundle to `src/internal/webui/dist/`, and the `src/internal/webui` Go package embeds that directory via `//go:embed` for `src/internal/server` and future exporters. `src/internal/webui/dist/` is committed; rerun `make web` after editing anything under `src/internal/webui/frontend/`.

`src/cmd/root.go` is wiring only: Cobra registers seven subcommands (`init`, `validate`, `scan`, `graph`, `serve`, `export-html`, `export-ai`). `init`, `validate`, `scan`, `graph`, and `serve` delegate into `src/internal/*`; the remaining export commands are still stubs.

### Packages that do not exist yet (deliberately)

v0.1 starts small to avoid pulling in too much schema complexity too early. When you need one of these, create the package under `src/internal/`. **Keep these exact names** — `src/cmd/root.go` and future docs assume them:

- `src/internal/config` — loads `mapture.yaml`.
- `src/internal/schema` — embeds CUE definitions for config and catalog validation.
- `src/internal/ui` — shared CLI reporting and output styling.
- `src/internal/server` — local HTTP explorer UI.
- `src/internal/exporter/html` — self-contained HTML report.
- `src/internal/exporter/ai` — `.mapture/ai/` bundle.

### Design invariants

- **Catalog is the source of truth.** The validator rejects unknown team / domain / event IDs referenced from comments. Don't add code paths that silently tolerate unknown IDs.
- **Event usage blocks are not event definitions.** `@event.domain` on listeners, bridges, publishers, and subscribers describes the usage site; only definition blocks should be forced to match the catalog event domain/owner.
- **Comments are flat `@key value` tags, not JSON.** Do not introduce structured JSON/YAML inside comments.
- **CLI output must go through `src/internal/ui`.** Keep stage headers, warnings, errors, summaries, and path formatting centralized instead of scattering `fmt.Printf` formatting across commands.
- **Node IDs are `type:name`.** This is the stable identity across graph, exports, and AI bundles. Never strip the prefix.
- **One binary, no runtime deps.** Frontend assets must be embedded via `embed` when the server/HTML exporter lands.
- **Comments-first, not static-analysis-first.** Tree-sitter / AST parsing is explicitly a later enhancement.
- **v1 enums are closed.** Constants for node/edge types live in `src/internal/graph/graph.go`; other enums belong in the validator when it's created. Extending an enum requires updating the matching docs and code together.
- **Don't leap milestones.** If a change clearly belongs to a later milestone, defer it and note why.

## Naming note

The original product spec uses the working name **ArchMap**, but the repo and intended site (`mapture.dev`) use **Mapture**. Scaffolding commits to `mapture` as the binary and module name. Rename blast radius if this turns out wrong: `go.mod` module path, `src/main.go` import, CLI `Use` in `src/cmd/root.go`, the `.mapture/` artifact directory, and this file. Confirm with the user before renaming.

## Project rules

Rules and conventions the team has adopted as the project grows. Managed by the `agent-docs` skill — invoke it when adding or updating a rule here so AGENTS.md stays organized and tool-specific symlinks stay correct.

- **Comments must earn their keep.** Add comments only when they explain behavior, tradeoffs, or non-obvious intent that helps humans and agents maintain the code. Do not add traceability comments that only point at spec sections.
- **`src/cmd/` is wiring only.** Subcommand files should parse flags and delegate to `src/internal/*`. Business logic in `src/cmd/` is a code smell.
- **Top-level `scripts/` is for repo operations.** Build, release, and CI helper scripts belong there, not in `src/`.
- **Testing flows must be one command away.** Prefer `make testing-*` and `./scripts/go.sh <feature> <fixture>` wrappers over telling contributors to remember raw `go run` paths, ports, or output files.
- **Pre-commit must stay fast and structural.** It should auto-run formatting checks plus linting, but leave the full example-backed test gauntlet to pre-push and CI.
- **Pre-push and CI must exercise `examples/`.** Broken fixtures under `examples/invalid/` are part of the guardrail suite and should fail predictably.
- **Public OSS project.** Every user-facing string, error message, and README section is read by strangers. Write accordingly.
- **One frontend, committed bundle.** The explorer UI lives in `src/internal/webui/frontend/` and is bundled to `src/internal/webui/dist/` by `make web`. Always commit the regenerated `src/internal/webui/dist/` together with any frontend source change so `go build` alone produces a working binary. Never hand-edit files under `src/internal/webui/dist/`.

<!-- agent-docs:rules:end -->
