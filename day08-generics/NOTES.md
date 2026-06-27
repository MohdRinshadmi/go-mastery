# Day 08 — Generics Cheatsheet

Quick reference. Standard library only (`cmp`, `slices`, `maps`) — no `x/exp`.

---

## Type parameter syntax

```go
func Name[T any](x T) T { ... }          // one type param
func Map[T, U any](s []T, f func(T) U) []U  // two params, same constraint
func Sum[T cmp.Ordered](xs []T) T { ... }   // constrained
```

`[T any]` is the **type parameter list** — square brackets, before the value params.
Each name (`T`, `U`) gets a **constraint** (`any`, `comparable`, `cmp.Ordered`, custom).

---

## Constraint kinds

| Constraint        | Type set                                   | Operations allowed              | Use for                          |
|-------------------|--------------------------------------------|---------------------------------|----------------------------------|
| `any`             | every type                                 | assign, pass, return only       | containers, pass-through utils   |
| `comparable`      | types usable with `==` / `!=`              | `==`, `!=`                      | sets, dedup, map keys            |
| `cmp.Ordered`     | ints, floats, strings                      | `< <= > >=`, `==`               | min/max, sort, ordering          |
| custom with `~`   | whatever you list, by underlying type      | depends on what's in the set    | "any numeric type", named types  |

```go
// custom constraint — note the ~ (underlying type)
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}
```

Without `~`, `int | int32` matches *only* those exact predeclared types — a defined
`type UserID int` would NOT qualify. With `~int`, it does.

---

## Generic struct + method receiver

```go
type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T)   { s.items = append(s.items, v) }
func (s *Stack[T]) IsEmpty() bool { return len(s.items) == 0 }

func (s *Stack[T]) Pop() (T, bool) {  // comma-ok, never a bare T
	if s.IsEmpty() {
		var zero T
		return zero, false
	}
	n := len(s.items) - 1
	v := s.items[n]
	s.items = s.items[:n]
	return v, true
}

var st Stack[int]   // instantiation: Stack[int] is a concrete type
st.Push(1)
v, ok := st.Pop()   // v is int, no assertion
```

- The receiver carries the type parameter: `(s *Stack[T])`.
- Methods may NOT add their own new type parameters — only the receiver's.

---

## Multiple type parameters

```go
type Pair[A, B any] struct {
	First  A
	Second B
}

func Zip[A, B any](as []A, bs []B) []Pair[A, B] { ... }
```

---

## Type inference rule

The compiler infers type args from the **value arguments**:

```go
Map([]int{1, 2}, func(x int) int { return x * 2 }) // T=int, U=int inferred
```

Inference **fails** — be explicit — when a type param appears only in the **result**,
nowhere in the arguments, or is an ambiguous untyped constant:

```go
Zero[int]()              // nothing to infer from
Parse[float64]("3.14")   // T only in return type
```

---

## When to use vs not

| Use generics                                   | Don't — do this instead                        |
|------------------------------------------------|------------------------------------------------|
| Reusable data structure (Stack/Set/Cache)      | One concrete type → write it for that type     |
| Same algorithm across 3+ types, same behavior  | 1–2 types → two small functions are clearer    |
| Precise numeric/ordered constraint             | Behavior differs by type → interface + dispatch|
| Map/Filter/Reduce-style utilities              | Unconstrained `[T any]` adding no safety → `any`|
| Internal implementation detail                 | Public API surface → export an interface       |

---

## Monomorphization / GC shape one-liner

Go doesn't fully monomorphize like C++; it does **GC shape stenciling** — one
instantiation per memory/GC shape (all pointers share one + a dictionary; each scalar
gets its own) — so scalar generics are ~as fast as hand-written, pointer-constrained
ones pay a small indirect cost.

---

## Key terms

- **Type parameter** — the placeholder type in `[T any]`, bound to a concrete type at the call site.
- **Constraint** — an interface bounding a type parameter, specifying its allowed operations.
- **Type set** — the set of concrete types a constraint admits (e.g. `cmp.Ordered`'s ints/floats/strings).
- **Underlying type / `~`** — `~int` matches any type whose underlying type is `int`, including named types.
- **`comparable`** — the constraint of types usable with `==`/`!=`; excludes slices, maps, funcs.
- **Monomorphization** — generating a separate specialized copy of code per type argument (C++/Rust).
- **GC shape stenciling** — Go's hybrid: one instantiation per GC shape, with a runtime dictionary for type-specific info.
- **Instantiation** — creating a concrete version of a generic (`Stack[int]`) from a generic definition.
- **Type inference** — the compiler deducing type arguments from value arguments at the call site.
