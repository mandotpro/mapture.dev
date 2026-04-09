package billingdb

// @arch.node database billing-db
// @arch.name Billing Database
// @arch.domain billing
// @arch.owner team-commerce
// @arch.description Stores payment intents, capture attempts, and gateway reconciliation references written by payment-service. It is read during retry and webhook handling, and inconsistent gateway state is the failure mode operators inspect here first.
type Store struct {
	writer interface {
		Insert(string, string) error
	}
}

func (s *Store) RecordCapture(orderID string) error {
	return s.writer.Insert("captures", orderID)
}
