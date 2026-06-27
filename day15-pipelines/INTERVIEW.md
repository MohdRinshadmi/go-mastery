# Day 15 — Pipelines & Cancellation: Interview Q&A

Model answers in `<details>`.

---

**1. What is a pipeline stage's signature, and why does each stage own a goroutine?**

<details>
<summary>Answer</summary>

A stage is `func(in <-chan T) <-chan U` (often `func(ctx, in) <-chan U`): it takes
a receive-only input channel and returns a receive-only output channel, running a
goroutine inside that reads from `in`, transforms, and writes to `out`. It owns a
goroutine because each stage must run **concurrently** with the others — that's
what lets data stream through with only one item per stage in flight (bounded
memory) and lets stages run in parallel. The directional channels encode the
producer/consumer contract.
</details>

---

**2. How does closing cascade through a pipeline, and what breaks if a stage forgets to close?**

<details>
<summary>Answer</summary>

Each stage does `defer close(out)`. When the source closes its output, the next
stage's `for v := range in` drains and ends, its goroutine returns, its own `defer
close(out)` fires — and so on down the chain. The closes cascade naturally. If a
stage forgets to close its output, the next stage's `range` waits forever for a
close that never comes → that stage and everything downstream hang (deadlock).
</details>

---

**3. How does an early-exiting consumer leak goroutines, and how does context fix it?**

<details>
<summary>Answer</summary>

If the consumer `break`s early, it stops receiving. The last stage blocks on its
send (no receiver), which blocks the stage feeding it, all the way up — every
upstream goroutine strands. A bare channel has no "reader left" signal. Context
fixes it: each stage `select`s `case out <- v:` against `case <-ctx.Done():`, and
the consumer `cancel()`s when it leaves. The cancel closes `ctx.Done()`, every
blocked send loses the race to it and returns, `defer close(out)` runs, and the
whole pipeline tears down.
</details>

---

**4. How do you parallelize a single slow stage?**

<details>
<summary>Answer</summary>

Fan-out that stage into M worker goroutines all reading from the same input
channel, then fan-in their outputs back into one channel (the Day 14 merge
pattern). Each worker is identical; the merge uses a `WaitGroup` and a coordinator
that closes the merged output after all workers finish. This turns the bottleneck
stage into a parallel sub-pipeline while the rest stays sequential — streaming
map-reduce with bounded resources. Parallelize only the stage the profiler flags.
</details>

---

**5. When would you add buffering to a stage's channel?**

<details>
<summary>Answer</summary>

When adjacent stages have **mismatched, bursty speeds** and a small buffer lets a
fast stage keep working through a slow stage's brief stalls, improving throughput.
Size it by **profiling**, not guessing — buffering trades memory and latency for
smoothing. Unbuffered-everywhere is the correct default (bounded memory, clear
synchronization); add buffers deliberately where measurement shows a bottleneck.
</details>

---

**6. How would you test that a pipeline leaks no goroutines?**

<details>
<summary>Answer</summary>

Record `runtime.NumGoroutine()` at a baseline, run the pipeline including an early
consumer exit, give cancelled stages a moment to unwind, then assert the count
returns to baseline. More robustly, use `go.uber.org/goleak` (e.g.
`goleak.VerifyTestMain`) which fails the test if any goroutine outlives it. The key
test case is **early consumer exit**, since that's where missing cancellation
leaks.
</details>

---

**7. When is an in-process channel pipeline the wrong tool?**

<details>
<summary>Answer</summary>

When you need **durability, restart-safety, or horizontal scale** beyond one
process: if a crash must not lose in-flight items, if data volume exceeds one
machine, or if stages must scale independently, you want a real streaming/queue
system (Kafka, NATS, SQS, Flink). The channel pipeline is the single-process
version of the same dataflow idea — great for in-memory streaming with bounded
resources, wrong when the work must survive process boundaries or restarts.
</details>

---

**8. Why is `defer cancel()` essential in the consumer even if you also call `cancel()` on the early break?**

<details>
<summary>Answer</summary>

`defer cancel()` guarantees teardown on **every** return path — normal completion,
an error mid-loop, or a panic — not just the one early-break path you remembered.
The explicit `cancel()` at the break stops upstream *promptly* (before `main`
continues), while the `defer` is the safety net for all other exits. Calling cancel
twice is harmless, so you keep both.
</details>

---

**9. What's the difference between a pipeline's cancellation discipline and a worker pool's close discipline?**

<details>
<summary>Answer</summary>

Close discipline (Day 14) is about **normal completion**: the sender closes each
channel so downstream `range`s end — it propagates "no more data" *forward*.
Cancellation discipline (Day 15) is about **early/abnormal stop**: a context
propagates "stop now" *backward/everywhere* so stages quit even when there's still
data to produce. A robust pipeline needs both: `defer close(out)` for clean drain,
and `select` on `ctx.Done()` for early teardown.
</details>

---

**10. Why does "only one item per stage in flight" give bounded memory, and when does that break?**

<details>
<summary>Answer</summary>

With unbuffered channels, a stage can't produce its next item until the next stage
takes the current one, so at most one item sits at each hop regardless of total
input size — you can stream a billion items through a fixed-size pipeline. It
breaks if you add large buffers (memory grows with buffer size) or if a stage
accumulates state (e.g. sorting/grouping requires holding many items). Then memory
is bounded by the buffer/state, not by "one per stage."
</details>
