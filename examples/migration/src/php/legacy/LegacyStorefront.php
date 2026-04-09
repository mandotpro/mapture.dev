<?php

namespace App\Legacy;

/**
 * @arch.node service legacy-storefront
 * @arch.name Legacy Storefront
 * @arch.domain legacy
 * @arch.owner team-legacy
 * @arch.description Legacy monolith checkout entrypoint that still accepts customer orders while the strangler migration is in progress. Its primary input is the pre-existing storefront checkout form, and the key failure mode is letting legacy writes drift away from the modern orders pipeline during cutover.
 * @arch.stores_in database storefront-db
 */
final class LegacyStorefront
{
    public function placeOrder(array $checkout): void
    {
        $this->legacyBus->publish('legacy.order.created', ['checkout' => $checkout]);

        /**
         * @event.id legacy.order.created
         * @event.role trigger
         * @event.domain legacy
         * @event.owner team-legacy
         * @event.producer App\Legacy\LegacyStorefront::placeOrder
         * @event.phase post-commit
         * @event.notes The monolith still emits this deprecated event so orders-service can keep ingesting checkout traffic until the last write paths move off the legacy app.
         */
    }
}
