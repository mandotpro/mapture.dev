/**
 * @arch.node api carrier-api
 * @arch.name Carrier API
 * @arch.domain shipping
 * @arch.owner team-operations
 * @arch.description Outbound integration for buying carrier labels and fetching tracking numbers from the shipping vendor. It is called once payment clears, and vendor-side timeouts are the failure mode shipping must recover from cleanly.
 */
export class CarrierApiClient {
  purchaseLabel(orderId: string) {
    return `tracking-${orderId}`;
  }
}
