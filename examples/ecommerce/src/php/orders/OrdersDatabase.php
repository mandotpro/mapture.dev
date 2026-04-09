<?php

namespace App\Orders;

/**
 * @arch.node database orders-db
 * @arch.name Orders Database
 * @arch.domain orders
 * @arch.owner team-commerce
 * @arch.description Primary relational store for checkout writes and order lifecycle state. It is written before any event fan-out begins, and stale transaction state is the failure mode the orders flow must avoid.
 */
final class OrdersDatabase
{
    public function insertOrder(array $record): void
    {
        $this->connection->insert('orders', $record);
    }
}
