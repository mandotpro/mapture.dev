# Task 004: Comment Scanner

## Goal
Implement a language-agnostic text scanner that walks directories, reads configured file extensions, and extracts `@arch.*` and `@event.*` tags from comment blocks, associating them with the exact file and line number.

## Context
Mapture treats code comments as the primary vehicle for declaring architecture metadata. Before we can build a graph or validate cross-references, we need the raw metadata payloads from the codebase. This covers layer 3 from `validation-layers.md`.

## Requirements

### 1. Directory Traversal (`src/internal/scanner`)
- Implement a walker that starts at configured `scan.include` paths.
- Safely ignore paths matching `scan.exclude`.
- Only process files with extensions enabled in `languages` (e.g. `.php`, `.go`, `.ts`).

### 2. Comment Extraction
- Extract blocks of comments:
  - `/** ... */` style (PHP, TS/JS)
  - `// ...` contiguous line blocks (Go, TS/JS)
- Ignore comment blocks that do not contain at least one `@arch.` or `@event.` tag.

### 3. Tag Parsing & Layer 3 Validation
- Implement a parser that extracts flat `@<namespace>.<key> <value>` pairs from the identified comment blocks. Ensure it safely handles multi-spaces.
- Build an intermediate "Raw Block" representation.
- Apply **Validation Layer 3 (Comment shape)**:
  - Fail if an `@arch.node` block is missing required fields (`name`, `domain`, `owner`).
  - Fail if an `@event.id` block is missing required fields (`role`, `domain`).
  - Fail if unknown keys exist within the namespace.
  - Fail on duplicate keys inside the same contiguous block.
  - Verify enums matching those in `enums.md` (`NodeType`, `EventRole`, etc).

### 4. Source Attachment
- Store the file path (relative to the repo root) and the starting line number of the block.
- Pass the collection of "Raw Blocks" to the next phase in the pipeline.

## Definition of Done
- Given `examples/demo`, running the scanner produces a structured collection of parsed comment objects in memory.
- An explicitly malformed comment (e.g., missing `@arch.owner`) throws a structured Layer 3 error.
- Multi-line comments and single-line comment streams are parsed dependably without needing AST parsing.
- Unit tests verify scanning of sample `.go`, `.php`, and `.ts` snippets.
