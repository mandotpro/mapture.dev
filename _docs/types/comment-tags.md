# Mapture Comment Tag Reference

_Complete specification for `@arch.*` and `@event.*` structured comment tags. These are the primary metadata layer that the scanner parses from source files. See PRD §14._

> **Format principle:** Tags are flat `@key value` pairs on separate lines inside a block or line comment. They must never contain embedded JSON, YAML, or structured data. Keep comments readable for humans in pull requests.

---

## Format rules

- Tags use the form `@namespace.key value`.
- The value is everything after the first whitespace following the key.
- Tags are parsed from contiguous comment blocks (PHP `/** */`, Go `//`, TypeScript `/** */` or `//`).
- A single comment block should define exactly one node or one event occurrence.
- Unknown tag keys cause a Layer 3 validation error.
- Duplicate tags within the same block are a Layer 3 validation error.

---

## `@arch.*` tags — Architecture nodes

Used to declare a node (service, API, database, or event) and its relationships to other nodes.

### Required node declaration tags

| Tag                       | Format                     | Description                                                            |
|---------------------------|----------------------------|------------------------------------------------------------------------|
| `@arch.node`              | `<NodeType> <id>`          | Declares this comment block as a node. See `enums.md` → NodeType.     |
| `@arch.name`              | `<human-readable string>`  | Display name of the node shown in graphs and exports.                  |
| `@arch.domain`            | `<domain-id>`              | Domain this node belongs to. Must exist in `domains.yaml`.             |
| `@arch.owner`             | `<team-id>`                | Team that owns this node. Must exist in `teams.yaml`.                  |

All four are required on every node-declaring comment block. Missing any one is a Layer 3 error.

### Optional node metadata tags

| Tag                | Format              | Description                                                              |
|--------------------|---------------------|--------------------------------------------------------------------------|
| `@arch.description`| `<prose string>`    | Short prose description of the node's purpose (used in AI summaries).    |
| `@arch.version`    | `<semver or int>`   | Version of this component, if versioned.                                 |
| `@arch.tags`       | `<comma-separated>` | Free-form labels for filtering in the UI.                                |
| `@arch.status`     | `active \| deprecated \| experimental` | Lifecycle status of this node.                         |

### Relation tags (edges)

Each relation tag adds a directed edge from this node to a target node. Multiple relation tags may appear in the same block.

| Tag                 | Format                  | EdgeType produced | Description                                               |
|---------------------|-------------------------|--------------------|-----------------------------------------------------------|
| `@arch.calls`       | `<NodeType> <id>`       | `calls`            | This node calls the target service or API.                |
| `@arch.depends_on`  | `<NodeType> <id>`       | `depends_on`       | Generic dependency when no specific relation type fits.   |
| `@arch.stores_in`   | `<NodeType> <id>`       | `stores_in`        | This node writes to the target database.                  |
| `@arch.reads_from`  | `<NodeType> <id>`       | `reads_from`       | This node reads from the target database.                 |

The target `<NodeType> <id>` in a relation tag must match a declared node (Layer 6 validation). Unknown targets produce a configurable error or warning.

### Examples

**PHP:**
```php
/**
 * @arch.node service checkout-service
 * @arch.name Checkout Service
 * @arch.domain orders
 * @arch.owner team-commerce
 * @arch.description Places orders and orchestrates the checkout flow.
 *
 * @arch.calls api payment-api
 * @arch.stores_in database orders-db
 */
final class CheckoutService {}
```

**Go:**
```go
// @arch.node database orders-db
// @arch.name Orders Database
// @arch.domain orders
// @arch.owner team-commerce
// @arch.description Primary relational store for the orders domain.
package ordersdb
```

**TypeScript:**
```ts
/**
 * @arch.node api payment-api
 * @arch.name Payment API
 * @arch.domain billing
 * @arch.owner team-billing
 */
export class PaymentApiClient {}
```

---

## `@event.*` tags — Event occurrences

Used to annotate code points where events are triggered or consumed. Event occurrence comments reference event IDs defined in `events.yaml`.

### Always-required event tags

| Tag             | Format          | Description                                                              |
|-----------------|-----------------|--------------------------------------------------------------------------|
| `@event.id`     | `<event-id>`    | Canonical event ID. Must exist in `events.yaml`.                         |
| `@event.role`   | `<EventRole>`   | Role of this code point. See `enums.md` → EventRole.                     |
| `@event.domain` | `<domain-id>`   | Domain of the code producing or consuming the event.                     |

### Conditionally required event tags

| Condition             | Required tag        | Format               | Description                                      |
|-----------------------|---------------------|----------------------|--------------------------------------------------|
| `@event.role trigger` | `@event.producer`   | `<function or path>` | Identifier of the function/class doing dispatch. |
| `@event.role listener`| `@event.consumer`   | `<function or path>` | Identifier of the function/class handling it.    |

### Optional event tags

| Tag               | Format                 | Description                                                         |
|-------------------|------------------------|---------------------------------------------------------------------|
| `@event.owner`    | `<team-id>`            | Team responsible at this usage site (defaults to catalog ownerTeam).|
| `@event.phase`    | `<EventPhase>`         | Transactional phase of the emission. See `enums.md` → EventPhase.  |
| `@event.topic`    | `<string>`             | Message broker topic / queue name (e.g. Kafka topic).               |
| `@event.version`  | `<integer>`            | Event contract version being used at this site.                     |
| `@event.notes`    | `<prose string>`       | Free-form notes for this usage site (e.g. retry behavior, SLA).    |

### Examples

**Trigger (PHP):**
```php
/**
 * @event.id order.placed
 * @event.role trigger
 * @event.domain orders
 * @event.producer checkout.place_order
 * @event.phase post-commit
 */
$bus->dispatch(new OrderPlaced($orderId));
```

**Listener (TypeScript):**
```ts
/**
 * @event.id order.placed
 * @event.role listener
 * @event.domain billing
 * @event.consumer capture_payment
 * @event.topic orders.events
 */
eventBus.on("order.placed", handleCapturePayment)
```

**Bridge (Go):**
```go
// @event.id order.placed
// @event.role bridge-out
// @event.domain orders
// @event.notes Forwards to external payment gateway topic.
func forwardOrderPlaced(e Event) {}
```

---

## Attachment rules

The scanner attaches each comment block to a nearby source location. Valid attachment targets by role:

| Tag type            | Valid attachment targets                                             |
|---------------------|---------------------------------------------------------------------|
| `@arch.*` nodes     | Class, function, package declaration, file header                   |
| `@event.* trigger`  | Dispatch call, publish call, function containing the dispatch       |
| `@event.* listener` | Event handler function, subscription registration                   |
| `@event.* definition` | Standalone block or class representing the event schema           |

Blocks attached to no recognizable symbol produce a Layer 5 warning.

---

## Changelog

| Version | Change |
|---------|--------|
| v0.1    | `@arch.*` and `@event.*` tag sets defined |
