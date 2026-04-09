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
make test
make lint
make web
make audit-public
go run src/main.go --help
```

Useful fixture commands:

```bash
make validate demo
make scan ecommerce
make graph migration
make serve ecommerce
```

## Frontend changes

The web explorer source lives under `src/internal/webui/frontend/`.

If you change anything there:

1. Rebuild the embedded bundle with `make web`
2. Commit both the source changes and `src/internal/webui/dist/`

The committed bundle is part of the shipped single-binary experience.

## Pull request expectations

- Keep PR titles and final squash-merge titles in Conventional Commit style:
  - `fix: ...`
  - `feat: ...`
  - `feat!: ...`
- Stable releases are automated from squash-merged titles on `main`
- Include tests or fixture coverage for behavior changes
- Update docs and examples when behavior changes
- Keep generated assets in sync

## Release notes and versioning

Stable releases are managed through an automated release PR flow. Canary builds are published separately from `main`.

If your change is user-visible, make sure the PR description explains:

- what changed
- why it changed
- whether docs or examples changed
- whether release notes should call it out

## Security and support

- Security issues: see [SECURITY.md](./SECURITY.md)
- Questions and usage help: see [SUPPORT.md](./SUPPORT.md)
- Conduct expectations: see [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md)
