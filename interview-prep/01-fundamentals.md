# Phase 1 — Go Fundamentals (Days 01–05)

Zero values, `defer`/`iota`/shadowing, slice & map internals, structs & pointers, receivers, error values, files & JSON. Self-quiz: read, answer aloud, then expand.

---

### 1. What is the zero value of a `string`, a `slice`, and a `map`? Which are safe to use immediately?

<details><summary>Answer</summary>

`""` for `string`, `nil` for both `slice` and `map`. A `nil` slice is fully usable: `len`, `range`, and `append` all work (`append` allocates the backing array on first use). A `nil` map is safe to **read** (returns the value's zero value) but **panics on write** — you must `make` it first. So `var s []int; s = append(s, 1)` is fine, but `var m map[string]int; m["x"] = 1` panics.
</details>

---

### 2. Why are unused variables (and imports) a compile error in Go? What's the design reasoning?

<details><summary>Answer</summary>

Go treats them as errors, not warnings, because warnings get ignored and rot. An unused variable is almost always a bug (a typo, a forgotten use, a shadow) or dead code that confuses the next reader. Failing the build forces it to zero. The cost is mild friction while drafting — handled with `_ = x` or just deleting the line. It's the same philosophy as `gofmt`: remove the room for argument.
</details>

---

### 3. Difference between `var x = 5` and `x := 5`? Where can each be used?

<details><summary>Answer</summary>

Both declare and infer the type. `:=` (short declaration) is **only valid inside a function** and requires at least one new variable on the left. `var` works at **package scope and function scope**, and lets you declare without initializing (`var x int`) or pin an explicit type (`var x int64 = 5`). Use `var` for package-level state and zero-value declarations; use `:=` for the common in-function case.
</details>

---

### 4. What does `iota` do, and why does Go have no `enum` keyword?

<details><summary>Answer</summary>

`iota` is a per-`const`-block counter that starts at 0 and increments by one per `ConstSpec` line, letting you generate sequential constants. Go has no `enum` keyword because typed constants plus `iota` cover the need without a new language feature — keeping the language small is a core Go value.

```go
type Weekday int
const (
    Sunday Weekday = iota // 0
    Monday                // 1
    Tuesday               // 2
)
```
The tradeoff: Go enums aren't exhaustive-checked and aren't a closed set, so an arbitrary `Weekday(99)` is legal — validate at boundaries.
</details>

---

### 5. Explain `defer` execution order. What does this print?

```go
defer fmt.Println("A")
defer fmt.Println("B")
fmt.Println("C")
```

<details><summary>Answer</summary>

Deferred calls run **LIFO** (stack order) when the surrounding function returns. So this prints `C`, then `B`, then `A`. LIFO is exactly right for cleanup: you release resources in reverse order of acquisition (close the file you opened last, first). Note the **arguments are evaluated at the `defer` statement**, not when it runs — `defer fmt.Println(i)` captures `i`'s current value, while `defer func(){ fmt.Println(i) }()` reads `i` at return time.
</details>

---

### 6. What is variable shadowing and why is it dangerous?

<details><summary>Answer</summary>

Shadowing is declaring a new variable with `:=` in an inner scope that reuses an outer variable's name; the inner one hides the outer for that block. It's dangerous with `err`: a common bug is `if x, err := f(); err == nil { ... }` where the inner `err` shadows an outer one you meant to set, so the outer stays `nil` and a failure slips through. `go vet -vettool=shadow` and careful `=` vs `:=` discipline catch it.
</details>

---

### 7. Why does Go use multiple return values / returned errors instead of exceptions?

<details><summary>Answer</summary>

Errors are values, so the control flow is explicit and local — you can see at every call site whether failure is handled, and the type checker reminds you. Exceptions create invisible non-local jumps that are easy to forget and hard to reason about in concurrent code. You gain explicitness and composability (`errors.Is/As`, wrapping); you lose brevity (the `if err != nil` boilerplate) and you can't "bubble up" automatically. `panic`/`recover` exists for truly exceptional, programmer-error situations — not routine failure.
</details>

---

### 8. What are the three internal fields of a slice, and what happens on `append` with vs without spare capacity?

<details><summary>Answer</summary>

A slice is a 3-word header: **pointer** to the backing array, **len**, and **cap**. With spare capacity (`len < cap`), `append` writes into the existing array in place and returns a header with `len+1` — fast, no allocation, and it **mutates shared backing memory**. Without spare capacity, `append` allocates a **new, larger** array (growth is roughly ~2× for small slices, tapering toward ~1.25× for large ones), copies, and the new slice no longer aliases the old backing array. That capacity-dependent aliasing is the source of most slice bugs.
</details>

---

### 9. Show the slice aliasing gotcha: when is `b := a[1:3]` dangerous?

<details><summary>Answer</summary>

`b` shares `a`'s backing array, so writes through `b` corrupt `a` — and an `append` to `b` that fits in capacity overwrites `a`'s later elements:

```go
a := []int{1, 2, 3, 4, 5}
b := a[1:3]        // len 2, cap 4 — still pointing into a
b = append(b, 99)  // fits in cap → overwrites a[3]!
// a is now [1 2 3 99 5]
```
Defenses: a full three-index slice `a[1:3:3]` caps `b` so any append reallocates, or copy with `append([]int{}, a[1:3]...)`.
</details>

---

### 10. Why does Go randomize map iteration order? What if you need deterministic order?

<details><summary>Answer</summary>

The runtime deliberately randomizes the starting bucket on each `range` so nobody writes code that accidentally depends on iteration order — a guarantee Go never made and one that would freeze the map's internal layout. If you need order, collect the keys into a slice and `sort.Strings(keys)`, then iterate the slice. Determinism is your job, not the map's.
</details>

---

### 11. How do you implement a set in Go idiomatically, and why that design?

<details><summary>Answer</summary>

`map[string]struct{}`. The `struct{}` value is **zero-width** — it allocates no space for the value, so you're paying only for keys. Use `_, ok := set[k]` to test membership and `set[k] = struct{}{}` to add. `map[string]bool` also works and reads slightly cleaner, but it stores a byte per entry and invites the "is `false` absent or present?" ambiguity; `struct{}` makes "present" the only meaningful state.
</details>

---

### 12. `make([]int, 0, 1000)` vs `var s []int` — what's the allocation difference when appending 1000 items?

<details><summary>Answer</summary>

`make([]int, 0, 1000)` pre-allocates a backing array with capacity 1000, so all 1000 appends hit spare capacity — **one allocation total**. `var s []int` starts nil with cap 0, so it grows by reallocation-and-copy multiple times (roughly log₂(1000) ≈ 10 growth events, each copying the prior contents). When you know the size, preallocating cap is a free, large win — fewer allocations and no repeated copies.
</details>

---

### 13. What does the three-index slice `s[1:3:3]` do, and when do you use it?

<details><summary>Answer</summary>

`s[low:high:max]` sets `len = high-low` and `cap = max-low`. So `s[1:3:3]` gives len 2, cap 2 — meaning the **next append must reallocate** instead of writing into `s`'s tail. You use it when handing a sub-slice to code that may append, to guarantee it can't clobber the parent's backing array. It's the safe way to share a window of a slice.
</details>

---

### 14. Value receiver vs pointer receiver — what's the difference and when do you choose each?

<details><summary>Answer</summary>

A **value receiver** operates on a copy, so it can't mutate the original and is cheap only for small types. A **pointer receiver** shares the original, can mutate it, and avoids copying large structs. Rule of thumb: use a pointer receiver if the method mutates the receiver, if the struct is large, or if the type holds a `sync.Mutex` (copying a mutex is a bug). For consistency, if any method needs a pointer receiver, give them **all** pointer receivers so the method set stays uniform.
</details>

---

### 15. Can you return the address of a local variable in Go? Why is this safe (unlike C)?

<details><summary>Answer</summary>

Yes, and it's safe. Go's **escape analysis** detects that the variable's address outlives the function, so the compiler allocates it on the **heap** instead of the stack; the GC keeps it alive as long as a reference exists. In C the same code returns a dangling pointer to a reclaimed stack frame. You don't manage the lifetime — the compiler and GC do.

```go
func newCounter() *int { c := 0; return &c } // c escapes to the heap — fine
```
</details>

---

### 16. Explain the Go 1.22 loop-variable fix. Show the old bug and the workaround.

<details><summary>Answer</summary>

Before 1.22, the loop variable was **one variable reused** across iterations, so closures/goroutines capturing it all saw the final value:

```go
for _, v := range items {
    go func() { fmt.Println(v) }() // pre-1.22: often prints last item N times
}
```
The pre-1.22 fix was to shadow per iteration: `v := v` inside the loop. In **Go 1.22+** the loop variable is scoped **per iteration**, so the bug is gone for `for ... range` and three-clause `for`. You still must understand it for older code and to grasp closure capture in general.
</details>

---

### 17. Why does `for _, v := range s { v = 99 }` not modify the slice? How do you modify in place?

<details><summary>Answer</summary>

`v` is a **copy** of each element, so assigning to it changes only the copy. To mutate the slice, index it: `for i := range s { s[i] = 99 }` (or `s[i] *= 2`). The `range` two-value form is read-only over copies by design; the index form gives you the addressable element.
</details>

---

### 18. Difference between `%w` and `%v` in `fmt.Errorf`? When deliberately choose `%v`?

<details><summary>Answer</summary>

`%w` **wraps** the error — it embeds it in a chain so `errors.Is`/`errors.As` can later unwrap and match it. `%v` just formats the error's text, breaking the chain. You choose `%v` deliberately when you want to **hide the underlying error from callers** — e.g., not leaking that "user not found" was really a Postgres `ErrNoRows`, or not exposing an internal error type as part of your package's API contract. Wrapping is the default; `%v` is the conscious decision to stop the chain.
</details>

---

### 19. Why `errors.Is(err, ErrNotFound)` instead of `err == ErrNotFound`? And `Is` vs `As`?

<details><summary>Answer</summary>

`err == ErrNotFound` only matches if `err` *is literally* that sentinel — it fails the moment someone wraps it with `%w`. `errors.Is` walks the **wrap chain** and returns true if any layer matches, so it survives wrapping. Use **`errors.Is`** to test for a specific **sentinel value** (`ErrNotFound`); use **`errors.As`** to extract a specific **error type** so you can read its fields:

```go
var verr *ValidationError
if errors.As(err, &verr) { log.Println(verr.Field) } // need the data
if errors.Is(err, ErrNotFound) { return 404 }        // just need identity
```
</details>

---

### 20. When is `panic` appropriate, and where is `recover` actually used in production?

<details><summary>Answer</summary>

`panic` is for **programmer errors and truly unrecoverable invariants** — a nil pointer that should never be nil, a corrupt internal state, impossible `default` cases — not for routine failures like "file missing." `recover` (only effective inside a deferred function) is used at **trust boundaries**: an HTTP server's middleware recovers a per-request panic so one bad handler doesn't crash the whole process, logs it, and returns 500. A library generally should not swallow panics silently.
</details>

---

### 21. A function logs an error *and* returns it. Why is that a smell?

<details><summary>Answer</summary>

It produces **double-reporting**: the error gets logged at every layer it passes through, so one failure becomes five log lines and noise drowns signal. The rule is **handle an error exactly once** — either log it (because you're handling it and stopping it here) **or** return it (because the caller will decide), not both. Decide at the boundary that has the context to act; everyone below just wraps-and-returns.
</details>

---

### 22. What's wrong with `defer f.Close()` on a file you *wrote* to?

<details><summary>Answer</summary>

`Close` can return an error — for a written file that error may be a **flush failure**, meaning your data never hit disk — and `defer f.Close()` silently discards it. For write paths, close explicitly and check the error (and `Sync` if durability matters):

```go
if err := f.Close(); err != nil { return fmt.Errorf("close: %w", err) }
```
`defer f.Close()` is fine for read-only files where a close error is harmless.
</details>

---

### 23. When `os.ReadFile` vs `bufio.Scanner`? What goes wrong if you pick wrong?

<details><summary>Answer</summary>

`os.ReadFile` slurps the **whole file into memory** — simple and fine for small/config files, but it'll exhaust RAM on a multi-GB log. `bufio.Scanner` streams line by line with bounded memory — right for large or unbounded input. Pick `ReadFile` for big files and you OOM; pick `Scanner` and forget that its default token limit is 64KB, and you silently truncate long lines (raise it with `scanner.Buffer`).
</details>

---

### 24. Why design a function to take `io.Reader` instead of a filename or `*os.File`?

<details><summary>Answer</summary>

`io.Reader` is the smallest abstraction over "a stream of bytes," so the function works with a file, a network connection, an HTTP body, a `bytes.Buffer`, or a `strings.Reader` — **without change**. That makes it trivially testable (`strings.NewReader("...")` in a unit test, no temp files) and composable with the whole `io` ecosystem (`io.Copy`, `gzip.NewReader`, etc.). "Accept the narrowest interface that does the job" is core Go design.
</details>

---

### 25. Why is a struct field missing from JSON output? Name three causes. And how do you distinguish "absent" from "present-but-zero"?

<details><summary>Answer</summary>

Three classic causes: (1) the field is **unexported** (lowercase) so `encoding/json` can't see it; (2) it has `json:"-"` which excludes it; (3) it has `omitempty` and holds a zero value, so it's dropped. To tell **absent** from **present-but-zero** on decode, use a **pointer** field (`*int`): `nil` means the key was absent, a non-nil pointer to `0` means it was explicitly present and zero. (`json.RawMessage` or a custom `UnmarshalJSON` are heavier alternatives.) Also note: JSON numbers decoded into `interface{}` become **`float64`**, which can surprise you with large integers.
</details>
