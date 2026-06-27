# Day 12 debugging — the WaitGroup that didn't wait

**Phase 3 · Concurrency · `wg.Add` inside the goroutine (Add/Wait race)**

## Symptom

We run 50 tasks concurrently and use a `sync.WaitGroup` to wait for all of them
before reporting the total. `Wait()` returns... but the work isn't actually
finished. Run it:

```bash
cd bugged
go run .
```

```
ran 8 rounds of 50 tasks; worst round completed only 0/50 before Wait returned
BUG: WaitGroup returned EARLY in 8/8 rounds (Add ran after Wait)
```

`Wait()` is supposed to block until every task calls `Done()`. Instead it sails
straight through while tasks are still running. In real code this means you read
half-written results, close a channel too early, or report "job complete" before
it is.

## Hint

Two tools name the bug instantly:

```bash
go vet ./...        # "WaitGroup.Add called from inside new goroutine"
go run -race .      # WARNING: DATA RACE on the WaitGroup counter
```

Look at *where* `wg.Add(1)` is called relative to the `go` statement. What is the
counter's value at the moment `Wait()` runs, if none of the goroutines have been
scheduled yet?

## How to reproduce

`go run .` in `bugged/` — prints an early-return in all 8 rounds, every run.
`go run -race .` reports a data race every run. `go vet ./...` flags it
statically. All deterministic; the program never hangs.

---

<details>
<summary><strong>Solution &amp; why</strong></summary>

### Root cause

`wg.Add(1)` is called **inside** each goroutine instead of **before** `go`:

```go
for i := 0; i < tasks; i++ {
    go func(id int) {
        wg.Add(1)        // BUG: runs only once the goroutine is scheduled
        defer wg.Done()
        ...
    }(i)
}
wg.Wait()
```

The loop can launch all 50 goroutines and reach `wg.Wait()` **before the
scheduler has run a single one of them**. At that instant the counter is still 0,
so `Wait()` sees "nothing to wait for" and returns immediately. The tasks then
run (and `Add`/`Done`) afterward, against a WaitGroup nobody is watching.

It's also a genuine data race: `Add` (a write to the counter) runs concurrently
with `Wait` (a read) with no happens-before relationship between them. `go vet`
catches it statically; `-race` catches it at runtime.

### The fix

Call `Add` in the loop, **before** launching the goroutine, so the counter is
fully accounted for before any goroutine — or `Wait` — runs:

```go
for i := 0; i < tasks; i++ {
    wg.Add(1)            // FIX: before `go`
    go func(id int) {
        defer wg.Done()
        ...
    }(i)
}
wg.Wait()                // now blocks until all 50 Done() calls
```

`fixed/` completes all tasks in every round and is clean under `-race`.

### The rules

> 1. **`wg.Add(n)` happens before `go`** — establish the count before anything can run.
> 2. **`defer wg.Done()`** — so it runs even if the goroutine panics.
> 3. Pass the `WaitGroup` by **pointer**, never by value (a copy's `Done` is invisible to `Wait`).

Verify: `go vet ./... && go run -race .` in `fixed/` is clean.

</details>
