# Day 04 Notes — Quick Reference

## The error interface
```go
type error interface { Error() string }
```
Anything with `Error() string` is an error. Return it **last**; `nil` == success.

## The heartbeat
```go
result, err := doSomething()
if err != nil {
    return nil, err   // or handle / wrap
}
// result is valid here
```

## Creating errors
```go
errors.New("something failed")          // static
fmt.Errorf("user %d not found", id)     // formatted
fmt.Errorf("loading config: %w", err)   // WRAPPING (preserves the cause)
```
Style: lowercase, no trailing punctuation.

## Sentinel errors
```go
var ErrNotFound = errors.New("not found")

if errors.Is(err, ErrNotFound) { ... }  // NOT err == ErrNotFound
```
Stdlib sentinels: `io.EOF`, `sql.ErrNoRows`, `os.ErrNotExist`.

## Wrapping: %w vs %v
| Verb | Effect | Use when |
|------|--------|----------|
| `%w` | links the cause; `errors.Is`/`As` still find it | caller may inspect the cause |
| `%v` | flattens to a string; chain is lost | hide internal types at an API boundary |

## Inspecting the chain
```go
errors.Is(err, ErrNotFound)             // identity vs a sentinel (walks chain)

var ve *ValidationError
if errors.As(err, &ve) {                // extract a TYPE (walks chain)
    fmt.Println(ve.Field)
}
```

## Custom error type (carries data)
```go
type ValidationError struct {
    Field string
    Msg   string
}
func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed on %s: %s", e.Field, e.Msg)
}
```

## panic / recover (NOT error handling)
```go
func safe() (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("recovered: %v", r)
        }
    }()
    mightPanic()
    return nil
}
```
- Panic = "the world is broken" (nil dep at startup, impossible state).
- Return errors for anything a caller could handle.
- A panic in any goroutine with no recover kills the whole process.

## defer + Close on writes
```go
func write(path string) (err error) {
    f, err := os.Create(path)
    if err != nil { return err }
    defer func() {
        if cerr := f.Close(); err == nil { err = cerr }
    }()
    // ... write ...
    return nil
}
```
Plain `defer f.Close()` discards Close's error — fine for reads, risky for writes.

## Key terms
- **Sentinel error** — package-level error value compared with `errors.Is`.
- **Wrapping** — `%w` attaches a cause while adding context.
- **`errors.Is`** — chain-walking identity check against a target value.
- **`errors.As`** — chain-walking type extraction into a target pointer.
- **Custom error type** — a struct implementing `error` that carries data.
- **panic / recover** — stack unwind / stop the unwind (only in a `defer`).
- **Named return** — lets a `defer` set/override the returned error.
