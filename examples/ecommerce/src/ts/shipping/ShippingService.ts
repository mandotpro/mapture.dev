import { CarrierApiClient } from "./CarrierApiClient";

/**
 * @arch.node service shipping-service
 * @arch.name Shipping Service
 * @arch.domain shipping
 * @arch.owner team-operations
 * @arch.description Turns captured payments into shipments and carrier label requests. Its primary trigger is payment.captured, and the key failure mode is avoiding duplicate labels when retries race with warehouse processing.
 * @arch.calls api carrier-api
 * @arch.depends_on event shipment-created-event
 */
export class ShippingService {
  constructor(private readonly carrierApi: CarrierApiClient) {}

  /**
   * @event.id payment.captured
   * @event.role listener
   * @event.domain shipping
   * @event.owner team-operations
   * @event.consumer shipping.createShipment
   * @event.topic commerce.payment-captured
   * @event.notes Shipping waits for this event because labels should only be purchased after billing makes payment capture durable.
   */
  createShipment(orderId: string) {
    this.carrierApi.purchaseLabel(orderId);

    /**
     * @event.id shipment.created
     * @event.role trigger
     * @event.domain shipping
     * @event.owner team-operations
     * @event.producer shipping.createShipment
     * @event.phase post-commit
     * @event.notes Shipping emits this once the label and tracking record exist so notifications and warehouse consumers can move forward from the same shipment state.
     */
  }
}
