---
id: 036
title: Scenario presets for explorer workflows
milestone: v0.4.0
status: todo
prd: §18, §29
depends_on: [031, 033]
---

## Goal
Add named explorer presets that package useful reading modes for common architecture tasks without forcing users to manually tune every filter and control.

## Why
The explorer is now powerful enough to overwhelm first-time users. Presets are the lowest-friction way to make the app useful for:

- onboarding
- async/event review
- storage review
- boundary review
- incident impact reading

This should make the explorer easier to use, not more configurable for its own sake.

## Scope

### 1. Introduce preset objects
Each preset should define a compact, frontend-consumable recipe:

- view mode
- density mode
- relation visibility
- node-type emphasis
- optional structure/boundary emphasis

### 2. Ship a small default set
Start with a short opinionated list:

- `System Overview`
- `Event Flow`
- `Boundary Review`
- `Storage Footprint`
- `Impact Review`

### 3. Keep presets layered on existing controls
Selecting a preset should update existing explorer state. It should not invent a separate rendering engine.

### 4. Preserve room for future repo defaults
This task is about runtime presets first. Repo-configurable default presets, ordering, or enable/disable behavior can later be layered through the UI config task.

## Acceptance

- presets are visible and selectable in the explorer
- each preset produces a meaningfully different reading mode
- presets compose with search and active filters without breaking the graph
- preset logic stays in the presentation layer, not in the backend export format

## Notes

- keep the preset set small and intentional
- the goal is fast answers, not a long menu of decorative views
