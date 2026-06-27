# Day 01 — Installation, Modules, Variables, Constants, Functions

> Mentor note: Today is the foundation. It looks easy. It is not. The difference between a junior and a senior shows up in *how* you declare a variable, *where* you put a constant, and *how* you design a function signature. Read every "Senior take" box.

---

## 0. Installation & toolchain

You don't need to memorize this, but you must understand it.

- `go` is a single binary that is compiler + build system + test runner + formatter + package manager. This is deliberate. In Java/Node you stitch together Maven/npm/eslint/jest. Go ships them as one tool. **Less choice = less bikeshedding = faster teams.** That philosophy runs through the whole language.
- Key commands you'll use daily:
  - `go run main.go` — compile + run in one step (dev only).
  - `go build` — produce a binary.
  - `go test ./...` — run all tests in all subpackages.
  - `go fmt ./...` — auto-format. **There is one correct format. You never argue about style in Go.** This kills an entire category of PR comments.
  - `go vet ./...` — static analysis for likely bugs.
  - `go mod tidy` — sync dependencies.

**Senior take:** The fact that formatting is non-negotiable (`gofmt`) is a *feature*. On a Go team nobody reviews whitespace. You review logic. Embrace it.

---

## 1. Go Modules

### Theory
A **module** is a collection of packages with a version, defined by a `go.mod` file at its root. It's how Go knows your project's name and its dependencies.

### Why it exists
Before modules (pre-2018), Go used `GOPATH` — all code lived in one global folder and there was no real dependency versioning. It was painful. Modules gave Go reproducible builds with pinned versions (like `package.json` + lockfile, but simpler).

### Create one
```bash
go mod init github.com/yourname/projectname
```
This makes `go.mod`:
```
module github.com/yourname/projectname

go 1.22
```
The module path (`github.com/yourname/projectname`) is also the **import prefix**. If you have a package in `internal/auth/`, others import it as `github.com/yourname/projectname/internal/auth`.

### When to use / not
- Always use modules. There is no modern Go without them.
- `go.sum` is the lockfile (cryptographic hashes). **Commit both `go.mod` and `go.sum`.** Never `.gitignore` them.

**Senior take:** Name your module after where it actually lives (the repo URL). New devs name it `myapp` and then can't publish or be imported cleanly. Use the real path from day one.

---

## 2. Variables

### Theory
Go is statically typed. Every variable has a type known at compile time. But Go infers types so you rarely write them out.

### The four ways to declare
```go
var a int = 10        // explicit type + value
var b = 10            // type inferred (still int)
var c int             // zero value (c == 0)
d := 10               // short declaration — ONLY inside functions
```

### The single most important Go concept: ZERO VALUES
There is **no "uninitialized" / null-garbage** in Go. Every type has a zero value:

| Type | Zero value |
|------|-----------|
| `int`, `float64` | `0` |
| `string` | `""` (empty, not nil) |
| `bool` | `false` |
| pointers, slices, maps, channels, interfaces, funcs | `nil` |

This means `var count int` is *immediately safe to use* — it's 0. No `NullPointerException` for value types. This is a deliberate safety design.

### When to use which
- `:=` inside functions — this is what you'll write 90% of the time.
- `var x Type` when you want the zero value (e.g. `var buf bytes.Buffer`).
- `var x = ...` at package level (you **cannot** use `:=` outside a function).

### Common mistakes
1. **Unused variables are a compile ERROR.** Not a warning. Go refuses to build.
   ```go
   x := 5 // declared and not used  -> BUILD FAILS
   ```
   Why? Unused vars are almost always a bug or dead code. Go forces you to clean up.
2. **Shadowing** — re-declaring in an inner scope by accident:
   ```go
   err := doA()
   if cond {
       err := doB() // NEW err, shadows outer. Outer err never updated. Classic bug.
   }
   ```
   `go vet` and the `-vet=shadow` linter catch this. Watch for it.

**Senior take:** Prefer `:=`, but reach for `var` when zero value *is* the intent. `var sb strings.Builder` reads better than `sb := strings.Builder{}`.

---

## 3. Constants

### Theory
Values fixed at compile time. Declared with `const`. Can be numbers, strings, booleans, runes — **not** slices, maps, or anything computed at runtime.

```go
const Pi = 3.14159
const MaxRetries = 3
const AppName = "checkout-service"
```

### iota — Go's enum mechanism
```go
type Weekday int
const (
    Sunday Weekday = iota // 0
    Monday                // 1
    Tuesday               // 2
)
```
`iota` resets to 0 in each `const` block and increments per line. This is how Go does enums (it has no `enum` keyword).

### Why it exists / when to use
- Constants are inlined by the compiler → zero runtime cost, and they can't be accidentally mutated.
- Use them for: magic numbers, config defaults, enum-like sets, bit flags.

### When NOT to use
- Anything needing runtime computation (`const t = time.Now()` ❌ — won't compile).
- Values that legitimately change per environment (use config, not consts).

**Senior take:** Untyped constants (`const MaxRetries = 3`) are more flexible than typed ones — they adapt to context (can be used as int, int64, float). Only add a type when you want to *restrict* usage (like the `Weekday` enum above).

---

## 4. Functions

### Theory
First-class citizens. Can be passed around, returned, assigned to variables, closed over.

### Basic
```go
func add(a int, b int) int {
    return a + b
}
// shorthand when consecutive params share a type:
func add(a, b int) int { return a + b }
```

### Multiple return values — THE Go signature
```go
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}
```
This `(result, error)` pattern is the backbone of all idiomatic Go. You'll write it thousands of times:
```go
result, err := divide(10, 2)
if err != nil {
    return err // handle immediately
}
```

### Named return values
```go
func split(sum int) (x, y int) {
    x = sum * 4 / 9
    y = sum - x
    return // "naked return" — returns x, y
}
```
Use sparingly. Good for documenting what's returned; bad when it hurts readability with naked returns in long functions.

### Variadic
```go
func sum(nums ...int) int { // nums is []int
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}
sum(1, 2, 3) // 6
```

### Closures
```go
func counter() func() int {
    count := 0
    return func() int {
        count++
        return count
    }
}
c := counter()
c() // 1
c() // 2  -- count is "remembered"
```

### defer — cleanup that always runs
```go
func readFile() error {
    f, err := os.Open("data.txt")
    if err != nil {
        return err
    }
    defer f.Close() // runs when function returns, no matter how
    // ... use f
    return nil
}
```
`defer` runs in **LIFO** order and is your tool for guaranteed cleanup (closing files, unlocking mutexes, closing DB rows).

### Common mistakes
1. Ignoring the error return: `result, _ := divide(...)` — sometimes fine, usually a code smell. Justify every `_`.
2. `defer` inside a loop → defers pile up until the function returns, not each iteration. Can leak resources.
3. Returning `nil, nil` from a `(T, error)` function — ambiguous. Pick a convention.

### Performance implications
- Go functions are cheap to call but **not free**. The compiler inlines small functions automatically.
- Returning large structs by value copies them. For big structs in hot paths, return pointers — but don't pointer-everything; small structs are faster by value (cache-friendly, no heap allocation). We'll measure this in Phase 2.

**Senior take:** A function should do *one thing* and its signature should tell the whole story. `func ProcessOrder(o Order) (Receipt, error)` — I know exactly what goes in, what comes out, and that it can fail. Design the signature first, body second.

---

## Expert Thinking Mode — how different levels see "a function"

- **Beginner:** "A function is reusable code I call to get a result."
- **Senior:** "A function is a contract. Its signature is documentation. Errors are part of the return type, not exceptions thrown into the void."
- **Staff:** "Function boundaries are where I draw testability and dependency lines. Pure functions are trivial to test; functions that reach out to the network are not — so I shape signatures to inject dependencies."
- **Architect:** "Function signatures across a codebase form an API surface. Consistency here (always `(T, error)`, context first, etc.) is what lets a 200-engineer org move without constant friction."

---

## Real-world use

- **Stripe / payments:** `(result, error)` everywhere — a charge either succeeds with a receipt or fails with a typed error. No exceptions silently bubbling up.
- **Constants & iota:** order states (`Pending`, `Paid`, `Shipped`) are modeled as `iota` enums across every commerce backend.
- **defer:** every DB query, file handle, and mutex in production Go is cleaned up with `defer`. It's the #1 leak-prevention tool.
- **Zero values:** Go's `sync.Mutex`, `bytes.Buffer`, `strings.Builder` are all designed so their zero value is ready to use — no constructor needed. That's idiomatic API design you'll copy.

---

## Interview Questions (answer these in your head, we'll discuss)

1. What is the zero value of a `string`, a `slice`, and a `map`? Which are safe to use immediately, which will panic?
2. Why are unused variables a compile error in Go? What's the design reasoning?
3. Difference between `var x = 5` and `x := 5`? Where can you use each?
4. What does `iota` do and why does Go not have an `enum` keyword?
5. Explain `defer` execution order. What does this print?
   ```go
   defer fmt.Println("A")
   defer fmt.Println("B")
   fmt.Println("C")
   ```
6. What is variable shadowing and why is it dangerous?
7. Why does Go use multiple return values instead of exceptions for errors?

---

## Your tasks for today

Go to `../exercises/`. There are **3 beginner exercises** + **1 intermediate challenge** with starter files. Fill them in, run them, and tell me when done. I will review each like a production PR.

Don't open `../solutions/` until you've tried. I'll know. 😄

---

## Day 01 companion files

- [Debugging challenge](../debugging/README.md) — variable shadowing turns a failed withdrawal into "all succeeded."
- [Pitfalls](../PITFALLS.md) — Trap → Why it bites → Fix.
- [Interview questions](../INTERVIEW.md) — with model answers.
- [Notes / cheatsheet](../NOTES.md) — quick reference.
- [Resources](../RESOURCES.md) — curated links.
