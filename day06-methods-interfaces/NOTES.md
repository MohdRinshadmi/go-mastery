# Day 06 Notes — Methods & Interfaces (Cheatsheet)

## Receiver rules

| | Value receiver `(r T)` | Pointer receiver `(r *T)` |
|---|---|---|
| Receives | a copy | the original |
| Can mutate the receiver | no | yes |
| Cost on large structs | copies whole struct | one pointer |
| Callable on a `T` value | yes | yes, **if addressable** (auto `&`) |
| Callable on a `*T` | yes (auto deref) | yes |
| In method set of `T` | yes | no |
| In method set of `*T` | yes | yes |

Default for structs in production: **pointer receivers**. Don't mix kinds on one
type — pick one.

## Method-set rule (T vs *T)

```
Method set of  T  = methods with receiver  T
Method set of *T  = methods with receiver  T  AND  *T
```

Consequence: a pointer satisfies more interfaces than the value. When a type
doesn't satisfy an interface, try `&value`.

## Method declaration

```go
type Rectangle struct{ W, H float64 }

func (r Rectangle) Area() float64      { return r.W * r.H } // value
func (r *Rectangle) Scale(f float64)   { r.W *= f; r.H *= f } // pointer
```

## Interface declaration & satisfaction

```go
type Stringer interface {
    String() string
}

// Implicit: no "implements" keyword. Having String() is enough.
type Color struct{ R, G, B uint8 }
func (c Color) String() string { return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B) }

var s Stringer = Color{255, 0, 0} // satisfies structurally
```

Interface composition (embed interfaces):

```go
type ReadWriter interface {
    io.Reader
    io.Writer
}
```

Compile-time assertion that a type satisfies an interface:

```go
var _ Stringer = Color{}     // fails to build if Color lacks String()
var _ io.Writer = (*MyW)(nil)
```

## Type assertion & type switch

```go
// Assertion — single result PANICS on mismatch
f := r.(*os.File)

// Comma-ok — safe, no panic
f, ok := r.(*os.File)
if ok { /* use f */ }

// Type switch — branch on dynamic type
switch v := i.(type) {
case int:    fmt.Println("int", v)
case string: fmt.Println("string", v)
case nil:    fmt.Println("nil")
default:     fmt.Printf("%T\n", v)
}
```

## Empty interface / any

```go
var x any = 42          // any == interface{} (Go 1.18+ alias)
func log(args ...any) {} // holds any type; loses static type safety
```
Reach for `any` only for genuinely unknown types (JSON, logging). Otherwise
prefer concrete types or generics (Day 08).

## Nil-interface one-liner reminder

> An interface is `(type, value)`. A nil concrete pointer boxed into an interface
> is `(*T, nil)` — **not** `nil`. On the success path, `return nil` literally;
> never return a typed nil into an `error`/interface.

## io.Reader / io.Writer signatures

```go
type Reader interface {
    Read(p []byte) (n int, err error) // fills p; returns count + io.EOF when done
}
type Writer interface {
    Write(p []byte) (n int, err error) // writes p; err != nil if short write
}
type Closer interface {
    Close() error
}
```
`Read` may return `n > 0` **and** `io.EOF` together — process the bytes before
checking the error.

## Key terms

- **Receiver** — the value or pointer a method is attached to (`(r T)` / `(r *T)`).
- **Method set** — the set of methods callable on a type; determines which
  interfaces it satisfies. `T` gets value methods; `*T` gets both.
- **Structural typing** — a type satisfies an interface by having the methods, with
  no explicit `implements` declaration (compile-time duck typing).
- **Interface value** — the runtime representation of an interface: a two-word
  pair `(type descriptor, value)`.
- **Type assertion** — extracting the concrete type from an interface value
  (`v.(T)`); panics without comma-ok on mismatch.
- **Type switch** — `switch v := i.(type)`; branches on the dynamic type.
- **Empty interface / `any`** — `interface{}`; satisfied by every type.
- **Method value** — `t.Method`, receiver bound, callable as `f(args)`.
- **Method expression** — `T.Method`, receiver unbound, callable as `f(t, args)`.
- **Embedding** — placing a type inside a struct/interface so its methods are
  **promoted** to the outer type (composition, not inheritance).
- **Composition over inheritance** — building behavior by combining small types
  and interfaces instead of class hierarchies.
- **Addressable value** — a value whose address can be taken (variables, slice
  elements, struct fields); required to auto-call pointer methods. Composite
  literals and map index expressions are not addressable.
