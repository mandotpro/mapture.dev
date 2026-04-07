# Mapture Config Schema

_Complete field specification for `mapture.yaml` — the per-repo configuration file that controls how Mapture scans, validates, and exports a repository. See PRD §12._

> **Config discovery:** The `mapture` binary walks up from the current directory to find `mapture.yaml`. If not found, all commands except `init` will fail with a descriptive error.

---

## File location

```
<repo-root>/
  mapture.yaml        ← repository config (checked into source control)
  architecture/
    teams.yaml
    domains.yaml
    events.yaml
```

---

## Full schema

### `version`

| Field     | Type      | Required | Default | Valid values |
|-----------|-----------|----------|---------|--------------|
| `version` | `integer` | ✅        | —       | `1`          |

Config file format version. Must be `1` in v1. Future breaking changes will increment this.

---

### `catalog`

| Field         | Type     | Required | Default          | Description                                         |
|---------------|----------|----------|------------------|-----------------------------------------------------|
| `catalog.dir` | `string` | ✅        | `./architecture` | Path (relative to `mapture.yaml`) to catalog files. |

The catalog directory must contain at minimum `teams.yaml` and `domains.yaml`. `events.yaml` is required if any `@event.*` tags exist in the scanned source.

---

### `scan`

| Field           | Type       | Required | Default | Description                                                  |
|-----------------|------------|----------|---------|--------------------------------------------------------------|
| `scan.include`  | `string[]` | ✅        | —       | Paths to scan for comment tags. Relative to `mapture.yaml`.  |
| `scan.exclude`  | `string[]` | —        | `[]`    | Paths to exclude from scanning. Globs supported.             |

**Well-known excludes** (recommended to always include):

```yaml
scan:
  exclude:
    - ./vendor
    - ./node_modules
    - ./dist
    - ./build
    - ./.git
```

---

### `languages`

Controls which file extensions the scanner processes. All default to `false` unless set.

| Field                 | Type      | Required | Default | Description                         |
|-----------------------|-----------|----------|---------|-------------------------------------|
| `languages.php`       | `boolean` | —        | `false` | Scan `.php` files.                  |
| `languages.go`        | `boolean` | —        | `false` | Scan `.go` files.                   |
| `languages.typescript`| `boolean` | —        | `false` | Scan `.ts` and `.tsx` files.        |
| `languages.javascript`| `boolean` | —        | `false` | Scan `.js` and `.jsx` files.        |

At least one language must be enabled for the scanner to produce output.

---

### `comments`

| Field             | Type     | Required | Default | Valid values | Description                           |
|-------------------|----------|----------|---------|--------------|---------------------------------------|
| `comments.style`  | `string` | —        | `tags`  | `tags`       | Comment parsing strategy. `tags` is the only supported value in v1. |

---

### `validation`

| Field                              | Type       | Required | Default | Description                                                              |
|------------------------------------|------------|----------|---------|--------------------------------------------------------------------------|
| `validation.failOnUnknownDomain`   | `boolean`  | —        | `true`  | Exit non-zero if a comment references a domain not in the catalog.        |
| `validation.failOnUnknownTeam`     | `boolean`  | —        | `true`  | Exit non-zero if a comment references a team not in the catalog.          |
| `validation.failOnUnknownEvent`    | `boolean`  | —        | `true`  | Exit non-zero if an `@event.id` is not in the events catalog.             |
| `validation.failOnUnknownNode`     | `boolean`  | —        | `true`  | Exit non-zero if a relation tag targets an undeclared node.               |
| `validation.requireMetadataOn`     | `string[]` | —        | `[]`    | Event roles that must have a comment annotation in every usage. Values from `EventRole`. |
| `validation.warnOnOrphanedNodes`   | `boolean`  | —        | `false` | Warn if any declared node has no edges.                                   |
| `validation.warnOnDeprecatedEvents`| `boolean`  | —        | `true`  | Warn if any `@event.id` references a deprecated event.                    |

---

## Full example

```yaml
version: 1

catalog:
  dir: ./architecture

scan:
  include:
    - ./services
    - ./pkg
    - ./src
  exclude:
    - ./vendor
    - ./node_modules
    - ./dist
    - ./build

languages:
  php: true
  go: true
  typescript: true

comments:
  style: tags

validation:
  failOnUnknownDomain: true
  failOnUnknownTeam: true
  failOnUnknownEvent: true
  failOnUnknownNode: true
  requireMetadataOn:
    - trigger
    - listener
  warnOnOrphanedNodes: false
  warnOnDeprecatedEvents: true
```

---

## Generated by `mapture init`

Running `mapture init .` produces a starter `mapture.yaml` with sensible defaults and all fields commented with descriptions. The generated file is idempotent — running `init` on an existing config will not overwrite it.

---

## Changelog

| Version | Change |
|---------|--------|
| v0.1    | Initial config schema. `failOnUnknownNode` and `warnOn*` fields added beyond the PRD baseline. |
