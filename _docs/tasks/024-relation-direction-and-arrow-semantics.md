---
id: 024
title: Standardize relation direction and arrow semantics in the explorer
milestone: v0.3.0
status: todo
prd: §15, §17, §18, §29
depends_on: [005, 020, 021, 023]
---

## Why
Some of the current edge directions feel wrong, especially for
`depends_on`. Even when the underlying graph is technically valid, the
visual explorer must make every relation feel semantically correct. If
edge direction is confusing, the whole graph becomes harder to trust.

## Scope
- Define the expected visual semantics for all relation types currently
  exposed in the explorer:
  - `calls`
  - `depends_on`
  - `stores_in`
  - `reads_from`
  - `emits`
  - `consumes`
- Verify that graph edge direction, arrowhead placement, and label text
  agree with those semantics.
- Fix any relation that is currently rendered in the wrong direction or
  with the wrong source/target mental model.
- Decide whether some relations need distinct line styles or arrow
  variants to feel correct, especially:
  - dependency-style edges
  - storage/read edges
  - event flow edges
- Document the canonical meaning so future exporters and UI work do not
  drift.

## Implementation Notes
- This task is about semantic correctness first, styling second.
- If the graph model or validator currently produces an inverted edge
  for some relation, fix that at the model/source layer rather than only
  compensating in the UI.
- Mermaid/exported graph output should eventually align with the same
  relation semantics.

## Acceptance
- Every supported relation reads naturally in the explorer
- `depends_on` no longer feels visually inverted
- Edge labels, arrowheads, and direction all agree with the same mental
  model
- The expected semantics are documented in code and/or task notes

