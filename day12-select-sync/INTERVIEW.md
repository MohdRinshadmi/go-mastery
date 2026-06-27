# Day 12 — select, buffered channels, WaitGroup, Once: Interview Q&A

Model answers in `<details>`.

---

**1. What is the difference between a buffered and unbuffered channel? When does each block?**

<details>
<summary>Answer</summary>

An unbuffered channel (`make(chan T)`) has capacity 0: a send blocks until a
receiver is ready and a receive blocks until a sender sends — a rendezvous. A
buffered channel (`make(chan T, n)`) holds up to `n` values: a send blocks only
when the buffer is **full**, a receive blocks only when it's **empty**. Buffered
channels decouple producer and consumer speed up to the buffer size; unbuffered
channels also act as a synchronization point.
</details>

---

**2. Why would you use a channel with a buffer of exactly 1? Give a concrete use case.**

<details>
<summary>Answer</summary>

To let a one-shot producer always complete its send even if the consumer might
stop listening — preventing a goroutine leak. Classic async-with-timeout:

```go
func asyncWork() <-chan result {
    ch := make(chan result, 1) // buffer of 1
    go func() { ch <- doWork() }() // never blocks, even if caller times out
    return ch
}
```

If the caller selects on this channel vs a timeout and the timeout wins, the
worker still lands its value in the buffer and exits instead of blocking forever.
</details>

---

**3. What does `select` do when multiple cases are ready simultaneously?**

<details>
<summary>Answer</summary>

It picks one **uniformly at random** — this is guaranteed by the spec. The
randomness prevents starvation (one always-ready channel can't permanently
monopolize the select) and stops you from accidentally relying on case order.
</details>

---

**4. Why is `time.After` in a `for/select` loop a resource leak? How do you fix it?**

<details>
<summary>Answer</summary>

`time.After(d)` allocates a new timer (and underlying goroutine) on every call;
the timer isn't garbage-collected until it fires after `d`. In a hot loop you
accumulate thousands of pending timers. Fix: create one `time.NewTimer` (or a
`time.Ticker` for periodic work) outside the loop, `defer timer.Stop()`, and
`Reset` it each iteration after draining its channel.
</details>

---

**5. What are the rules for `sync.WaitGroup.Add`? What happens if you call it inside the goroutine?**

<details>
<summary>Answer</summary>

`Add(n)` must be called **before** the goroutines it counts are launched, so the
counter is correct before `Wait` can run. `Done()` (usually deferred) decrements;
`Wait()` blocks until the counter hits 0. Calling `Add` inside the goroutine
races with `Wait`: if `Wait` runs before the goroutine is scheduled, the counter
is 0 and `Wait` returns immediately, letting you proceed while work is unfinished.
`go vet` and `-race` both catch this.
</details>

---

**6. When would you use a WaitGroup vs channels to wait for goroutines?**

<details>
<summary>Answer</summary>

Use a `WaitGroup` when you only need "wait for N goroutines to finish" and don't
need to collect data from them. Use channels when the goroutines produce results
or errors you need to receive. Often you combine them: a `WaitGroup` for
completion plus a channel for results — or skip the hand-rolling and use
`errgroup` (Day 14) when the work can fail.
</details>

---

**7. What does `sync.Once` guarantee, and why is naive "check-then-initialize" wrong?**

<details>
<summary>Answer</summary>

`once.Do(f)` runs `f` exactly once across all goroutines, and every caller blocks
until that first run completes — so subsequent callers see the fully-initialized
result. The naive version:

```go
if config == nil { config = load() }
```

is a race: two goroutines can both observe `nil` and both call `load()` (double
init, plus a data race on `config`). `Once` provides the correct internal
synchronization and a happens-before guarantee for what `f` wrote.
</details>

---

**8. How do you implement a timeout on a channel receive?**

<details>
<summary>Answer</summary>

`select` over the channel and a timer channel:

```go
select {
case v := <-ch:
    use(v)
case <-time.After(2 * time.Second): // or time.NewTimer in a loop
    return errors.New("timeout")
}
```

In production prefer threading a `context.WithTimeout` (Day 13) and selecting on
`ctx.Done()`, which composes cancellation across a whole call tree rather than a
single receive.
</details>

---

**9. What's the `done`-channel pattern, and what replaced it?**

<details>
<summary>Answer</summary>

A goroutine loops in a `select` over its work channel **and** a `done` channel;
closing `done` signals every worker to return:

```go
for {
    select {
    case j := <-jobs: process(j)
    case <-done:      return
    }
}
```

`close(done)` is a broadcast — all receivers wake at once. `context.Context`
(Day 13) largely replaced raw `done` channels because it also carries deadlines
and propagates cancellation down a call tree, but the underlying mechanism is the
same and many libraries still expose `done` channels.
</details>

---

**10. Why does `defer wg.Done()` matter rather than calling `wg.Done()` at the end?**

<details>
<summary>Answer</summary>

If the goroutine panics or returns early before a plain `wg.Done()`, the counter
never decrements and `wg.Wait()` blocks forever — hanging the whole program.
`defer wg.Done()` runs during stack unwinding even on panic, guaranteeing the
decrement. Same reasoning as `defer mu.Unlock()`.
</details>
