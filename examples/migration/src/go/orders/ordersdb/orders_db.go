package ordersdb

// @arch.node database orders-db
// @arch.name Orders Database
// @arch.domain orders
// @arch.owner team-platform
// @arch.description Dedicated store for the modern orders write model and outbox state. Orders-service writes here before emitting the new canonical event, and duplicate legacy imports are the failure mode the team audits against this data.
type Store struct {
	writer interface {
		Insert(string, string) error
	}
}

func (s *Store) SaveOrder(orderID string) error {
	return s.writer.Insert("orders", orderID)
}
