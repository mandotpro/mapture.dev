# mapture.dev

> Repo-native architecture mapping that stays close to the code.

[![CI](https://github.com/mandotpro/mapture.dev/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/mandotpro/mapture.dev/actions/workflows/ci.yml)
[![Canary](https://github.com/mandotpro/mapture.dev/actions/workflows/canary.yml/badge.svg?branch=main)](https://github.com/mandotpro/mapture.dev/actions/workflows/canary.yml)
[![Release](https://img.shields.io/github/v/release/mandotpro/mapture.dev?display_name=tag)](https://github.com/mandotpro/mapture.dev/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/mandotpro/mapture.dev)](https://github.com/mandotpro/mapture.dev/blob/main/go.mod)
[![License](https://img.shields.io/github/license/mandotpro/mapture.dev)](./LICENSE)

mapture.dev is an experimental architecture graph tool for repositories that want a lightweight, reviewable source of truth for system structure. It combines `mapture.yaml` plus flat `@arch.*` and `@event.*` code comments, validates the result, and renders it as a polished CLI, JSON graph exports, and an interactive explorer.

> Status: early preview. mapture.dev is under active development and not production-ready yet, but the validator, export pipeline, examples, and local explorer are ready for evaluation and feedback.

![mapture.dev explorer on the ecommerce example](./.github/assets/explorer-ecommerce-hero.png)

## 3-minute quickstart

Clone the repo and run the current examples locally:

```bash
git clone https://github.com/mandotpro/mapture.dev.git
cd mapture.dev

go run src/main.go validate examples/demo
go run src/main.go serve examples/ecommerce
```

Then open the local explorer and inspect the bundled example graph.
For the repo’s day-to-day wrappers and testing helpers, run `make help`.
Release packaging and distribution scripts live under `scripts/release/`.

For local development:

```bash
make build
make serve ecommerce
```

`make build` refreshes the embedded web bundle automatically when the frontend is stale. `make serve <fixture>` always rebuilds the embedded web bundle first, then rebuilds the binary, so the local explorer reflects the latest frontend and backend changes. Use `make web` only when you want a frontend-only rebuild.

Once installed, `mapture --help` and `mapture --version` show the current version, detected release channel, install source, and whether a newer build is available for that channel.

## Install

### Quick install with curl

Install the latest stable release into `~/.local/bin`:

```bash
curl -fsSL https://raw.githubusercontent.com/mandotpro/mapture.dev/main/scripts/install.sh | bash
```

Install the rolling canary channel:

```bash
curl -fsSL https://raw.githubusercontent.com/mandotpro/mapture.dev/main/scripts/install.sh | bash -s -- --channel canary
```

Override the install directory when needed:

```bash
curl -fsSL https://raw.githubusercontent.com/mandotpro/mapture.dev/main/scripts/install.sh | MAPTURE_INSTALL_DIR=/usr/local/bin bash
```

The curl installer supports macOS and Linux on `amd64` and `arm64`.
If the latest stable release does not have prebuilt archives yet, the installer will tell you and you can use Homebrew, source install, or the canary channel until the next stable cut lands.

### Homebrew

Tap the dedicated Homebrew repository once:

```bash
brew tap mandotpro/mapture
```

Install the rolling canary channel today:

```bash
brew install mandotpro/mapture/mapture-canary
```

Stable `mapture` Homebrew packages are published from semver releases cut on the `0.x` branch.
Both channels install the same `mapture` binary, so switch channels by uninstalling the other formula first.

### Prebuilt archives

- Stable semver binaries are published on [GitHub Releases](https://github.com/mandotpro/mapture.dev/releases).
- Rolling canary prereleases are published at [the canary release](https://github.com/mandotpro/mapture.dev/releases/tag/canary).

### Build from source

```bash
go install github.com/mandotpro/mapture.dev/cmd/mapture@latest
```

Install the current `main` branch from source:

```bash
GOPROXY=direct go install github.com/mandotpro/mapture.dev/cmd/mapture@main
```

For a reproducible stable source install, prefer an explicit semver tag once the current `v0.x.y` line is published:

```bash
go install github.com/mandotpro/mapture.dev/cmd/mapture@v0.x.y
```

Notes:

- Use `GOPROXY=direct` with `@main` for source-installed canary/dev builds. Branch installs are cached aggressively by the public Go proxy, so direct fetches are more predictable for the moving `main` branch.
- `@latest` follows the newest Go-visible module version. Until the next plain `v0.x.y` stable tag is published, that may still resolve to a recent `main` pseudo-version instead of the latest stable release.
- Source installs use Go module version metadata. Release archives and Homebrew builds keep the channel version injected at build time.

### Upgrade an existing install

`mapture update` upgrades the current binary in place and follows the active release channel by default.

- Homebrew installs delegate to `brew upgrade`
- Go installs delegate to `go install`
- Direct binary installs download the matching GitHub release asset and replace the current executable
- `mapture --help` and `mapture --version` will suggest `Run: mapture update` when a newer build is available for the detected channel

Examples:

```bash
mapture update
mapture update --channel stable
mapture update --channel canary
```

For troubleshooting or quick confirmation:

```bash
mapture --version
mapture version
```

That output includes the current version, channel, detected install source, and resolved binary path so it is easier to spot stale Homebrew canaries or older direct installs.
## What mapture.dev does today

- Validates catalog ownership, domains, events, and architecture references
- Scans Go, PHP, TypeScript, and JavaScript comment blocks for `@arch.*` and `@event.*` tags
- Builds a normalized graph with deterministic node and edge identities
- Exports a shareable JSON Graph Format artifact plus explorer-facing visualisation JSON
- Serves an interactive local explorer UI for browsing the graph
- Ships example fixtures for demo, ecommerce, migration, and invalid validation cases

## Current limitations

- Comments-first only. No AST or Tree-sitter source analysis yet.
- The public graph and UI are still evolving under pre-`v1.0.0` versioning.
- Mermaid export is being rebuilt on top of the JGF export model and is not part of the current CLI surface.
- AI bundle export is planned, but not yet implemented.
- Release channels are early: canary builds are convenient for evaluation, not stability guarantees.

## Why comments-first

mapture.dev is designed for teams that want architecture metadata to live close to the code and stay reviewable in pull requests.

That means:

- no heavy source instrumentation
- no separate modeling tool to keep in sync
- one small config file for ownership and domain references
- portable annotations that work across mixed-language repos

## Supported source languages

- Go
- PHP
- TypeScript
- JavaScript

## Examples

- [`examples/demo/`](./examples/demo/) — smallest end-to-end example
- [`examples/ecommerce/`](./examples/ecommerce/) — richer multi-language flow with services, APIs, databases, and events
- [`examples/migration/`](./examples/migration/) — incremental modernization scenario
- [`examples/invalid/`](./examples/invalid/) — intentionally broken fixtures used by validation tests

## Release channels

- Stable semver releases are cut from maintenance branches such as `0.x` and `1.x`.
- Merges into `main` do not publish immediately; one canary build is published nightly if `main` changed since the previous canary.
- Stable releases are intentional: merge a PR into `0.x` or `1.x` with title `release: patch`, `release: minor`, or `release: vX.Y.Z`, or run the manual stable-release workflow.
- Homebrew canary installs are synced to `mandotpro/mapture` from the canary workflow.
- Stable release assets and the stable Homebrew formula are published after the tag is created.

## Maintainer release flow

For the next stable version on a maintenance branch:

```bash
./scripts/release/plan-release.sh 0.x patch
```

That prints the next version and the PR title to use.

Recommended flow:

1. Merge the intended changes into `0.x` or `1.x`.
2. Merge a PR into that branch with title `release: patch`, `release: minor`, or `release: vX.Y.Z`.
3. The merged PR creates the tag.
4. The `Release` workflow publishes the archives and updates Homebrew.

Manual fallback:

- Run the `Stable Release` workflow from GitHub Actions.
- Provide the target branch and either a bump or an explicit version.

## Contributing and support

- [Contributing guide](./CONTRIBUTING.md)
- [Support guide](./SUPPORT.md)
- [Security policy](./SECURITY.md)
- [Code of Conduct](./CODE_OF_CONDUCT.md)

## Further reading

- [Product spec](./_docs/mapture-dev-prd-v1.md)
- [Type and schema docs](./_docs/types/)
- [Task history and roadmap notes](./_docs/tasks/)

## License

[MIT](./LICENSE)
