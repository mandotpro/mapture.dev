# Task 030: Config-Driven Custom Tags Integration

## Goal
Establish robust configuration schema enforcement for free-form `tags`. Centrally defining tags in the main repo guarantees complex secondary filtering outputs correctly within the UI cleanly and typo-free.

## Context
When adopting Mapture globally, organizations slice codebases beyond domains via horizontal metadata like `tier-1`, `pci-compliant`, or `deprecated-soon`. These free-form definitions must be rigidly centrally managed; if an engineer typos a tag (`pci` vs `pci-compliance`), filtering fails dynamically and reporting breaks.

## Requirements
- Edit `_docs/types/config-schema.md` (and corresponding `cuelang` checks) so `mapture.yaml` can define `tags: [array of exact allowed strings]`.
- Natively parse source block annotations holding `@arch.tags` matrices.
- Perform a strict verification blocking the build flow with "Unknown Tag" validation errors if a developer creates unindexed tags.
- Wire these strictly defined traits through the output graph models so the Svelte Flow layouts dynamically generate actionable checkbox filters based universally on the known list.

## Definition of Done
- `mapture.yaml` successfully provisions valid operational tags safely via CUE structures.
- Erroneous tags strictly fail structural validation testing.
- The UI gracefully aggregates tags providing unique multi-dimensional view options (e.g., visually isolating all elements tagged `pci-compliant` natively mapped).
