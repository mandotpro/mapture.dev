/**
 * @arch.node service shipping-service
 * @arch.name Shipping Service
 * @arch.domain shipping
 * @arch.owner team-operations
 * @arch.description Creates shipments after payment succeeds and requests carrier labels.
 * @arch.calls api carrier-api
 * @arch.depends_on event shipment-created-event
 */
export class ShippingService {
  /**
   * @event.id payment.captured
   * @event.role listener
   * @event.domain shipping
   * @event.owner team-operations
   * @event.consumer shipping.createShipment
   * @event.topic commerce.payment-captured
   */
  createShipment(orderId: string) {
    /**
     * @event.id shipment.created
     * @event.role trigger
     * @event.domain shipping
     * @event.owner team-operations
     * @event.producer shipping.createShipment
     * @event.phase post-commit
     */
  }
}
