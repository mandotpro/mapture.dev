/**
 * Shared event contracts used by multiple domains in the monorepo.
 */

/**
 * @arch.node event order-placed-event
 * @arch.name Order Placed Event
 * @arch.domain orders
 * @arch.owner team-commerce
 * @arch.description Shared contract emitted when checkout commits a new order and downstream commerce workflows should begin. Billing, inventory, and notifications all depend on it, and version drift is the failure mode this contract keeps visible.
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
 * @arch.description Shared contract for a successfully captured payment after billing durably records the gateway result. Shipping and notifications consume it next, and schema drift would create duplicate shipment or messaging side effects.
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
 * @arch.description Shared shipment contract emitted once shipping stores a carrier label and tracking number. Notifications rely on it for customer updates, and missing tracking fields are the failure mode this contract makes explicit.
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
 * @arch.description Internal delivery contract published after notification-service hands work to the messaging pipeline. Subscribers use it for telemetry and support timelines, and delivery payload drift is the failure mode this shared definition prevents.
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
