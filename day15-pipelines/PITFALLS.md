# Day 15 — Pipelines & Cancellation: Pitfalls

Concurrency gotchas as **Trap → Why → Fix**.

---

### 1. No cancellation path → early consumer leaks every upstream stage

**Trap:** Stages only `out <- v` with no `select` on a done/ctx channel. The
consumer `break`s after a few results.

**Why:** Once the consumer stops receiving, the last stage blocks on its send,
which blocks the stage before it, all the way up. Every upstream goroutine
strands — a leak per early exit.

**Fix:** Thread a `context` (or `done` channel) and `select { case out <- v:
case <-ctx.Done(): return }` in **every** stage. The consumer `defer cancel()`s.

---

### 2. A stage that forgets to `close` its output

**Trap:** A producing goroutine returns (or its `range in` ends) without
`close(out)`.

**Why:** The next stage's `for v := range out` waits for a close that never
comes → it hangs forever (and so does everything downstream).

**Fix:** `defer close(out)` at the top of every stage's goroutine. The close
cascades down as each stage's input drains.

---

### 3. Closing a channel from the receiving side, or twice

**Trap:** A downstream stage closes its *input*, or two goroutines close the same
output.

**Why:** Only the sending side closes, exactly once. A second close or a
close-then-send **panics**.

**Fix:** Each stage closes only the channel it sends on (its `out`), via `defer
close(out)`, exactly once.

---

### 4. Sharing a mutable value across stages by pointer

**Trap:** Stages pass `*T` down the pipeline and each mutates the pointed-to
value.

**Why:** Multiple stages touching the same memory concurrently is a data race
(`-race` flags it). Pipelines work because each item is *owned* by one stage at a
time.

**Fix:** Send values (or freshly-allocated copies); transfer ownership, don't
share. If you must share, synchronize — but that defeats the pipeline's point.

---

### 5. `cancel()` not deferred (or only called on the happy path)

**Trap:** Calling `cancel()` only after a successful loop, not on early `break`/
error/`return`.

**Why:** Any path that leaves without calling `cancel` leaks the context's
resources and orphans upstream stages — exactly the leak you were trying to avoid.

**Fix:** `defer cancel()` right after creating the context, **and** call `cancel()`
explicitly at an early break if you want upstream to stop *before* `main` returns.
Double-cancel is harmless.

---

### 6. Unbuffered everywhere when stages have very different speeds

**Trap:** Every stage channel is unbuffered, so the whole pipeline runs at the
speed of the slowest stage even when a small buffer would smooth bursts.

**Why:** With zero buffering each hop is a strict rendezvous; a bursty fast stage
keeps stalling on a slow one.

**Fix:** Add buffering **deliberately**, sized by profiling, to decouple stages
with mismatched speeds. Don't buffer blindly — measure first.

---

### 7. Pipelining trivially cheap work

**Trap:** Building a multi-stage channel pipeline for per-item work that's a few
nanoseconds (e.g. adding 1).

**Why:** Each stage is a goroutine + channel hop; the channel/scheduling overhead
can dwarf the actual work, making the pipeline *slower* than a plain loop.

**Fix:** Pipeline when stages do real work (I/O, heavy compute) or when streaming
bounded memory matters. For trivial transforms, batch items or just loop.
