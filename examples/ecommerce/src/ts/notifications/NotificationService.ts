import { EmailApiClient } from "./EmailApiClient";

/**
 * This file intentionally demonstrates pub/sub rather than a direct callback.
 * The service publishes notification.sent after vendor hand-off and does not know which subscribers consume that lifecycle event.
 */
/**
 * @arch.node service notification-service
 * @arch.name Notification Service
 * @arch.domain notifications
 * @arch.owner team-engagement
 * @arch.description Transforms commerce and operations events into customer-facing email requests. It starts from order, payment, and shipment events, and the critical failure mode is avoiding customer sends when upstream data is incomplete or the delivery vendor rejects the payload.
 * @arch.calls api email-api
 * @arch.depends_on event notification-sent-event
 */
export class NotificationService {
  constructor(private readonly emailApi: EmailApiClient) {}

  /**
   * @event.id order.placed
   * @event.role listener
   * @event.domain notifications
   * @event.owner team-engagement
   * @event.consumer notifications.sendOrderConfirmation
   * @event.topic commerce.order-placed
   * @event.notes Notification-service listens here because checkout should not block on email rendering or vendor latency once the order is safely committed.
   */
  sendOrderConfirmation(orderId: string) {
    this.emailApi.enqueue("order-confirmation", orderId);

    /**
     * @event.id notification.sent
     * @event.role publisher
     * @event.domain notifications
     * @event.owner team-engagement
     * @event.topic notifications.lifecycle
     * @event.notes Published after the confirmation request is accepted by the delivery pipeline so any number of telemetry subscribers can react independently.
     */
  }

  /**
   * @event.id payment.failed
   * @event.role listener
   * @event.domain notifications
   * @event.owner team-engagement
   * @event.consumer notifications.sendPaymentFailure
   * @event.topic commerce.payment-failed
   * @event.notes Notification-service consumes payment failures so billing stays focused on capture logic while the customer gets the retry guidance they need.
   */
  sendPaymentFailure(orderId: string) { this.emailApi.enqueue("payment-failure", orderId); }

  /**
   * @event.id shipment.created
   * @event.role listener
   * @event.domain notifications
   * @event.owner team-engagement
   * @event.consumer notifications.sendShipmentUpdate
   * @event.topic commerce.shipment-created
   * @event.notes Shipping updates are consumed here so the customer only receives tracking information after a real carrier label exists.
   */
  sendShipmentUpdate(orderId: string) { this.emailApi.enqueue("shipment-update", orderId); }
}
