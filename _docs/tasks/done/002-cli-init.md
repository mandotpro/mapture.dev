# Task 002: CLI `init` Command

## Goal
Implement the `mapture init [path]` command to interactively bootstrap a target directory. It should provide a rich, wizard-like setup experience (using a library like `charmbracelet/huh` or `survey`) to automatically detect project features and ask the user for confirmation via checkboxes and inputs, ending with fully scaffolded `mapture.yaml` and catalog files.

## Context
When a user wants to adopt Mapture, they need a starting point. The `init` command should dynamically generate the repository structure based on user input and project state, preventing the user from needing to manually write configuration files from scratch.

## Requirements

### 1. `init` command interactive interface
- Implemented inside `src/cmd/` using a capable TUI prompt library.
- Accepts an optional `[path]` argument (defaults to `.`).
- **Interactive Wizard Steps:**
  1. **Source Directories:** Prompts for `scan.include` dirs. Suggests defaults (e.g., `./src`, `./cmd`, `./pkg`) based on what exists in the directory.
  2. **Excluded Directories:** Prompts for `scan.exclude` dirs. Pre-fills with `.git`, `node_modules`, `vendor`, etc.
  3. **Language Detection:** Automatically detects languages used in the project (checking for `.go`, `.php`, `.ts`/`.js` files). Presents these in a checkbox list so the user can formally toggle/confirm which languages Mapture should scan.
  4. **Validation settings:** simple radio buttons or toggles to define strictness (e.g., `failOnUnknownTeam`, `failOnUnknownDomain`).
- Safe execution: must check if any target files already exist before the wizard writes data, gracefully warning and prompting to skip or merge, ensuring it never blind-overwrites existing catalogs or configs.

### 2. Scaffold `mapture.yaml`
- Write out a full `mapture.yaml` customized strictly based on the user's interactive choices.
- The file should include helpful commented-out text explaining the fields.

### 3. Scaffold Catalog directory
- Create the default `catalog.dir` (typically `./architecture/`).
- Generate `teams.yaml` with 2 example teams.
- Generate `domains.yaml` with 2 example domains referencing those teams.
- Generate `events.yaml` with 1 example event showcasing the schema.
- Content of these files should exactly match the lean "Example" structures in `_docs/types/catalog-schemas.md`.

## Definition of Done
- Running `go run src/main.go init .` opens a rich interactive terminal UI.
- The UI automatically selects `ts` / `go` / `php` if those files exist in the tree.
- The resulting `mapture.yaml` accurately specifies the languages, includes, excludes, and toggles selected in the wizard.
- The 4 required dummy catalog and configuration files are generated cleanly without overwriting existing files.
