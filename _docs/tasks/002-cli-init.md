# Task 002: CLI `init` Command

## Goal
Implement the `mapture init [path]` command to safely bootstrap a target directory with empty example files for `mapture.yaml`, `teams.yaml`, `domains.yaml`, and `events.yaml`.

## Context
When a user wants to adopt Mapture, they need a starting point. The `init` command should generate the repository structure required by the validation layers without forcing the user to copy-paste.

## Requirements

### 1. `init` command interface
- Implemented inside `src/cmd/` (or `cmd/`).
- Accepts an optional `[path]` argument (defaults to `.`).
- Safe execution: must check if any target files already exist and gracefully skip or merge, ensuring it never overwrites existing catalogs or configs.

### 2. Scaffold `mapture.yaml`
- Write out a full default `mapture.yaml` to the target path.
- The file should include helpful commented-out text explaining the fields (especially `scan.include`/`exclude` and `catalog.dir`).

### 3. Scaffold Catalog directory
- Create the default `catalog.dir` (typically `./architecture/`).
- Generate `teams.yaml` with 2 example teams.
- Generate `domains.yaml` with 2 example domains referencing those teams.
- Generate `events.yaml` with 1 example event showcasing the schema.
- Content of these files should exactly match the "Example" structures in `_docs/types/catalog-schemas.md`.

## Definition of Done
- Creating an empty directory and running `go run src/main.go init .` creates the exact 4 files necessary to have a working, layer-1 compliant Mapture structure.
- Running `init .` a second time prints "Warning: teams.yaml already exists, skipping".
- The examples generated must pass the upcoming validation steps unmodified.
