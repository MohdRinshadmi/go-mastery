# Day 11 — Goroutines & Channels: Pitfalls

Concurrency gotchas as **Trap → Why → Fix**.

---

### 1. Send on an unbuffered channel with no guaranteed receiver

**Trap:** You launch N goroutines that each `out <- v`, but only receive some of
the values (first-wins, early break). The rest block forever.

**Why:** An unbuffered send is a rendezvous — it does not return until a receiver
takes the value. A goroutine stuck on a send is a leak: it holds its stack, and
anything it references, for the life of the process.

**Fix:** Buffer to the number of one-shot senders (`make(chan T, n)`), or thread
a `context`/`done` channel and `select { case out <- v: case <-done: return }`
so abandoned senders can exit.

---

### 2. `main()` returns before the goroutine runs

**Trap:** `go fmt.Println("hi")` then `main` ends — "hi" sometimes never prints.

**Why:** When `main` returns the runtime kills every other goroutine, mid-flight,
with no cleanup. There is no implicit "wait for children."

**Fix:** Synchronize before exit — receive on a channel, `wg.Wait()` (Day 12), or
`<-ctx.Done()`. Never rely on a `time.Sleep` to "give it time"; that's a race.

---

### 3. Loop variable captured by reference in a goroutine

**Trap (pre-Go 1.22):**
```go
for i := 0; i < 5; i++ {
    go func() { fmt.Println(i) }() // often prints 5,5,5,5,5
}
```

**Why:** Before Go 1.22 the loop variable `i` was a single shared variable; all
closures captured the *same* `i`, read after the loop finished. Go 1.22+ gives
each iteration its own `i`, fixing *this* form — but the same trap returns the
moment you capture *any* shared mutable variable from a goroutine.

**Fix:** Pass the value as an argument: `go func(i int){ ... }(i)`. It's explicit,
version-independent, and survives refactors. Build/test with `-race`.

---

### 4. Forgetting to `close`, so `range ch` hangs

**Trap:** A producer sends a fixed number of values then returns without
`close(ch)`; the consumer `for v := range ch` waits for a close that never comes
→ deadlock (`fatal error: all goroutines are asleep`).

**Why:** `range` over a channel only ends when the channel is *closed and drained*.
Without a close it blocks waiting for the next value forever.

**Fix:** The **sender** closes when done, ideally `defer close(out)` at the top of
the producing goroutine. Never close from the receiver.

---

### 5. Closing a channel with multiple senders

**Trap:** Several sender goroutines each `close(ch)` when "done" → `panic: close
of closed channel`. Or one closes while another is still sending → `panic: send
on closed channel`.

**Why:** A channel may be closed exactly once, and only the side that owns sending
may close it. With many senders there is no single owner.

**Fix:** Use a `sync.WaitGroup` (Day 12) so one coordinator goroutine closes after
all senders finish: `go func(){ wg.Wait(); close(ch) }()`. Or `sync.Once`.

---

### 6. nil channel operations block forever

**Trap:** `var ch chan int` (never `make`d), then `ch <- 1` or `<-ch` hangs the
goroutine permanently.

**Why:** Operations on a nil channel block forever by spec. Often happens when a
struct field channel was never initialized.

**Fix:** Always `make` channels before use. (A nil channel *is* useful on purpose
inside `select` to disable a case — Day 12 — but only there, with a comment.)

---

### 7. Directionless channels in signatures let teams wire producer→producer

**Trap:** Passing `chan T` everywhere; two functions both *send* on a channel
nobody receives from, and it compiles fine.

**Why:** A bidirectional `chan T` carries no contract about who produces and who
consumes, so the compiler can't catch a mis-wiring.

**Fix:** Annotate direction at every boundary: producers take `chan<- T`,
consumers take `<-chan T`. The compiler then enforces the dataflow.
