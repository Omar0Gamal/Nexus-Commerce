# ADR 0006: Inventory Reservation via Redis TTL Locks

Date: 2026-10-01
Status: Accepted

## Context

Concurrent checkouts on low-stock variants can oversell inventory if stock is only decremented when an external payment gateway webhook fires after redirect completion.

## Decision

Reserve stock at checkout initiation (`POST /api/v1/checkout/intent`) by writing a 15-minute TTL key in Redis:

```text
lock:inv:{variant_id}:{session_id} -> {quantity}
```

When the payment webhook confirms capture, PostgreSQL inventory is decremented inside a transaction and the Redis lock is deleted. If checkout is abandoned, the key expires automatically without database writes.

## Consequences

- Prevents concurrent checkout overselling without holding long-lived PostgreSQL row locks.
- Abandoned checkouts release reserved stock automatically after 15 minutes.
- Requires Redis AOF persistence (`--appendonly yes`) so active reservation locks survive process restarts.
