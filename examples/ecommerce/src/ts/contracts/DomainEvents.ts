/**
 * Shared event contracts used by multiple domains in the monorepo.
 */

/**
 * @arch.node event order-placed-event
 * @arch.name Order Placed Event
 * @arch.domain orders
 * @arch.owner team-commerce
 * @arch.description Canonical contract for a newly placed order.
 * @event.id order.placed
 * @event.role definition
 * @event.domain orders
 * @event.owner team-commerce
 * @event.version 1
 */
export interface OrderPlacedEvent {
  type: "order.placed";
  orderId: string;
  totalAmount: number;
}

/**
 * @arch.node event payment-captured-event
 * @arch.name Payment Captured Event
 * @arch.domain billing
 * @arch.owner team-commerce
 * @arch.description Contract used by shipping and notifications after a successful payment capture.
 * @event.id payment.captured
 * @event.role definition
 * @event.domain billing
 * @event.owner team-commerce
 * @event.version 1
 */
export interface PaymentCapturedEvent {
  type: "payment.captured";
  orderId: string;
  paymentId: string;
}

/**
 * @arch.node event shipment-created-event
 * @arch.name Shipment Created Event
 * @arch.domain shipping
 * @arch.owner team-operations
 * @arch.description Contract sent when a carrier label is created.
 * @event.id shipment.created
 * @event.role definition
 * @event.domain shipping
 * @event.owner team-operations
 * @event.version 1
 */
export interface ShipmentCreatedEvent {
  type: "shipment.created";
  orderId: string;
  trackingNumber: string;
}

/**
 * @arch.node event notification-sent-event
 * @arch.name Notification Sent Event
 * @arch.domain notifications
 * @arch.owner team-engagement
 * @arch.description Internal delivery contract used by notification telemetry consumers.
 * @event.id notification.sent
 * @event.role definition
 * @event.domain notifications
 * @event.owner team-engagement
 * @event.version 1
 */
export interface NotificationSentEvent {
  type: "notification.sent";
  orderId: string;
  channel: "email";
}
