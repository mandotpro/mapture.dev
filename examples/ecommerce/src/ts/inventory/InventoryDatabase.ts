/**
 * @arch.node database inventory-db
 * @arch.name Inventory Database
 * @arch.domain inventory
 * @arch.owner team-operations
 * @arch.description Holds per-SKU availability snapshots and reservation counters used by inventory-service. It is updated during reserve and release flows, and lock contention on hot stock is the failure mode the inventory path has to survive.
 */
export class InventoryDatabase {
  reserve(orderId: string) {
    return `reservation:${orderId}`;
  }
}
