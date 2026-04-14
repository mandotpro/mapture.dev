# Contributing to Mapture

Thanks for contributing to Mapture.

Mapture is an experimental, comments-first architecture graph tool. The repo is public, but the project is still early. Contributions that improve correctness, usability, release quality, examples, and docs are especially valuable.

## Before you open a PR

1. Open an issue first for non-trivial changes.
2. Keep changes focused. Large mixed PRs are difficult to review and difficult to release.
3. Make sure every user-facing claim matches the current implementation. Do not document planned commands or exporters as if they already work.

## Local setup

Prerequisites:

- Go `1.25.x`
- Node.js `20+` for rebuilding the embedded web bundle
- Git

Common commands:

```bash
make help
make test
make lint
make build
make serve ecommerce
make web
make audit-public
go run src/main.go --help
```

Release/distribution automation scripts are grouped under `scripts/release/`.


Useful fixture commands:

```bash
make validate demo
make scan ecommerce
make export-json-graph migration
make export-json-visualisation migration
make serve ecommerce
```

`make build` auto-refreshes the embedded web bundle when the frontend sources are newer than `src/internal/webui/dist/`. `make serve <fixture>` always rebuilds the embedded web bundle first, then rebuilds the binary, so local explorer testing always uses the latest app state. Use `make web` when you want an explicit frontend-only rebuild without building a binary.

## Frontend changes

The web explorer source lives under `src/internal/webui/frontend/`.

If you change anything there:

1. Rebuild the embedded bundle with `make web`
2. Commit both the source changes and `src/internal/webui/dist/`

The committed bundle is part of the shipped single-binary experience.

## Graph JSON Schema

The exported graph JSON schema is stable. If making structural changes to the graph payload:
- Bump the `schemaVersion` field across the project's Go structs, CUE schemas, and TypeScript interfaces.
- Write a migration note for any breaking change inside the PR description.

## Pull request expectations

- Keep PR titles and final squash-merge titles in Conventional Commit style:
  - `fix: ...`
  - `feat: ...`
  - `feat!: ...`
- Include tests or fixture coverage for behavior changes
- Update docs and examples when behavior changes
- Keep generated assets in sync

## Release notes and versioning

- `main` feeds the nightly canary channel only, and only when `main` changed since the previous canary build.
- Stable releases are cut from maintenance branches such as `0.x` and `1.x`.
- Stable tags are created intentionally:
  - merge a PR titled `release: patch`
  - merge a PR titled `release: minor`
  - or merge a PR titled `release: vX.Y.Z`
- Manual fallback is available in GitHub Actions through the `Stable Release` workflow.
- `./scripts/release/plan-release.sh <branch> <patch|minor>` prints the next version and the matching PR title.
- The Homebrew tap in `mandotpro/homebrew-mapture` is updated automatically from the same release pipelines when the required repository variable and token secret are present.

If your change is user-visible, make sure the PR description explains:

- what changed
- why it changed
- whether docs or examples changed
- whether release notes should call it out

## Security and support

- Security issues: see [SECURITY.md](./SECURITY.md)
- Questions and usage help: see [SUPPORT.md](./SUPPORT.md)
- Conduct expectations: see [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md)
