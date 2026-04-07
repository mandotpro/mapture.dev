# Task 003: CLI skeleton, Config, and Catalog Loader

## Goal
Establish the foundational command line structure for `mapture` and implement robust parsing for both the repository configuration (`mapture.yaml`) and the central architecture catalogs (`teams.yaml`, `domains.yaml`, `events.yaml`).

## Context
Before we can scan code for architecture comments, the tool must be able to load its configuration to know *where* to look, and it must load the canonical catalogs to know *what* exists. This fulfills PRD §12 and §13, and Validation Layers 1 & 2 from `validation-layers.md`.

## Requirements

### 1. CLI Skeleton (`src/cmd/`)
- Initialize a `cobra` command-line application (if not already present).
- Define root command (`mapture`).
- Scaffold subcommands as defined in PRD: `init`, `validate`, `scan`, `graph`, `serve`, `export-html`, `export-ai`.
- Subcommands should just output a `TODO: not implemented` message for now, except for wiring them up to print help.

### 2. Config Parser (`src/internal/config`)
- Create standard structs representing the schema in `_docs/types/config-schema.md`.
- Implement a discovery function that walks up from the current directory to find `mapture.yaml`.
- Provide a strict YAML parsing step that halts on unknown keys or invalid schema types.
- Apply default values (e.g., `catalog.dir` defaults to `./architecture`).

### 3. Catalog Loader (`src/internal/catalog`)
- Create data models matching `_docs/types/catalog-schemas.md` for Team, Domain, and Event.
- Load the files from the directory specified by `catalog.dir`.
- Execute **Validation Layer 2 (Catalog internal consistency)**:
  - Check for uniqueness of all IDs.
  - Verify `[a-z0-9-]+` formats for IDs.
  - Check cross-references: Domain `ownerTeams` must exist in Teams.
  - Check Event `ownerTeam` and `domain`.
- Provide a strongly typed `Catalog` struct containing maps indexable by `id` for fast lookup during later stages.

## Definition of Done
- `mapture validate examples/demo` successfully locates the config and loads the catalogs without errors.
- Any invalid YAML or missing catalog file results in a structured error logging (Validation Layer 1).
- Duplicate IDs in catalogs throw clear Validation Layer 2 errors.
- Unit tests cover configuration defaults and catalog schema edge cases.
