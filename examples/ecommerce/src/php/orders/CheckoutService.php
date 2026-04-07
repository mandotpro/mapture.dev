<?php

namespace App\Orders;

/**
 * @arch.node service checkout-service
 * @arch.name Checkout Service
 * @arch.domain orders
 * @arch.owner team-commerce
 * @arch.description Accepts a cart, places the order, and orchestrates downstream commerce workflows.
 * @arch.calls api payment-api
 * @arch.calls service inventory-service
 * @arch.calls service notification-service
 * @arch.stores_in database orders-db
 * @arch.depends_on event order-placed-event
 */
final class CheckoutService
{
    public function placeOrder(array $cart): void
    {
        /**
         * @event.id order.placed
         * @event.role trigger
         * @event.domain orders
         * @event.owner team-commerce
         * @event.producer App\Orders\CheckoutService::placeOrder
         * @event.phase post-commit
         * @event.notes Fired after the order row is committed so billing and inventory see a stable order id.
         */
    }

    public function cancelOrder(int $orderId): void
    {
        /**
         * @event.id order.cancelled
         * @event.role trigger
         * @event.domain orders
         * @event.owner team-commerce
         * @event.producer App\Orders\CheckoutService::cancelOrder
         * @event.phase post-commit
         * @event.notes Used to release any reserved stock and stop notification work.
         */
    }
}
