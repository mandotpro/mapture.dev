---
id: 033
title: First-class tags support
milestone: v0.3.0
status: done
prd: §17, §18, §29
depends_on: [030]
---

## Goal
Add first-class repo tags as a controlled vocabulary, validate them across config and code comments, carry them through the JGF export and derived visualisation export, and make them usable in the explorer for tag-driven flow analysis.

## Why
Domains and owners are not enough once a repo grows. Teams need cross-cutting metadata such as:

- `critical-path`
- `pci`
- `customer-facing`
- `legacy`
- `migration-target`

If tags stay free-form and unvalidated, they stop being useful quickly. The explorer also needs a way to follow flows across domain and ownership boundaries without inventing a second metadata system.

## Scope

### 1. Define repo-wide tag vocabulary
Extend `mapture.yaml` with an optional top-level `tags:` list:

```yaml
tags:
  - critical-path
  - pci
  - customer-facing
  - legacy
  - migration-target
```

Rules:

- tags are optional
- tags are repo-level vocabulary and must be kebab-case
- duplicate top-level tags are rejected at config load time
- `teams[].tags` and `domains[].tags` remain supported, but every referenced tag must exist in the top-level vocabulary
- legacy `teams.yaml` / `domains.yaml` still work and obey the same vocabulary rules

### 2. Parse direct tags from comments
Support:

- `@arch.tags tag-a, tag-b, tag-c`
- `@event.tags tag-a, tag-b`

Rules:

- comma-separated values
- trim whitespace
- normalize case/spacing
- dedupe and sort deterministically
- keep the existing flat single-line syntax; no nested structure

### 3. Validate and compute effective tags
Validation should report dedicated `unknown_tag` diagnostics when tags are referenced outside the configured vocabulary.

Each graph node should carry:

- `tags`: direct tags from the node or event comments
- `effectiveTags`: union of direct tags plus inherited domain tags plus inherited owner-team tags

Both lists should be deduped and exported deterministically.

### 4. Carry tags through JGF and visualisation
JGF is the canonical artifact.

Required export behavior:

- repo tag vocabulary under `graph.metadata.mapture.catalog.tags`
- direct node tags in node metadata
- effective node tags in node metadata
- visualisation export derived from JGF preserves the same tag fields for the explorer

### 5. Add explorer filtering and narrowing
Tags should become a first-class explorer filter alongside search, teams, domains, and types.

Behavior:

- add a `Tags` pill in the left filter rail
- tag options are derived from the currently visible node set, using `effectiveTags`
- each tag option shows a visible-node count
- selecting one or more tags filters by `effectiveTags`
- OR semantics within selected tags
- AND semantics against search, teams, domains, and types
- active filter badges and reset behavior include tags
- the available tag list narrows as the visible graph narrows, so users can follow tagged flows
- node inspector `Tags` row shows `effectiveTags`

## Acceptance

- `mapture.yaml` can declare a repo-wide allowed tag vocabulary
- `teams[].tags` and `domains[].tags` validate against that vocabulary
- `@arch.tags` and `@event.tags` are parsed, normalized, and validated
- unknown tags fail validation with `unknown_tag`
- graph nodes carry `tags` and `effectiveTags`
- JGF carries the tag vocabulary plus node tags
- visualisation export preserves the same tag data
- explorer filters can narrow by tags using `effectiveTags`
- tag options recompute from the visible graph as filtering changes
- output ordering is deterministic

## Notes

- tags remain optional
- tags are a repo-level controlled vocabulary, not free-form labels
- legacy split catalogs remain supported, but single-file `mapture.yaml` is the default path
- Mermaid, AI export, and future policy work should consume this same tag model from JGF rather than inventing their own
