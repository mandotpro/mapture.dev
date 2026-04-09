package contracts

// @arch.node event legacy-order-created-event
// @arch.name Legacy Order Created Event
// @arch.domain legacy
// @arch.owner team-legacy
// @arch.status deprecated
// @arch.description Deprecated monolith order-created contract kept alive during the migration window so orders-service can continue ingesting legacy checkout traffic.
// @event.id legacy.order.created
// @event.role definition
// @event.domain legacy
// @event.owner team-legacy
type LegacyOrderCreatedEvent struct{}

// @arch.node event orders-created-event
// @arch.name Orders Created Event
// @arch.domain orders
// @arch.owner team-platform
// @arch.description Canonical modern order-created contract emitted once orders-service durably records a migrated or native order.
// @event.id orders.created
// @event.role definition
// @event.domain orders
// @event.owner team-platform
type OrdersCreatedEvent struct{}
