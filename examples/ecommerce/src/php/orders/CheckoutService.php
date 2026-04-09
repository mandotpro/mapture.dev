<?php

namespace App\Orders;

/**
 * @arch.node service checkout-service
 * @arch.name Checkout Service
 * @arch.domain orders
 * @arch.owner team-commerce
 * @arch.description Accepts a validated storefront cart, persists the order, and kicks off billing, inventory, and notification work. Its primary input is the checkout cart plus buyer context, and the critical failure mode is preventing any downstream side effects before the order commit succeeds.
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
        $this->eventBus->dispatch('order.placed', ['cart' => $cart]);

        /**
         * @event.id order.placed
         * @event.role trigger
         * @event.domain orders
         * @event.owner team-commerce
         * @event.producer App\Orders\CheckoutService::placeOrder
         * @event.phase post-commit
         * @event.notes Checkout emits this only after the order transaction commits so billing, inventory, and notifications all receive a stable order id and totals snapshot.
         */
    }

    public function cancelOrder(int $orderId): void
    {
        $this->eventBus->dispatch('order.cancelled', ['orderId' => $orderId]);

        /**
         * @event.id order.cancelled
         * @event.role trigger
         * @event.domain orders
         * @event.owner team-commerce
         * @event.producer App\Orders\CheckoutService::cancelOrder
         * @event.phase post-commit
         * @event.notes Checkout publishes cancellations so inventory can release held stock and downstream customer messaging can stop before fulfillment advances.
         */
    }
}
