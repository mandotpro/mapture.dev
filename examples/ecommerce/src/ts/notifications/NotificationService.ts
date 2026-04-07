/**
 * @arch.node service notification-service
 * @arch.name Notification Service
 * @arch.domain notifications
 * @arch.owner team-engagement
 * @arch.description Builds transactional email payloads from commerce and operations events.
 * @arch.calls api email-api
 * @arch.depends_on event notification-sent-event
 */
export class NotificationService {
  /**
   * @event.id order.placed
   * @event.role listener
   * @event.domain notifications
   * @event.owner team-engagement
   * @event.consumer notifications.sendOrderConfirmation
   * @event.topic commerce.order-placed
   */
  sendOrderConfirmation(orderId: string) {
    /**
     * @event.id notification.sent
     * @event.role publisher
     * @event.domain notifications
     * @event.owner team-engagement
     * @event.topic notifications.lifecycle
     * @event.notes Published after the order confirmation email is accepted by the delivery pipeline.
     */
  }

  /**
   * @event.id payment.failed
   * @event.role listener
   * @event.domain notifications
   * @event.owner team-engagement
   * @event.consumer notifications.sendPaymentFailure
   * @event.topic commerce.payment-failed
   */
  sendPaymentFailure(orderId: string) {}

  /**
   * @event.id shipment.created
   * @event.role listener
   * @event.domain notifications
   * @event.owner team-engagement
   * @event.consumer notifications.sendShipmentUpdate
   * @event.topic commerce.shipment-created
   */
  sendShipmentUpdate(orderId: string) {}
}
