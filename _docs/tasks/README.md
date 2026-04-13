# Active Backlog

## Product Goal
Mapture should produce one trustworthy architecture model from source comments and repo config, then reuse the same exported JSON Graph Format artifact everywhere:

1. validate and inspect architecture locally
2. serve the explorer from live repo data or a saved export
3. open the explorer offline from a static bundle plus JSON
4. convert the same JSON into Mermaid and AI-oriented artifacts
5. power an MCP server so users can ask architecture questions without a second scanner

The rule for the remaining backlog is simple:

**build once, export once, consume everywhere**

That means:

- the scanner and validator remain the only source-of-truth builders
- every downstream surface should consume the same JGF export or a derived transform from it
- no consumer should invent its own private graph shape
- UI work should sit on top of the JGF export via a shared visualisation transform, not bypass it

## Ordered Work Queue

| ID | Story | Why it comes now |
| --- | --- | --- |
| 032 | Export-driven Mermaid and diagnostics outputs | Mermaid and CI diagnostics should read from the same export/diagnostics model instead of bespoke code paths. |
| 034 | AI export from JGF graph export | AI artifacts should be generated from the shared JGF export, not from a second traversal. |
| 035 | MCP server over JGF graph export | Chat-with-your-architecture depends on the JGF export existing first. |
| 036 | Scenario presets for explorer workflows | Presets make the explorer easier to use once the data contract is stable. |
| 037 | Cross-boundary validations | Policy enforcement should come after tags and the JGF export contract are settled. |
| 038 | Configurable explorer UI defaults and visual tuning | Repo-level UI tuning is useful, but not more important than the shared data pipeline. |
| 039 | CLI output polish, color system, and terminal UX | The terminal is part of the product and should become clearer before broader adoption. |

## Non-goals For This Slice

- adding another scanner path for AI or MCP
- introducing a second graph JSON format for the web UI
- making UI session state part of `mapture.yaml`
- turning canary/stable release work into the product data model

## Notes

- Task IDs are intentionally contiguous again so the active queue reads in execution order.
- Older backlog sketches were folded into the stories below instead of kept as parallel docs.
- `tags` and `scenario presets` stay in the backlog as requested.
- Completed stories live under `_docs/tasks/done/`.
