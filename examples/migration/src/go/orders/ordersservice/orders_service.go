package ordersservice

// @arch.node service orders-service
// @arch.name Orders Service
// @arch.domain orders
// @arch.owner team-platform
// @arch.description Owns the modern order write model and translates legacy checkout traffic into the new orders pipeline. Its primary trigger is the deprecated legacy.order.created event during migration, and the critical failure mode is duplicating orders when both write paths are exercised for the same checkout.
// @arch.reads_from database orders-db
// @arch.stores_in database orders-db
type Service struct {
	writeModel interface {
		UpsertLegacyOrder(string) error
		AppendOutbox(string) error
	}
}

// @event.id legacy.order.created
// @event.role listener
// @event.domain orders
// @event.owner team-platform
// @event.consumer orders.importLegacyOrder
// @event.topic legacy.orders
// @event.notes Orders-service listens to the deprecated monolith event so legacy checkouts still land in the modern pipeline while storefront routes are strangled one by one.
func (s *Service) ImportLegacyOrder(orderID string) { _ = s.writeModel.UpsertLegacyOrder(orderID) }

// @event.id orders.created
// @event.role trigger
// @event.domain orders
// @event.owner team-platform
// @event.producer orders.importLegacyOrder
// @event.phase post-commit
// @event.notes Orders-service emits the new canonical event after the dedicated order record is durable so downstream consumers can migrate off the legacy stream.
func (s *Service) emitOrderCreated(orderID string) { _ = s.writeModel.AppendOutbox(orderID) }
