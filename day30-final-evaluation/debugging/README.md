# Day 30 debugging — the order→payment wiring that leaks a goroutine per timeout

**Phase 6 · microservices wiring · goroutine leaks under timeout**

> Stdlib only. The "PaymentService" is a slow function and the inter-service hop
> is a channel, so the leak is real and reproducible offline — and visible via
> `runtime.NumGoroutine()`.

## Symptom

The OrderService calls the (slow) PaymentService in a goroutine and waits for the
result, giving up after a timeout — a standard resilience pattern. The system
works, orders that time out are reported correctly... but in production the
process's goroutine count climbs forever. Every order whose payment exceeds the
timeout permanently leaks a goroutine; under sustained load the process eventually
exhausts memory and falls over.

```bash
cd bugged
go run -race .
```

Expected: goroutine count returns to ~baseline after the work finishes.
Actual: ~200 goroutines leaked — one per timed-out order — and they never go away.

## Hint

The payment goroutine reports its result by **sending on a channel**. The caller
receives from that channel — *until it times out*, at which point it stops
receiving. Now think about what the payment goroutine does ~50ms later when the
slow charge finally returns and it tries to send. Who's listening? What does a
send on an unbuffered channel do when there's no receiver?

## How to reproduce

`go run -race .` in `bugged/`. It places 200 orders with a timeout shorter than
the payment latency (so every payment "times out"), waits for the slow payments
to finish, then prints goroutines before vs after.

---

<details>
<summary><strong>Solution &amp; why</strong></summary>

### Root cause

The result channel is **unbuffered**, and the caller stops receiving once it times
out:

```go
resultCh := make(chan PaymentResult) // unbuffered
go func() { resultCh <- charge(orderID) }() // late send blocks forever
select {
case res := <-resultCh:
    return res, true
case <-time.After(timeout):
    return PaymentResult{}, false // caller walks away, stops receiving
}
```

A send on an unbuffered channel blocks until a receiver is ready. After the
timeout fires, there is no receiver — ever. So when the slow `charge` finally
returns, the goroutine blocks on `resultCh <- ...` **forever**. It's parked,
holding its stack, counted in `runtime.NumGoroutine()`, and never reclaimed (a
blocked goroutine is not garbage — something still "could" receive). One leak per
timed-out order; under load this is an unbounded resource leak that ends in OOM.
This is the single most common goroutine-leak shape in Go service code.

### The fix

Make the channel **buffered with capacity 1** so the goroutine's single send
always succeeds, even when nobody is receiving:

```go
resultCh := make(chan PaymentResult, 1) // buffered, cap 1
go func() { resultCh <- charge(orderID) }() // always fits -> goroutine exits
```

The buffer slot is always free (only one send happens), so the send never blocks;
the goroutine completes and is reclaimed whether or not the caller timed out.

The fuller, production version *also* threads a `context` into the downstream call
so the goroutine can **abort the work** on timeout instead of finishing it and
discarding the result:

```go
ctx, cancel := context.WithTimeout(parent, timeout)
defer cancel()
go func() { resultCh <- chargeCtx(ctx, orderID) }() // resultCh still buffered(1)
```

Rules:

> 1. A goroutine that sends a result the caller might abandon needs a **buffered
>    channel (cap 1)** — otherwise a late send blocks forever and leaks.
> 2. Every goroutine you start needs a guaranteed way to **finish** (a receiver, a
>    buffer slot, or a `ctx.Done()` exit). "Fire and forget" is "fire and leak."
> 3. Watch `runtime.NumGoroutine()` (or a goroutine-count metric / `pprof` goroutine
>    profile) — a steadily climbing count is a leak, not "busy."

`fixed/` returns to baseline goroutine count; the bugged version leaks one per timeout.

</details>
