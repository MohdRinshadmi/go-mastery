# Day 12 — Buffered Channels, select, WaitGroups & sync.Once

> Mentor note: Yesterday you learned the fundamentals. Today you learn the tools that make concurrency *composable*. `select` is Go's concurrency superpower — it's what lets you write expressive, non-blocking, timeout-aware concurrent code. `WaitGroup` is the "wait for N goroutines" hammer you'll reach for daily. `sync.Once` is deceptively simple and solves a hard class of bugs. Pay close attention to the deadlock section.

---

## 1. Buffered Channels

### Theory
A **buffered channel** has capacity > 0: `make(chan T, N)`. A send blocks only when the buffer is **full**. A receive blocks only when the buffer is **empty**.

```go
ch := make(chan int, 3) // capacity 3

ch <- 1 // doesn't block — buffer has room
ch <- 2
ch <- 3
// ch <- 4 // WOULD block — buffer full

v := <-ch // doesn't block — buffer has data
fmt.Println(v) // 1 (FIFO)
```

### The mental model

```
Unbuffered:  sender --------handshake--------- receiver  (must meet)
Buffered:    sender -> [slot1|slot2|slot3] -> receiver   (decoupled)
```

Buffered channels **decouple** producer speed from consumer speed — up to the buffer capacity. Beyond capacity, the sender blocks.

### When to use buffered channels

1. **Known fixed work:** `make(chan error, n)` to collect errors from n goroutines without blocking each one.
2. **Rate limiting / bounding:** a buffered channel as a semaphore (Day 14).
3. **Small performance buffer:** smooth out brief bursts in producer/consumer speed mismatch.
4. **Avoiding goroutine leak:** if a producer sends one result and might not be consumed (e.g. timeout path), a buffer of 1 prevents the goroutine from hanging.

### When NOT to use buffered channels

- **As a workaround for design problems.** If you need a buffer of 1000 to avoid blocking, that's a sign your producer and consumer are badly matched — fix the design.
- **"Just add a buffer" is the concurrency version of "just add more RAM."** It hides the problem.
- **Don't buffer if you want synchronization.** A buffer of 1 might work for now and fail when the system speeds up.

### The "buffer of 1 to prevent goroutine leak" pattern

```go
// Without buffer: if the caller times out and stops receiving,
// the goroutine hangs forever trying to send.
func asyncWork() <-chan result {
    ch := make(chan result, 1) // buffer of 1: goroutine can always send
    go func() {
        ch <- doWork() // never blocks, even if caller doesn't receive
    }()
    return ch
}
```

This is a common, practical pattern. Remember it.

**Senior take:** The moment you reach for a large buffer, write a comment explaining *why* that capacity is correct. "I benched it" and "the max burst is 50 items/sec and the consumer drains at 100/sec" are good reasons. "It stopped blocking" is not.

---

## 2. select

### Theory
`select` lets a goroutine wait on **multiple channel operations** simultaneously. It picks the first one that's ready. If multiple are ready, it picks one at random (uniform distribution — this is spec-guaranteed).

```go
select {
case v := <-ch1:
    fmt.Println("received from ch1:", v)
case ch2 <- 42:
    fmt.Println("sent to ch2")
}
```

### Why select exists
Without `select`, you'd have to poll or use separate goroutines per channel. `select` is how you write event-driven code in Go: "do whichever of these is ready first."

### The `default` case — non-blocking operations

```go
select {
case v := <-ch:
    fmt.Println("got:", v)
default:
    fmt.Println("nothing ready") // runs immediately if no case is ready
}
```

`default` makes `select` non-blocking. Use it for:
- **Try-send / try-receive:** don't block, just skip if the channel isn't ready.
- **Polling:** check if work is available, do something else if not.

**Common mistake:** Using `default` in a hot loop without a sleep is a busy-wait:
```go
for {
    select {
    case v := <-ch:
        process(v)
    default:
        // NO sleep here → 100% CPU usage. Add time.Sleep or restructure.
    }
}
```

### Timeouts with `time.After`

```go
select {
case v := <-ch:
    fmt.Println("got:", v)
case <-time.After(1 * time.Second):
    fmt.Println("timed out after 1s")
}
```

`time.After(d)` returns a `<-chan time.Time` that fires after duration `d`. Used in a select case, it gives you a timeout. **Simple and idiomatic.**

**Production note:** `time.After` creates a new timer every call. If you call it in a tight loop (e.g., in a `for/select`), each iteration leaks a timer until it fires. Use `time.NewTimer` with explicit `Stop()` in loops:

```go
timer := time.NewTimer(1 * time.Second)
defer timer.Stop()
select {
case v := <-ch:
    fmt.Println("got:", v)
case <-timer.C:
    fmt.Println("timed out")
}
```

### select with done channel pattern

The most important `select` pattern in production Go:

```go
func worker(jobs <-chan Job, done <-chan struct{}) {
    for {
        select {
        case j, ok := <-jobs:
            if !ok {
                return // jobs channel closed
            }
            process(j)
        case <-done:
            return // we were told to stop
        }
    }
}
```

This goroutine processes jobs but can be cancelled at any time via `done`. You'll write this pattern constantly until Day 13's `context.Context` replaces `done` channels.

### Common deadlocks with select

1. **select with no cases → blocks forever:**
   ```go
   select {} // valid Go, blocks the goroutine forever (useful in some rare cases)
   ```

2. **All cases on nil channels → blocks forever:**
   ```go
   var ch1, ch2 chan int
   select {
   case v := <-ch1: // nil channel: never ready
   case v := <-ch2: // nil channel: never ready
   // no default
   }
   // DEADLOCK
   ```

3. **Forgetting default in a non-concurrent context:**
   ```go
   ch := make(chan int)
   select {
   case v := <-ch: _ = v
   // no sender, no default → DEADLOCK
   }
   ```

**Senior take:** `select` with a nil channel case is *intentional* sometimes — a nil case is permanently disabled. This is how you dynamically enable/disable cases in a `for/select` loop. Powerful, but explain it with a comment.

---

## 3. sync.WaitGroup

### Theory
`sync.WaitGroup` lets you wait for a collection of goroutines to finish. It has three methods:
- `Add(n)` — increment the counter by n (before launching goroutines).
- `Done()` — decrement the counter by 1 (call in each goroutine, typically via `defer`).
- `Wait()` — block until the counter reaches 0.

```go
var wg sync.WaitGroup

for i := 0; i < 5; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done() // always defer — so it runs even if the goroutine panics
        doWork(id)
    }(i)
}

wg.Wait() // blocks until all 5 goroutines call Done
fmt.Println("all done")
```

### Why it exists
Channels work for waiting too (`done := make(chan struct{})`), but WaitGroup is simpler when you just need "wait for N goroutines" without passing data.

### When to use WaitGroup vs channels
- **WaitGroup:** when you just need "wait for all goroutines" and don't need to collect results.
- **Channels:** when goroutines produce results or errors you need to collect.
- **Both:** WaitGroup for completion, channel for errors (or use `errgroup` from Day 14).

### Common mistakes

1. **Add inside the goroutine — RACE:**
   ```go
   // BAD: if the goroutine runs before Add, Wait() sees counter = 0 and returns too early
   go func() {
       wg.Add(1) // WRONG — must call Add BEFORE go
       defer wg.Done()
       doWork()
   }()
   wg.Wait()
   ```
   Rule: **Always `wg.Add(n)` before `go`.**

2. **Passing WaitGroup by value:**
   ```go
   func doWork(wg sync.WaitGroup) { // BUG: copy of wg, Done() doesn't affect caller
       defer wg.Done()
   }
   // Fix: pass pointer
   func doWork(wg *sync.WaitGroup) {
       defer wg.Done()
   }
   ```

3. **Reusing WaitGroup before Wait() returns:**
   ```go
   wg.Add(1)
   go func() { defer wg.Done(); doWork() }()
   wg.Add(1) // fine if before Wait
   wg.Wait()
   wg.Add(1) // safe only AFTER Wait returns
   ```

**Senior take:** `defer wg.Done()` is non-negotiable. If `doWork()` panics, Done() still runs and Wait() doesn't block forever. Without defer, a panicking goroutine hangs your whole program at Wait().

---

## 4. sync.Once

### Theory
`sync.Once` guarantees that a function runs exactly once, regardless of how many goroutines call it concurrently.

```go
var once sync.Once
var config *Config

func getConfig() *Config {
    once.Do(func() {
        config = loadExpensiveConfig() // runs only the first time
    })
    return config
}
```

### Why it exists
The naive "check-then-initialize" pattern is a classic race condition:
```go
if config == nil {      // goroutine A checks: nil
    // goroutine B also checks: nil — both see nil!
    config = loadConfig() // both initialize — double init, race condition
}
```
`sync.Once` solves this with the correct internal synchronization.

### When to use
- Lazy initialization of expensive resources: DB connections, config, caches.
- One-time setup in tests.
- Singleton patterns (though Go generally avoids singletons).

### When NOT to use
- When you need reset/re-run semantics — Once is forever. You can't reset it.
- When the initialization can fail and you need to retry. Once with an error doesn't retry:
  ```go
  var once sync.Once
  var db *DB
  var dbErr error
  func getDB() (*DB, error) {
      once.Do(func() {
          db, dbErr = openDB() // if this fails, it fails forever
      })
      return db, dbErr
  }
  ```
  If your initialization can fail, consider `sync.Once` + a `errors` package, or `golang.org/x/sync/singleflight`.

**Senior take:** `sync.Once` is the correct singleton implementation in Go. It beats `init()` functions (which always run at package load), `var` with init (can't handle errors), and check-then-set (race condition). Use it.

---

## Common Deadlocks — Reference Guide

| Scenario | Why it deadlocks | Fix |
|---|---|---|
| `ch <- v` with no receiver and no buffer | Send blocks; no goroutine will ever receive | Add receiver goroutine or buffer |
| `for range ch` with channel never closed | Range waits for close; sender doesn't close | Always `close(ch)` when sender is done |
| `wg.Add(1)` inside the goroutine | Race: Wait() may return before Add | Move Add before `go` |
| `select {}` with all nil channels | All cases permanently not ready | Add `default` or use real channels |
| Goroutine A waits for B; B waits for A | Classic deadlock | Restructure ownership or use context |
| All goroutines blocked on channel ops | Typical "forgot to close" or "forgot a receiver" | Trace the channel graph |

---

## Expert Thinking Mode

- **Beginner:** "Buffered channels are like queues. Select is a switch for channels."
- **Senior:** "`select` is a multiplexer — I use it to combine a work channel, a done signal, and a timeout into a single blocking wait. The `default` case is a trap: never use it in a loop without a sleep or you burn CPU. `time.After` in a loop leaks timers — use `time.NewTimer`."
- **Staff:** "WaitGroup is for fire-and-collect-N-goroutines. But as soon as I need error propagation, I reach for `errgroup` (Day 14). `sync.Once` is my lazy-init standard — I even use it in tests to set up expensive state once across all test cases."
- **Architect:** "The `done`-channel pattern (`case <-done: return`) was the pre-context approach to goroutine cancellation. We've largely replaced it with `context.Context` (Day 13), but understanding done-channels is important because many third-party libraries still use them, and it's the mental model behind context."

---

## Real-world use

- **HTTP multiplexing:** `select` over multiple upstream responses to return the fastest one (or timeout if none respond quickly). Used in API gateways.
- **Rate limiters:** A buffered channel as a semaphore pool (capacity = max concurrent requests). Workers take a slot, do work, return the slot.
- **Graceful shutdown:** A `done` channel that gets closed when SIGTERM arrives. Every goroutine has a `case <-done` that exits cleanly.
- **Plugin/extension loading:** `sync.Once` initializes the plugin registry exactly once on first access, regardless of which goroutine triggers it.

---

## Common Race Conditions & Production Pitfalls

### Race: WaitGroup.Add in goroutine
See "Common mistakes" above. The race window is tiny — flaky in testing, silent in production for years, then suddenly fails under load.

### Pitfall: time.After leak in for/select loops
```go
// LEAK: each iteration creates a new timer. Old timers aren't GC'd until they fire.
for {
    select {
    case msg := <-ch:
        handle(msg)
    case <-time.After(5 * time.Second): // NEW timer every iteration
        checkHealth()
    }
}
```
In a high-throughput loop this creates thousands of goroutines. Fix with `time.NewTimer` + reset.

### Pitfall: sync.Once with error is sticky
If `once.Do` runs an initialization that fails, it will NOT retry. The failure is permanent for the lifetime of the `once`. Design around this: if the error is transient, don't use Once.

---

## Interview Questions

1. What is the difference between a buffered and unbuffered channel? When does each one block?
2. Why would you use a channel with a buffer of exactly 1? Give a concrete use case.
3. What does `select` do when multiple cases are ready simultaneously?
4. Why is `time.After` in a `for/select` loop a resource leak? How do you fix it?
5. What are the rules for `sync.WaitGroup.Add`? What happens if you call it inside the goroutine?
6. When would you use a WaitGroup vs channels to wait for goroutines?
7. What does `sync.Once` guarantee? Why is the naive "check-then-initialize" pattern wrong?

---

## Your tasks for today

Go to `../exercises/`. There are **3 beginner exercises** + **1 intermediate challenge**. Pay special attention to deadlock scenarios — the exercises are designed to trigger your first real deadlocks if you get the pattern wrong.

---

## Day 12 companion files

Self-study materials for this day (all in the day folder):

- [Debugging challenge](../debugging/README.md) — `wg.Add` inside the goroutine: a WaitGroup that returns before the work finishes (`bugged/` vs `fixed/`, caught by `go vet` and `-race`).
- [Pitfalls](../PITFALLS.md) — 7 select/WaitGroup/Once/buffer traps as Trap → Why → Fix.
- [Interview Q&A](../INTERVIEW.md) — 10 questions with model answers.
- [Notes](../NOTES.md) — buffered channels, select, WaitGroup, Once quick reference + key terms.
- [Resources](../RESOURCES.md) — curated links (select patterns, sync docs, timer leaks).
