# ADR 0005: Redis-Backed Worker Queue

Date: 2026-09-30
Status: Accepted

## Context

Transactional emails, outbound merchant webhooks, and PDF invoice generation must execute asynchronously outside HTTP request lifecycles without provisioning dedicated message broker infrastructure (RabbitMQ or Kafka).

## Decision

Implement an in-process worker pool (`app/backend-api/internal/worker/worker.go`) backed by Redis lists. Producers enqueue JSON payloads via `LPUSH` and worker goroutines dequeue jobs via blocking `BRPOP` operations.

## Consequences

- Reuses the existing Redis 7 instance with no additional broker containers.
- Failed jobs are routed to a Redis dead-letter queue after retry exhaustion.
- Does not provide built-in DAG scheduling or priority queues beyond separate Redis list keys.
