# Day 14 — Worker Pools, Fan-Out/In, errgroup: Pitfalls

Concurrency gotchas as **Trap → Why → Fix**.

---

### 1. Never closing `results` → consumer hangs

**Trap:** Producer closes `jobs`, workers return, but nothing closes `results`.
The consumer's `for r := range results` blocks forever after the last value.

**Why:** `range` over a channel only ends on close-and-drain. No close = the
consumer waits for a value that never comes → deadlock.

**Fix:** One coordinator closes `results` after all senders finish:
`go func(){ wg.Wait(); close(results) }()`.

---

### 2. Never closing `jobs` → workers (and `wg.Wait`) hang

**Trap:** The producer forgets `close(jobs)` after sending all work.

**Why:** Each worker's `for j := range jobs` waits for more jobs forever, so no
worker returns, so `wg.Wait()` never unblocks and the coordinator never closes
`results`. The whole pool stalls.

**Fix:** The producer closes `jobs` exactly once when there's no more work.

---

### 3. Closing a channel from the consumer or from multiple goroutines

**Trap:** The receiver closes `results`, or every worker calls `close(results)`.

**Why:** Only the sending side may close, exactly once. A second close panics
(`close of closed channel`); a send after close panics (`send on closed
channel`). Multiple senders have no single owner.

**Fix:** Close from one coordinator after `wg.Wait()`. Receivers never close.

---

### 4. Unbounded fan-out

**Trap:** `for _, x := range millions { go work(x) }`.

**Why:** Millions of goroutines blow up memory and, worse, hammer the downstream
(DB/API) with millions of concurrent calls — a self-inflicted DDoS. "Goroutines
are cheap" is true; the thing they call is not.

**Fix:** A **bounded** worker pool (fixed N), or `errgroup` with `g.SetLimit(n)`.
Choose N from downstream capacity, not your CPU count, for I/O work.

---

### 5. Worker blocked on `results <- x` after the consumer quits

**Trap:** The consumer stops early (an error, enough results) and stops draining
`results`. Workers block forever on their next send → leak.

**Why:** An unbuffered/full `results` send needs a receiver. If the receiver is
gone, senders strand.

**Fix:** Give workers a cancellation path: `select { case results <- x: case
<-ctx.Done(): return }`, and cancel when the consumer is done. `errgroup`'s shared
context does this for you.

---

### 6. `errgroup` without its context (no cancellation on error)

**Trap:** Using a plain `errgroup.Group{}` (or ignoring the returned `ctx`) so a
failing task doesn't stop the others.

**Why:** Without the shared context, the first error doesn't cancel sibling tasks
— they keep running (and maybe leak) after the result is already doomed.

**Fix:** `g, ctx := errgroup.WithContext(parent)`; pass that `ctx` into every
task and select on it. The first non-nil error cancels `ctx`; `g.Wait()` returns
it.

---

### 7. `wg.Add` inside the worker goroutine

**Trap:** `go func(){ wg.Add(1); defer wg.Done(); ... }()`.

**Why:** Races with `wg.Wait()` / the close-coordinator; the count may be 0 when
`Wait` runs, closing `results` while workers still send → panic or lost data.

**Fix:** `wg.Add(1)` in the loop before `go`.
