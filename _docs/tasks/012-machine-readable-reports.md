---
id: 008
title: Machine-readable validate report + CI exit codes
milestone: v0.2.0
status: todo
prd: §15 (CLI behavior), §29 v0.2.0 success
depends_on: [004]
---

## Why
PRD §15 says validate must return a machine-readable report and a non-zero
exit code on failure. Today `mapture validate` always exits 0 and prints a
single human line: `config and catalog OK ...` — unusable in CI.

## Scope
- Add flags: `--format text|json` (default `text`), `--out <file>`,
  `--quiet`
- Define a stable JSON schema:
  ```json
  {
    "schemaVersion": 1,
    "summary": {"errors": N, "warnings": N, "byLayer": {...}},
    "issues": [{"severity":"error","layer":4,"code":"unknown-domain","file":"...","line":10,"message":"..."}]
  }
  ```
- Exit code policy:
  - 0 → no errors (warnings allowed)
  - 1 → user errors (validation failures)
  - 2 → tool errors (config/catalog could not be loaded)
- Color the text format (only when stdout is a TTY)
- Document the schema in a CUE file under `internal/schema/` so the
  output is itself validatable

## Acceptance
- `mapture validate examples/demo --format json` writes a parseable JSON
  document on stdout and exits 0
- Introducing a known-bad fixture causes exit code 1 and `errors > 0`
- A missing `mapture.yaml` causes exit code 2
- The JSON schema is locked behind a snapshot test
