# Day 12 — select, buffered channels, WaitGroup, Once: Pitfalls

Concurrency gotchas as **Trap → Why → Fix**.

---

### 1. `wg.Add(1)` inside the goroutine

**Trap:**
```go
go func() { wg.Add(1); defer wg.Done(); work() }()
wg.Wait()
```

**Why:** The goroutine may not be scheduled before `Wait()` runs. At that moment
the counter is 0, so `Wait` returns early — you proceed while work is unfinished.
It's also a data race between `Add` and `Wait` (`go vet` + `-race` both flag it).

**Fix:** `wg.Add(1)` in the loop **before** `go`. The count must be established
before any goroutine — or `Wait` — can run.

---

### 2. Passing `sync.WaitGroup` (or any sync type) by value

**Trap:** `func work(wg sync.WaitGroup)` — `Done()` decrements a *copy*.

**Why:** `WaitGroup`, `Mutex`, `Once` contain internal state that must not be
copied. A copy's `Done` is invisible to the original's `Wait` → `Wait` blocks
forever (or returns wrong).

**Fix:** Pass a pointer: `func work(wg *sync.WaitGroup)`. `go vet`'s `copylocks`
check catches accidental copies.

---

### 3. `default:` in a `for/select` with no sleep — busy-wait

**Trap:**
```go
for {
    select {
    case v := <-ch: process(v)
    default: // nothing here
    }
}
```

**Why:** When `ch` isn't ready, `default` fires *immediately* and the loop spins
at 100% CPU, starving other goroutines.

**Fix:** Usually you don't want `default` at all — a bare `select` blocks until a
case is ready (that's what you want). Use `default` only for genuine try-send/
try-receive, and if you must poll, add a `time.Sleep` or a `time.Ticker` case.

---

### 4. `time.After` inside a `for/select` loop

**Trap:**
```go
for {
    select {
    case m := <-ch:        handle(m)
    case <-time.After(5*time.Second): checkHealth()
    }
}
```

**Why:** Each iteration creates a **new** timer; the old one isn't garbage-
collected until it fires. A hot loop spawns thousands of pending timers
(goroutine + memory) — a slow leak.

**Fix:** Create one `time.NewTimer` (or `time.Ticker`) outside the loop, `defer
timer.Stop()`, and `Reset` it per iteration. Tickers are ideal for periodic work.

---

### 5. Buffer of N "to stop the blocking"

**Trap:** A producer blocks, so you bump `make(chan T, 1000)` until it doesn't.

**Why:** A large buffer hides a producer/consumer speed mismatch — it doesn't fix
it. Under real load the buffer fills and you're back to blocking, now with extra
latency and memory. "It stopped blocking" is not a capacity justification.

**Fix:** Size buffers for a *reason* you can state (measured burst vs drain rate,
or the known fixed number of one-shot senders). Otherwise fix the design
(backpressure, more consumers, a worker pool — Day 14).

---

### 6. `sync.Once` used where initialization can fail

**Trap:**
```go
once.Do(func() { db, err = openDB() }) // if openDB fails, err is permanent
```

**Why:** `Once` runs the function exactly once, success or failure. A transient
error (DB not up yet) becomes a *permanent* error for the process lifetime — no
retry ever happens.

**Fix:** Use `Once` only for initialization that can't meaningfully fail. For
fallible init that should retry, use a mutex + recompute, or
`golang.org/x/sync/singleflight`.

---

### 7. `select` with all cases on nil (or no) channels

**Trap:** `select {}` or a `select` whose every case reads a never-`make`d (nil)
channel and has no `default`.

**Why:** A nil-channel case is never ready; with no ready case and no `default`,
the goroutine blocks forever → deadlock (`all goroutines are asleep`).

**Fix:** Ensure at least one case can become ready, or add a `default`/timeout.
(Setting a channel to nil to *disable* a case is a deliberate, commented pattern —
not this accident.)
