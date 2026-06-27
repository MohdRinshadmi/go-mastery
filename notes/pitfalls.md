# Go Pitfalls — The Consolidated Gotchas List

> These are the bugs that pass in dev, look fine in review, and corrupt data in prod. Grouped by topic. Each entry: **Trap → Why it bites → Fix.** If you can recite the fix from the trap alone, you're ready.

---

## Slices & maps (Day 02)

**Slice aliasing on append.**
- *Trap:* `b := a[:3]; b = append(b, 99)` silently overwrites `a[3]`.
- *Why:* While `cap > len`, append writes into the **shared backing array** — no new allocation, no copy.
- *Fix:* Three-index slice to cap-limit (`a[:3:3]`) or copy first (`append([]int{}, a...)`).

**`range` copies values.**
- *Trap:* `for _, p := range pts { p.X = 99 }` leaves `pts` unchanged.
- *Why:* The loop variable is a **copy** of each element, not a reference into the slice.
- *Fix:* Index it: `for i := range pts { pts[i].X = 99 }`.

**Discarding append's result.**
- *Trap:* `append(s, x)` on its own line — the new element vanishes.
- *Why:* `append` returns a **new slice header**; ignoring it discards updated len/ptr.
- *Fix:* Always `s = append(s, x)`.

**Writing to a nil map panics.**
- *Trap:* `var m map[string]int; m["k"] = 1` → `panic: assignment to entry in nil map`.
- *Why:* The zero value of a map is nil; reads are safe but writes need an allocated map.
- *Fix:* `m := make(map[string]int)` (or a map literal) before writing.

**Modifying a map during range.**
- *Trap:* Adding/deleting keys while ranging gives nondeterministic, partially-visible results.
- *Why:* Map iteration order is randomized and the spec doesn't define behavior for added keys mid-range.
- *Fix:* Collect keys/changes first, then mutate after the loop. (Delete of the current key is the one defined exception.)

**Relying on map iteration order.**
- *Trap:* Tests/CSV/audit logs that depend on key order flake.
- *Why:* Go **deliberately randomizes** map order to stop you depending on it.
- *Fix:* Extract keys, `sort` them, iterate the sorted slice.

---

## Structs, methods, interfaces (Days 03, 06)

**Value receiver can't mutate.**
- *Trap:* `func (u User) SetName(n string) { u.Name = n }` does nothing visible.
- *Why:* A value receiver operates on a **copy** of the struct.
- *Fix:* Use a pointer receiver: `func (u *User) SetName(...)`. Keep receivers consistent across a type.

**nil interface != nil.**
- *Trap:* Returning a nil `*MyError` as `error` makes `err != nil` true even though "nothing went wrong".
- *Why:* An interface is nil only when **both** its type and value are nil; here the type (`*MyError`) is set.
- *Fix:* Return a literal `nil` for the error, or check before wrapping — don't stuff a typed nil pointer into an `error`.

**Comparing structs with uncomparable fields.**
- *Trap:* `a == b` panics or won't compile if the struct contains a slice/map/func field.
- *Why:* `==` requires every field to be comparable.
- *Fix:* Write an explicit `Equal` method or use `reflect.DeepEqual` (tests only — it's slow).

---

## Errors (Day 04)

**Comparing wrapped errors with `==`.**
- *Trap:* `if err == ErrNotFound` returns false once the error has been wrapped with `%w`.
- *Why:* `==` only matches the top-level value; wrapping nests the sentinel deeper in the chain.
- *Fix:* `errors.Is(err, ErrNotFound)` (and `errors.As` for typed errors) — they walk the chain.

**Wrapping with no context.**
- *Trap:* `if err != nil { return err }` everywhere; prod log just says `"EOF"`.
- *Why:* No breadcrumb of *where* it failed.
- *Fix:* `return fmt.Errorf("parsing header: %w", err)` — add the operation at each layer.

**`%v` where you meant `%w`.**
- *Trap:* `fmt.Errorf("x: %v", err)` flattens the cause so callers can't `errors.Is` it.
- *Why:* `%v` stringifies and drops the wrap link; `%w` preserves it.
- *Fix:* Use `%w` when callers may inspect the cause; use `%v` deliberately to hide internals at an API boundary.

**Deferred `Close()` error ignored.**
- *Trap:* `defer f.Close()` on a file you **wrote** to silently drops a failed close (= possibly lost data).
- *Why:* `defer f.Close()` discards the returned error.
- *Fix:* Named return + `defer func(){ if cerr := f.Close(); err == nil { err = cerr } }()`.

**Log-and-return the same error.**
- *Trap:* Logging then returning makes the error appear N times as it bubbles up.
- *Why:* Each layer logs the same failure.
- *Fix:* Pick one — handle (log) **or** return. Usually return; log once at the top.

---

## JSON & encoding (Day 05)

**Unexported fields aren't marshaled.**
- *Trap:* `json.Marshal` silently omits lowercase fields; they come back as zero on Unmarshal.
- *Why:* encoding/json uses reflection, which can only see **exported** fields.
- *Fix:* Capitalize fields you want serialized; use a `json:"name"` tag to control the wire name.

**Forgetting `omitempty` / `-` semantics.**
- *Trap:* Zero values clutter output, or a secret field leaks into JSON.
- *Why:* Default is "always include, field name = Go name".
- *Fix:* `json:"x,omitempty"` to drop zero values; `json:"-"` to never serialize.

---

## Concurrency: goroutines & channels (Days 11–12)

**Loop variable capture (and the Go 1.22 change).**
- *Trap (pre-1.22):* `for i := ...; { go func(){ println(i) }() }` — all goroutines print the last `i`.
- *Why:* The closure captured the **single shared** loop variable, read after the loop finished.
- *Fix:* Go **1.22+** gives each iteration its own copy, so this specific case is fixed. Pre-1.22 (or for clarity), rebind `i := i` or pass `i` as an argument. Still understand it — old code and other shared-capture cases exist.

**Unbuffered channel deadlock.**
- *Trap:* `ch := make(chan int); ch <- 1` with no concurrent receiver → `fatal error: all goroutines are deadlocked`.
- *Why:* An unbuffered send blocks until a receiver is ready; main is the only goroutine.
- *Fix:* Launch the receiver in a goroutine first, or use a buffered channel if decoupling is intended.

**Forgetting to close → `range` hangs.**
- *Trap:* `for v := range ch` never returns because the sender never `close`s.
- *Why:* `range` exits only when the channel is closed **and** drained.
- *Fix:* `defer close(out)` in the sole sender. Remember: only the sender closes; send-after-close panics.

**Closing from the wrong end / twice.**
- *Trap:* Multiple senders each `close(ch)` → `panic: close of closed channel`.
- *Why:* Closing is a one-time broadcast owned by the sender.
- *Fix:* Use a `WaitGroup` so a single coordinator closes after all senders finish, or `sync.Once`.

**main() exits before goroutines finish.**
- *Trap:* `go work()` then return from main — work may never run.
- *Why:* The runtime kills all goroutines when main returns.
- *Fix:* Synchronize with `WaitGroup`, a channel, or context before returning.

**Goroutine leaks.**
- *Trap:* A goroutine started in a handler blocks forever on a channel/`ctx` that never fires.
- *Why:* No exit condition; it holds memory/locks/FDs until OOM.
- *Fix:* Give every goroutine an owner and an exit (`ctx.Done()`, closed channel, finished work). Test with `goleak`/`runtime.NumGoroutine()`.

---

## Concurrency: context & shared state (Days 13–14)

**Forgetting to cancel context.**
- *Trap:* `ctx, _ := context.WithTimeout(...)` — dropping `cancel` leaks the timer/goroutine.
- *Why:* `cancel` releases the resources backing the deadline even when it fires normally.
- *Fix:* `ctx, cancel := context.WithTimeout(...); defer cancel()` — always, even on the happy path.

**Data races on shared state.**
- *Trap:* Many goroutines doing `counter++` (or read+write a map) with no lock.
- *Why:* `counter++` is read-modify-write, not atomic; concurrent map writes panic outright.
- *Fix:* Guard with `sync.Mutex`/`RWMutex`, or use `atomic.Int64`. **Run `go test -race` in CI** — you cannot find these by reading code.

**Copying a `sync.Mutex` by value.**
- *Trap:* Passing a struct-with-mutex by value copies the lock; the copy guards nothing.
- *Why:* Each copy is an independent lock.
- *Fix:* Use pointer receivers / pass pointers. `go vet` catches this.

**Forgetting `defer mu.Unlock()` on an early return.**
- *Trap:* Returning while holding the lock → next caller deadlocks.
- *Why:* The lock is never released on that path.
- *Fix:* `mu.Lock(); defer mu.Unlock()` so it unlocks on every return (and panic).

**Storing `context.Context` in a struct.**
- *Trap:* Caching a ctx as a field then reusing it after it's cancelled.
- *Why:* Context is per-call-tree; a stored one goes stale.
- *Fix:* Pass `ctx` explicitly as the first parameter of each call.

---

## Concurrency: timers & select (Days 12, 15, 25)

**`time.After` in a loop leaks.**
- *Trap:* `for { select { case <-time.After(d): ...; case <-ch: ... } }` allocates a fresh timer each iteration.
- *Why:* `time.After` creates a `Timer` that isn't GC'd until it fires; under load these pile up.
- *Fix:* Hoist a `time.NewTimer`/`NewTicker` outside the loop and `Reset`/`Stop` it (or just `defer ticker.Stop()`).

**select with a nil channel that never fires.**
- *Trap:* Forgetting that a `case <-nilCh` blocks forever (sometimes intended, often a bug).
- *Why:* Operations on a nil channel block permanently.
- *Fix:* Set a channel to nil to *intentionally* disable a select case; otherwise ensure it's `make`d.

---

## HTTP & shutdown (Days 16–25)

**Not closing `resp.Body`.**
- *Trap:* `resp, _ := http.Get(url)` without `defer resp.Body.Close()` leaks connections.
- *Why:* The underlying TCP connection isn't returned to the pool until the body is closed.
- *Fix:* `defer resp.Body.Close()` immediately after the error check.

**No graceful shutdown.**
- *Trap:* Process killed mid-request drops in-flight work and connections.
- *Why:* `ListenAndServe` doesn't drain on signal by itself.
- *Fix:* Catch SIGTERM, call `srv.Shutdown(ctx)` with a deadline to finish in-flight requests, then exit.
