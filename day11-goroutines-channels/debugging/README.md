# Day 11 debugging — the replicas that never go home

**Phase 3 · Concurrency · goroutine leak on an unbuffered channel**

## Symptom

We query several "replicas" concurrently and return the *first* answer (a common
"hedged request" pattern). Each replica runs in its own goroutine and sends its
result on a shared channel; we receive one value (the winner) and return.

Functionally it looks fine — the right value comes back. But the process is
quietly bleeding goroutines. Run it:

```bash
cd bugged
go run .
```

Output:

```
after 20 rounds of 5 replicas: 80 goroutines still alive
GOROUTINE LEAK detected: 80 stranded goroutines blocked on `out <- id`
```

80 = 20 rounds × 4 losers each. In a real server, every request would strand
N−1 goroutines forever. Memory climbs, the goroutine profile balloons, and
eventually you OOM. This is the classic "works in dev, dies in prod at 2 a.m."

## Hint

Ask the Day 11 question: **"What stops each goroutine?"** The winner is received,
so its `out <- id` completes. What about the four losers? The channel is
*unbuffered*, and you only ever receive once. Where are the losers blocked?

Confirm the leak count grows with `rounds` and inspect with a goroutine profile:

```bash
go run .            # watch the leaked count
GODEBUG=schedtrace=1000 go run .   # optional: scheduler shows parked goroutines
```

## How to reproduce

`go run .` in `bugged/` — prints a non-zero (and growing-with-rounds) leak count
every time. It is deterministic, not flaky.

---

<details>
<summary><strong>Solution &amp; why</strong></summary>

### Root cause

An **unbuffered** channel send blocks until a receiver is ready. We launch 5
senders but receive exactly **one** value. The 4 losers reach `out <- id`, find
no receiver (we already returned), and **block there forever** — a textbook
goroutine leak. Nothing ever cancels or drains them.

The bug is invisible to functional testing: the return value is correct. Only a
goroutine count (`runtime.NumGoroutine()`), a leak detector (`go.uber.org/goleak`),
or a pprof goroutine profile reveals it.

### The fix

Give every one-shot sender a slot so it can complete its send and exit, even
though we only consume one value:

```go
out := make(chan int, replicas) // buffered to the number of senders
```

Now each goroutine's `out <- id` always succeeds — there's a free buffer slot —
so all five return promptly. `fixed/` prints `0 goroutines still alive`.

Other correct fixes, depending on the situation:

- **Context / done-channel + `select`** on the send (Day 13/15): the losers see
  cancellation and `return` instead of blocking. Best when the number of senders
  is unbounded or long-lived.
- **Drain the channel**: keep receiving the remaining N−1 values (wasteful, but
  it unblocks the senders).

Buffering to the *known, fixed* sender count is the simplest correct fix for a
hedged/first-wins fan-out of one-shot goroutines.

### The rule

> Every goroutine you launch needs an exit condition you can state in one
> sentence. A send on an unbuffered channel is only an exit if a receive is
> *guaranteed* to happen. "Someone will probably read it" is how leaks ship.

Verify: `go run -race .` in `fixed/` is clean and reports no leak.

</details>
