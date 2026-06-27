# Debugging Challenge — The Nil-Interface Gotcha

A `validate` function reports success on the empty path, yet the caller insists
validation *failed*. The code compiles, runs, and lies to you. This is the
signature gotcha of Day 06.

## Symptom

`validate("alice")` should return "no error", but the caller's `err != nil`
check is `true` and prints:

```
validation failed: <nil>
```

A `<nil>` error that is somehow not nil. The empty-string (failure) case looks
fine, which makes the bug even sneakier — it only bites on the success path.

## Repro

Bugged (wrong output):

```sh
cd /Users/ioss/Documents/StudyProjects/GO/day06-methods-interfaces/debugging/bugged
go run .
```

Expected (buggy) output:

```
=== bugged ===
invalid input: name: must not be empty
validation failed: <nil>
```

Fixed (correct output):

```sh
cd /Users/ioss/Documents/StudyProjects/GO/day06-methods-interfaces/debugging/fixed
go run .
```

Expected (correct) output:

```
=== fixed ===
invalid input: name: must not be empty
alice is valid
```

## Hint

Look at the *declared type* of the variable `validate` returns. The function's
return type is `error` (an interface), but the local variable is a concrete
`*ValidationError`. What exactly gets stored in the interface when you return a
nil pointer of a concrete type? Print `%T` on the returned value to see.

<details>
<summary>Solution & why</summary>

An interface value in Go is a **two-word pair**: `(type, value)`.

- A genuinely nil interface is `(nil, nil)` — both words empty.
- `var e *ValidationError` is a nil *pointer*, but it still has a concrete type.

When the bugged `validate` does `return e` on the success path, Go boxes that
nil pointer into the `error` interface. The result is `(*ValidationError, nil)`
— the **type word is set** even though the value word is nil. Comparing that
interface to `nil` checks *both* words, so `err != nil` is `true`. Printing it
calls `Error()` (or formats the nil), giving the misleading `<nil>`.

```go
// BUG: typed nil leaks into the interface
func validate(name string) error {
    var e *ValidationError      // (*ValidationError, nil) once returned
    if name == "" {
        e = &ValidationError{...}
    }
    return e
}
```

The fix is to never let a typed nil escape into the interface. Return the
untyped `nil` literal directly on the success path, which produces a true
`(nil, nil)` interface:

```go
// FIX: untyped nil -> genuine nil interface
func validate(name string) error {
    if name == "" {
        return &ValidationError{...}
    }
    return nil
}
```

**Rules of thumb:**

- A function returning `error` should `return nil` literally on success — never
  return a concrete `*T` variable that might be nil.
- If you must keep a concrete error variable, convert with a guard:
  `if e != nil { return e }; return nil`.
- `go vet` cannot always catch this; treat "typed nil returned as an interface"
  as a code-review smell.

</details>
