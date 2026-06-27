# Day 14 — Worker Pools, Fan-Out/In, errgroup: Interview Q&A

Model answers in `<details>`.

---

**1. Draw a worker pool. Who closes `jobs`? Who closes `results`? Why that ordering?**

<details>
<summary>Answer</summary>

A producer feeds a `jobs` channel; N worker goroutines `range jobs` and write to a
`results` channel; a consumer `range results`.

- The **producer** closes `jobs` when there's no more work → each worker's `range
  jobs` ends and the worker returns.
- A **single coordinator** closes `results` *after* `wg.Wait()` (all workers done):
  `go func(){ wg.Wait(); close(results) }()` → the consumer's `range results` ends.

Ordering matters: close `results` too early (while workers still send) → panic and
lost data; never close it → consumer hangs. Close `jobs` first to stop the workers,
then `results` once they've all returned.
</details>

---

**2. What's the difference between fan-out and fan-in?**

<details>
<summary>Answer</summary>

**Fan-out:** multiple goroutines read from the *same* channel, distributing work
across them (that's the worker pool — N workers over one `jobs` channel).
**Fan-in:** merge multiple input channels into one output channel, typically with a
goroutine per input copying into the shared output and a coordinator that closes
the output after a `WaitGroup`. Fan-out to parallelize, fan-in to collect —
together they're a parallel map-reduce.
</details>

---

**3. How can a worker pool leak goroutines, and how do you prevent it?**

<details>
<summary>Answer</summary>

A worker blocks forever on `results <- x` because the consumer stopped draining
(e.g. it returned after an error), and nothing closes or drains the channel.
Prevent it with a cancellation path: workers `select` on `ctx.Done()` while
sending, and the consumer cancels when it's done. `errgroup.WithContext` gives you
this shared cancellation for free. Also: always close `jobs` (else workers never
return) and `results` (else the consumer hangs).
</details>

---

**4. Why does `errgroup` need a context? What does the first error trigger?**

<details>
<summary>Answer</summary>

`g, ctx := errgroup.WithContext(parent)` ties the group to a derived context. When
the first `g.Go` task returns a non-nil error, the group **cancels that ctx**.
Tasks that are selecting on `ctx.Done()` then bail out instead of running to
completion or leaking. `g.Wait()` waits for all tasks and returns the **first**
non-nil error. Without the context, a failure wouldn't stop the siblings.
</details>

---

**5. How do you bound concurrency, and how do you choose the bound?**

<details>
<summary>Answer</summary>

Run a fixed number N of workers pulling from a shared channel, or use
`errgroup`'s `g.SetLimit(N)`. Choose N by workload: **CPU-bound** work →
~`runtime.NumCPU()` (more just adds context-switch overhead beyond real
parallelism); **I/O-bound** work → higher (dozens/hundreds, since workers mostly
wait), but capped by **downstream capacity** — the pool exists to protect the DB/
API you're calling, so N is a capacity-planning decision, ideally benchmarked.
</details>

---

**6. Who is allowed to close a channel, and what happens if you get it wrong?**

<details>
<summary>Answer</summary>

Only the **sending** side closes, and exactly **once**. Closing from the receiver,
closing twice, or sending after close all **panic** (`close of closed channel` /
`send on closed channel`). With multiple senders there's no single owner, so a
coordinator closes after a `WaitGroup` confirms all senders finished.
</details>

---

**7. Why is `wg.Add()` inside the goroutine a bug?**

<details>
<summary>Answer</summary>

It races with `wg.Wait()`: the goroutine may not be scheduled before `Wait` runs,
so the counter is 0 and `Wait` returns early. In a pool, the close-coordinator then
closes `results` while workers are still sending → panic or lost results. Always
`wg.Add(1)` before the `go`.
</details>

---

**8. When would you prefer `errgroup.SetLimit` over a hand-rolled worker pool?**

<details>
<summary>Answer</summary>

Almost always, when the work can fail. `errgroup` with `SetLimit(n)` gives you a
bounded pool **plus** first-error capture, context cancellation of siblings, and
clean `Wait()` — the exact plumbing that's error-prone by hand (a common source of
leaked goroutines). Reach for a hand-rolled pool only when you need a behavior
errgroup doesn't model (e.g. custom result streaming, per-worker state, or
collecting *all* errors rather than the first).
</details>

---

**9. How do you collect results *and* errors from a pool without leaking?**

<details>
<summary>Answer</summary>

Two common shapes: (a) a result channel plus `errgroup` for errors/cancellation —
workers `select` between sending a result and `ctx.Done()`; or (b) a `results`
channel of a struct `{Value, Err}` so each item carries its own error, with the
consumer deciding what to do. Either way every send has a cancellation path and the
sender closes `results` after `wg.Wait()`, so no worker strands.
</details>

---

**10. Your pool processes a million items. What questions does a staff engineer ask?**

<details>
<summary>Answer</summary>

What's the **downstream's** capacity? — set N to that, not to CPU count. Is there
**backpressure** when the queue fills (does the producer block, or buffer
unboundedly)? What are the **failure semantics** — retry with backoff, dead-letter,
or fail-fast via errgroup? And ultimately: at this scale, is in-process concurrency
even right, or do you need a real distributed queue (Kafka/SQS) with horizontal
workers? Concurrency bound is a scaling and failure-isolation decision.
</details>
