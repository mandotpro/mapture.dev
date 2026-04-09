# Task 027: Realistic Examples & Test Fixtures Alignment

## Goal
`examples/ecommerce` is already a near-complete, realistic multi-service fixture. This task has two narrow objectives:

1. Tighten the annotation quality of `examples/ecommerce` so its real user flows read like production documentation rather than tag demos.
2. Add one new self-contained fixture, `examples/migration`, that models a real strangler-fig migration and exercises the deprecated-event warning path — the one realistic scenario `ecommerce` does not cover.

This is not a refactor. No renames, no new node types, no new components added to fit arbitrary tag coverage goals.

## Context

### What the ecommerce fixture already models
Reading the current tree (`examples/ecommerce/src/**`) against the scanner output, the fixture already covers these real user flows end-to-end:

- **Happy-path checkout** — `CheckoutService` (PHP) → `payment-api` → `payment-service` (Go) via `@arch.calls`; `payment-service` emits `payment.captured`; `shipping-service` (TS) consumes it and emits `shipment.created`; `notification-service` (TS) consumes each lifecycle event and builds the customer email.
- **Payment failure path** — `payment-service` emits `payment.failed`; `notification-service.sendPaymentFailure` consumes it.
- **Order cancellation / stock release** — `CheckoutService.cancelOrder` emits `order.cancelled`; `inventory-service.releaseForCancelledOrder` (TS) consumes it.
- **External payment reconciliation (async outbound)** — `payment-service.forwardCapturedPayment` re-emits `payment.captured` with role `bridge-out` to a finance reconciliation stream, representing the real "internal commerce event → external finance ledger" flow.
- **External webhook intake (async inbound)** — `stripe-webhook-api` receives Stripe traffic and emits `stripe.webhook.received` with role `bridge-in`, representing the real Stripe webhook integration.
- **External HTTP calls** — `carrier-api` (shipping label vendor) and `email-api` (transactional email vendor) are modelled as external `api` nodes reached via `@arch.calls`, which is the canonical way to model an outbound HTTP client boundary in Mapture.
- **Internal pub/sub lifecycle projection** — `NotificationService.sendOrderConfirmation` publishes `notification.sent` (role `publisher`) and `NotificationAuditProjector.project` subscribes to it (role `subscriber`), building internal audit views. This is the realistic pattern where the producer does not know its consumers — distinct from the direct `trigger`/`listener` hand-offs used elsewhere.

Every `@event.role` value the scanner supports (`trigger`, `listener`, `publisher`, `subscriber`, `bridge-in`, `bridge-out`) is already used, and each use corresponds to a real architectural motif. Do not add components just to demonstrate a role.

### Scanner grammar this task must stay faithful to
- **`@arch.*`** keys: `node, name, domain, owner, description, version, tags, status, calls, depends_on, stores_in, reads_from` (`src/internal/scanner/scanner.go:23`). There is no `@arch.summary` tag; UI side-panel text comes from `@arch.description` (`validator.go:200` populates `Node.Summary` from it).
- **`@event.role`** values: `definition, trigger, listener, bridge-out, bridge-in, publisher, subscriber` (`scanner.go:27`). In the validator, `trigger`/`publisher`/`bridge-out` all map to the `emits` edge; `listener`/`subscriber`/`bridge-in` all map to the `consumes` edge (`validator.go:287-290`). The only required-field differences are: `trigger` requires `@event.producer`, `listener` requires `@event.consumer`.
- **Edge semantics** (`src/internal/graph/graph.go:17-33`): `calls` is the synchronous outbound call; `depends_on` is a logical dependency on another node (the fixture uses it both for service→service and service→event references); `stores_in`/`reads_from` target databases; `emits`/`consumes` are derived from event-role blocks. External HTTP boundaries are modelled as `@arch.calls api <client>` pointing at a dedicated `api` node — `bridge-in`/`bridge-out` are **event** roles for async boundaries, not an alternative spelling of an HTTP call.
- **Deprecation warnings** fire only for events with `status: deprecated` or `deprecated: true`, gated by `warnOnDeprecatedEvents` (`validator.go:275`). There is no node-level deprecation check today.

### Test wiring constraint
`examples/demo` and `examples/ecommerce` are both load-bearing test fixtures: `scanner_test.go`, `mermaid_test.go`, `validator_test.go`, `cmd/root_test.go`, and `server_test.go` all reference them by path. Any structural change to those directories is a change to the test suite.

## Requirements

### 1. Tighten `examples/ecommerce` annotation quality
No new components, no renames, no role swaps. The goal is that a developer browsing the fixture in `mapture serve` learns the flow from the side-panel text alone.

- Audit every `@arch.description` and `@event.notes` across `examples/ecommerce/src/**`. Each node description should be 1–3 sentences covering purpose, primary inputs, and the one interesting failure mode. Each event-role block should state *why* this producer or consumer reacts to the event, not just restate the role.
- Add a short top-of-file comment on `NotificationService.ts` and `NotificationAuditProjector.ts` explaining the pub/sub distinction those two files are intentionally demonstrating (producer does not know its consumers; the projector is one of potentially many subscribers). This is documentation about Mapture's modelling vocabulary; do not change any tags.
- Replace any remaining placeholder method bodies with one realistic line of code per method (an idiomatic call, DB query, or event dispatch). The file must still look like the language it claims to be — real imports, real signatures, real error handling where it is the clearest way to express the failure mode.
- `mapture validate examples/ecommerce` must continue to report zero errors and zero warnings after these edits.

Do **not**:
- Add a "second emitter" or alternate `publisher` block "for coverage" — the fixture already has a real publisher/subscriber pair.
- Add a `bridge-out` event for outbound email — the real external boundary is `email-api`, and it is already correctly modelled as `@arch.calls api email-api`.
- Alter `@arch.depends_on` usages in the fixture. Several of them reference event IDs that do not exist in `events.yaml` (e.g. `order-placed-event` vs catalog `order.placed`); validation tolerates this today, and investigating or "fixing" it is out of scope for this task.

### 2. Introduce `examples/migration` — strangler-fig pattern
A new self-contained fixture that models the one realistic scenario not covered by `ecommerce`: a partial migration from a legacy monolith to a modern service, with the legacy event stream marked deprecated so downstream consumers visibly get the warning.

Flow modelled (strangler fig):
- `legacy-storefront` — a PHP monolith in domain `legacy`, owned by `team-legacy`. Stores its state in a `storefront-db` (MySQL-style) database via `@arch.stores_in`. Emits `legacy.order.created` on successful checkout (role `trigger`). The event's catalog entry has `status: deprecated`.
- `orders-service` — a Go service in domain `orders`, owned by `team-platform`. During the migration window it still listens to `legacy.order.created` (role `listener`) so new orders flowing through the monolith continue to land in the modern pipeline. It emits `orders.created` (role `trigger`, status `active`) going forward, and stores its own state in `orders-db`.
- `orders-api` — a thin Go `api` node in domain `orders`, the modern write path. `@arch.calls service orders-service` for the normal case, and `@arch.calls service legacy-storefront` for the subset of endpoints not yet migrated. This cross-domain call is the whole point of the strangler-fig pattern and must be explicit in the fixture.

Catalog shape:
- `teams.yaml`: `team-legacy`, `team-platform`.
- `domains.yaml`: `legacy` (owned by `team-legacy`), `orders` (owned by `team-platform`). `allowedOutboundDomains` wired so `orders → legacy` is permitted (the real migration call) and the reverse is permitted for `legacy.order.created` consumption.
- `events.yaml`: `legacy.order.created` with `status: deprecated`; `orders.created` with `status: active`.
- `mapture.yaml`: `warnOnDeprecatedEvents: true`, matching the `ecommerce` config.

Validation outcome:
- `mapture validate examples/migration` must exit zero, report zero errors, and report exactly the deprecated-event warning(s) produced by `orders-service` consuming `legacy.order.created`. This is the assertion the new fixture exists to protect.

Out of scope for this task: node-level deprecation, team-boundary enforcement rules, or any new validator feature. If we want those, file separate tasks.

### 3. Wire the new fixture into the test suite, leave the others alone
- Add one Go test (in the existing `cmd/root_test.go` or `validator_test.go` style, whichever already covers "fixture validates with expected warnings") that loads `examples/migration`, asserts zero errors, and asserts at least one warning whose code is the deprecated-event code and whose message references `legacy.order.created`.
- Do **not** delete or restructure `examples/demo`. It is the minimum-footprint fixture used by multiple tests; leave its file paths intact. If its annotations have drifted from the current scanner grammar, fix in place — do not rename files.
- Do **not** change any existing assertions against `examples/ecommerce` unless §1 forces a measurable change in node/edge counts. If it does, regenerate `src/internal/exporter/mermaid/testdata/ecommerce-billing.golden.mmd` and update the count to match; otherwise leave test code untouched.

## Definition of Done
- `examples/ecommerce` still validates with zero errors and zero warnings. Every node has a description that reads like real documentation; every event-role block explains *why* that producer or consumer reacts to the event.
- `examples/migration` exists, validates with zero errors and exactly the expected deprecated-event warning(s), and its strangler-fig structure (legacy PHP monolith, modern Go service, cross-domain migration call) is clearly visible from the source files alone.
- One new test asserts the `examples/migration` warning outcome.
- `examples/demo` is untouched structurally; no existing test path needed updating.
- No invented tag names (`@arch.summary`) and no invented event IDs (`payment.cleared`) appear anywhere in the repo.
