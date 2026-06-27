# Day 04 — Error Handling (the Go way)

> Mentor note: This is the most important day in Phase 1. Coming from Python/Java/JS, your instinct is `try/catch` and `throw`. **Go has none of that.** Errors are ordinary values you return and check. This feels verbose at first ("why am I writing `if err != nil` everywhere?") and then one day you realize you can see every failure path in a function just by reading it top to bottom — no invisible control flow leaping out of nested calls. That clarity is the whole point. Internalize this today; everything in Phases 2–6 builds on it.

---

## 1. The `error` type

### Theory
`error` is just an interface with one method:

```go
type error interface {
    Error() string
}
```

Anything with an `Error() string` method is an error. That's it. Functions that can fail return `error` as their **last** return value, and `nil` means success.

```go
func readConfig(path string) (Config, error) {
    // ... on failure:
    return Config{}, errors.New("config not found")
    // ... on success:
    return cfg, nil
}
```

### Why it exists
Exceptions create **invisible control flow**: any line might throw, jumping out of the function to some `catch` far away. In a big codebase you can't tell what throws without reading every callee. Go makes failure **explicit and local**: the error is right there in the return signature, and you handle it on the next line. Trade verbosity for honesty.

### The heartbeat of Go
```go
result, err := doSomething()
if err != nil {
    return nil, err   // or handle/log/wrap
}
// use result — guaranteed valid here
```
You'll write this thousands of times. Embrace it; don't fight it.

---

## 2. Creating errors

```go
errors.New("something failed")                 // static message
fmt.Errorf("user %d not found", id)            // formatted
fmt.Errorf("loading config: %w", err)          // WRAPPING (note %w) — see below
```

**Senior take:** Error strings are lowercase, no trailing punctuation (`"connection refused"`, not `"Connection refused."`). Why? Errors get wrapped: `"loading config: connection refused"`. Capitalized middles read wrong. `go vet` and linters enforce this.

---

## 3. Sentinel errors

A package-level error value you can compare against:

```go
var ErrNotFound = errors.New("not found")

func Get(id string) (*User, error) {
    // ...
    return nil, ErrNotFound
}

// caller:
u, err := Get("x")
if errors.Is(err, ErrNotFound) {   // use errors.Is, NOT err == ErrNotFound
    // handle the "missing" case specifically
}
```

Standard library examples you'll use: `io.EOF`, `sql.ErrNoRows`, `os.ErrNotExist`.

**Why `errors.Is` not `==`?** Because errors get wrapped (next section). `==` only matches the exact top-level value; `errors.Is` walks the whole wrap chain. Always `errors.Is`.

---

## 4. Wrapping with `%w` — adding context without losing the cause

When you pass an error up, add context about *where you were*:

```go
func loadUser(id string) (*User, error) {
    data, err := db.Query(id)
    if err != nil {
        return nil, fmt.Errorf("loadUser %s: %w", id, err)
    }
    ...
}
```

`%w` (not `%v`) **wraps**: the new error carries a message *and* a link to the original. The final string might read:

```
loadUser 42: query failed: connection refused
```

— a breadcrumb trail from the high-level operation down to the root cause. And callers can still inspect the original:

- `errors.Is(err, sql.ErrNoRows)` — is this (anywhere in the chain) that sentinel?
- `errors.As(err, &target)` — extract a specific error *type* from the chain.

**Senior take:** Wrap with `%w` when the caller might want to programmatically inspect the cause; use `%v` to flatten it to a plain string when you deliberately want to *hide* internal error types from callers (an API boundary). Choosing `%w` vs `%v` is an API design decision, not a style one.

---

## 5. Custom error types

When an error needs to carry **data**, make it a struct:

```go
type ValidationError struct {
    Field string
    Msg   string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed on %s: %s", e.Field, e.Msg)
}
```

Extract it with `errors.As`:

```go
var ve *ValidationError
if errors.As(err, &ve) {
    fmt.Println("bad field:", ve.Field)   // access the structured data
}
```

`errors.As` walks the wrap chain looking for an error of that type and, if found, assigns it to your pointer. This is how you handle different failure categories differently (validation → 400, not-found → 404, db down → 500).

---

## 6. panic / recover — NOT your error handling

`panic` unwinds the stack like an exception. `recover` (only useful inside a `defer`) stops the unwind.

```go
func safe() (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("recovered from panic: %v", r)
        }
    }()
    mightPanic()
    return nil
}
```

### When to panic
- **Truly unrecoverable** programmer errors / impossible states (a `default:` in a switch that should be exhaustive, a nil dependency at startup).
- Almost never in library code for *expected* failures.

### When NOT to panic
- Anything a caller could reasonably handle: file missing, bad input, network down. Return an `error`.

### The one place recover is standard
Top of a server's request handler / worker goroutine: recover so one bad request doesn't crash the whole process. (We'll do this for real in Phase 4 middleware.) **A panic in any goroutine with no recover kills the entire program.**

**Senior take:** "Don't panic" is a real Go proverb. If your reflex is to `throw`, return an error instead. Panics are for "the world is broken," not "this request failed."

---

## 7. `defer` for cleanup (reinforced)

`defer` guarantees cleanup runs on every return path, including after an error:

```go
f, err := os.Open(path)
if err != nil {
    return err
}
defer f.Close()   // runs no matter how we leave the function
```

Watch out: to capture a `Close()` error you must do it explicitly (named return + defer closure). Plain `defer f.Close()` discards its error — fine for reads, risky for writes (a failed `Close` on a write can mean lost data).

---

## Common mistakes

1. `if err != nil { return err }` everywhere with **no added context** — when it fails in prod you get `"EOF"` with no idea where. Wrap: `fmt.Errorf("parsing header: %w", err)`.
2. Using `==` to compare wrapped errors. Use `errors.Is`.
3. `panic` for ordinary failures. Return errors.
4. Ignoring errors with `_` to "make it compile." Every `_ =` on an error is a decision you must justify.
5. Logging **and** returning the same error (it gets logged 5× as it bubbles up). Pick one: handle it (log) **or** return it. Usually return; log once at the top.
6. Returning `nil, nil` from `(T, error)` — ambiguous. If there's no value, that's usually an error or a sentinel.

## Performance

- Creating errors is cheap, but `fmt.Errorf` allocates. In ultra-hot paths that error on every call (rare), sentinel errors avoid allocation. Don't micro-optimize the happy path away over this.
- `errors.Is/As` walk the chain — negligible cost, don't worry about it.

---

## Expert Thinking Mode — "a function failed"

- **Beginner:** "I'll throw/return an error string and move on."
- **Senior:** "What does the caller need to *do* differently for each failure? That decides sentinel vs typed error vs plain string. I wrap with context at every layer so the logs tell a story."
- **Staff:** "Errors are part of my package's public contract. Which errors do I promise callers can `errors.Is`? Changing that is a breaking change. I hide internal causes at API boundaries with `%v`."
- **Architect:** "Across services, errors become status codes, retries, and alerts. Error taxonomy (retryable vs fatal vs client-fault) is a system design concern — it drives circuit breakers and SLOs."

---

## Real-world use

- **Stripe/payments:** typed errors (`CardError`, `RateLimitError`) so callers branch correctly — decline vs retry vs bug.
- **`sql.ErrNoRows`:** every Go DB layer checks `errors.Is(err, sql.ErrNoRows)` to turn "no row" into a 404 instead of a 500.
- **gRPC/HTTP boundaries:** internal wrapped errors get mapped to status codes; the wrap chain feeds structured logs and traces.
- **recover in middleware:** every production HTTP framework recovers panics per-request so one bad handler can't take down the server.

---

## Interview Questions

1. Why does Go use returned error values instead of exceptions? What do you gain and lose?
2. Difference between `%w` and `%v` in `fmt.Errorf`? When would you deliberately choose `%v`?
3. Why `errors.Is(err, ErrNotFound)` instead of `err == ErrNotFound`?
4. When do you use `errors.As` vs `errors.Is`?
5. When is `panic` appropriate? Where is `recover` actually used in production?
6. A function logs an error and also returns it. Why is that a smell?
7. What's wrong with `defer f.Close()` on a file you wrote to?

---

## Your tasks

`../exercises/` — three beginner exercises (sentinel errors, wrapping, a custom error type) plus a challenge: a tiny config-loader that distinguishes "file missing" (use a sentinel) from "invalid value" (custom typed error with the field name), wrapping with `%w` along the way and letting the caller branch with `errors.Is` / `errors.As`. Bring it for PR review.

---

## Day 04 companion files

- [Debugging challenge](../debugging/README.md) — a wrapped sentinel never matches `==`, so a soft "not found" leaks as a hard error.
- [Pitfalls](../PITFALLS.md) — Trap → Why it bites → Fix.
- [Interview questions](../INTERVIEW.md) — with model answers.
- [Notes / cheatsheet](../NOTES.md) — quick reference.
- [Resources](../RESOURCES.md) — curated links.
