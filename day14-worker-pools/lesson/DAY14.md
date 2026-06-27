# Day 14 — Worker Pools, Fan-Out, Fan-In, errgroup

> Mentor note: This is the day concurrency becomes *useful*. Spawning a goroutine per item sounds great until you have 10 million items and you OOM or hammer a downstream service into the ground. **Bounded concurrency** — a fixed pool of workers pulling from a queue — is the single most common production concurrency pattern in Go. Master the worker pool and you've got 80% of real-world concurrent backend code.

---

## 1. The worker pool

The shape: a `jobs` channel feeds N worker goroutines; each worker sends to a `results` channel.

```go
func workerPool(jobs <-chan int, results chan<- int, workers int) {
    var wg sync.WaitGroup
    for w := 0; w < workers; w++ {     // fixed number of workers
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := range jobs {       // pull until jobs is closed
                results <- process(j)
            }
        }()
    }
    go func() { wg.Wait(); close(results) }() // close results when all workers done
}
```

Key mechanics:
- **`range jobs`** in each worker: loops until the channel is closed and drained. This is how workers know to stop.
- **Producer closes `jobs`** when there's no more work — signals workers to exit.
- **`close(results)` after `wg.Wait()`** — so the consumer's `range results` terminates cleanly. Closing too early loses data; never closing hangs the consumer.

### Why bounded?
- Caps memory (N goroutines, not millions).
- Caps load on downstream (DB, API) — natural rate limiting.
- N is tunable: CPU-bound work → ~`runtime.NumCPU()`; I/O-bound work → higher (dozens/hundreds) since workers mostly wait.

**Senior take:** "Goroutines are cheap" is true (~2KB each) but downstream resources are not. A worker pool exists to protect the *thing you're calling*, not just your own memory. The pool size is a capacity-planning decision, not a default.

---

## 2. Fan-Out / Fan-In

- **Fan-Out**: multiple goroutines read from the *same* channel, distributing work (that's the worker pool — N workers fanning out over `jobs`).
- **Fan-In**: merge multiple result channels into one. The classic merge:

```go
func fanIn(channels ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup
    for _, c := range channels {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for v := range c {
                out <- v
            }
        }(c)
    }
    go func() { wg.Wait(); close(out) }()
    return out
}
```

Fan-out to parallelize, fan-in to collect. Together they're a parallel map-reduce.

---

## 3. Error handling in concurrency — `errgroup`

The hard part of concurrent code is errors: if one goroutine fails, how do you (a) capture the error, (b) cancel the others, (c) wait for cleanup? Doing this by hand with channels + context + WaitGroup is error-prone. `golang.org/x/sync/errgroup` packages it:

```go
import "golang.org/x/sync/errgroup"

func fetchAll(ctx context.Context, urls []string) error {
    g, ctx := errgroup.WithContext(ctx) // ctx is cancelled when any task errors
    for _, url := range urls {
        url := url
        g.Go(func() error {             // each task returns an error
            return fetch(ctx, url)
        })
    }
    return g.Wait()                     // returns the FIRST non-nil error
}
```

- First error cancels the shared `ctx` → other tasks see `ctx.Done()` and bail.
- `g.Wait()` blocks until all return and gives you the first error.
- `g.SetLimit(n)` bounds concurrency — turning errgroup into a worker pool with error handling for free. This is the modern idiomatic choice.

**Senior take:** Reach for `errgroup` whenever you fan out work that can fail. Hand-rolled error+cancel plumbing is a classic source of leaked goroutines (a worker blocked forever sending to a channel nobody reads). `errgroup` + context is the safe default.

---

## Common mistakes (these are the production pitfalls)
1. **Goroutine leak**: a worker blocked on `results <- x` because the consumer stopped reading (e.g., after an error) and nobody closes/drains. Always have a cancellation path (context) and ensure consumers drain or producers select on `ctx.Done()`.
2. **Closing a channel from the consumer or from multiple goroutines** → panic. *The sender* closes, and only once. With multiple senders, close after a `WaitGroup`.
3. **Sending on a closed channel** → panic.
4. **Unbounded fan-out** (`for _, x := range millions { go work(x) }`) → memory blowup + downstream meltdown. Use a pool / `SetLimit`.
5. **Forgetting to close `jobs`** → workers `range` forever → `wg.Wait()` never returns → deadlock.
6. **WaitGroup misuse**: `wg.Add` inside the goroutine (race) instead of before launching it.
7. Ignoring per-item errors — one failed job silently dropped.

## Performance
- Pool size: benchmark it. Too few → underutilized; too many → context-switch overhead + downstream overload.
- Channel ops have a cost; for huge fan-out, batching items reduces channel traffic.
- For pure CPU-bound parallelism, `GOMAXPROCS` (default = cores) caps real parallelism regardless of goroutine count.

---

## Expert Thinking Mode — "process a million items concurrently"

- **Beginner:** "`for _, x := range items { go process(x) }`."  (OOMs / melts the DB.)
- **Senior:** "Bounded worker pool. Producer closes jobs; senders close results after WaitGroup. Context cancellation on error. Run under `-race`."
- **Staff:** "What's the downstream's capacity? Pool size = that, not my CPU count. Backpressure when the queue fills. errgroup with SetLimit + per-item retry/backoff. Observability on queue depth."
- **Architect:** "At this scale is in-process concurrency even right, or do I need a real queue (Kafka/SQS) with horizontal workers (Phase 6/7)? Concurrency model is a scaling and failure-isolation decision."

---

## Real-world use

- **Image/video processing, web crawlers, batch ETL** — all worker pools bounded by downstream capacity.
- **`errgroup`** is everywhere parallel calls happen: fetching from several services to assemble one response, parallel DB lookups.
- **Cloudflare/Uber** bound concurrency to protect downstreams; an unbounded fan-out that DDoSed an internal service is a classic incident postmortem.

---

## Interview Questions

1. Draw a worker pool. Who closes `jobs`? Who closes `results`? Why the ordering?
2. What's the difference between fan-out and fan-in?
3. How can a worker pool leak goroutines, and how do you prevent it?
4. Why does `errgroup` need a context? What does the first error trigger?
5. How do you bound concurrency, and how do you choose the bound (CPU vs I/O work)?
6. Who is allowed to close a channel, and what happens if you get it wrong?
7. Why is `wg.Add()` inside the goroutine a bug?

---

## Your tasks

`../exercises/`: (1) implement a bounded worker pool that squares numbers, (2) implement `fanIn` merging channels, (3) challenge: a concurrent URL-status checker using `errgroup` with `SetLimit` that returns a `map[url]status` and cancels on the first hard error. Run with `-race`. Reference in `../solutions/`.
