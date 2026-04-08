---
id: 026
title: Add semantic filter visuals and improve active-filter traceability
milestone: v0.3.0
status: todo
prd: §10, §18, §29
depends_on: [014, 020, 021]
---

## Why
The current filter rail works functionally, but it does not yet provide
enough visual traceability. Active filters should be easier to read,
easier to distinguish, and more obviously connected to the graph
semantics they represent.

## Scope
- Add visual tokens/icons to the filter system so users can distinguish:
  - node-type filters
  - team/owner filters
  - domain filters
  - relation filters
  - preset/flow filters
- Give active filters a stronger semantic treatment using gentle colors,
  icons, and consistent pill structure.
- Ensure active badges communicate both:
  - what is currently scoped
  - what category the badge belongs to
- Make filter state easier to parse when multiple filters are active at
  once.
- Keep the rail compact and minimal; do not reintroduce large sidebars
  or heavy settings panels.

## Implementation Notes
- Reuse the same semantic color/icon system across:
  - filter pills
  - active badges
  - layout/preset selectors where relevant
- Preserve accessibility and text clarity; icons should support labels,
  not replace them.
- Favor subtle, low-noise visuals over bright dashboard styling.

## Acceptance
- Users can distinguish filter categories without reading every label
- Active badges remain compact but clearly communicate scope/type
- Multiple active filters are still easy to scan
- Filter visuals feel consistent with the node/edge semantics of the
  explorer rather than as generic UI chips

