# Day 08 — Generics: Write Once, Use With Any Type

> Mentor note: Generics landed in Go 1.18 (March 2022), after years of careful design. The Go team delayed until they had a design that felt *Go-like*: simple, readable, no surprises. If you've used generics in TypeScript, Java, or Rust, you'll notice Go's version is intentionally more constrained. That's a feature. Today you'll learn when they're the right tool and — just as importantly — when they're overkill.

---

## 0. The Problem Generics Solve

Before Go 1.18:

```go
// Had to write a separate function for every type
func SumInts(s []int) int { ... }
func SumFloat64s(s []float64) float64 { ... }

// Or use any — and lose all type safety
func Sum(s []interface{}) interface{} { ... }
```

The first approach is boilerplate. The second throws away compile-time safety. Generics give you: **write once, typed at the call site.**

---

## 1. Type Parameters — the syntax

```go
// T is a type parameter — constrained to any type
func Map[T, U any](slice []T, fn func(T) U) []U {
    result := make([]U, len(slice))
    for i, v := range slice {
        result[i] = fn(v)
    }
    return result
}

// Call site — compiler infers the types
doubled := Map([]int{1, 2, 3}, func(x int) int { return x * 2 })
// doubled: []int{2, 4, 6}

lengths := Map([]string{"go", "rust", "java"}, func(s string) int { return len(s) })
// lengths: []int{2, 4, 4}
```

The `[T, U any]` part is the **type parameter list**. Square brackets, come before the regular parameters. Each parameter has a **constraint** — here `any` means "any type at all."

---

## 2. Constraints — what you can do with T

A constraint is an interface that specifies what operations are allowed on T.

### `any` — no restrictions
```go
func First[T any](slice []T) (T, bool) {
    if len(slice) == 0 {
        var zero T
        return zero, false
    }
    return slice[0], true
}
```
With `any`, you can only do things that work on every type: assign, pass around, return. You cannot `+`, `-`, `<`, `>`, or compare with `==`.

### `comparable` — supports == and !=
```go
func Contains[T comparable](slice []T, item T) bool {
    for _, v := range slice {
        if v == item {
            return true
        }
    }
    return false
}

Contains([]int{1, 2, 3}, 2)          // true
Contains([]string{"a", "b"}, "c")    // false
```

**Note:** slices, maps, and functions are NOT comparable — you can't use `==` on them. So `Contains[[]int]` won't compile.

### `constraints.Ordered` — supports < > <= >=
```go
import "golang.org/x/exp/constraints"

func Min[T constraints.Ordered](a, b T) T {
    if a < b {
        return a
    }
    return b
}

Min(3, 5)         // 3  (int)
Min(3.14, 2.71)   // 2.71 (float64)
Min("apple", "banana") // "apple" (string)
```

### Custom constraints with `~` (underlying type)
```go
// Integer matches int, int8, int16, int32, int64
type Integer interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64
}

func Abs[T Integer](x T) T {
    if x < 0 {
        return -x
    }
    return x
}

type MyInt int
Abs(MyInt(-5)) // works — ~int matches MyInt's underlying type
```

The `~int` syntax means "any type whose underlying type is int" — including named types like `type UserID int`.

---

## 3. Generic Types (Structs)

```go
// A type-safe stack — no interface{} required.
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(item T)     { s.items = append(s.items, item) }
func (s *Stack[T]) IsEmpty() bool   { return len(s.items) == 0 }

func (s *Stack[T]) Pop() (T, bool) {
    if s.IsEmpty() {
        var zero T
        return zero, false
    }
    n := len(s.items) - 1
    item := s.items[n]
    s.items = s.items[:n]
    return item, true
}

// Usage — type is inferred
var intStack Stack[int]
intStack.Push(1)
intStack.Push(2)
v, _ := intStack.Pop() // v is int, no type assertion needed
```

Notice: methods on generic types include the type parameter in the receiver: `(s *Stack[T])`.

---

## 4. The `constraints` Package

Go's standard library has `golang.org/x/exp/constraints` (experimental) and Go 1.21 added `slices` and `maps` packages with generic functions. The key predefined constraints:

```go
constraints.Ordered    // types supporting < > <= >= == (int, float, string, etc.)
constraints.Integer    // signed and unsigned integer types
constraints.Float      // floating point types
constraints.Complex    // complex number types
constraints.Signed     // signed integers only
constraints.Unsigned   // unsigned integers only
```

As of Go 1.21, the standard library also has:
```go
import "slices"
slices.Contains([]int{1,2,3}, 2)          // true
slices.Max([]int{3,1,4,1,5})              // 5
slices.Sort([]string{"banana", "apple"})  // sorts in place

import "maps"
maps.Keys(map[string]int{"a": 1, "b": 2}) // []string{"a", "b"} (order undefined)
```

---

## 5. Multiple Type Parameters

```go
// Pair: two types, possibly different.
type Pair[A, B any] struct {
    First  A
    Second B
}

func Zip[A, B any](as []A, bs []B) []Pair[A, B] {
    n := len(as)
    if len(bs) < n {
        n = len(bs)
    }
    result := make([]Pair[A, B], n)
    for i := 0; i < n; i++ {
        result[i] = Pair[A, B]{as[i], bs[i]}
    }
    return result
}

zipped := Zip([]string{"a", "b"}, []int{1, 2})
// []Pair[string, int]{{First:"a", Second:1}, {First:"b", Second:2}}
```

---

## 6. When Generics HELP vs When They HURT

This is the most important judgment call.

### Use generics when:
- You're writing a data structure (stack, queue, set, cache) that should work with any type.
- You have the *exact same algorithm* applied to multiple types and you'd otherwise copy-paste.
- The standard library equivalent doesn't exist yet (pre-1.21: `slices.Sort` etc.).
- You want to express a constraint precisely (e.g., "any numeric type").

### Do NOT use generics when:
- You have 1-2 concrete types. Just write two functions. It's clearer.
- The `any`-constrained generic adds no type safety benefit over an interface approach.
- You want to dispatch differently based on type — use a type switch or interface instead.
- The constraint is more complex than the benefit. If you're fighting the type system, interfaces are probably better.
- Readability suffers. `func Process[T Processor[V], V any, R Result[V]](...)` — stop, go back to interfaces.

**The Go team's own guidance (Rob Pike, Ian Lance Taylor):**
> "If you find yourself writing the same code 3+ times with different types and no behavioral difference, generics are the right tool."

### Real example: don't over-generic
```go
// BAD — pointless generic
func ToString[T any](v T) string {
    return fmt.Sprintf("%v", v)
}

// GOOD — just use fmt.Sprintf directly, or accept any
func ToString(v any) string {
    return fmt.Sprintf("%v", v)
}
```

```go
// ALSO BAD — only one concrete type in practice
func Process[T OrderProcessor](p T) error { ... }

// GOOD — just use the interface directly
func Process(p OrderProcessor) error { ... }
```

---

## 7. Performance: Monomorphization and GC Shape Stenciling

This is a detail most Go devs don't need daily, but good to know:

**C++ / Rust generics:** full monomorphization — the compiler generates a specialized version of every function for every type argument. Fast, but large binary.

**Go generics:** uses a hybrid called "GC shape stenciling":
- All pointer types share one instantiation (because they have the same GC shape).
- Primitive types (int, float64, etc.) get their own instantiation.
- This means Go generics don't necessarily generate as many copies as C++ templates.

**Practical implication:**
- Generic code over concrete scalars (int, float) will be roughly as fast as direct code.
- Generic code over interface-constrained types may be slightly slower due to indirect dispatch.
- For most use cases: don't worry about this. Profile before optimizing (Day 10).

---

## 8. Type Inference

Go's compiler infers type arguments at the call site when possible:

```go
func Map[T, U any](s []T, fn func(T) U) []U { ... }

// ✓ Inferred — compiler sees []int and func(int)int → T=int, U=int
doubled := Map([]int{1, 2, 3}, func(x int) int { return x * 2 })

// ✓ Also inferred
lengths := Map([]string{"a", "bb"}, func(s string) int { return len(s) })

// ✗ Sometimes you must be explicit — when inference is ambiguous
var result []float64 = Map[int, float64]([]int{1, 2}, func(x int) float64 { return float64(x) * 1.5 })
```

---

## Common mistakes

1. **Using `any` when you actually need an operation.** With `[T any]` you can only assign, pass, and return `T` — you cannot `==`, `<`, or `+`. If your function compares or orders values, the constraint must be `comparable` or `constraints.Ordered`. Reaching for `any` and then casting back defeats the entire point of generics.
2. **Assuming `comparable` covers everything.** Slices, maps, and functions are not comparable, so `Contains[[]int]` won't compile. And `comparable` permits `==` but NOT `<` — ordering needs `constraints.Ordered`. Pick the *narrowest* constraint that supports the operations you actually use.
3. **Forgetting `~` in custom constraints.** `interface { int | int32 }` matches *only* `int` and `int32` — a named type `type UserID int` will NOT satisfy it. Write `~int | ~int32` so the constraint matches any type whose *underlying* type is one of those. Omitting `~` is the most common "why won't my named type work" bug.
4. **Returning a bare zero with `T` instead of the comma-ok form.** `func Get[T any](m map[string]T, key string) T` can't signal "missing" — the zero value of `T` is ambiguous (is `0`/`""`/`nil` a real value or a miss?). Return `(T, bool)` so the caller can tell.
5. **Reaching for generics with only one or two concrete types.** If a function is only ever called with `int`, just write it for `int`; if there are two types with the *same* behavior, two small functions are often clearer than a generic. Generics earn their keep at 3+ types or for reusable data structures — not as a reflex.
6. **Generics in public APIs that should be interfaces.** A signature like `func Process[T Processor[V], V any](...)` is hard to read, hard to mock, and hard to evolve. Generics are an implementation tool; export interfaces or concrete types and keep the type-parameter machinery inside.

---

## Expert Thinking Mode

- **Beginner:** "Generics let me write one function that works for int and string."
- **Senior:** "I reach for generics when I have duplicate code that differs only in type and has identical behavior. I define tight constraints — not `any` — to make misuse a compile error."
- **Staff:** "I standardize our team's generic utility types (Result[T], Option[T], Paginated[T]) to avoid everyone inventing them differently. But service layer code uses interfaces, not generics — the behaviors differ."
- **Architect:** "Generics are a implementation-detail tool, not an API design tool. Public package APIs should almost always use interfaces or concrete types — generics in APIs make them harder to understand, harder to mock, and harder to evolve."

---

## Real-world use

- **Go standard library (1.21+):** `slices`, `maps`, `cmp` packages are entirely generic. `slices.Sort`, `slices.Contains`, `maps.Keys` — code you'd have copy-pasted before.
- **Hashicorp:** Uses generic cache and container types internally to avoid the `interface{}` + type-assertion soup that pre-1.18 code required.
- **Google's internal Go:** `Result[T]` and `Option[T]` types via generics, replacing the `(T, error)` tuple pattern in some contexts for ergonomics.
- **samber/lo library:** A lodash-like library for Go, entirely generic. `lo.Map`, `lo.Filter`, `lo.Reduce`, `lo.Uniq` — great reference for idiomatic generic functions.

---

## Interview Questions

1. What problem do generics solve? Give a concrete before/after example.
2. What is the difference between `any`, `comparable`, and `constraints.Ordered` as constraints?
3. What does `~int` mean in a constraint? Why is it needed?
4. When should you use a generic function vs an interface-based function?
5. Can you use generics to replace all interface usage? Why or why not?
6. How does Go's generic implementation (GC shape stenciling) differ from C++'s monomorphization?
7. A junior writes `func Get[T any](m map[string]T, key string) T` — what's wrong, and how do you fix it?

---

## Your tasks for today

Go to `../exercises/`. Implement a generic Set type, a generic Result type, generic Map/Filter/Reduce functions, and a challenge where you build a type-safe generic event bus. Try everything before opening `../solutions/`.

## Day 08 companion files

- [Debugging challenge](../debugging/README.md) — the bare-zero generic lookup bug.
- [Pitfalls](../PITFALLS.md) — Trap → Why → Fix.
- [Interview questions](../INTERVIEW.md) — with model answers.
- [Notes / cheatsheet](../NOTES.md) — quick reference.
- [Resources](../RESOURCES.md) — curated links.
