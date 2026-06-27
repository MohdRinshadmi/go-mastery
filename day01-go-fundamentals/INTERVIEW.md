# Day 01 Interview Questions — Fundamentals

Self-quiz: read the question, answer out loud, then expand the solution.

---

**1. What is the zero value of a `string`, a `slice`, and a `map`? Which are safe to use immediately, and which will panic?**

<details><summary>Answer</summary>

`string` → `""` (empty string), `slice` → `nil`, `map` → `nil`. The empty string
and a `nil` slice are immediately usable: you can range over a nil slice (zero
iterations) and `append` to it. A `nil` map is safe to *read* (returns the zero
value) but **panics on write** — you must `make` it first. So: string and slice
safe; nil map safe to read, panics on write.
</details>

---

**2. Why are unused variables a compile error in Go (not a warning)?**

<details><summary>Answer</summary>

Because an unused variable is almost always either dead code or a bug (you
computed something and forgot to use it). Go's designers chose to make the
codebase self-cleaning: the compiler forces you to remove cruft immediately
rather than letting warnings accumulate and get ignored. The same applies to
unused imports.
</details>

---

**3. Difference between `var x = 5` and `x := 5`? Where can you use each?**

<details><summary>Answer</summary>

Both infer the type as `int`. `var x = 5` is a declaration that works at both
package scope and inside functions. `x := 5` is short variable declaration and is
**only** legal inside a function body. `:=` also requires at least one new
variable on its left-hand side.
</details>

---

**4. What does `iota` do, and why doesn't Go have an `enum` keyword?**

<details><summary>Answer</summary>

`iota` is a per-`const`-block counter that starts at 0 and increments by 1 for
each ConstSpec line. It's Go's idiomatic way to build enumerated constants:
`const ( Sunday = iota; Monday; Tuesday )` gives 0, 1, 2. Go favours a small set
of orthogonal primitives over many keywords; `iota` plus a named integer type
covers the enum use case without a dedicated keyword.
</details>

---

**5. Explain `defer` execution order. What does this print?**
```go
defer fmt.Println("A")
defer fmt.Println("B")
fmt.Println("C")
```

<details><summary>Answer</summary>

Deferred calls run in **LIFO** (last-in, first-out) order when the surrounding
function returns. So the prints happen as `C` (immediate), then `B`, then `A`.
Output:
```
C
B
A
```
</details>

---

**6. What is variable shadowing and why is it dangerous?**

<details><summary>Answer</summary>

Shadowing is declaring a new variable in an inner scope with the same name as one
in an outer scope, so the inner name "hides" the outer for the rest of that
block. It's dangerous because you often *think* you're updating the outer
variable (e.g. `err`) but you've created a new one — the outer stays at its old
value, so checks downstream see stale data. `go vet -vet=shadow` helps catch it.
</details>

---

**7. Why does Go use multiple return values for errors instead of exceptions?**

<details><summary>Answer</summary>

Returning `(result, error)` makes every failure path explicit and local: you can
see what can fail just by reading the function top-to-bottom, with no invisible
control flow leaping out of a deeply nested call. You gain honesty and
readability; you pay with some verbosity (`if err != nil` repeated). Go bets that
explicitness is worth it at scale.
</details>

---

**8. When would you reach for `var x Type` instead of `x := ...`?**

<details><summary>Answer</summary>

When the **zero value is exactly what you want**. Idioms like
`var buf bytes.Buffer`, `var sb strings.Builder`, `var mu sync.Mutex` are ready
to use at their zero value — writing `bytes.Buffer{}` is noisier. Also use `var`
when you need package-level declarations (where `:=` isn't allowed) or when you
want the type written out for clarity.
</details>

---

**9. What's the difference between a typed and an untyped constant?**

<details><summary>Answer</summary>

An untyped constant (`const n = 3`) has a *default* type but adapts to whatever
context uses it — it can be assigned to an `int`, `int64`, or `float64` without an
explicit conversion. A typed constant (`const n int = 3`) is locked to that type
and won't implicitly convert. Prefer untyped unless you intend to restrict usage.
</details>

---

**10. What's wrong with `defer f.Close()` inside a `for` loop?**

<details><summary>Answer</summary>

`defer` runs when the *function* returns, not at the end of each loop iteration.
In a long loop the deferred `Close()` calls pile up and the file descriptors stay
open until the whole function exits — a resource leak. Extract the body into a
helper function (so each `defer` runs per call) or close explicitly each
iteration.
</details>

---

**11. Why is the module path usually a repo URL like `github.com/you/app`?**

<details><summary>Answer</summary>

The module path is also the **import prefix**. Naming it after the real repo
location means the module can be published and imported by others cleanly
(`github.com/you/app/internal/auth`). Naming it `myapp` works locally but can't
be fetched or imported by another project without renaming — a breaking change.
</details>
