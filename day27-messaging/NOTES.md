# Day 27 — Quick Reference (messaging / event-driven)

## Sync vs async
- **Sync (REST/gRPC):** request → wait → response. Immediate consistency, tight coupling, cascading failures.
- **Async (messaging):** emit event → broker → consumers later. Decoupled, resilient, scalable; eventual consistency.
- Use async for: deferrable work, fan-out, load spikes, independent deploys.

## Kafka vs RabbitMQ
| | Kafka | RabbitMQ |
|---|---|---|
| Model | Append-only **log**, retained | Queue/broker, removed on ack |
| Consume | Offset-based, **replayable** | Once, then gone |
| Throughput | Very high (streams) | High |
| Ordering | Per-partition | Per-queue (weaker) |
| Best for | Event streaming, sourcing, replay | Task queues, RPC, rich routing |

## Kafka concepts
- **Topic** — named stream, split into **partitions** for parallelism.
- **Partition** — ordered append-only log; ordering only *within* it. Key → partition.
- **Consumer group** — shares a topic's work; one partition → one consumer. Scale up to #partitions.
- **Offset** — consumer position; committing == "processed up to here."

## Delivery guarantees
- **At-most-once** — commit before processing; may lose. Rare.
- **At-least-once** — process then commit; may **duplicate**. The default.
- **Exactly-once** — narrow/expensive; in practice = at-least-once + **idempotent consumer** ("effectively once").

## Idempotency (non-negotiable)
```go
if alreadyProcessed(evt.EventID) { return nil } // dedup (atomic check+record)
markPaid(evt.OrderID)                            // or naturally idempotent op
recordProcessed(evt.EventID)
```
Techniques: dedup on event ID · naturally idempotent op (`SET status=paid`) · upsert by business key.

## Production concerns
- **DLQ** — after N retries, route aside; without it a poison message blocks the partition.
- **Consumer lag** — backlog vs latest offset; monitor it, scale or shed when it grows.
- **Schema evolution** — events are a contract; backward-compatible changes + schema registry.
- **Outbox pattern** — write row + event row in one DB tx, relay publishes; solves dual-write.
- **Ordering vs parallelism** — more partitions = more parallelism, ordering only per key.

## Key terms
**Broker** · **Topic** · **Partition** · **Offset** · **Consumer group** · **At-least-once** · **Idempotency / dedup key** · **DLQ (dead-letter queue)** · **Consumer lag** · **Poison message** · **Outbox pattern** · **Choreography vs orchestration** · **Saga / compensating action** · **Schema registry**.

> Go clients: `segmentio/kafka-go`, `rabbitmq/amqp091-go`. The runnable exercises use an in-memory bus so they work offline; real broker code lives in `solutions/broker_reference.go` (build-ignored) with a `docker-compose.yml`.
