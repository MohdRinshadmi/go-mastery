# Challenge 02 — "but I returned nil!"

**Phase 2 · Core engineering · interfaces & errors**

## Symptom

A `validateUser` function returns `error`. When the user is valid it returns nil. The caller does the textbook check:

```go
if err := validateUser(u); err != nil {
    fmt.Println("invalid:", err)
    return
}
fmt.Println("user is valid")
```

But for a perfectly valid user it prints `invalid: <nil>` — the error branch fires even though the validation found nothing wrong. Run it:

```bash
cd bugged
go run .
```

Expected: `user is valid`
Actual: `invalid: <nil>`

## Hint

An interface value in Go is a pair: `(type, value)`. It is `nil` only when **both** halves are nil. What concrete type does `validateUser` return its "no error" value as? Print `fmt.Printf("%T\n", err)` in the caller — if it says something other than `<nil>`, you've found it.

## How to reproduce

`go run .` in `bugged/`. The bug is deterministic.

---

<details>
<summary><strong>Solution &amp; why</strong></summary>

### Root cause

The classic **typed-nil-in-an-interface** trap. The buggy code declares a concrete error variable and returns it:

```go
func validateUser(u User) error {
    var verr *ValidationError // concrete pointer type, currently nil
    if u.Name == "" {
        verr = &ValidationError{"name required"}
    }
    return verr // returning a *ValidationError, even when it's nil
}
```

When `verr` is a nil `*ValidationError`, the `return verr` statement *boxes* it into the `error` interface. The resulting interface value is `(type: *ValidationError, value: nil)`. The **type half is not nil**, so the interface as a whole is **not equal to `nil`**. The caller's `err != nil` is therefore true, and you get the phantom error with a `<nil>` value.

This bites people constantly with custom error types and with functions that return `(*MyError)` and rely on implicit conversion. It's one of the most-asked Go interview questions for a reason.

### The fix

Never return a typed nil pointer as an interface. Return the untyped `nil` literal explicitly when there's no error:

```go
func validateUser(u User) error {
    if u.Name == "" {
        return &ValidationError{"name required"}
    }
    return nil // untyped nil -> interface is genuinely nil
}
```

Now the success path returns a true nil interface `(type: nil, value: nil)`, and `err != nil` is false.

Rule of thumb:

> Functions that return `error` should return the `nil` literal for success — never a nil concrete-typed variable. Don't declare `var err *MyError` and return it; branch and `return nil` directly.

If you genuinely must hold a concrete error variable, convert deliberately: `if verr != nil { return verr }; return nil`.

</details>
