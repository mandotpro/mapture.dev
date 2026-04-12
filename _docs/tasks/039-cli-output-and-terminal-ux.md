---
id: 039
title: CLI output polish, color system, and terminal UX
milestone: v0.5.0
status: todo
prd: §12, §29
depends_on: [030]
---

## Goal
Make the Mapture CLI feel intentionally designed in the terminal and easier to trust at a glance:

- readable at a glance
- consistent across commands
- color-aware without becoming noisy
- safe for CI, pipes, and `NO_COLOR`
- explicit about version, channel, install source, and update state

The CLI should communicate progress, success, warnings, and failures clearly, while keeping machine-readable outputs untouched.

## Why
The current command surface is functionally solid, but it still reads as plain scaffolding in several places:

- stage/progress output is useful but visually flat
- validation diagnostics are harder to scan than they should be
- command success/failure states do not have a strong visual hierarchy
- install/update/version guidance is not obvious when users are on stale binaries or mixed channels
- help output does not clearly guide users toward the most useful commands
- users cannot easily tell whether they are already on the latest canary or stable release
- users cannot easily tell where the current binary came from

For a public CLI tool, this matters. The terminal is part of the product.

## Scope

### 1. Introduce one shared terminal styling layer
Create a single internal terminal formatting package or reporter layer used by command-oriented output.

It should own:

- ANSI color policy
- semantic styles like `success`, `warning`, `error`, `muted`, `heading`, `accent`
- glyph/icon policy with ASCII-safe fallback
- TTY detection
- `NO_COLOR` support
- forced `--color` policy if introduced later

This should replace ad hoc formatting decisions spread across command code.

### 2. Improve high-value command output first
Focus on commands users run most often:

- `validate`
- `serve`
- `graph`
- `export-json`
- `version`
- `update`

Expected improvements:

- stronger stage headings
- clearer success/error summaries
- better warning emphasis
- easier scanning of counts and result state
- tighter spacing and more consistent line structure

### 3. Productized help and version surfaces
`mapture --help`, `mapture help`, and `mapture --version` should behave like product entry points, not raw Cobra defaults.

Expected behavior:

- first line: `mapture.dev - <version>`
- second line: compact metadata such as channel, install source, and resolved executable path
- tagline aligned to the repo marketing copy
- short update suggestion when a newer build is available for the active channel

This richer header should appear on help and version only, not before every command.

### 4. Keep machine-readable outputs clean
Human formatting must never leak into structured outputs.

That means:

- JSON written by `export-json` stays pure JSON
- Mermaid output stays plain Mermaid
- only human-facing status/progress/reporting paths use colors and terminal formatting

### 5. Improve diagnostics readability
Validation diagnostics should become easier to scan in terminal output.

Good improvements include:

- color by severity
- stable severity badges
- clearer file/line formatting
- slightly better grouping or spacing
- summary at the end that matches the same style system

Do not make diagnostics overly decorative. The goal is speed of reading.

### 6. Cover update/install channel messaging
The CLI should communicate release-channel state clearly where relevant.

Examples:

- `mapture update` should clearly say whether it is updating `stable` or `canary`
- `mapture --help` and `mapture --version` should tell the user whether a newer version is available
- stale install guidance should be easier to understand
- if a user is on an older binary without newer commands, the docs should make the expected upgrade path obvious

This task does not need to solve impossible cases inside already-old binaries, but the current product/docs should reduce confusion going forward.

## Acceptance

- human-facing CLI commands use a shared formatting layer instead of ad hoc strings
- colors are enabled only when appropriate for a terminal context
- `NO_COLOR` disables color output cleanly
- CI/piped output remains readable and safe
- validation output is easier to scan for severity, location, and summary
- command help feels more intentional and more product-ready
- `help` and `version` clearly surface version, channel, install source, and upgrade status
- structured outputs remain unchanged and machine-safe

## Implementation Notes

- prefer subtle, restrained color use over rainbow output
- treat color as reinforcement, not the only source of meaning
- use semantic styling, not command-specific hard-coded ANSI escapes
- design for both dark and light terminals
- keep ASCII fallback viable for non-Unicode terminals if glyphs are used
- avoid turning the CLI into a TUI; this is formatting and communication polish, not an interactive shell app

## Suggested Design Direction

Use a small semantic palette:

- success: green
- warning: amber
- error: red
- accent/info: blue or cyan
- muted metadata: grey

Potential formatting patterns:

- stage prefix: muted bullet/glyph + section title
- success line: green badge + concise outcome
- warning/error diagnostics: severity badge first, then message, then location
- final summary: compact counts with severity coloring
- branded header: product name, version, then compact metadata line
- outdated notice: one short line plus `Run: mapture update`

## Test Plan

- unit tests for color policy:
  - TTY vs non-TTY
  - `NO_COLOR`
  - explicit disabled mode if added
- command tests for `validate` and `serve`:
  - output still contains expected text
  - no ANSI sequences in non-TTY/test mode unless explicitly enabled
- smoke checks for:
  - `mapture --help`
  - `mapture --version`
  - `mapture validate examples/demo`
  - `mapture serve examples/ecommerce`
  - `mapture update --channel canary`

## Non-goals

- building a full interactive TUI
- changing JSON, Mermaid, or canonical export payloads
- adding per-user CLI theme configuration in `mapture.yaml`
