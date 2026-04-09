<?php

namespace App\Legacy;

/**
 * @arch.node database storefront-db
 * @arch.name Storefront Database
 * @arch.domain legacy
 * @arch.owner team-legacy
 * @arch.description Primary relational store behind the legacy storefront checkout flow. It is written before the deprecated event is emitted, and transaction drift during the migration window is the failure mode the legacy team still has to contain.
 */
final class StorefrontDatabase
{
    public function insertOrder(array $record): void
    {
        $this->connection->insert('legacy_orders', $record);
    }
}
