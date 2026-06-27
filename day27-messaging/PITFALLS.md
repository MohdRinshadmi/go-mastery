# Day 27 — Pitfalls (Kafka / RabbitMQ / event-driven)

Format: **Trap → Why → Fix**.

### 1. Non-idempotent consumer under at-least-once
**Trap:** The consumer applies a side effect on every delivery, so a redelivered event double-charges / double-emails.
**Why:** Brokers guarantee at-least-once, not exactly-once; a lost ack causes redelivery.
**Fix:** Dedupe on a unique event ID (atomic check+record), use naturally idempotent ops, or upsert by business key.

### 2. Committing the offset before processing
**Trap:** A crash between committing and finishing the work loses the message — it's never reprocessed.
**Why:** Committing the offset tells the broker "done"; if you commit first you've lied.
**Fix:** Process, *then* commit/ack. That gives at-least-once (duplicates), which your idempotent consumer handles.

### 3. No dead-letter queue
**Trap:** One "poison" message that always fails blocks the partition/queue forever, stalling everything behind it.
**Why:** The broker keeps redelivering the same failing message; the consumer never makes progress.
**Fix:** After N retries, route the message to a DLQ for inspection and let the stream move on.

### 4. Assuming global ordering across partitions
**Trap:** Code relies on event B arriving after event A when they're on different partitions — it doesn't hold.
**Why:** Kafka guarantees order only *within* a partition; across partitions there's no ordering.
**Fix:** Key events that must be ordered by the same entity (e.g. `order_id`) so they land on one partition.

### 5. Ignoring consumer lag
**Trap:** Consumers silently fall behind; the backlog grows for hours before anyone notices.
**Why:** Without a lag metric, "slowly losing" looks identical to "healthy."
**Fix:** Monitor consumer lag (distance from the latest offset). Growing lag → scale consumers (up to #partitions) or shed load.

### 6. Breaking schema compatibility
**Trap:** A producer changes the event shape and old consumers crash on deploy.
**Why:** Events are a contract over time; consumers on the old schema can't parse the new bytes.
**Fix:** Make changes backward-compatible (add fields, don't remove/renumber); use a schema registry (Avro/Protobuf) and compatibility checks.

### 7. Using a queue where you need a synchronous answer
**Trap:** "Is this in stock right now?" routed through async messaging — the caller can't get an immediate reply.
**Why:** Messaging is fire-and-forget + eventual; it doesn't return a value to the producer.
**Fix:** Use sync gRPC/REST for queries needing an immediate answer; use events for fan-out work that can happen later.

### 8. No outbox → "DB committed but publish failed"
**Trap:** You write the order to the DB then publish the event; the publish fails and the event is lost forever.
**Why:** Two separate systems (DB + broker) with no shared transaction can partially succeed.
**Fix:** Outbox pattern — write the row and an event row in one DB transaction; a relay publishes from the outbox.
