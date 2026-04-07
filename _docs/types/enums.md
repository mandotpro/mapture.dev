# Mapture Enums

_Source of truth for all closed-value sets used across catalog schemas, comment tags, graph models, and validation. All enums are closed in v1 — adding a value requires updating both this file and the corresponding constant in `src/internal/graph/graph.go`. See PRD §16._

> **Version policy:** Values marked `v1` are stable. `experimental` values may change without notice. `deprecated` values are accepted but will produce warnings.

---

## NodeType

Enum identifier: `NodeType`  
Used in: `@arch.node <type> <id>`, graph nodes, edge `from`/`to` prefixes.

| Value      | Description                                          | Status | Since |
|------------|------------------------------------------------------|--------|-------|
| `service`  | A running process or microservice                    | stable | v0.1  |
| `api`      | An HTTP/gRPC/GraphQL API surface or client           | stable | v0.1  |
| `database` | A persistent data store (SQL, NoSQL, cache, queue)   | stable | v0.1  |
| `event`    | A domain / integration / system event                | stable | v0.1  |

**Node ID convention:** `<NodeType>:<id>` — e.g. `service:checkout-service`, `database:orders-db`.  
IDs must be lowercase kebab-case. Colons are the only separator between type prefix and name.

---

## EdgeType

Enum identifier: `EdgeType`  
Used in: graph edges, `@arch.*` relation tags.

| Value        | Tag equivalent          | Description                                         | Status | Since |
|--------------|-------------------------|-----------------------------------------------------|--------|-------|
| `calls`      | `@arch.calls`           | Service/API calls another service or API            | stable | v0.1  |
| `depends_on` | `@arch.depends_on`      | Generic dependency (use when no specific type fits) | stable | v0.1  |
| `stores_in`  | `@arch.stores_in`       | Component writes to a data store                    | stable | v0.1  |
| `reads_from` | `@arch.reads_from`      | Component reads from a data store                   | stable | v0.1  |
| `emits`      | _(via event trigger)_   | Component emits/publishes an event                  | stable | v0.1  |
| `consumes`   | _(via event listener)_  | Component consumes/subscribes to an event           | stable | v0.1  |

---

## EventKind

Enum identifier: `EventKind`  
Used in: `events.yaml` → `kind` field.

| Value         | Description                                                             | Status | Since |
|---------------|-------------------------------------------------------------------------|--------|-------|
| `domain`      | Business domain event meaningful to the bounded context                 | stable | v0.1  |
| `integration` | Cross-system integration event crossing organizational or system bounds | stable | v0.1  |
| `system`      | Infrastructure/platform event (e.g. health, scaling, lifecycle)        | stable | v0.1  |
| `internal`    | Low-level technical event not intended for external consumers           | stable | v0.1  |

---

## EventVisibility

Enum identifier: `EventVisibility`  
Used in: `events.yaml` → `visibility` field.

| Value        | Description                                                       | Status | Since |
|--------------|-------------------------------------------------------------------|--------|-------|
| `internal`   | Only consumable within the owning domain                          | stable | v0.1  |
| `public`     | Consumable by any domain within the system                        | stable | v0.1  |
| `private`    | Restricted to explicitly listed consumers or producers            | stable | v0.1  |
| `deprecated` | Previously published but no longer recommended; triggers warning  | stable | v0.1  |

---

## EventStatus

Enum identifier: `EventStatus`  
Used in: `events.yaml` → `status` field.

| Value          | Description                                                        | Status | Since |
|----------------|--------------------------------------------------------------------|--------|-------|
| `active`       | Event is in production use                                         | stable | v0.1  |
| `deprecated`   | Event is being phased out; `replacedBy` should be set              | stable | v0.1  |
| `experimental` | Event schema is unstable; breaking changes may occur               | stable | v0.1  |

---

## EventPhase

Enum identifier: `EventPhase`  
Used in: `@event.phase` comment tag (optional).

| Value         | Description                                                        | Status | Since |
|---------------|--------------------------------------------------------------------|--------|-------|
| `pre-commit`  | Emitted before a data store transaction is committed               | stable | v0.1  |
| `post-commit` | Emitted after a transaction has successfully committed             | stable | v0.1  |
| `async`       | Emitted asynchronously; no transactional guarantee                 | stable | v0.1  |
| `integration` | Emitted as part of a cross-system integration flow                 | stable | v0.1  |

---

## EventRole

Enum identifier: `EventRole`  
Used in: `@event.role` comment tag (required).

| Value        | Description                                                                   | Requires          | Status | Since |
|--------------|-------------------------------------------------------------------------------|-------------------|--------|-------|
| `definition` | Declares the event in a central or documentation location                     | —                 | stable | v0.1  |
| `trigger`    | Code point that dispatches / publishes the event                              | `@event.producer` | stable | v0.1  |
| `listener`   | Code point that handles / consumes the event                                  | `@event.consumer` | stable | v0.1  |
| `bridge-out` | Forwards an internal event out of the domain boundary                         | —                 | stable | v0.1  |
| `bridge-in`  | Receives an event from outside the domain boundary and translates it inward   | —                 | stable | v0.1  |
| `publisher`  | Generic publisher (use when trigger semantics don't apply, e.g. message bus)  | —                 | stable | v0.1  |
| `subscriber` | Generic subscriber (use when listener semantics don't apply)                  | —                 | stable | v0.1  |

---

## Changelog

| Version | Change |
|---------|--------|
| v0.1    | Initial enum set defined |
