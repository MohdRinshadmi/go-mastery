# Day 08 — Generics Interview Questions

Ten questions. The first seven are the lesson's; the last three go deeper. Try to
answer aloud before expanding each `<details>`.

---

### 1. What problem do generics solve? Give a concrete before/after example.

<details>
<summary>Answer</summary>

Before generics you had two bad options: copy-paste a function per type
(`SumInts`, `SumFloat64s`) which is boilerplate, or accept `interface{}` and lose
all compile-time type safety plus pay for boxing and type assertions.

```go
// Before: lose type safety
func Sum(s []interface{}) interface{} { ... } // runtime assertions everywhere

// After: write once, typed at the call site
func Sum[T cmp.Ordered](s []T) T { ... }
```

Generics give you "write once, typed at the call site": one implementation, full
static typing, no assertions, and (for scalars) no boxing.

</details>

---

### 2. What is the difference between `any`, `comparable`, and `cmp.Ordered` as constraints?

<details>
<summary>Answer</summary>

They are three progressively smaller type sets:

- **`any`** — every type. You can only assign, pass, and return. No `==`, no `<`.
- **`comparable`** — types usable with `==` and `!=` (excludes slices, maps, funcs).
  Permits equality, **not** ordering.
- **`cmp.Ordered`** — integers, floats, and strings: types supporting
  `< <= > >=` (and ordered `==`).

Rule of thumb: pick the narrowest one that supports the operations you actually use.
`comparable ⊂` (loosely) the equality-capable types; `cmp.Ordered` is the ordered
subset.

</details>

---

### 3. What does `~int` mean in a constraint? Why is it needed?

<details>
<summary>Answer</summary>

`~int` means "any type whose **underlying type** is `int`," including defined types
like `type UserID int`. Plain `int` in a constraint matches *only* the predeclared
`int`, so `UserID` would be rejected.

```go
type Integer interface { ~int | ~int8 | ~int16 | ~int32 | ~int64 }
type UserID int
Abs(UserID(-5)) // works only because of ~int
```

It's needed so generic code works with the named types people define on top of
primitives — without it, the most common "why won't my named type satisfy this
constraint" bug appears.

</details>

---

### 4. When should you use a generic function vs an interface-based function?

<details>
<summary>Answer</summary>

Use a **generic** when the algorithm is identical across types and only the type
differs — data structures (Stack, Set, Cache), and util functions (Map/Filter/Reduce).
The win is type safety and no boxing.

Use an **interface** when behavior *differs* by type and you want dynamic dispatch
(a `Writer`, a `Processor`, a payment backend). Each implementer behaves differently;
that's polymorphism, not parametricity.

Heuristic: "same code, different type" → generic. "Different code behind a common
shape" → interface.

</details>

---

### 5. Can you use generics to replace all interface usage? Why or why not?

<details>
<summary>Answer</summary>

No. Generics give *parametric* polymorphism (one body, many types) resolved at
compile time. Interfaces give *ad-hoc / subtype* polymorphism — different
implementations dispatched at runtime, including types you don't control and
heterogeneous collections (`[]io.Writer` holding several concrete types).

Generics can't hold a mixed-type collection, can't be satisfied by an external type
without that type knowing your constraint, and can't dispatch to different behavior.
The two tools are complementary; in fact constraints *are* interfaces.

</details>

---

### 6. How does Go's generic implementation (GC shape stenciling) differ from C++'s monomorphization?

<details>
<summary>Answer</summary>

**C++/Rust** fully monomorphize: a separate specialized copy of the function for
every type argument. Fast, but binary bloat and longer compiles.

**Go** uses **GC shape stenciling**: it generates one instantiation per *GC shape*,
not per type. All pointer-shaped types share a single instantiation (they have the
same memory/GC layout) and pass a hidden dictionary for type-specific info; each
distinct scalar shape (int, float64, …) gets its own.

Result: fewer copies than C++, scalar generics roughly as fast as hand-written code,
while pointer/interface-constrained generics may pay a small indirect-dispatch cost
via the dictionary.

</details>

---

### 7. A junior writes `func Get[T any](m map[string]T, key string) T` — what's wrong, and how do you fix it?

<details>
<summary>Answer</summary>

It can't signal "missing." A missing key makes `m[key]` return the **zero value** of
`T`, which is ambiguous — `0`, `""`, or `nil` may be legitimate stored values. With
`[T any]` there's no value you can reserve as a "not found" sentinel.

Fix: thread the map's comma-ok through the signature.

```go
func Get[T any](m map[string]T, key string) (T, bool) {
	v, ok := m[key]
	return v, ok
}
```

(This is exactly the day-08 debugging challenge.)

</details>

---

### 8. What is the difference between a generic type and an interface?

<details>
<summary>Answer</summary>

A **generic type** (`Stack[T]`) is a *template*: `Stack[int]` and `Stack[string]`
are distinct, concrete types produced at compile time. There is no boxing — a
`Stack[int]` stores real `int`s — and the element type is fixed per instantiation.

An **interface** (`io.Writer`) is a *single runtime type* holding a (type, value)
pair via a method table. It can hold any implementer, mixed together, dispatched
dynamically.

So: generic type = compile-time specialization, homogeneous, no dispatch; interface
= runtime abstraction, heterogeneous, dynamic dispatch. A constraint is an interface
*used* to bound a type parameter — the roles overlap in syntax but differ in purpose.

</details>

---

### 9. Can a method have its own new type parameter? Explain.

<details>
<summary>Answer</summary>

**No.** Go methods cannot introduce their *own* type parameters. A method may use the
type parameters of its receiver:

```go
func (s *Stack[T]) Push(v T) { ... } // T comes from the receiver — fine
```

but you cannot write `func (s *Stack[T]) Map[U any](fn func(T) U) []U` — that's a
compile error. The reason is interface satisfaction: a method set must be fixed for a
type to satisfy an interface, but a parametric method would represent infinitely many
signatures, breaking method-set checking and the runtime method table.

Workaround: make it a top-level generic *function* instead.

```go
func Map[T, U any](s *Stack[T], fn func(T) U) []U { ... }
```

</details>

---

### 10. What is type inference, and when does it fail?

<details>
<summary>Answer</summary>

Type inference lets the compiler deduce type arguments from the ordinary arguments,
so you write `Map(xs, fn)` instead of `Map[int, string](xs, fn)`. It works by
unifying the types of the passed values against the type parameters.

It **fails** (and you must specify explicitly) when a type parameter doesn't appear
in any argument, or appears only in the *result*:

```go
func Zero[T any]() T { var z T; return z }
x := Zero[int]() // must be explicit — nothing to infer T from

func Parse[T any](s string) T { ... }
n := Parse[int]("42") // T only in the return type — explicit
```

It also fails for untyped-constant ambiguity (e.g. inferring `T` from a literal `3`
that could be several numeric types) and when a needed argument is an untyped `nil`.
When inference can't pin every type parameter, supply them in `[ ]` at the call site.

</details>
