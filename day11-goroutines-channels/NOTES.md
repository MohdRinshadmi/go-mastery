# Day 11 — Goroutines & Channels: Quick Reference

## Goroutines

```go
go f(x)                 // launch concurrently
go func() { ... }()     // anonymous
```

- ~2 KB stack, grows/shrinks; ~1 µs to create; millions feasible.
- Scheduled M:N — G goroutines onto M OS threads via P processors (`GOMAXPROCS`).
- `main` returning kills all goroutines. **Always synchronize before exit.**
- Every goroutine needs a stated exit condition (done / cancelled / channel closed).
- `runtime.NumGoroutine()` — count live goroutines (leak check in tests).

## Channels

```go
ch := make(chan int)        // unbuffered (capacity 0)
ch <- v                     // send (blocks until receiver ready)
v := <-ch                   // receive (blocks until sender ready)
v, ok := <-ch               // ok == false when closed & drained
close(ch)                   // sender signals "no more values"
for v := range ch { ... }   // receive until closed & drained
```

### Blocking rules

| Operation | Unbuffered | Closed channel | nil channel |
|---|---|---|---|
| send `ch <- v` | blocks until receiver | **panic** | blocks forever |
| receive `<-ch` | blocks until sender | zero value, `ok=false` | blocks forever |
| `close(ch)` | — | **panic** (double close) | **panic** |

### Close discipline
- Only the **sender** closes; only **once**.
- Multiple senders → one coordinator closes after `wg.Wait()` (Day 12).
- `defer close(out)` at the top of a producer goroutine is the idiom.

### Directional channels
```go
func produce(out chan<- int)   // send-only  (compiler-enforced)
func consume(in <-chan int)     // receive-only
```
Annotate direction in every function signature — it's a compile-time contract.

## Patterns

```go
// signal-only channel (no data)
done := make(chan struct{})
go func() { defer close(done); work() }()
<-done

// generator returns receive-only channel
func gen(nums ...int) <-chan int {
    out := make(chan int)
    go func() { defer close(out); for _, n := range nums { out <- n } }()
    return out
}
```

## Loop-variable capture
- Go 1.22+: each iteration gets its own loop variable (the classic `5,5,5,5,5`
  bug is fixed for the loop var itself).
- Still pass shared values as args: `go func(i int){...}(i)`. Always test `-race`.

---

## Key terms

- **Goroutine** — lightweight runtime-scheduled execution unit (~2 KB stack).
- **M:N scheduler** — multiplexes many goroutines (G) onto few OS threads (M) via processors (P).
- **GOMAXPROCS** — number of P; max goroutines running in parallel (default = CPU cores).
- **Channel** — typed, internally-synchronized conduit between goroutines.
- **Unbuffered channel** — capacity 0; send/receive rendezvous (handshake).
- **Rendezvous / handshake** — the synchronization an unbuffered send+receive performs.
- **Directional channel** — `chan<-` (send-only) / `<-chan` (receive-only).
- **Goroutine leak** — a goroutine that never exits (usually blocked on a channel).
- **CSP** — Communicating Sequential Processes; Hoare's model behind Go's channels.
- **happens-before** — memory-model ordering; a channel send happens-before its receive completes.
