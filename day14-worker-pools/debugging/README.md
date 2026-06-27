# Day 14 debugging — the worker pool that never finishes

**Phase 3 · Concurrency · results channel never closed (consumer deadlock)**

## Symptom

A textbook bounded worker pool: a producer feeds `jobs`, 4 workers square the
numbers and write to `results`, and the consumer `range`s over `results` to sum
them. It computes the right values... and then **hangs forever**. Run it:

```bash
cd bugged
go run .
```

```
DEADLOCK detected: consumer is blocked on `range results`
  cause: results channel was never closed after wg.Wait()
  workers all returned (jobs closed), but range results never sees a close
exit status 1
```

(The program ships with a 2-second watchdog so it reports and exits instead of
hanging your terminal. In real code there's no watchdog — the goroutine just
hangs and the process wedges.)

## Hint

Trace the **close discipline** of every channel:

- `jobs` — who closes it? (The producer does — good. That's why workers' `range
  jobs` ends and they return.)
- `results` — who closes it? Workers are the senders... and nothing closes it.

What does `for r := range results` do after the last value, if `results` is never
closed?

## How to reproduce

`go run .` in `bugged/` — after draining the last result the consumer's `range
results` blocks waiting for a close. The watchdog reports the deadlock after ~2s
and exits non-zero. Deterministic every run.

---

<details>
<summary><strong>Solution &amp; why</strong></summary>

### Root cause

`results` is **never closed**. The pool gets the `jobs` side right — the producer
closes `jobs`, so each worker's `range jobs` ends and the workers return. But
`range` over a channel only terminates when the channel is **closed and drained**.
With no close on `results`, the consumer drains all 20 values and then blocks on
the next receive forever — a deadlock. (If it were the *only* goroutine left, Go
would `fatal error: all goroutines are asleep - deadlock!`; here the watchdog
catches it first.)

### The fix

The **senders** close the channel — and only after *all* of them are done. With
multiple worker-senders, a single coordinator goroutine waits on the WaitGroup and
closes once:

```go
go func() {
    wg.Wait()       // all workers have returned
    close(results)  // exactly one close, by the owner of sending
}()
```

Now the consumer's `for r := range results` ends cleanly when the last result is
drained. `fixed/` collects all 20 results and is clean under `-race`.

### Why not close earlier / from the consumer / per worker?

- **From the consumer:** the receiver must never close — sending on a closed
  channel panics, and other workers may still be sending.
- **Per worker (each calls `close`):** the first close succeeds, the next worker's
  `results <- x` or second `close` **panics**.
- **Before `wg.Wait()`:** you'd close while workers are still sending → panic and
  lost results.

### The rules

> 1. **The sender closes; only once.** With many senders, one coordinator closes
>    after `wg.Wait()`.
> 2. **Close `jobs` to stop the workers; close `results` to stop the consumer.**
>    Forget the first → workers `range` forever → `wg.Wait()` never returns.
>    Forget the second → consumer `range`s forever.
> 3. A leak-free pool has a stop signal for every `range`.

Verify: `go run -race .` in `fixed/` collects all results and exits cleanly.

</details>
