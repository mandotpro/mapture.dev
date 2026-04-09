# Task 006: Rich CLI Output Framework

## Goal
Implement a unified, standardized terminal UI (TUI) layer using modern styling (such as `charmbracelet/lipgloss` or `pterm`) to give commands like `mapture validate` clean, readable, and highly informative output.

## Context
As `mapture validate` traverses from fetching the configuration (Layer 1) to full Graph connection (Layer 6), it handles hundreds of files and references. A plain text dump of log files is overly hostile to users. To ensure wide adoption, the tool must act like a modern CLI framework, presenting clean summaries, nicely formatted file-path traces, and standardized iconography for success, warnings, and errors.

## Requirements

### 1. Unified Reporter / UI Package (`src/internal/ui`)
- Create a dedicated Go package that exposes structured printing capabilities instead of scattering `fmt.Printf` or `log.Fatal` everywhere.
- Abstract the styling for:
  - **Headers/Stages:** e.g., `[✓] Loaded catalogs` or `[⚡] Scanning sources...`
  - **Warnings:** Unified yellow text/icon indicating non-fatal issues (like a deprecation use).
  - **Errors:** Unified red text/icon indicating compilation stops.
  - **Paths:** Specifically highlight relative file paths and line numbers so IDE users can clearly read where errors occurred.

### 2. Enhanced `mapture validate` Output
- Update the `validate` command execution to use this structured Reporter.
- Rather than panicking on the very first error (where possible), aggregate non-breaking errors and print them cleanly as a compiled list at the end of the run.
- Display a robust status summary footer: E.g., `Validation Failed: 3 Errors, 1 Warning found in 24 files`.

### 3. Consistent Tooling Aesthetics
- Ensure the color palettes, indents, and visual hierarchies match the interactive wizard used in `mapture init` (Task 002).
- If running in CI environments (e.g. without TTY support), safely degrade the rich colors to clean plaintext structures to prevent ANSI artifact bugs.

## Definition of Done
- Commands like `validate` now print beautiful, step-by-step progress natively.
- Any simulated bad reference gracefully ends the CLI process with a highly legible, colored error pointing explicitly to the offending `filename` and `line number`.
- The `src/internal/ui` provides a clear internal interface that future commands (like `serve` or `graph`) can instantly leverage for standardized consistency.
