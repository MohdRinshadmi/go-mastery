# Day 01 Pitfalls — Variables, Constants, iota, Functions

Each one is **Trap → Why it bites → Fix**. These are the gotchas that make Day 01
"easy" code wrong.

---

### 1. Variable shadowing with `:=`

**Trap**
```go
err := doA()
if cond {
    err := doB() // NEW err, shadows the outer one
    _ = err
}
// outer err still holds doA()'s result, not doB()'s
```

**Why it bites** `:=` only needs *one* new name on the left to be legal; it
silently re-declares everything else in the current scope. The outer variable you
thought you were updating never changes — your error check reads a stale value.

**Fix** Use `=` when you mean "assign to the variable I already have." Run
`go vet`. Keep variable scopes small.

---

### 2. Unused variables / imports are compile **errors**

**Trap**
```go
x := computeThing() // declared and not used -> BUILD FAILS
```

**Why it bites** Coming from languages where this is a warning, you expect it to
run. Go refuses to compile — an unused variable is usually dead code or a bug.

**Fix** Remove it, or assign to `_` if you deliberately want to discard it
(`_ = x`). For imports, let `goimports`/`gofmt` clean them up.

---

### 3. `:=` outside a function

**Trap**
```go
package main
count := 0   // syntax error: non-declaration statement outside function body
```

**Why it bites** `:=` is a *statement*; package level only allows declarations.

**Fix** At package scope use `var count = 0` (or `var count int`).

---

### 4. `iota` resets per `const` block — and keeps counting across skipped lines

**Trap**
```go
const (
    A = iota // 0
    B        // 1
    _        // 2 (skipped, but iota still advanced)
    D        // 3
)
const (
    X = iota // 0 again — new block resets iota
)
```

**Why it bites** People assume `iota` is a global counter or that it restarts
when a value repeats. It's per-`const`-block and increments on *every* line,
including blank/`_` ones.

**Fix** Remember: one block = one counter starting at 0, +1 per ConstSpec line.
Use `_ = iota` to skip the 0 value when you want enums to start at 1.

---

### 5. Typed vs untyped constants

**Trap**
```go
const ratio = 3        // untyped — flexible
const count int = 3    // typed int
var f float64 = ratio  // OK (untyped adapts)
var g float64 = count  // COMPILE ERROR: cannot use int as float64
```

**Why it bites** A typed constant won't implicitly convert; an untyped one
adapts to context. Over-typing constants makes them rigid.

**Fix** Leave constants untyped unless you specifically want to *restrict* usage
(e.g. an enum type like `type Weekday int`).

---

### 6. `defer` inside a loop

**Trap**
```go
for _, path := range paths {
    f, _ := os.Open(path)
    defer f.Close() // piles up — none close until the function returns
}
```

**Why it bites** `defer` fires at *function* return, not loop-iteration end. In a
long loop you leak file descriptors until the whole function exits.

**Fix** Move the work into its own function so each `defer` runs per call, or
close explicitly at the end of each iteration.

---

### 7. Returning `nil, nil` from `(T, error)`

**Trap**
```go
func find(id int) (*User, error) {
    return nil, nil // "not found"? caller has to nil-check the value too
}
```

**Why it bites** Ambiguous contract — the caller can't tell "no value" from
"success." Easy to forget the extra nil check and panic later.

**Fix** Pick a convention: return a sentinel error (`ErrNotFound`) or a
guaranteed-valid value. Don't make callers guess.
