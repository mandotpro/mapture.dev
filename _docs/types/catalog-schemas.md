# Mapture Catalog Schemas

_Field-level specifications for the three v1 catalog YAML files: `teams.yaml`, `domains.yaml`, `events.yaml`. These files live under the `catalog.dir` path defined in `mapture.yaml` (default: `./architecture/`)._

---

## Team

**File:** `architecture/teams.yaml`  
**Purpose:** Canonical ownership registry. Used to validate `@arch.owner` tags and `ownerTeams` references in domains.

### Schema

| Field   | Type     | Required | Constraints                         | Description                                          |
|---------|----------|----------|-------------------------------------|------------------------------------------------------|
| `id`    | `string` | ✅        | Unique across all teams; kebab-case | Machine identifier for the team. Used in cross-refs. |
| `name`  | `string` | ✅        | Non-empty                           | Human-readable team name.                            |
| `email` | `string` | —        | Valid email                         | Team distribution list or primary contact alias.     |

### Example

```yaml
teams:
  - id: team-commerce
    name: Commerce Team
    email: commerce@example.com

  - id: team-billing
    name: Billing Team
    email: billing@example.com
```

### Validation rules

- `id` must be unique within the file.
- `id` must be kebab-case (`[a-z0-9-]+`).
- `id` is the cross-reference key used in `domains.ownerTeams`, `events.ownerTeam`, and `@arch.owner`.

---

## Domain

**File:** `architecture/domains.yaml`  
**Purpose:** Canonical business or technical domain registry. Used to validate `@arch.domain` and `@event.domain` tags.

### Schema

| Field           | Type       | Required | Constraints                         | Description                                    |
|-----------------|------------|----------|-------------------------------------|------------------------------------------------|
| `id`            | `string`   | ✅        | Unique; kebab-case                  | Machine identifier for the domain.             |
| `name`          | `string`   | ✅        | Non-empty                           | Human-readable domain name.                    |
| `ownerTeams`    | `string[]` | ✅        | Min 1 item; each must exist in teams | Team IDs that own this domain.                |
| `description`   | `string`   | —        | Prose text                          | Short description of the domain's responsibility. |

### Example

```yaml
domains:
  - id: orders
    name: Orders
    ownerTeams: [team-commerce]
    description: Handles the full lifecycle of customer orders.

  - id: billing
    name: Billing
    ownerTeams: [team-billing]
    description: Manages payment capture and invoicing.
```

### Validation rules

- `id` must be unique within the file.
- Each `ownerTeams` entry must exist as a team `id`.

---

## Event

**File:** `architecture/events.yaml`  
**Purpose:** Canonical event catalog. Used to validate `@event.id`, `@event.domain`, and deprecation policy.

### Schema

| Field          | Type              | Required | Constraints                            | Description                                                    |
|----------------|-------------------|----------|----------------------------------------|----------------------------------------------------------------|
| `id`           | `string`          | ✅        | Unique; dot-namespaced (`domain.action`) | Canonical event identifier. Used in `@event.id` tags.        |
| `name`         | `string`          | ✅        | Non-empty                              | Human-readable event name.                                     |
| `domain`       | `string`          | ✅        | Must exist in domains catalog          | Domain that owns and governs this event.                       |
| `ownerTeam`    | `string`          | ✅        | Must exist in teams catalog            | Team responsible for the event contract.                       |
| `kind`         | `EventKind`       | ✅        | See `enums.md` → EventKind             | Semantic classification of the event.                          |
| `visibility`   | `EventVisibility` | ✅        | See `enums.md` → EventVisibility       | Who is permitted to consume this event.                        |
| `status`       | `EventStatus`     | ✅        | See `enums.md` → EventStatus           | Lifecycle status of the event.                                 |
| `description`  | `string`          | —        | Prose text                             | What this event represents and when it is emitted.             |
| `version`      | `integer`         | —        | Positive integer; default `1`          | Contract version. Increment on breaking schema changes.        |
| `payloadSchema`| `string`          | —        | URI or relative path                   | Link to the payload schema definition (JSON Schema, protobuf). |
| `deprecated`   | `boolean`         | —        | Default `false`                        | If `true`, triggers a warning on any `@event.id` usage.        |
| `replacedBy`   | `string`          | —        | Must exist as an event `id`            | The event that supersedes this one (used with `deprecated`).   |

### Example

```yaml
events:
  - id: order.placed
    name: Order Placed
    domain: orders
    ownerTeam: team-commerce
    kind: domain
    visibility: internal
    status: active
    version: 1
    description: Emitted when a customer successfully places an order.

  - id: order.placed.v1
    name: Order Placed (v1 legacy)
    domain: orders
    ownerTeam: team-commerce
    kind: domain
    visibility: deprecated
    status: deprecated
    deprecated: true
    replacedBy: order.placed
```

### Validation rules

- `id` must be unique within the file.
- `domain` must exist as a domain `id`.
- `ownerTeam` must exist as a team `id`.
- If `deprecated: true`, a `replacedBy` value is strongly recommended (warning if absent).
- If `status: deprecated`, `deprecated: true` should also be set (warning otherwise).

---

## Changelog

| Version | Change |
|---------|--------|
| v0.1    | Schemas for Team, Domain, Event defined. Kept lean: no tags, no allowed-domain lists. |
