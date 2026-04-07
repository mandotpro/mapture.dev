/**
 * @arch.node api payment-api
 * @arch.name Payment API
 * @arch.domain billing
 * @arch.owner team-billing
 */
export class PaymentApiClient {}

/**
 * @event.id order.placed
 * @event.role listener
 * @event.domain billing
 * @event.consumer capture_payment
 */
export function handleCapturePayment(): void {}
