# Day 04 Interview Questions — Error Handling

---

**1. Why does Go use returned error values instead of exceptions? What do you gain and lose?**

<details><summary>Answer</summary>

Returning `error` as the last value makes every failure path **explicit and
local** — you can see what can fail by reading a function top to bottom, with no
invisible control flow leaping out of a nested call. You gain honesty,
readability, and easier reasoning about partial failure. You lose some brevity:
the repeated `if err != nil` is more verbose than a single try/catch.
</details>

---

**2. Difference between `%w` and `%v` in `fmt.Errorf`? When would you deliberately choose `%v`?**

<details><summary>Answer</summary>

`%w` **wraps**: the new error carries both a message and a link to the original,
so `errors.Is`/`errors.As` can still find the cause. `%v` just formats the cause
into the string, flattening it with no link. Choose `%v` deliberately at an **API
boundary** when you want to *hide* internal error types from callers so they can't
couple to them.
</details>

---

**3. Why `errors.Is(err, ErrNotFound)` instead of `err == ErrNotFound`?**

<details><summary>Answer</summary>

`==` only matches the exact top-level error value. Once an error has been wrapped
with `%w`, the top-level value is the wrapper, not the sentinel, so `==` returns
false. `errors.Is` walks the entire wrap chain and matches the sentinel wherever
it is — so it keeps working through wrapping.
</details>

---

**4. When do you use `errors.As` vs `errors.Is`?**

<details><summary>Answer</summary>

`errors.Is` tests **identity** against a sentinel value ("is this error, anywhere
in the chain, `ErrNotFound`?"). `errors.As` finds an error of a specific **type**
in the chain and assigns it to your pointer so you can read its fields
(`var ve *ValidationError; errors.As(err, &ve)`). Use `Is` for sentinels, `As`
for typed errors carrying data.
</details>

---

**5. When is `panic` appropriate, and where is `recover` actually used in production?**

<details><summary>Answer</summary>

`panic` is for **truly unrecoverable** programmer errors or impossible states — a
nil dependency at startup, an unreachable `default` in an exhaustive switch — not
for expected failures like a missing file. The standard place for `recover` is at
the top of a server's per-request handler or a worker goroutine, so one bad
request can't crash the entire process (a panic in any goroutine with no recover
kills the program).
</details>

---

**6. A function logs an error and also returns it. Why is that a smell?**

<details><summary>Answer</summary>

Because the caller will likely log it too, and so will *its* caller — the same
error gets logged once per layer as it bubbles up, producing noisy, duplicated log
lines. The convention is: either **handle** the error (and log it) **or** return
it — usually return, and log once at the top of the stack.
</details>

---

**7. What's wrong with `defer f.Close()` on a file you wrote to?**

<details><summary>Answer</summary>

`Close` can return an error, and on a **write** that error can mean buffered data
never reached disk — silent data loss. Plain `defer f.Close()` discards that
error. For writes, capture it with a named return + deferred closure so a failed
`Close` is reported.
</details>

---

**8. What is a sentinel error, and what are the trade-offs versus a typed error?**

<details><summary>Answer</summary>

A **sentinel** is a package-level error value (`var ErrNotFound = errors.New(...)`)
that callers compare against with `errors.Is`. It's simple but carries no data and
becomes part of your public API (changing it is breaking). A **typed error** is a
struct implementing `error` that can carry fields (e.g. the offending field name)
and is extracted with `errors.As`. Use sentinels for simple categories, typed
errors when callers need structured data.
</details>

---

**9. Why are Go error strings lowercase with no trailing punctuation?**

<details><summary>Answer</summary>

Because errors are routinely wrapped into larger messages
(`"loading config: connection refused"`). A capitalized or period-terminated
fragment reads wrong in the middle of a wrapped chain. `go vet` and linters
enforce the convention.
</details>

---

**10. What does this print, and why?**
```go
func f() (err error) {
    defer func() { err = errors.New("from defer") }()
    return errors.New("from return")
}
```

<details><summary>Answer</summary>

It returns the error `"from defer"`. With a **named** return value, a deferred
function can read and overwrite the return value after `return` sets it. The
`return` statement assigns `err = "from return"`, then the deferred closure runs
and reassigns `err = "from defer"`, which is what the caller receives. This is the
same mechanism used to capture a `Close` error or to `recover` into `err`.
</details>

---

**11. How would you map a wrapped error to an HTTP status code at an API boundary?**

<details><summary>Answer</summary>

Inspect the chain and branch: `errors.Is(err, ErrNotFound)` → 404;
`var ve *ValidationError; errors.As(err, &ve)` → 400 (and you can include
`ve.Field`); otherwise → 500. This is exactly why wrapping with `%w` matters —
the boundary can still classify the cause even though intermediate layers added
context.
</details>
