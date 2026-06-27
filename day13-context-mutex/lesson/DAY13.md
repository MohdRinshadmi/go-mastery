# Day 13 — Context, Mutexes, and the Race Detector

> Mentor note: Days 11–12 gave you goroutines and channels. Today is about *control and safety*: how to cancel work that's no longer needed (`context`), how to protect shared memory when channels aren't the right tool (`sync.Mutex`), and how to catch the bugs you can't see by reading the code (`-race`). Concurrency bugs are the scariest in production — they're nondeterministic, pass in dev, and corrupt data at 3am under load. The race detector is your seatbelt. Use it always.

---

## 1. context.Context — cancellation & deadlines

### The problem
A request comes in. It fans out to a DB call, a cache call, and two HTTP calls. The client disconnects. Without coordination, all that work keeps running, wasting resources. Or one call hangs forever and the whole request never returns. `context` solves both: it's a **cancellation signal + deadline that propagates down a call tree**.

### The shape
```go
func handleRequest(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel() // ALWAYS cancel to release resources, even on the happy path

    result, err := doWork(ctx) // pass ctx down to everything
    ...
}

func doWork(ctx context.Context) (string, error) {
    select {
    case <-time.After(5 * time.Second): // simulated slow work
        return "done", nil
    case <-ctx.Done():                  // cancelled or timed out
        return "", ctx.Err()            // context.Canceled or context.DeadlineExceeded
    }
}
```

### The rules (memorize these)
1. **`ctx` is the first parameter**, always named `ctx`: `func F(ctx context.Context, ...)`.
2. **Never store a Context in a struct.** Pass it explicitly through calls.
3. **Always call `cancel()`** (defer it) — leaking a context leaks a goroutine/timer.
4. **Don't pass `nil`** — use `context.Background()` (top of main/request) or `context.TODO()` (placeholder).
5. **`ctx.Value` is for request-scoped data** (request ID, auth user) — NOT for passing optional function args. Overusing Values is an anti-pattern.

### Variants
- `context.WithCancel(parent)` — manual cancel.
- `context.WithTimeout(parent, d)` — cancel after duration.
- `context.WithDeadline(parent, t)` — cancel at a time.
- `context.WithValue(parent, key, val)` — attach request-scoped data.

**Senior take:** Every blocking call in a real service — DB query, HTTP request, channel receive — should be cancellable via context. A function that does I/O but doesn't take a `ctx` is a code-review reject. This is how you avoid goroutine leaks and runaway requests.

---

## 2. Mutexes — protecting shared memory

Channels are great for *handing off* data. But sometimes you just have shared state (a counter, a cache, a map) that multiple goroutines read/write. That's a `sync.Mutex`.

```go
type Counter struct {
    mu sync.Mutex
    n  int
}
func (c *Counter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock() // defer so it unlocks even if the body panics
    c.n++
}
func (c *Counter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.n
}
```

### RWMutex — many readers, one writer
`sync.RWMutex` lets unlimited concurrent **readers** (`RLock`) but exclusive **writers** (`Lock`). Use when reads vastly outnumber writes (a config cache, a read-mostly map).

### atomic — lock-free for single values
For a simple counter, `sync/atomic` (or Go 1.19+ `atomic.Int64`) is faster than a mutex:
```go
var n atomic.Int64
n.Add(1)
n.Load()
```

### Channels vs Mutexes — which?
- **Channel**: transferring ownership of data, coordinating goroutines, pipelines, signaling.
- **Mutex**: protecting a small piece of shared state with simple read/write.
- Go proverb: *"Don't communicate by sharing memory; share memory by communicating"* — prefer channels for coordination. But don't force a channel where a mutex-guarded counter is simpler and faster. Use the right tool.

**Senior take:** The mutex should be **unexported and live next to the data it guards** (same struct). Lock for the shortest span possible. Never return a value *and* hold expectations that the caller locks — encapsulate locking inside the methods. A leaked lock (forgot to Unlock) = deadlock = hung service.

---

## 3. Data races and the `-race` detector

A **data race**: two goroutines access the same memory concurrently, at least one writes, with no synchronization. The result is undefined — corrupted values, torn reads, crashes. The terrifying part: the code *looks* fine and usually works in testing.

```go
// RACE: many goroutines incrementing a plain int with no lock
counter := 0
for i := 0; i < 1000; i++ {
    go func() { counter++ }() // counter++ is read-modify-write, NOT atomic
}
```

You cannot reliably find these by reading code. Go ships a detector:

```bash
go run -race main.go
go test -race ./...
```

It instruments memory accesses and **reports the exact two stack traces** that raced. Run it in CI on every concurrent package. It has overhead (~5-10x), so it's a testing tool, not a production build flag.

**Senior take:** "It works on my machine" is how race bugs ship. The rule on my teams: any package with goroutines runs under `-race` in CI, no exceptions. A clean `-race` run is the only proof your synchronization is correct.

### Concurrent map access
A bare `map` accessed by multiple goroutines (one writing) **panics at runtime** ("concurrent map writes") — Go actively detects this. Fix with a mutex around the map, or `sync.Map` for specific high-contention read-mostly cases (benchmark before choosing `sync.Map`; a plain map+RWMutex often wins).

---

## Common mistakes
1. Not calling `cancel()` → context/goroutine/timer leak.
2. Storing `context.Context` in a struct field.
3. Copying a `sync.Mutex` by value (passing a struct-with-mutex by value copies the lock → broken). `go vet` catches this. Use pointer receivers.
4. Locking too coarsely (whole function incl. slow I/O) → kills concurrency. Or too finely → races.
5. Forgetting `defer mu.Unlock()` and returning early with the lock held → deadlock.
6. Using `ctx.Value` as a general argument-passing mechanism.
7. Reading a shared variable in a loop condition while another goroutine writes it — race; use atomic or a channel.

## Performance
- Atomics > mutex > channel for a simple counter (but measure; contention changes everything).
- RWMutex helps only when reads dominate and the critical section is non-trivial; for tiny sections its extra bookkeeping can be slower than a plain Mutex.
- Lock contention shows up in pprof as time in `sync.(*Mutex).Lock` — a signal to shard the lock or rethink the design.

---

## Expert Thinking Mode — "make this concurrent code correct"

- **Beginner:** "It compiles and printed the right answer once. Ship it."
- **Senior:** "Run it under `-race`. Every I/O call takes a ctx. Locks are encapsulated, minimal, and deferred. I can name the synchronization point for every shared variable."
- **Staff:** "What's the cancellation story end to end — does client disconnect propagate to the DB? Is the lock a contention bottleneck at 10k rps? Channel vs mutex chosen by profiling, not taste."
- **Architect:** "Cancellation, timeouts, and backpressure are system properties. Context deadlines tie into request budgets and circuit breakers across services. Shared mutable state is minimized by design; immutability and message-passing scale better than locks."

---

## Real-world use

- **Every Go HTTP/gRPC server** threads `r.Context()` (which cancels on client disconnect) through handlers into DB and downstream calls.
- **Database drivers** (`database/sql`, pgx) take `ctx` on every query so a timed-out request stops querying.
- **`-race` in CI** is standard at Google/Uber/Cloudflare; it has caught countless prod-bound races.
- **RWMutex-guarded config** that's read on every request and reloaded occasionally is a ubiquitous pattern.

---

## Interview Questions

1. What problem does `context.Context` solve? Give two concrete uses.
2. What are the rules for using context (first param, cancel, no nil, no struct storage, Value usage)?
3. What is a data race? Why can't you reliably find it by reading code?
4. How does `go test -race` work and why isn't it a production build flag?
5. When do you choose a channel vs a mutex?
6. Why must you not copy a `sync.Mutex` by value? What catches it?
7. What happens when multiple goroutines write a plain map concurrently?

---

## Your tasks

`../exercises/` has: (1) a `SafeCounter` with a deliberate race for you to fix and prove clean under `-race`, (2) a `fetchWithTimeout(ctx, ...)` to implement using `select` + `ctx.Done()`, and (3) a challenge: a concurrency-safe in-memory cache (`map` + `RWMutex`) with `Get`/`Set` that passes `go test -race`. Run everything with `-race` and bring me the output.

---

## Day 13 companion files

Self-study materials for this day (all in the day folder):

- [Debugging challenge](../debugging/README.md) — a data race on a shared map (lost updates + `fatal error: concurrent map writes`), fixed with an `RWMutex` (`bugged/` vs `fixed/`, proved with `-race`).
- [Pitfalls](../PITFALLS.md) — 7 context/mutex/race traps as Trap → Why → Fix.
- [Interview Q&A](../INTERVIEW.md) — 10 questions with model answers.
- [Notes](../NOTES.md) — context, mutex, RWMutex, atomic, race-detector quick reference + key terms.
- [Resources](../RESOURCES.md) — curated links (context blog, memory model, race detector).
