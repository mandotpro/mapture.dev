# Task 001: Project Restructure

## Goal
Move all Go implementation code (`cmd`, `internal`, `main.go`) into a dedicated `src/` directory, while leaving the `examples/` directory outside of `src/` to serve as a clean, predictable integration testing and demonstration sandbox.

## Context
By isolating the Go source inside `src/` and moving `examples/` alongside it in the repository root, we ensure that the examples and demo catalogs are completely decoupled from the binary's internal structure. This acts as a reliable test bed to validate commands like `mapture init` and `mapture validate` without polluting the core project codebase.

## Requirements

### 1. Structure the `src/` directory
- Create `src/` at the repository root.
- Move the `cmd/` and `internal/` packages into `src/`.
- Move `main.go` into `src/`.
- Update `go.mod` (if necessary, though the module name `mapture.dev` should remain the same; typical Go projects might use `src/` occasionally but typically leave the root `go.mod` in place or move it down to `src/`). Ensure Go build paths still resolve (e.g. `go build ./src/...`).

### 2. Protect the `examples/` directory
- Ensure `examples/` remains at the repo root.
- The `examples/demo` folder serves as the ultimate E2E test target. Mapture commands run from `src/` will target `../examples/demo`.

## Definition of Done
- All application source files live neatly inside `/src`.
- We can still run `go run src/main.go` from the root, passing it `examples/demo` for parsing.
- Integration tests or scripts cleanly separate "the tool" (`src/`) from "the target" (`examples/`).
