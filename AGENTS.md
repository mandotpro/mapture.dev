# AGENTS.md

This file is the single source of truth for AI coding agents working in this repository — Claude Code, Codex, Cursor, Copilot, Aider, Jules, Windsurf, and any future tool that reads [agents.md](https://agents.md). Tool-specific files like `CLAUDE.md` are symlinks to this file; do not edit them directly.

Mapture is an MIT-licensed, single-binary, repo-native architecture graph tool written in Go. It turns a small YAML catalog plus `@arch.*` / `@event.*` comment tags in source files into a normalized architecture graph, an interactive UI, static HTML, and an AI-ready bundle.

**The canonical product spec is [`_docs/mapture-dev-prd-v1.md`](./_docs/mapture-dev-prd-v1.md). Read the relevant PRD section before changing behavior** — section numbers are cited inline throughout the code (e.g. `// See PRD §13.2`) and should stay that way so intent is traceable.

## Commands

```bash
go mod tidy                  # after editing go.mod / adding deps
go build ./src/...          # compile the application packages
go vet ./src/...            # static checks
go run src/main.go --help   # smoke-test the CLI
go run src/main.go validate examples/demo   # dogfood against the bundled example

go test ./src/...           # (no tests yet — scaffold stage)
go test ./src/internal/catalog -run TestLoad   # single test, once tests exist
```

The `examples/demo/` tree is the canonical fixture: the minimal end-to-end example from PRD §33 (catalog YAMLs + annotated PHP/Go/TS sources). Use it as the test fixture rather than inventing new ones.

## Architecture

Three layers, all normalized through one graph model:

1. **Catalog** (`src/internal/catalog`) — YAML files under `architecture/` (`teams.yaml`, `domains.yaml`, `events.yaml`) are the source of truth for teams, domains, and events. Comments reference catalog IDs; they do not redefine them. See PRD §13.
2. **Scanner** (planned, `src/internal/scanner`) — walks include paths, parses flat `@arch.*` / `@event.*` tag comments, attaches each block to a nearby source location, and emits typed nodes/edges. Comments-only in v0.1 — no AST or Tree-sitter. See PRD §14, §22.
3. **Graph** (`src/internal/graph`) — the normalized `Node`/`Edge`/`Graph` model is the shared payload between scanner output, validator input, and every exporter. Node identity is `type:name` (e.g. `service:checkout-service`) across the entire pipeline. See PRD §17.

`src/cmd/root.go` is wiring only: Cobra registers seven subcommands (`init`, `validate`, `scan`, `graph`, `serve`, `export-html`, `export-ai`) that currently dispatch to `todo()` stubs. Real logic lands in `src/internal/*` as features are built — grow `src/internal/*`, not `src/cmd/`.

### Packages that do not exist yet (deliberately)

v0.1 starts small (PRD §30 risk: "too much schema complexity too early"). When you need one of these, create the package under `src/internal/` and cite the PRD section in the doc comment. **Keep these exact names** — `src/cmd/root.go` and future docs assume them:

- `src/internal/config` — loads `mapture.yaml`. PRD §12.
- `src/internal/scanner` — comment parser + source attachment. PRD §14, §22.
- `src/internal/validator` — six-layer validation (config → catalog → comment shape → catalog consistency → attachment → graph). PRD §15.
- `src/internal/server` — local HTTP explorer UI. PRD §10, §18.
- `src/internal/exporter/mermaid` — Mermaid flowchart. PRD §18.
- `src/internal/exporter/html` — self-contained HTML report. PRD §10.
- `src/internal/exporter/ai` — `.mapture/ai/` bundle. PRD §19.

### Design invariants

- **Catalog is the source of truth.** The validator rejects unknown team / domain / event IDs referenced from comments (PRD §15 layer 4). Don't add code paths that silently tolerate unknown IDs.
- **Comments are flat `@key value` tags, not JSON.** Do not introduce structured JSON/YAML inside comments. PRD §14.
- **Node IDs are `type:name`.** This is the stable identity across graph, exports, and AI bundles. Never strip the prefix.
- **One binary, no runtime deps.** Frontend assets must be embedded via `embed` when the server/HTML exporter lands. PRD §9.
- **Comments-first, not static-analysis-first.** Tree-sitter / AST parsing is explicitly a v0.4+ enhancement. PRD §22.
- **v1 enums are closed** (PRD §16). Constants for node/edge types live in `src/internal/graph/graph.go`; other enums belong in the validator when it's created. Extending an enum requires updating both the code constant and the PRD section.
- **Don't leap milestones.** Features have assigned versions in PRD §29 (v0.1 → v1.0). If a change belongs to a later milestone, defer it and note why.

## Naming note

The PRD body uses the working name **ArchMap**, but the repo, the PRD title, and the intended site (`mapture.dev`) all use **Mapture**. Scaffolding commits to `mapture` as the binary and module name. Rename blast radius if this turns out wrong: `go.mod` module path, `src/main.go` import, CLI `Use` in `src/cmd/root.go`, the `.mapture/` artifact directory, and this file. Confirm with the user before renaming.

## Project rules

Rules and conventions the team has adopted as the project grows. Managed by the `agent-docs` skill — invoke it when adding or updating a rule here so AGENTS.md stays organized and tool-specific symlinks stay correct.

- **Cite PRD sections in doc comments** when implementing a feature (`// See PRD §15 layer 3.`). Keeps intent traceable back to the spec.
- **`src/cmd/` is wiring only.** Subcommand files should parse flags and delegate to `src/internal/*`. Business logic in `src/cmd/` is a code smell.
- **Public OSS project.** Every user-facing string, error message, and README section is read by strangers. Write accordingly.

<!-- agent-docs:rules:end -->
