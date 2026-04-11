---
id: 038
title: Configurable explorer UI defaults and visual tuning
milestone: v0.5.0
status: todo
prd: §18, §29
depends_on: [031, 036]
---

## Goal
Expose the most valuable explorer defaults through optional `mapture.yaml` settings so teams can tune the starting experience without turning user session state into repo config.

## Why
Different repos need different defaults, but that does not justify hard-coding everything in the frontend forever.

Examples:

- event-heavy repos want `Event Flow` by default
- large repos need more initial zoom-out
- some teams want edge labels hidden unless focused
- some want a different light/dark default or stronger domain emphasis

## Scope

### 1. Keep settings optional
If a repo does not configure UI defaults, the explorer should behave exactly as it does today.

### 2. Add high-value repo defaults only
Good candidates:

- default view
- default density
- available views/presets
- viewport fit padding
- min/max zoom
- edge label policy
- node colors
- optional theme preference default (`system`, `light`, `dark`)

Bad candidates:

- current search text
- selected node
- open menus
- last pan/zoom
- personal workbench positions

### 3. Carry defaults through the canonical export
The backend should pass repo-level UI defaults through the export so every consumer sees the same starting contract.

### 4. Preserve design-system direction
Any config added here should remain compatible with the existing token-based UI system and light/dark theming.

## Acceptance

- `mapture.yaml` can define optional explorer defaults without becoming required
- those defaults appear in the canonical export and are applied by the explorer
- user session state remains local, not written into repo config
- old repos with no new keys continue to render unchanged

## Notes

- this task is intentionally later than presets and export unification
- repo-level defaults are useful only after the explorer contract is stable
