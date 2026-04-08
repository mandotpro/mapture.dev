---
id: 023
title: Fix event consumption semantics and make producer-consumer links readable
milestone: v0.3.0
status: todo
prd: §14, §17, §18, §29
depends_on: [004, 005, 020, 021]
---

## Why
The event side of the graph is still hard to read. `consumes` links do
not currently communicate the event relationship clearly enough, and the
visual result does not make the producer -> event -> consumer path feel
obvious. This is one of the most important explorer use cases, so event
semantics need dedicated treatment.

## Scope
- Audit the current graph meaning for:
  - `emits`
  - `consumes`
  - event definition / event node linking
- Define one canonical visual model for event flow:
  - producer points to event
  - consumer relationship is readable relative to the event node
- If needed, introduce an explicit reverse/back-reference treatment so a
  consumed event visually reads as "this node depends on that event"
  rather than as an arbitrary line.
- Improve edge labels and/or edge decorations for event links so the
  difference between `emits` and `consumes` is obvious at a glance.
- Ensure presets such as producer-to-consumer flow use this improved
  event link model consistently.

## Implementation Notes
- Do not invent new graph concepts unless the validator/graph model
  genuinely needs them.
- Start by clarifying the direction semantics in one place, then make
  renderer behavior follow that model.
- If event links need asymmetric arrowheads or special edge components,
  keep that scoped to event semantics rather than all edges.

## Acceptance
- Producer -> event -> consumer flow is immediately understandable on
  `examples/ecommerce`
- `emits` and `consumes` are visually distinct
- Event presets highlight full event flow without ambiguous direction
- Users can follow a consumed event back to the originating event node

