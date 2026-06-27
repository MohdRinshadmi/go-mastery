# Day 04 Pitfalls — Error Handling

**Trap → Why it bites → Fix.**

---

### 1. Comparing wrapped errors with `==`

**Trap**
```go
if err == ErrNotFound { ... } // false when err is wrapped with %w
```

**Why it bites** `==` checks the top-level value for identity. Once an error is
wrapped (`fmt.Errorf("...: %w", ErrNotFound)`), the top-level value is the
wrapper, so the sentinel comparison fails and your special-case branch is skipped.

**Fix** Use `errors.Is(err, ErrNotFound)` — it walks the whole wrap chain.

---

### 2. Wrapping with `%v` when you needed `%w`

**Trap**
```go
return fmt.Errorf("loading config: %v", err) // flattens to a string
// caller's errors.Is(err, ErrX) now fails
```

**Why it bites** `%v` formats the cause into the message but does **not** link it,
so `errors.Is`/`errors.As` can no longer find it. Programmatic inspection breaks.

**Fix** Use `%w` when callers might inspect the cause. Use `%v` deliberately only
to *hide* internal error types at an API boundary.

---

### 3. Returning errors with no added context

**Trap**
```go
if err != nil { return err } // bubbles up bare "EOF" with no breadcrumb
```

**Why it bites** In production you get a useless one-word error and no idea which
layer produced it.

**Fix** Wrap with where-you-were: `return fmt.Errorf("parsing header: %w", err)`.

---

### 4. Using `panic` for ordinary, recoverable failures

**Trap**
```go
if missing { panic("file not found") } // caller can't handle it gracefully
```

**Why it bites** `panic` unwinds the stack like an exception and, in a goroutine
with no `recover`, kills the whole process. File-missing/bad-input/network-down
are expected failures a caller should handle.

**Fix** Return an `error`. Reserve `panic` for truly unrecoverable / impossible
states (nil dependency at startup, unreachable `default`).

---

### 5. Logging *and* returning the same error

**Trap**
```go
if err != nil {
    log.Println(err) // logged here...
    return err       // ...and again at every layer above
}
```

**Why it bites** The same error gets logged once per layer as it bubbles up —
noisy, duplicated logs that make incidents harder to read.

**Fix** Pick one: **handle** it (log) **or** **return** it. Usually return, and log
once at the top of the stack.

---

### 6. `defer f.Close()` on a file you wrote to

**Trap**
```go
f, _ := os.Create(path)
defer f.Close() // discards Close's error
```

**Why it bites** A failed `Close()` on a write can mean buffered data never
reached disk — silent data loss — and the plain `defer` throws that error away.

**Fix** Capture it with a named return and a deferred closure:
```go
func write() (err error) {
    f, err := os.Create(path)
    if err != nil { return err }
    defer func() {
        if cerr := f.Close(); err == nil { err = cerr }
    }()
    ...
}
```

---

### 7. Returning `nil, nil` from `(T, error)`

**Trap**
```go
return nil, nil // "not found"? caller must nil-check the value too
```

**Why it bites** Ambiguous contract; the caller can't distinguish "no value" from
"success," and is likely to forget the extra nil check and panic.

**Fix** Return a sentinel (`ErrNotFound`) or a guaranteed-valid value. Don't make
callers guess.

---

### 8. Capitalized / punctuated error strings

**Trap**
```go
errors.New("Connection refused.") // wrong style
```

**Why it bites** Errors get wrapped into larger messages
(`"dialing db: Connection refused."`), where a capital middle and trailing period
read wrong. `go vet`/linters flag it.

**Fix** Lowercase, no trailing punctuation: `errors.New("connection refused")`.
