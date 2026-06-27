# Phase 3 — Concurrency (Days 11–15)

Goroutines & the scheduler, channels, `select`, `sync` primitives, `context`, the memory model, worker pools, fan-out/in, pipelines. Self-quiz: answer aloud, then expand.

---

### 1. Goroutine vs OS thread? How does the scheduler map between them (GMP)?

<details><summary>Answer</summary>

A goroutine is a **user-space, runtime-scheduled** coroutine with a tiny (~2–8KB growable) stack; an OS thread is heavier (~1–2MB stack, kernel-scheduled). The runtime multiplexes many goroutines onto few threads via the **GMP model**: **G** = goroutine, **M** = OS thread (machine), **P** = processor (a scheduling context, count = `GOMAXPROCS`). An M needs a P to run Gs; each P has a local run queue (plus a global queue), and idle Ps **steal** work from busy ones. On a blocking syscall the M detaches from its P so another M keeps the P busy. That's why a million goroutines is cheap.
</details>

---

### 2. "Don't communicate by sharing memory; share memory by communicating" — in practice?

<details><summary>Answer</summary>

Instead of multiple goroutines locking and mutating a shared variable, pass ownership of the data through a **channel** so exactly one goroutine touches it at a time. The channel both transfers the value *and* synchronizes, eliminating the race by construction:

```go
results := make(chan int)
go func() { results <- compute() }() // producer owns the value until it sends
total := <-results                   // consumer now owns it
```
It's a guideline, not a law — a simple counter is cleaner with a `sync.Mutex` or `atomic`. Use channels for *ownership transfer and orchestration*, mutexes for *protecting a small piece of shared state*.
</details>

---

### 3. What is an unbuffered channel? What happens on a send with no receiver ready?

<details><summary>Answer</summary>

An unbuffered channel (`make(chan T)`) has **zero capacity**, so a send and a receive must **rendezvous**: the send **blocks** until another goroutine is ready to receive (and vice versa). This makes it a synchronization point — when the send completes, you *know* the receiver got it. If no receiver ever shows up and it's the only goroutine, you get a deadlock (`fatal error: all goroutines are asleep`). A **buffered** channel only blocks the sender when the buffer is full and the receiver when it's empty.
</details>

---

### 4. What is a goroutine leak? How do you detect and prevent one?

<details><summary>Answer</summary>

A goroutine leak is a goroutine that **blocks forever** (on a channel send/receive that never completes) and is never collected — it holds memory and references indefinitely, an unbounded slow leak. Common cause: a producer sending into a channel whose consumer left early. **Detect** with `runtime.NumGoroutine()` in tests, goroutine pprof profiles, or tools like `goleak`. **Prevent** by giving every goroutine a guaranteed exit: a `ctx.Done()` case in its `select`, closing channels to signal completion, or buffering so a send can't block. Every goroutine you start must have an answer to "how does this end?"
</details>

---

### 5. Receiving from a closed channel vs sending to one?

<details><summary>Answer</summary>

**Receiving** from a closed channel never blocks: it drains any buffered values, then returns the element's **zero value** immediately, with the comma-ok form telling you (`v, ok := <-ch` → `ok == false`). **Sending** to a closed channel **panics**, and so does closing an already-closed or nil channel. Hence the rule: the **sender** (or a single coordinator) closes, never the receiver, and only once — typically via `defer close(ch)` in the lone producer.
</details>

---

### 6. Why do directional channels (`chan<-`, `<-chan`) exist? Real use case?

<details><summary>Answer</summary>

They let a function signature **declare and enforce direction**, so the compiler stops a stage from misusing a channel. A pipeline stage takes a receive-only input and a send-only output:

```go
func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() { defer close(out); for n := range in { out <- n * n } }()
    return out
}
```
Returning `<-chan int` guarantees callers can only read it, making "who closes this?" unambiguous and preventing a consumer from accidentally sending or closing.
</details>

---

### 7. Buffered vs unbuffered — when does each block? Why a buffer of exactly 1?

<details><summary>Answer</summary>

Unbuffered: sender blocks until a receiver is ready; receiver blocks until a sender is ready (rendezvous). Buffered(n): sender blocks only when the buffer is **full**; receiver blocks only when **empty**. A **buffer of 1** is the classic decoupling tool: it lets a producer deposit one value and move on without waiting for the consumer — used for **signal/notify** patterns (a non-blocking "work is available" ping) and to hold a single most-recent value so a fast producer isn't gated on a slow consumer for one item.
</details>

---

### 8. What does `select` do when multiple cases are ready? How do you avoid blocking forever?

<details><summary>Answer</summary>

When several cases are ready, `select` picks **one uniformly at random** — this prevents starvation, so you can't rely on ordering. A `default` case makes the `select` **non-blocking** (runs immediately if nothing else is ready). To bound waiting, add a timeout or cancellation case:

```go
select {
case v := <-ch:   handle(v)
case <-ctx.Done(): return ctx.Err()  // never block forever
}
```
A `select{}` with no cases blocks the goroutine permanently — sometimes used to park `main`.
</details>

---

### 9. Why is `time.After` in a `for/select` loop a leak, and how do you fix it?

<details><summary>Answer</summary>

`time.After(d)` allocates a **new timer every loop iteration**, and each timer lives (holding its channel and runtime entry) until it fires — under a hot loop you accumulate many uncollected timers. Fix by creating **one** `time.Timer`/`time.Ticker` outside the loop and `Reset`-ing it, and `Stop`-ing it on exit:

```go
t := time.NewTimer(d); defer t.Stop()
for {
    t.Reset(d)
    select {
    case v := <-ch: handle(v)
    case <-t.C:     onTimeout()
    }
}
```
</details>

---

### 10. Rules for `sync.WaitGroup.Add`? Why is `Add` inside the goroutine a bug?

<details><summary>Answer</summary>

Call `Add` **before launching** the goroutine (on the parent/spawning goroutine), call `Done` exactly once when each finishes (usually `defer wg.Done()`), and `Wait` after all `Add`s. If you call `wg.Add(1)` *inside* the new goroutine, there's a **race**: `Wait` may run before that goroutine has even scheduled and incremented the counter, so `Wait` sees zero and returns early — missing in-flight work. The counter must be raised before there's any chance `Wait` runs.
</details>

---

### 11. When WaitGroup vs channels to wait for goroutines?

<details><summary>Answer</summary>

Use a **`WaitGroup`** when you only need to know "all N goroutines finished" and don't care about their results — it's the lightweight join. Use **channels** when you need to **collect results** as they arrive, stream values, or coordinate (fan-in, pipelines, first-result-wins). Often you combine them: workers send results on a channel, and a separate goroutine `wg.Wait()`s then `close()`s the results channel so the consumer's `range` terminates cleanly.
</details>

---

### 12. What does `sync.Once` guarantee? Why is naive check-then-init wrong?

<details><summary>Answer</summary>

`once.Do(f)` guarantees `f` runs **exactly once** across all goroutines, and that every caller **blocks until that first run completes** — so no one sees a half-initialized result. The naive `if x == nil { x = init() }` is a data race: two goroutines can both read nil and both initialize (double work, possibly inconsistent state), and there's no happens-before edge guaranteeing the write is visible. `sync.Once` provides both the mutual exclusion and the memory ordering for free.
</details>

---

### 13. What is a data race, and why can't you reliably find it by reading code?

<details><summary>Answer</summary>

A data race is two goroutines accessing the same memory concurrently, at least one writing, with no synchronization establishing an order. You can't reliably spot it by reading because it's a **timing** bug — the code may run correctly thousands of times and corrupt state only under a specific interleaving on a specific scheduler/CPU. Worse, its behavior is **undefined**, so the compiler may reorder around it. You find races by *running* under `go test -race`, which instruments accesses and reports the conflicting goroutines and stacks.
</details>

---

### 14. When channel vs mutex? Why can't you copy a `sync.Mutex`?

<details><summary>Answer</summary>

Reach for a **channel** to transfer ownership of data or orchestrate goroutine lifecycles; reach for a **mutex** to guard a small piece of shared state with simple critical sections (a counter, a map). A mutex is usually faster and clearer for "many readers/writers of one struct." You must not **copy** a `Mutex` by value because the copy carries the original's internal lock state — copying a *locked* mutex yields two corrupt half-locks. So embed mutexes in structs you pass by pointer; `go vet`'s `copylocks` check flags accidental copies.
</details>

---

### 15. Mutex vs RWMutex vs atomic — how to choose? What happens on concurrent plain-map writes?

<details><summary>Answer</summary>

`sync.Mutex` — full mutual exclusion, the default. `sync.RWMutex` — many concurrent **readers** *or* one writer; worth it only on **read-heavy** data where read concurrency outweighs its extra bookkeeping (under write-heavy load a plain Mutex is often faster). `sync/atomic` — lock-free single-word operations (counters, flags, pointer swaps) — fastest but limited to primitives. Concurrent writes to a plain `map` are an unsynchronized race that the runtime **deliberately detects and crashes** with `fatal error: concurrent map writes` (not recoverable) — use a `Mutex`-guarded map or `sync.Map`.
</details>

---

### 16. What problem does `context.Context` solve? Give two uses and the rules.

<details><summary>Answer</summary>

`context` carries **cancellation, deadlines/timeouts, and request-scoped values** down a call tree, so a cancelled or timed-out request can signal every goroutine doing work for it to stop. Two uses: (1) **timeout** a DB/HTTP call (`context.WithTimeout`), and (2) **propagate cancellation** when a client disconnects so you stop wasted work. Rules: pass `ctx` as the **first parameter** (`ctx context.Context`), never store it in a struct, never pass `nil` (use `context.Background()`/`TODO()`), always `defer cancel()` to release resources, and use `context.Value` only for request-scoped metadata (request IDs), never for optional function params.
</details>

---

### 17. How do you implement `fetchWithTimeout(ctx, ...)`?

<details><summary>Answer</summary>

Run the work in a goroutine that reports on a channel, then `select` between that channel and `ctx.Done()`:

```go
func fetchWithTimeout(ctx context.Context, url string) (Result, error) {
    ch := make(chan Result, 1) // buffered so the goroutine can't leak on timeout
    go func() { ch <- fetch(url) }()
    select {
    case r := <-ch:    return r, nil
    case <-ctx.Done(): return Result{}, ctx.Err() // DeadlineExceeded or Canceled
    }
}
```
The buffer of 1 is essential: on timeout we return without receiving, and the buffer lets the goroutine's send complete instead of blocking forever.
</details>

---

### 18. Explain Go's memory model / happens-before. Why does it matter?

<details><summary>Answer</summary>

The memory model defines when a write by one goroutine is **guaranteed visible** to a read by another — via "happens-before" edges established by synchronization. A channel **send happens-before** the corresponding receive completes; a `Mutex` `Unlock` happens-before a subsequent `Lock`; `once.Do` completion happens-before any later `Do` returns. Without such an edge, there's **no guarantee** a write is ever seen (CPU caches, compiler reordering) — that's why an unsynchronized "I set the flag" can spin forever. You don't add memory barriers manually; you use channels/`sync`/`atomic`, which provide the ordering.
</details>

---

### 19. Draw a worker pool. Who closes `jobs`, who closes `results`, and why the ordering?

<details><summary>Answer</summary>

A fixed set of N worker goroutines `range` over a shared `jobs` channel and send to a `results` channel. The **producer** closes `jobs` when it's done sending — that's what lets each worker's `range` terminate. A coordinator does `wg.Wait()` for all workers, *then* closes `results`, so the consumer's `range results` ends only after every worker has finished:

```go
go func() { wg.Wait(); close(results) }()
```
Ordering matters: closing `results` before workers finish would panic on their next send; never closing `jobs` leaks the workers forever.
</details>

---

### 20. Fan-out vs fan-in? How do you bound concurrency and choose the bound?

<details><summary>Answer</summary>

**Fan-out** = multiple goroutines reading the *same* input channel to parallelize work; **fan-in** = merging multiple input channels into one output (each source feeds a goroutine that forwards into the shared out, with a `WaitGroup` closing out when all are done). You **bound** concurrency with a fixed worker count or a semaphore (buffered channel / `errgroup.SetLimit`) to avoid spawning unbounded goroutines. Choose the bound by workload: for **CPU-bound** work, ~`GOMAXPROCS` (more just thrashes); for **I/O-bound** work, much higher (dozens–hundreds) since goroutines spend most time blocked.
</details>

---

### 21. Why does `errgroup` need a context, and what does the first error trigger?

<details><summary>Answer</summary>

`errgroup.WithContext` returns a group plus a derived `ctx`; when **any** goroutine in the group returns a non-nil error, the group **cancels that ctx**, signaling all the *other* goroutines (which are `select`ing on `ctx.Done()`) to abandon their work early. `g.Wait()` then returns the **first** error. So it gives you fail-fast fan-out: one failure stops the squad instead of letting siblings burn resources. `SetLimit(n)` additionally caps how many run at once.
</details>

---

### 22. What's a pipeline stage's signature, and how does closing cascade? How does an early-exiting consumer leak, and how does context fix it?

<details><summary>Answer</summary>

Each stage is `func(in <-chan T) <-chan U`: it owns a goroutine, `range`s its input, sends to a new output, and `defer close(out)`s. Closing **cascades**: when stage 1 closes its output, stage 2's `range` ends, so stage 2 closes its output, and so on to the consumer. The leak: if the consumer takes only the first few values and stops reading, upstream stages **block forever** on a send no one receives. **Context fixes it** — every stage adds a `case <-ctx.Done(): return` to its send `select`, so when the consumer cancels, all stages unblock and exit:

```go
select {
case out <- v:
case <-ctx.Done(): return
}
```
You test "no leaks" by snapshotting `runtime.NumGoroutine()` before/after (allowing the scheduler a moment to settle).
</details>
