<?php

/**
 * @arch.node service checkout-service
 * @arch.name Checkout Service
 * @arch.domain orders
 * @arch.owner team-commerce
 *
 * @arch.calls api payment-api
 * @arch.stores_in database orders-db
 */
final class CheckoutService
{
    public function placeOrder(int $orderId): void
    {
        /**
         * @event.id order.placed
         * @event.role trigger
         * @event.domain orders
         * @event.producer checkout.place_order
         * @event.phase post-commit
         */
        // $bus->dispatch(new OrderPlaced($orderId));
    }
}
