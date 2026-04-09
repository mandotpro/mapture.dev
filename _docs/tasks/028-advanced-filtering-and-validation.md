# Task 028: Advanced Filtering, Boundary Validation, and Tags Support

## Goal
Prepare the PRD additions and architectural foundation for handling massive monoliths. This introduces partial directory scanning, formal cross-boundary breach validations, and rigorous support for user-defined tags that sync directly into the UI filters.

## Context
When adopting Mapture in complex, million-line applications, teams struggle if they are forced to view or validate the entire codebase simultaneously. They need granular workflows:
1. **Partial Data Views:** The ability to target Mapture at specific directories to construct isolated domain sub-graphs.
2. **Path & Boundary Defenses:** The ability to validate if a team is illegally bypassing boundaries (e.g., Domain A synchronously calling Domain B's private database).
3. **Custom Dimensionality (`tags`):** The ability to slice the architecture by custom metadata (e.g., `tier-1`, `pci-compliant`) that acts as an additional robust filter layer in the UI.

## Requirements

### 1. Partial Monolith Scanning (Sub-tree Filtering)
- Define a feature allowing users to filter specific directories out of a full monolith so they only work with partial mapping data.
- The CLI (`mapture validate` and `mapture scan`) must gracefully handle unresolved edges that point into the excluded monolith logic without crashing, perhaps marking them as "external to scope".

### 2. Cross-Boundary Breach Validations
- Draft a formal PRD expansion detailing logical flow constraints. 
- Provide mechanisms for users to define illegal communication paths (e.g., forbidding synchronous cross-domain calls unless explicitly allowlisted). 
- If the Mapture scanner detects a `@arch.calls` crossing a restricted boundary, it must trigger a hard Validation Error.

### 3. First-Class `tags` Definitions
- Extend `mapture.yaml` (and the `cuelang` schema checking it) to host a centrally defined array of `tags` allowed within the repository.
- Support reading `tags` from the source code annotations (e.g., `@arch.tags pci, core`).
- Validate any scanned tags against the `mapture.yaml` allowlist; reject unknown tags immediately.
- Integrate the validated tags payload down into the Web output so the Svelte Flow UI (Task 014) can dynamically generate checkbox filters for these specific tags.

## Definition of Done
- A formal PRD update (or sub-document) is drafted detailing how path validation and boundary breach logic executes.
- `mapture.yaml`'s CUE schemas are securely extended to lock down custom tags.
- The UI specification formally incorporates tag-based filtering alongside domains and teams.
