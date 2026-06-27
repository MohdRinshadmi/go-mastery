# Day 04 Debugging Challenge — The sentinel that never matches

A `greet` function is supposed to treat a missing user as a *soft* case (return a
generic "Hello, stranger!") and only surface real errors. For a known user it
works; for an unknown ID it's supposed to greet the stranger — but instead the
"not found" error **leaks out** as a hard error.

This is the **comparing wrapped errors with `==`** trap from today's lesson.

## Reproduce

```bash
cd bugged
go run .
```

Observed output:

```
id=1 -> Hello, Ada!
id=99 ERROR: lookup id=99: user not found   <-- should be "Hello, stranger!"
```

## Hint

`lookup` returns the sentinel **wrapped** with `%w`. What is the concrete type of
the `err` value the caller receives? Does `err == ErrNotFound` compare the
sentinel, or the wrapper around it?

<details>
<summary>Solution &amp; why</summary>

`lookup` doesn't return `ErrNotFound` directly — it returns
`fmt.Errorf("lookup id=%d: %w", id, ErrNotFound)`, which is a `*fmt.wrapError`
wrapping the sentinel. The caller then does:

```go
if err == ErrNotFound { ... } // always false
```

`==` compares the **top-level** error value for identity. The top-level value is
the wrapper, not `ErrNotFound`, so the comparison is false and the "soft" branch
is never taken. The error leaks out as if it were unexpected.

**Fix:** use `errors.Is`, which walks the entire `%w` wrap chain looking for the
target:

```go
if errors.Is(err, ErrNotFound) {
    return "Hello, stranger!", nil
}
```

Now the wrapped sentinel is matched and the stranger is greeted:

```
id=1  -> Hello, Ada!
id=99 -> Hello, stranger!
```

Rules to internalize:
- **Never** compare errors with `==` against a sentinel; use `errors.Is`.
- Use `errors.As(err, &target)` when you need to extract a specific error *type*
  (to read its fields), not just test identity.
- Wrapping with `%w` is correct and desirable — it preserves the cause *and* lets
  `errors.Is`/`errors.As` still find it. The bug is in how the caller inspects it,
  not in the wrapping.
</details>
