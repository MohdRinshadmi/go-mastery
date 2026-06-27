# Day 12 — select, buffered channels, WaitGroup, Once: Quick Reference

## Buffered channels

```go
ch := make(chan int, 3)   // capacity 3
ch <- 1                   // blocks only when buffer is FULL
v := <-ch                 // blocks only when buffer is EMPTY (FIFO)
len(ch)  // items currently buffered
cap(ch)  // capacity
```

- Decouples producer/consumer speed up to `cap`.
- `buffer of 1` pattern: lets a one-shot producer always send even if the
  consumer stops listening → prevents a goroutine leak.
- Size buffers for a stated reason, not to "stop blocking."

## select

```go
select {
case v := <-ch1:   useRecv(v)
case ch2 <- x:      // try send
case <-time.After(d):  // timeout
default:            // non-blocking: runs if no case ready
}
```

- Multiple cases ready → one chosen **uniformly at random** (spec-guaranteed).
- No `default` → blocks until a case is ready.
- `default` in a `for` loop with no sleep → busy-wait (100% CPU). Avoid.
- A `nil` channel case is never ready — used deliberately to disable a case.
- `select {}` blocks forever.

### Timeout in a loop — avoid the `time.After` leak
```go
t := time.NewTimer(d)
defer t.Stop()
for {
    select {
    case m := <-ch: handle(m); if !t.Stop() { <-t.C }; t.Reset(d)
    case <-t.C:     onTimeout()
    }
}
// For periodic work prefer time.NewTicker.
```

## sync.WaitGroup

```go
var wg sync.WaitGroup
for i := 0; i < n; i++ {
    wg.Add(1)              // BEFORE go
    go func() { defer wg.Done(); work() }()
}
wg.Wait()                  // blocks until counter == 0
```

- **Add before `go`.** Add inside the goroutine races with Wait.
- **`defer Done()`** — survives panics.
- Pass by **pointer** (`*sync.WaitGroup`); never copy (`go vet` copylocks).
- Reuse only after `Wait()` returns.

## sync.Once

```go
var once sync.Once
once.Do(func() { config = load() }) // runs exactly once, others block until done
```

- Exactly-once init; correct singleton/lazy-init primitive.
- Cannot reset. Runs the function once even if it "fails" → don't use for
  fallible/retryable init.

## Deadlock cheat-sheet

| Scenario | Fix |
|---|---|
| send on unbuffered, no receiver | add receiver / buffer |
| `range ch` never closed | sender `close(ch)` when done |
| `wg.Add` inside goroutine | move Add before `go` |
| `select` all nil / no default | real channel or `default`/timeout |
| copy of WaitGroup/Mutex | pass pointer |

---

## Key terms

- **Buffered channel** — capacity > 0; send blocks only when full, receive only when empty.
- **Buffer-of-1 pattern** — prevents leak of a one-shot sender whose consumer may bail.
- **select** — wait on multiple channel ops; random choice among ready cases.
- **default case** — makes a select non-blocking (try-send/try-receive).
- **busy-wait** — spinning loop with `default` and no sleep, burning CPU.
- **WaitGroup** — counter to wait for N goroutines (`Add`/`Done`/`Wait`).
- **sync.Once** — runs a function exactly once across goroutines.
- **done channel** — close-to-broadcast cancellation; predecessor of context.
- **copylocks** — the `go vet` check that flags copying sync types by value.
