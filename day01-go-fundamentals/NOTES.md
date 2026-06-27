# Day 01 Notes — Quick Reference

## Toolchain commands
```bash
go run main.go      # compile + run (dev)
go build            # produce a binary
go test ./...       # run all tests, all subpackages
go fmt ./...        # auto-format (one true format)
go vet ./...        # static analysis for likely bugs
go mod tidy         # sync go.mod / go.sum
```

## Modules
```bash
go mod init github.com/you/project   # module path == import prefix
```
Commit both `go.mod` and `go.sum`. Name the module after the real repo path.

## Variables — four ways to declare
```go
var a int = 10   // explicit type + value
var b = 10       // inferred (int)
var c int        // zero value (0)
d := 10          // short decl — ONLY inside a function
```

## Zero values (no "uninitialized garbage")
| Type | Zero |
|------|------|
| `int`, `float64` | `0` |
| `string` | `""` |
| `bool` | `false` |
| pointer, slice, map, chan, interface, func | `nil` |

`var x int` is immediately usable (it's 0). A `nil` map panics on write.

## Constants & iota
```go
const Pi = 3.14159          // untyped — flexible
const Max int = 3           // typed — restricted

type Weekday int
const (
    Sunday Weekday = iota   // 0
    Monday                  // 1
    Tuesday                 // 2
)
```
- `iota` resets to 0 per `const` block, +1 per line (incl. blank/`_` lines).
- Constants are compile-time only: no `time.Now()`, no slices/maps.

## Functions
```go
func add(a, b int) int { return a + b }            // shared-type params

func divide(a, b float64) (float64, error) {       // the (T, error) signature
    if b == 0 { return 0, errors.New("division by zero") }
    return a / b, nil
}

func split(sum int) (x, y int) {                   // named returns
    x = sum * 4 / 9
    y = sum - x
    return                                          // naked return
}

func sum(nums ...int) int { /* nums is []int */ }   // variadic

func counter() func() int {                         // closure
    n := 0
    return func() int { n++; return n }
}
```

## defer
```go
defer f.Close()   // runs on function return, LIFO order
```
- LIFO: last deferred runs first.
- Avoid `defer` inside loops (defers pile up until function returns).

## Error-handling heartbeat
```go
result, err := doSomething()
if err != nil {
    return err
}
// result is valid here
```

## Key terms
- **Module** — versioned collection of packages defined by `go.mod`.
- **Zero value** — the default value a variable has with no initializer.
- **Shadowing** — inner-scope re-declaration that hides an outer variable.
- **`iota`** — per-block auto-incrementing constant generator (Go's enum tool).
- **Untyped constant** — constant with no fixed type that adapts to context.
- **Naked return** — `return` with no values, using named return variables.
- **Closure** — a function value that captures variables from its surrounding scope.
- **Variadic** — a function taking a variable number of trailing args (`...T`).
