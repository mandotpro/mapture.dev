/**
 * @arch.node service inventory-service
 * @arch.name Inventory Service
 * @arch.domain inventory
 * @arch.owner team-operations
 * @arch.description Reserves and releases stock in response to order lifecycle changes.
 * @arch.reads_from database inventory-db
 * @arch.stores_in database inventory-db
 * @arch.depends_on event order-placed-event
 */
export class InventoryService {
  /**
   * @event.id order.placed
   * @event.role listener
   * @event.domain inventory
   * @event.owner team-operations
   * @event.consumer inventory.reserveForOrder
   * @event.topic commerce.order-placed
   */
  reserveForOrder(orderId: string) {
    /**
     * @event.id inventory.reserved
     * @event.role trigger
     * @event.domain inventory
     * @event.owner team-operations
     * @event.producer inventory.reserveForOrder
     * @event.phase pre-commit
     */
  }

  /**
   * @event.id order.cancelled
   * @event.role listener
   * @event.domain inventory
   * @event.owner team-operations
   * @event.consumer inventory.releaseForCancelledOrder
   * @event.topic commerce.order-cancelled
   */
  releaseForCancelledOrder(orderId: string) {
    /**
     * @event.id inventory.released
     * @event.role trigger
     * @event.domain inventory
     * @event.owner team-operations
     * @event.producer inventory.releaseForCancelledOrder
     * @event.phase async
     */
  }
}
