package paymentservice

// @arch.node service payment-service
// @arch.name Payment Service
// @arch.domain billing
// @arch.owner team-commerce
// @arch.description Captures customer payments and turns gateway outcomes into internal commerce events.
// @arch.reads_from database billing-db
// @arch.stores_in database billing-db
// @arch.depends_on event payment-captured-event
type Service struct{}

// @event.id order.placed
// @event.role listener
// @event.domain billing
// @event.owner team-commerce
// @event.consumer payment.captureForOrder
// @event.topic commerce.order-placed
// @event.notes Billing starts payment capture as soon as checkout publishes the order.
func (s *Service) CaptureForOrder(orderID string) {}

// @event.id payment.captured
// @event.role trigger
// @event.domain billing
// @event.owner team-commerce
// @event.producer payment.captureForOrder
// @event.phase post-commit
// @event.notes Emitted after the payment record and ledger reference are committed.
func (s *Service) emitPaymentCaptured(orderID string) {}

// @event.id payment.failed
// @event.role trigger
// @event.domain billing
// @event.owner team-commerce
// @event.producer payment.captureForOrder
// @event.phase post-commit
// @event.notes Used by notifications when the customer must retry checkout.
func (s *Service) emitPaymentFailed(orderID string) {}

// @event.id payment.captured
// @event.role bridge-out
// @event.domain billing
// @event.owner team-commerce
// @event.topic finance.payment-captured
// @event.notes Forwards the internal payment result to an external finance reconciliation stream.
func (s *Service) forwardCapturedPayment(orderID string) {}
