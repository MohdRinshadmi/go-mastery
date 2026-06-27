# Day 08 — Generics Pitfalls (Trap → Why → Fix)

Seven traps that turn a clean generic into a compile error, a subtle bug, or
code that should never have been generic. The lesson's *Common mistakes* section
covers the headlines; this drills into each with code.

---

## 1. Using `any` when you need an operation

**Trap**

```go
func Max[T any](a, b T) T {
	if a > b { // compile error: invalid operation a > b (T may not be ordered)
		return a
	}
	return b
}
```

**Why.** `any` is the empty type set — the *only* things guaranteed for every type
are assign, pass, and return. You cannot `==`, `<`, `>`, or `+`. The compiler refuses
because some types in `any` (slices, structs with slices, funcs) don't support `>`.

**Fix.** Pick the narrowest constraint that supports the operation. For ordering, use
the std `cmp.Ordered`:

```go
import "cmp"

func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}
```

---

## 2. `comparable` does not give you ordering

**Trap**

```go
func Max[T comparable](a, b T) T {
	if a > b { // compile error: a > b not allowed; comparable only permits == and !=
		return a
	}
	return b
}
```

**Why.** `comparable` is exactly the set of types usable with `==` and `!=`. That is
*not* the same as orderable. Booleans, structs of comparables, and interface values
are comparable but have no `<`. Ordering is a strictly smaller set.

**Fix.** Use `comparable` only for equality (dedup, set membership, map keys). Use
`cmp.Ordered` when you need `< <= > >=`.

```go
func Contains[T comparable](s []T, x T) bool { /* == is fine */ }
func Sort[T cmp.Ordered](s []T)               { /* needs ordering */ }
```

---

## 3. Slices, maps, and funcs are not `comparable`

**Trap**

```go
func Contains[T comparable](s []T, x T) bool { ... }

Contains([][]int{{1}, {2}}, []int{1})
// compile error: []int does not satisfy comparable
```

**Why.** `==` on a slice, map, or func is illegal in Go (only `== nil` is allowed),
so those types are not in the `comparable` type set. The constraint correctly rejects
them at compile time rather than panicking at runtime.

**Fix.** Don't force `==` onto non-comparable types. Take an equality function, or
require the caller to provide a key:

```go
func ContainsFunc[T any](s []T, eq func(T) bool) bool {
	for _, v := range s {
		if eq(v) {
			return true
		}
	}
	return false
}
// or use the std slices.ContainsFunc
```

---

## 4. Forgetting `~` in a custom constraint

**Trap**

```go
type Integer interface {
	int | int32 | int64 // no ~
}

type UserID int
func Sum[T Integer](xs []T) T { ... }

Sum([]UserID{1, 2, 3})
// compile error: UserID does not satisfy Integer (UserID's underlying type is int)
```

**Why.** `int | int32` is a type set containing *exactly* those named types. A defined
type like `UserID` has underlying type `int` but is a *different* type, so it's not in
the set. This is the #1 "why won't my named type work" generic bug.

**Fix.** Add `~` to match by underlying type:

```go
type Integer interface {
	~int | ~int32 | ~int64
}
// now any type whose underlying type is int / int32 / int64 qualifies — UserID included
```

---

## 5. Returning a bare zero `T` instead of `(T, bool)`

**Trap**

```go
func Get[T any](m map[string]T, key string) T {
	return m[key] // missing key -> zero value of T, indistinguishable from a real zero
}
```

**Why.** With `[T any]` there is no value of `T` you can reserve as "missing" — every
value, *especially* the zero value, may be a legitimate stored value. A missing key and
a key storing `0`/`""`/`nil` come back identical. (This is the day-08 debugging challenge.)

**Fix.** Thread the comma-ok through the signature:

```go
func Get[T any](m map[string]T, key string) (T, bool) {
	v, ok := m[key]
	return v, ok
}
```

Same rule applies to `Pop`, `Find`, `First`, cache `Get` — never a bare `T`.

---

## 6. Generics for only one or two concrete types

**Trap**

```go
func ToString[T any](v T) string {
	return fmt.Sprintf("%v", v)
}
```

**Why.** This generic buys nothing: `T` is unconstrained, so the body can't use any
type-specific operation, and the type parameter just adds instantiation noise and
slower-to-read signatures. Generics earn their keep at 3+ types with *identical*
behavior, or for reusable data structures — not as a reflex.

**Fix.** Use `any` (an ordinary interface value) or just call `fmt.Sprintf` inline:

```go
func ToString(v any) string { return fmt.Sprintf("%v", v) }
```

If you only ever pass `int`, write it for `int`. Two types with the same behavior?
Two small functions are often clearer than one generic.

---

## 7. Generics in public APIs that should be interfaces

**Trap**

```go
// Exported, and painful to read / mock / evolve.
func Process[T Processor[V], V any, R Result[V]](p T) (R, error) { ... }
```

**Why.** Type parameters on a public signature leak implementation into the API
surface. Callers must understand your constraint machinery, mocks must satisfy the
type set, and you can't change the parameterization without breaking the API. When the
behaviors differ by type, you want dynamic dispatch, not a type set.

**Fix.** Generics are an *implementation* tool. Export an interface or concrete type;
keep the type parameters inside the package.

```go
type Processor interface {
	Process(ctx context.Context) (Result, error)
}

func Process(p Processor) (Result, error) { return p.Process(ctx) }
```

---

### One-line summary

> Pick the **narrowest constraint** that compiles, never use the zero value as a
> sentinel, reach for generics only at 3+ types or for data structures, and keep
> type parameters out of your public API.
