# Task 028: Partial Monolith Sub-Tree Scanning

## Goal
Enable Mapture to scan and validate isolated sub-directories within a large monolithic repository natively without requiring a full-repo parse.

## Context
When adopting Mapture in complex, million-line applications, teams struggle if they are forced to view or validate the entire codebase simultaneously. They need granular workflows where they target Mapture at specific directories (e.g., `./src/checkout`) to construct isolated domain sub-graphs immediately.

## Requirements
- Define CLI arguments (e.g., `mapture scan --target ./src/checkout`) or `mapture.yaml` scope injections to target sub-folders smoothly.
- Modify the scanner and Graph builder to gracefully tolerate "dangling validations". If nodes have edges pointing to services or elements that exist outside the targeted partial directory, validation layers cannot panic.
- Ensure the JSON representations format these external bounds logically without dropping nodes.

## Definition of Done
- `mapture validate --target ./paths` executes flawlessly without requiring monolith-level knowledge.
- Cross-directory boundaries remain fully transparent locally.
