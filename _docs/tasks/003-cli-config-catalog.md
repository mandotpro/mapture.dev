# Task 003: CLI skeleton, Config, and Catalog Loader

## Goal
Establish the foundational command line structure for `mapture` and implement robust parsing for both the repository configuration (`mapture.yaml`) and the central architecture catalogs (`teams.yaml`, `domains.yaml`, `events.yaml`).

## Context
Before we can scan code for architecture comments, the tool must be able to load its configuration to know *where* to look, and it must load the canonical catalogs to know *what* exists. This covers validation layers 1 and 2 from `validation-layers.md`.

## Requirements

### 1. CLI Skeleton (`src/cmd/`)
- Initialize a `cobra` command-line application (if not already present).
- Define root command (`mapture`).
- Scaffold subcommands: `init`, `validate`, `scan`, `graph`, `serve`, `export-html`, `export-ai`.
- Subcommands should just output a `TODO: not implemented` message for now, except for wiring them up to print help.

### 2. Config Parser (`src/internal/config`)
- Utilize the `cuelang` validation wrapper developed in Task **001.1** to assert schema definitions.
- Implement a location discovery loop that walks up the tree from the current working directory to find `mapture.yaml`.
- Wire the CUE parser errors into human-readable outputs and halt execution.
- Inject default configuration settings (e.g., `catalog.dir` -> `./architecture`) explicitly inside Go or within the `.cue` default values.

### 3. Catalog Loader (`src/internal/catalog`)
- Create data models matching `_docs/types/catalog-schemas.md` for Team, Domain, and Event.
- Load the files from the directory specified by `catalog.dir`.
- Unify files tightly via CUE to instantly execute schema layout assertions:
  - Check for uniqueness of all IDs.
  - Verify regex forms for IDs.
- Execute **Validation Layer 2 (Catalog cross-references)** structurally in Go or merged CUE:
  - Domain `ownerTeams` must map to valid IDs inside `teams.yaml`.
  - Event `ownerTeam` and `domain` must map to valid elements inside `teams.yaml` and `domains.yaml`.
- Expose a parsed and type-safe `Catalog` cache representation.

## Definition of Done
- `mapture validate examples/demo` successfully locates the config and loads the catalogs without errors.
- Any invalid YAML or missing catalog file results in a structured error logging (Validation Layer 1).
- Duplicate IDs in catalogs throw clear Validation Layer 2 errors.
- Unit tests cover configuration defaults and catalog schema edge cases.
