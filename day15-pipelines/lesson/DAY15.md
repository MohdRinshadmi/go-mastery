# Day 15 — Pipelines & the Phase 3 Capstone

> Mentor note: A pipeline is just stages connected by channels, each stage a goroutine that reads from the previous and writes to the next. It's how you stream-process data with bounded memory and natural parallelism — think a Unix pipe (`cat | grep | sort`) but type-safe and concurrent. The thing that separates a toy pipeline from a production one is **cancellation and cleanup**: every stage must shut down cleanly when downstream stops caring. Get that right and you never leak a goroutine again.

---

## 1. The pipeline pattern

Each stage: `func(in <-chan T) <-chan U` — takes a read channel, returns a read channel, runs a goroutine inside.

```go
// Stage 1: source
func gen(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, n := range nums { out <- n }
    }()
    return out
}

// Stage 2: transform
func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in { out <- n * n }
    }()
    return out
}

// Compose: gen -> square -> consume
for v := range square(gen(1, 2, 3)) {
    fmt.Println(v) // 1 4 9
}
```

Each stage closes its output when its input is drained (`defer close(out)`), which terminates the next stage's `range`. The closes cascade down the pipeline naturally.

**Why bounded memory:** only one item per stage is "in flight" (unbuffered) — you can stream a billion items through without loading them all. Add buffered channels to decouple stage speeds.

## 2. Cancellation — the part everyone forgets

If the consumer stops early (an error, a `break`, enough results), upstream stages block forever on `out <- v` → **goroutine leak**. Fix: thread a `context` (or a `done` channel) and `select` on it in every stage.

```go
func square(ctx context.Context, in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            select {
            case out <- n * n:        // normal send
            case <-ctx.Done():        // downstream gave up -> exit, no leak
                return
            }
        }
    }()
    return out
}
```

Now `cancel()` (deferred by the consumer) tears down the whole pipeline. **Every send and receive in a long-lived pipeline should `select` on `ctx.Done()`.**

**Senior take:** The leak test: "if my consumer returns early, does every upstream goroutine exit?" If you can't answer yes, you have a leak. Run with a goroutine-count check or `go.uber.org/goleak` in tests. A leak per request = a slow memory death in prod.

## 3. Adding parallelism to a stage

A slow stage (e.g. network I/O) becomes the bottleneck. Fan-out that stage into M workers, then fan-in their outputs (Day 14) — same merge pattern slots straight into a pipeline. Pipelines + fan-out/in = streaming map-reduce with bounded resources.

## Common mistakes
1. A stage that doesn't `close` its output → next stage's `range` hangs forever.
2. No cancellation path → early consumer exit leaks every upstream goroutine.
3. Closing a channel from the receiving side or twice → panic.
4. Sharing one mutable value across stages by pointer and mutating it → race (send copies/ownership, don't share).
5. Unbuffered everywhere when stages have very different speeds → throughput limited by the slowest; add buffering deliberately.

## Performance
- Buffer channels to smooth bursty stages; size by profiling, not guessing.
- Each stage is a goroutine + channel hop — for trivially cheap per-item work, the channel overhead can exceed the work; batch items or skip the pipeline.
- Parallelize only the stage the profiler says is the bottleneck.

---

## Expert Thinking Mode — "stream-process this data"

- **Beginner:** "Load it all into a slice, loop, transform." (Fine until the data doesn't fit in memory.)
- **Senior:** "Pipeline of stages, bounded memory, context cancellation in every stage, parallelize the slow stage with fan-out/in. Prove no goroutine leak."
- **Staff:** "Where's backpressure? If a downstream stalls, does the source slow down or buffer unboundedly? What's the failure semantics — drop, retry, dead-letter?"
- **Architect:** "In-process pipeline vs a distributed streaming system (Kafka Streams, Flink) — chosen by data volume, durability, and restart semantics. The channel pipeline is the single-process version of the same dataflow idea (Phase 6)."

---

## Real-world use

- **ETL / log processing / media transcoding**: read → decode → transform → encode → write, each a stage, the slow one fanned out.
- **Go's own tooling** uses this dataflow style; the classic Go blog "Pipelines and cancellation" is required reading.
- **Cancellation discipline** from pipelines is the same discipline that keeps request-scoped goroutines from leaking in every Go service.

---

## Interview Questions

1. What is a pipeline stage's signature and why does each stage own a goroutine?
2. How does closing cascade through a pipeline, and what breaks if a stage forgets to close?
3. How does an early-exiting consumer leak goroutines, and how does context fix it?
4. How do you parallelize a single slow stage?
5. When would you add buffering to a stage's channel?
6. How would you test that a pipeline leaks no goroutines?
7. When is an in-process channel pipeline the wrong tool?

---

## Phase 3 Capstone (in `../exercises/` and `../solutions/`)

Two deliverables that combine everything from Phase 3:

1. **Concurrent URL checker** — given a list of URLs, check their reachability with **bounded concurrency** and a per-request **timeout** (context), returning a `map[url]result`. (Uses a fake checker so it runs offline and deterministically — swap in real `http.Get` for homework.)

2. **Cancellable pipeline** — `gen → square → filterEven`, each stage `select`ing on `ctx.Done()`, with the consumer taking only the first few results and cancelling — and you proving (by design) that no stage leaks.

Run both with `-race`. Passing this completes Phase 3 — you'll have goroutines, channels, select, context, mutexes, worker pools, fan-out/in, and pipelines all under your belt. Phase 4 (building real backends) is next.
