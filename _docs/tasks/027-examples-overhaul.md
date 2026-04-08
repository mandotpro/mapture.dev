# Task 027: Real-World Example Flows & Test Fixtures Overhaul

## Goal
Redesign the `examples/` directory to contain narrative-driven, realistic architectural flows. The current demos are disorganized and fail to showcase the true expressive power of Mapture's tag system. We need pristine fixtures that tell a complete structural story, which simultaneously act as rigorous E2E test environments.

## Context
Our examples act as both marketing material (when users test `mapture validate examples/demo`) and as the canonical test suites for our CLI (e.g. `pre-push` hooks run strictly against `examples/`). If the examples are a messy jumble of generic nodes, the tool's output looks unimpressive, and test coverage remains superficial. We need interconnected, recognizable system designs that showcase synchronous calls, async event buses, boundaries, and database ownership.

## Requirements

### 1. Refactor `examples/ecommerce` as the Canonical Flow
Transform `examples/ecommerce` into an impeccably mapped flow covering diverse interaction types:
- **Synchronous Dependency:** A generic TS `checkout-api` that `@arch.depends_on` a Go `payment-service`.
- **Event-Driven Async:** The `payment-service` acts as a `@event.publisher` for `payment.cleared`.
- **Event Consumption:** A PHP `inventory-worker` acts as an `@event.subscriber`, updating a local database (`@arch.stores_in inventory-db`).
- **External Boundaries:** Demonstrate external API calls via `@event.bridge-out` pointing to `Stripe` or `SendGrid`.

### 2. Introduce a "Legacy Migration" Flow (`examples/migration`)
Establish a new directory demonstrating typical migration realities:
- A massive `/legacy-monolith` component.
- Clear structural deprecations using the validation constraints (e.g., triggering a warning if a node interacts with a deprecated component).
- Distinct Team ownership dividing domains (e.g., `checkout-team` vs `legacy-team`), showcasing Mapture's domain filtering effectively.

### 3. Polish Source File Annotations
- Replace the currently disjointed pseudo-code in `/examples/` with cleanly formatted, recognizable source code snippets matching standard Go, PHP, and TS syntaxes.
- Add rich markdown descriptions inside `@arch.summary` tags in the blocks so the UI Side-Panels feel informative when clicked.

### 4. Wire Directly to Test Suites
- Update the Go tests to ensure elements like `Stripe` are correctly asserted as external bridges.
- Ensure the newly mapped E-Commerce example executes successfully under `mapture validate examples/ecommerce` with absolutely zero errors (a true "Golden Path" test).

## Definition of Done
- Exploring `examples/ecommerce` via `mapture serve` visually produces a stunning, instantly understandable microservice diagram that resembles an actual production environment.
- Mapture's documentation features these explicitly designed scenarios to teach developers how to capture async boundaries and external components smoothly. 
- The `examples/demo` folder is either cleaned up to represent a tiny "Hello World" footprint, or deleted in favor of the more comprehensive `ecommerce` narrative.
