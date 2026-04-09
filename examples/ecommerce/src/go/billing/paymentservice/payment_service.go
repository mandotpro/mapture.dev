package paymentservice

// @arch.node service payment-service
// @arch.name Payment Service
// @arch.domain billing
// @arch.owner team-commerce
// @arch.description Turns placed orders and gateway callbacks into durable payment records and commerce events. It starts from order.placed and Stripe webhook input, and the critical failure mode is preventing duplicate captures when retries race with reconciliation.
// @arch.reads_from database billing-db
// @arch.stores_in database billing-db
// @arch.depends_on event payment-captured-event
type Service struct {
	store interface {
		SaveCapture(string) error
		SaveFailure(string) error
	}
	ledger interface {
		Publish(string) error
	}
}

// @event.id order.placed
// @event.role listener
// @event.domain billing
// @event.owner team-commerce
// @event.consumer payment.captureForOrder
// @event.topic commerce.order-placed
// @event.notes Billing reacts here because checkout should finish its write path first while capture starts immediately against a durable order record.
func (s *Service) CaptureForOrder(orderID string) { _ = s.store.SaveCapture(orderID) }

// @event.id payment.captured
// @event.role trigger
// @event.domain billing
// @event.owner team-commerce
// @event.producer payment.captureForOrder
// @event.phase post-commit
// @event.notes Payment-service emits this only after the payment row and reconciliation reference are committed so shipping and notifications never act on a transient gateway success.
func (s *Service) emitPaymentCaptured(orderID string) {
	_ = s.ledger.Publish("payment.captured:" + orderID)
}

// @event.id payment.failed
// @event.role trigger
// @event.domain billing
// @event.owner team-commerce
// @event.producer payment.captureForOrder
// @event.phase post-commit
// @event.notes Payment failures are published so notification-service can prompt the customer to retry without coupling email rendering into billing code.
func (s *Service) emitPaymentFailed(orderID string) { _ = s.store.SaveFailure(orderID) }

// @event.id payment.captured
// @event.role bridge-out
// @event.domain billing
// @event.owner team-commerce
// @event.topic finance.payment-captured
// @event.notes Billing bridges this event outward because finance reconciliation still depends on an external stream fed from the internal payment result.
func (s *Service) forwardCapturedPayment(orderID string) {
	_ = s.ledger.Publish("finance.payment-captured:" + orderID)
}
