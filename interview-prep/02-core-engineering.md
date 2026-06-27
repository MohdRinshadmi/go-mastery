# Phase 2 — Core Engineering (Days 06–10)

Methods & method sets, interfaces, composition & DI, generics, testing & mocking, benchmarking & profiling. Self-quiz: answer aloud, then expand.

---

### 1. Explain Go's implicit interface satisfaction. How does it differ from Java's `implements`?

<details><summary>Answer</summary>

A type satisfies an interface simply by **having the right methods** — there's no `implements` keyword and no declared relationship. This is **structural, not nominal**, typing: the interface belongs to the *consumer*, not the implementer. So you can define an interface in your package that an existing third-party type already satisfies, decoupling the two. Java requires the implementer to know about the interface up front; Go inverts that — the abstraction is declared where it's needed.
</details>

---

### 2. What is the method set of `T` vs `*T`? Give an example where it matters.

<details><summary>Answer</summary>

The method set of `T` includes only its **value-receiver** methods; the method set of `*T` includes **both** value- and pointer-receiver methods. This matters for interface satisfaction: if `Save()` has a pointer receiver, then `*T` satisfies the interface but `T` does **not**.

```go
type Saver interface{ Save() }
func (t *Thing) Save() {}
var s Saver = &Thing{} // ok
var s Saver = Thing{}  // compile error: Thing has no Save in its method set
```
You can call pointer methods on an addressable value directly (`t.Save()` auto-takes `&t`), but interface assignment is stricter.
</details>

---

### 3. Why does Go favor small interfaces? What's wrong with a 10-method interface?

<details><summary>Answer</summary>

Small interfaces (`io.Reader`, `io.Writer` — one method each) are easy to implement, easy to mock, and easy to compose. A 10-method interface forces every implementer and every fake to supply all 10, couples consumers to behavior they don't use, and is hard to satisfy with adapters. "The bigger the interface, the weaker the abstraction." Define interfaces at the **point of use** with only the methods that call site needs, then compose them.
</details>

---

### 4. Explain the nil interface gotcha. What does `(type, value)` mean internally?

<details><summary>Answer</summary>

An interface value is a **two-word pair: (type, value)**. It equals `nil` only when **both** words are nil. The trap: returning a typed nil pointer as an interface gives a non-nil interface (type is set, value is nil), so `== nil` is false:

```go
func get() error {
    var e *MyError = nil
    return e // interface = (*MyError, nil) — NOT nil!
}
if get() != nil { /* this runs, surprising everyone */ }
```
Fix: return a literal `nil` for the success path (`return nil`), never a typed nil pointer dressed as an interface.
</details>

---

### 5. When is a type assertion safe vs risky? What's the alternative?

<details><summary>Answer</summary>

The single-value form `v := i.(T)` **panics** if `i` doesn't hold a `T` — risky. The comma-ok form `v, ok := i.(T)` is safe: `ok` is false instead of panicking. For dispatch over several possible types, use a **type switch** (`switch v := i.(type)`), where `v` is typed as the matched case's type in each branch. Prefer comma-ok / type switch; reserve the panicking form for cases where a wrong type is a genuine bug you *want* to crash on.
</details>

---

### 6. Explain struct embedding. Is it inheritance?

<details><summary>Answer</summary>

Embedding puts one type inside another **without a field name**, and the outer type **promotes** the embedded type's fields and methods so you can call them directly. It is **composition, not inheritance**: there's no subtype polymorphism, no `super`, no overriding in the OOP sense (the outer type can *shadow* a method, but dispatch is static, not virtual). The embedded value doesn't know it's embedded. It's "has-a wearing the syntax of is-a."
</details>

---

### 7. What does "accept interfaces, return structs" mean, and why?

<details><summary>Answer</summary>

**Accept interfaces** as parameters so callers can pass any implementation (real, fake, decorated) — maximum flexibility and testability at the boundary. **Return concrete structs** so callers get the full, documented type with all its methods and fields, and so you don't prematurely lock your API behind a narrow abstraction. Returning an interface hides capabilities and forces type assertions to get them back. The asymmetry maximizes both flexibility (in) and clarity (out).
</details>

---

### 8. How do you test a service that depends on a database without hitting a real DB? What is the composition root?

<details><summary>Answer</summary>

Depend on a **`Store` interface**, not a concrete `*sql.DB`; in tests inject a hand-written fake or in-memory implementation. Because Go interfaces are implicit, the fake just needs the right methods — no mocking framework required. The **composition root** is the single place (usually `main`) where you construct the concrete dependencies and wire them together; everything below receives interfaces via constructor injection, so the wiring decision lives in exactly one spot and the business logic stays ignorant of which implementation it got.
</details>

---

### 9. What problem does the functional options pattern solve? When use it?

<details><summary>Answer</summary>

It gives you **extensible, defaulted, optional** configuration without breaking the API every time you add a knob. Instead of a giant constructor or a config struct, you pass variadic `Option` functions that mutate the config:

```go
func NewClient(opts ...Option) *Client {
    c := &Client{timeout: 30 * time.Second} // sane defaults
    for _, opt := range opts { opt(c) }
    return c
}
client := NewClient(WithTimeout(5*time.Second), WithRetries(3))
```
Use it when a type has several optional settings and you want backward-compatible evolution. For 1–2 required params, a plain constructor is simpler — don't over-engineer.
</details>

---

### 10. How do you implement a `CachedStore` that wraps any `Store` without generics?

<details><summary>Answer</summary>

The **decorator pattern**: `CachedStore` holds a `Store` interface and *is itself* a `Store`, so it composes transparently.

```go
type CachedStore struct {
    inner Store
    cache map[string]Value
}
func (c *CachedStore) Get(k string) (Value, bool) {
    if v, ok := c.cache[k]; ok { return v, true }
    v, ok := c.inner.Get(k)
    if ok { c.cache[k] = v }
    return v, ok
}
```
Because it satisfies the same interface, you can stack decorators (caching → logging → real store) freely. No generics needed — the interface *is* the abstraction.
</details>

---

### 11. How do you prevent circular imports in a layered Go architecture?

<details><summary>Answer</summary>

Make dependencies point **inward, one direction only**: transport → service → domain, and the **domain imports nothing** from the outer layers. Define interfaces in the layer that *consumes* them (e.g., the service declares the `Repository` interface it needs), so the concrete repository depends on the service's abstraction rather than the reverse. If two packages genuinely need each other, that's a design smell — extract the shared types into a third, lower package. Go's compiler hard-fails on cycles, which forces you to fix the layering.
</details>

---

### 12. What problem do generics solve? Give a concrete before/after.

<details><summary>Answer</summary>

They give **type-safe code reuse** over many types without `interface{}` boxing and runtime assertions. Before generics, a generic `Map` returned `[]interface{}` and forced casts (and lost type safety); after:

```go
func Map[T, U any](s []T, f func(T) U) []U {
    out := make([]U, len(s))
    for i, v := range s { out[i] = f(v) }
    return out
}
lengths := Map([]string{"a","bb"}, func(s string) int { return len(s) }) // []int, checked
```
The compiler verifies types and there's no boxing. The win is biggest for container types and algorithms.
</details>

---

### 13. Difference between `any`, `comparable`, and `constraints.Ordered`? What does `~int` mean?

<details><summary>Answer</summary>

`any` (alias for `interface{}`) allows **any** type — but you can do almost nothing with the value except pass it around. `comparable` permits `==`/`!=`, so it's the constraint for map keys and set members. `constraints.Ordered` permits `<`, `>`, etc. — needed for `Min`/`Max`/`Sort`. The `~` in `~int` means "any type whose **underlying type** is `int`," so a named type like `type Celsius int` still satisfies the constraint. Without `~`, only the exact `int` type qualifies.
</details>

---

### 14. When use a generic function vs an interface-based one? Can generics replace all interfaces?

<details><summary>Answer</summary>

Use **generics** when the logic is identical across types and you want to **preserve the concrete type** through the call (containers, `Map/Filter/Reduce`, `Min/Max`). Use **interfaces** when you want **runtime polymorphism / dynamic dispatch** — different behaviors behind one contract (an `io.Writer` that could be a file or a socket, chosen at runtime). Generics resolve at compile time and can't hold a heterogeneous collection of different concrete types behind one variable; interfaces can. They're complementary — neither replaces the other.
</details>

---

### 15. How does Go's generics implementation differ from C++ monomorphization? And fix `func Get[T any](m map[string]T, key string) T`.

<details><summary>Answer</summary>

Go uses **GC-shape stenciling**: it generates one instantiation per *memory layout / GC shape* (e.g., all pointer-shaped types share one) rather than one per concrete type like C++ monomorphization. That keeps binary size down at a small dispatch cost, versus C++'s zero-overhead-but-code-bloat approach. The `Get` bug: on a missing key it returns `T`'s zero value, indistinguishable from a present zero — fix by returning `(T, bool)`:

```go
func Get[T any](m map[string]T, key string) (T, bool) { v, ok := m[key]; return v, ok }
```
</details>

---

### 16. What is a table-driven test and why is it the Go idiom? What does `t.Run` add?

<details><summary>Answer</summary>

You define a slice of `struct{ name string; input ...; want ... }` cases and loop over them, running the same assertions for each — adding a case is one line, not a new function. `t.Run(tc.name, func(t *testing.T){...})` wraps each case as a **named subtest**, so failures report *which* case broke, you can run one with `-run TestX/case_name`, and you can mark subtests `t.Parallel()`. It's the idiom because it maximizes coverage-per-line and keeps failures legible.
</details>

---

### 17. `t.Error` vs `t.Fatal`? Why must tests be independent of order?

<details><summary>Answer</summary>

`t.Error` records a failure and **keeps going** (good for checking several independent assertions in one test). `t.Fatal` records and **stops the current test immediately** (use when continuing would panic — e.g., a nil result you're about to dereference). Tests must be order-independent because `go test` may reorder them and `t.Parallel()` runs them concurrently; any shared mutable state or assumed sequencing produces flaky, irreproducible failures. Each test sets up and tears down its own world.
</details>

---

### 18. What does `go test -race` do, and why isn't it a production build flag?

<details><summary>Answer</summary>

It enables the **race detector**, which instruments memory accesses at runtime and reports when two goroutines touch the same location concurrently with at least one write and no synchronization. It finds real data races that are invisible to code review because they depend on timing. It's not for production because it adds **5–10× CPU and large memory overhead** — you run it in CI and during testing, where the cost buys correctness, not in the hot path.
</details>

---

### 19. Is 100% coverage a good goal?

<details><summary>Answer</summary>

No — coverage measures *lines executed*, not *behaviors verified*. You can hit 100% with tests that assert nothing, and you can have rock-solid code at 75%. Chasing the last few percent often means testing trivial getters and error paths that can't realistically fire, at the cost of brittle tests. Aim coverage at **critical paths and tricky logic**; treat the number as a smoke detector for *untested important code*, not a target to max out.
</details>

---

### 20. How do you write and run a benchmark? What do `-benchmem`, ns/op, B/op, allocs/op mean?

<details><summary>Answer</summary>

A `func BenchmarkX(b *testing.B)` loops `b.N` times (the framework tunes `N`); run with `go test -bench=. -benchmem`. `-benchmem` adds memory stats. **ns/op** = nanoseconds per operation (speed), **B/op** = bytes allocated per op, **allocs/op** = number of heap allocations per op. You care about allocs because each one pressures the GC; cutting allocations often improves latency and tail-latency more than micro-optimizing CPU. Use `b.ResetTimer()` after expensive setup and assign results to a package-level sink so the compiler can't optimize the work away.
</details>

---

### 21. What's escape analysis, and how do you see what escapes to the heap? Name common causes of slow Go.

<details><summary>Answer</summary>

Escape analysis is the compiler deciding whether a value can stay on the **stack** (cheap, auto-freed) or must go to the **heap** (GC-managed). Reveal it with `go build -gcflags='-m'`. Common slowness: (1) needless heap allocations (escaping values, growing slices/maps without preallocating cap); (2) excessive interface boxing / reflection; (3) lock contention or oversharing; plus things like per-request JSON re-parsing and unbounded goroutine creation. The meta-rule: **measure before optimizing** — "optimize without measuring" is an anti-pattern because intuition about Go hotspots is usually wrong, and you risk complicating code for no real gain.
</details>

---

### 22. Walk me through the optimize loop. When is `sync.Pool` the right tool?

<details><summary>Answer</summary>

**Benchmark → profile (pprof) → identify the real hotspot → fix one thing → re-benchmark to confirm the gain → repeat.** Never skip the measure step at either end. `sync.Pool` is appropriate when you allocate and discard many **short-lived, reusable** objects of the same type in a hot path (e.g., per-request buffers) and profiling shows allocation pressure — it amortizes allocations across requests. It's a mistake when objects are long-lived, rarely reused, or hold state you forget to reset (the pool can return a dirty object); and note pool contents can be GC'd at any time, so never assume an entry persists.
</details>
