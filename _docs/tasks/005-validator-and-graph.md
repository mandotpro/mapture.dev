# Task 005: Graph Model and Strict Validations

## Goal
Transform the valid intermediate comment blocks into the normalized Graph Model, run all cross-referencing validation layers, and finalize output rendering for the `validate` command.

## Context
Once comments are scanned successfully (Layer 3), we need to ensure the assertions in those comments agree with the catalog and each other. This task ties the scanner outputs to the catalog inputs, converting everything to a standard `internal/graph` structure, fulfilling PRD §15 (Layers 4-6) and §17.

## Requirements

### 1. Internal Graph Model (`internal/graph`)
- Implement representations matching `_docs/types/graph-model.md` for `Node`, `Edge`, and `Graph`.
- Expose an interface to add Nodes and Edges incrementally.

### 2. Validation Layer 4 (Catalog cross-references) (`internal/validator`)
- Match the scanned blocks against the loaded `Catalog`.
- Throw errors/warnings according to rules defined:
  - `@arch.domain` and `@arch.owner` exist in actual catalogs.
  - `@event.id` matches an actual event.
  - Assert that domains match the definitions, flagging deprecations.

### 3. Validation Layer 5 (Source attachment) (Stub)
- In v0.1, we only issue warnings if no standard file structures surround comments. Implement a simple placeholder or skip entirely if language-agnostic parsing makes this difficult initially.

### 4. Validation Layer 6 (Graph integrity)
- Once the `Graph` is constructed, trace node IDs.
- Fail on duplicate generated node IDs.
- Fail if an edge `to` references a node ID that doesn't exist anywhere in the built graph (subject to `failOnUnknownNode` config).

### 5. Validate reporting
- Construct the structured `JSON` error and warning representations.
- Connect the `validate` command in standard format (non-zero exits upon error codes).

## Definition of Done
- `mapture validate examples/demo` successfully integrates Config, Catalog, Scanner, and Graph steps. It finishes with an exit 0 and outputs the graph size.
- Intentionally adding a bad cross-reference (e.g. an `@arch.domain` to a non-existent value) in `examples/demo` causes a Layer 4 error and a non-zero exit code.
- Implement exhaustive unit test suites covering the validation layers individually using mock "Raw Blocks" and Catalogs.
