---
id: 033
title: First-class tags support
milestone: v0.3.0
status: todo
prd: §17, §18, §29
depends_on: [030]
---

## Goal
Add centrally defined tags to `mapture.yaml`, validate their use in code comments, and carry them through the JGF export plus derived visualisation export so filtering and future policy work can rely on them.

## Why
Domains and owners are not enough once a repo grows. Teams need cross-cutting metadata such as:

- `critical-path`
- `pci`
- `customer-facing`
- `legacy`
- `migration-target`

If tags stay free-form and unvalidated, they stop being useful quickly.

## Scope

### 1. Define tags in repo config
Extend `mapture.yaml` with an optional list of allowed tags, for example:

```yaml
tags:
  - critical-path
  - pci
  - legacy
  - migration-target
```

### 2. Parse tags from source comments
Support `@arch.tags` on architecture blocks and make the scanner preserve order-insensitive sets deterministically.

### 3. Validate tags
Unknown tags should fail validation with a dedicated diagnostic.

Expected behavior:

- missing tag in config -> validation error
- duplicated tag on a node -> normalized away or warned once, but never duplicated in export

### 4. Export tags everywhere they matter
The JGF export should include:

- allowed tag definitions from config
- normalized per-node tags in the graph

This makes tags available to:

- explorer filters
- Mermaid filtering
- AI export
- future cross-boundary policy rules

## Acceptance

- `mapture.yaml` can declare allowed tags
- `@arch.tags` is scanned and validated
- unknown tags fail validation with a dedicated code
- tags appear in the JGF export and in the explorer filters
- output ordering is deterministic

## Notes

- tags should be optional, not required
- tags are repo-level vocabulary, not free-text labels
- this task intentionally lands before cross-boundary rules because tags may later become one of the policy inputs
