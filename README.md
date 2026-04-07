# Mapture

> Catalog your architecture. Validate it in code. Explore it as a graph.

Mapture is an MIT-licensed, single-binary, repo-native architecture graph
tool. It turns a small YAML catalog plus lightweight code comments into
validated dependency maps, interactive diagrams, and AI-ready
architecture context.

**Status:** early scaffolding — not yet usable. See
[`_docs/mapture-dev-prd-v1.md`](./_docs/mapture-dev-prd-v1.md) for the
full product spec and [`CLAUDE.md`](./CLAUDE.md) for a repo map.

## Idea in 30 seconds

1. Declare a tiny central catalog in `architecture/` (`teams.yaml`,
   `domains.yaml`, `events.yaml`).
2. Annotate code with flat `@arch.*` / `@event.*` tag comments.
3. Run `mapture` to validate, visualize, and export the graph.

```bash
mapture init .
mapture validate .
mapture serve .
mapture export-html . -o architecture-report.html
mapture export-ai .
```

A minimal end-to-end example lives in [`examples/demo/`](./examples/demo/).

## Why comments-first

Comments are language-agnostic, portable across PHP / Go / TypeScript,
close to the code, and reviewable in pull requests. Mapture parses them
into data and validates them against the catalog. See PRD §8.

## Roadmap

- **v0.1** — comments → normalized graph + Mermaid + static HTML
- **v0.2** — validation + CI integration
- **v0.3** — interactive explorer UI
- **v0.4** — stronger source attachment per language
- **v0.5** — event / workflow views
- **v1.0** — stable formats, docs, polish

Full breakdown in PRD §29.

## License

[MIT](./LICENSE)
