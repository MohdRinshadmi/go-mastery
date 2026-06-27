# Day 27 debugging — the customer who got charged twice

**Phase 6 · messaging · at-least-once delivery & idempotency**

> Stdlib only. The broker is a channel that redelivers one event, simulating the
> lost-ack redelivery every real broker (Kafka, RabbitMQ) does. No external deps.

## Symptom

A payment consumer reads `OrderPaid` events from a broker and charges the
customer. Most of the time it's fine. But occasionally — after a deploy, a
rebalance, or a network blip — a customer is charged **twice** for one order.

```bash
cd bugged
go run -race .
```

Expected: `order-1` charged `$49.99` once.
Actual: charged `$99.98` — the same event was processed twice.

## Hint

Brokers don't guarantee exactly-once delivery; the realistic guarantee is
**at-least-once**. If the consumer processes a message but crashes (or the
network drops) before the ack reaches the broker, the broker assumes failure and
**redelivers**. So your consumer *will* see some events more than once. Given
that, what must the consumer do before it applies a side effect? Look at what the
event carries that you're not using.

## How to reproduce

`go run -race .` in `bugged/`. The broker is told to redeliver `evt-1` once
(a lost ack), and the consumer charges on every delivery.

---

<details>
<summary><strong>Solution &amp; why</strong></summary>

### Root cause

The consumer treats every delivery as a brand-new event and applies the side
effect unconditionally:

```go
for e := range b.ch {
    led.charge(e.OrderID, e.Amount) // runs once PER DELIVERY, not per event
}
```

At-least-once delivery means duplicates are normal, not exceptional. "We use a
queue so we won't lose messages" is only half the contract — the other half is
"we will sometimes get the same message twice, so the consumer must tolerate it."
A non-idempotent consumer under at-least-once delivery is a latent double-charge
(or double-email, double-shipment) waiting for the next rebalance.

### The fix

Make the consumer **idempotent**: dedupe on the event's unique ID and skip events
you've already processed.

```go
if !processed.markIfNew(e.EventID) {
    continue // duplicate -> no-op
}
led.charge(e.OrderID, e.Amount)
```

The check-and-record must be **atomic** (one lock, or an `INSERT ... ON CONFLICT`
/ Redis `SETNX`), otherwise two concurrent deliveries can both pass the "have I
seen it?" check before either records it — and you're back to a double charge.

Other ways to reach idempotency, same idea:

> 1. Dedup on a unique **event ID** (shown here). Store seen IDs in Redis/DB with a TTL.
> 2. Make the operation **naturally idempotent**: `SET status=paid` instead of
>    `balance += amount`. Re-applying it changes nothing.
> 3. **Upsert** keyed by the business ID (`INSERT ... ON CONFLICT DO NOTHING`).

Rules:

> - Design every consumer assuming each event arrives **1 or more** times, possibly
>   out of order. At-least-once + idempotent consumer = "effectively once."
> - The dedup check and the side effect must be atomic, or the dedup itself races.
> - Don't chase broker-level "exactly-once" — it's narrow and expensive; idempotent
>   consumers are the practical answer.

`fixed/` charges exactly once; the bugged version charges per delivery.

</details>
