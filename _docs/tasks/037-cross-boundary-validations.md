---
id: 037
title: Cross-boundary validations
milestone: v0.4.0
status: todo
prd: §15, §17, §29
depends_on: [030, 033]
---

## Goal
Add explicit boundary-policy validation so teams can define which domain/team relationships are allowed and have Mapture flag architecture breaches automatically.

## Why
Once the shared model is stable, Mapture should do more than describe architecture. It should help protect it.

The key use cases are:

- preventing accidental cross-domain calls
- allowing known migration exceptions explicitly
- surfacing risky boundary erosion in CI and in the explorer

## Scope

### 1. Add policy configuration
Extend `mapture.yaml` with explicit boundary rules, for example:

- allowed outbound domains
- allowed inbound domains
- optional relation-type allowlists
- optional team-level policies where domain-level rules are not enough

Keep the first version simple and reviewable.

### 2. Validate graph relationships against policy
Apply policy checks to the normalized graph after graph construction.

Likely relation families:

- `calls`
- `depends_on`
- `reads_from`
- `stores_in`
- `emits`
- `consumes`

### 3. Export violations through the canonical diagnostics model
Boundary violations should appear like any other validation diagnostic so they are visible in:

- `validate --format text`
- `validate --format json`
- the canonical export
- the explorer
- AI and MCP consumers later

## Acceptance

- a repo can define allowed/disallowed cross-boundary relationships
- illegal relationships fail validation with clear diagnostics
- legal migration exceptions can be modeled explicitly
- violations appear in the canonical export diagnostics block

## Notes

- do not couple this to the explorer UI first; policy must live in the validator
- tags may later become part of policy conditions, which is why tags land earlier
