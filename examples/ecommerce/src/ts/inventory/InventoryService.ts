import { InventoryDatabase } from "./InventoryDatabase";

/**
 * @arch.node service inventory-service
 * @arch.name Inventory Service
 * @arch.domain inventory
 * @arch.owner team-operations
 * @arch.description Consumes order lifecycle events to reserve or release sellable stock in the inventory store. Its primary triggers are new orders and cancellations, and overselling due to missed releases is the failure mode it is designed to prevent.
 * @arch.reads_from database inventory-db
 * @arch.stores_in database inventory-db
 * @arch.depends_on event order-placed-event
 */
export class InventoryService {
  constructor(private readonly inventoryDb: InventoryDatabase) {}

  /**
   * @event.id order.placed
   * @event.role listener
   * @event.domain inventory
   * @event.owner team-operations
   * @event.consumer inventory.reserveForOrder
   * @event.topic commerce.order-placed
   * @event.notes Inventory listens immediately so a placed order reserves scarce stock before another checkout can oversell the same SKU.
   */
  reserveForOrder(orderId: string) {
    this.inventoryDb.reserve(orderId);

    /**
     * @event.id inventory.reserved
     * @event.role trigger
     * @event.domain inventory
     * @event.owner team-operations
     * @event.producer inventory.reserveForOrder
     * @event.phase pre-commit
     * @event.notes Inventory emits this reservation signal so downstream consumers can react without reading reservation state directly from the database.
     */
  }

  /**
   * @event.id order.cancelled
   * @event.role listener
   * @event.domain inventory
   * @event.owner team-operations
   * @event.consumer inventory.releaseForCancelledOrder
   * @event.topic commerce.order-cancelled
   * @event.notes Inventory consumes cancellations so reserved units are returned before the same stock is sold again through another checkout.
   */
  releaseForCancelledOrder(orderId: string) {
    this.inventoryDb.reserve(`release:${orderId}`);

    /**
     * @event.id inventory.released
     * @event.role trigger
     * @event.domain inventory
     * @event.owner team-operations
     * @event.producer inventory.releaseForCancelledOrder
     * @event.phase async
     * @event.notes This release event makes the stock correction visible to any future consumers after inventory finishes the asynchronous unwind.
     */
  }
}
