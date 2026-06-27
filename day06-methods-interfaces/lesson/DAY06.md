# Day 06 — Methods, Interfaces, and Why Go Beats Inheritance

> Mentor note: Today is where Go diverges sharply from every OOP language you've used. There's no class, no inheritance, no virtual dispatch table. What Go has instead is *simpler and more powerful*. The interface system is one of the most elegant ideas in language design. Understand this day deeply — it changes how you think about code.

---

## 0. The Big Picture

In Java/Python/C# you think: "What *is* this thing? It's a Dog, which is an Animal."
In Go you think: "What can this thing *do*? It can Speak(), so it satisfies Speaker."

Go uses **structural typing** (duck typing at compile time). You don't declare that a type implements an interface — the compiler checks it for you, silently. This has profound consequences for how you structure code.

---

## 1. Methods

### Theory
A method is a function with a **receiver** — the type it belongs to. No `class` keyword needed. You attach methods to any named type you define.

```go
type Rectangle struct {
    Width, Height float64
}

// Value receiver — gets a COPY of the struct
func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

// Pointer receiver — gets a POINTER, can mutate
func (r *Rectangle) Scale(factor float64) {
    r.Width *= factor
    r.Height *= factor
}
```

### Value receiver vs Pointer receiver — the core decision

| | Value receiver `(r Rect)` | Pointer receiver `(r *Rect)` |
|---|---|---|
| Gets | A copy | The original |
| Can mutate | No (copy dies) | Yes |
| Performance | Copy overhead for large structs | One pointer, no copy |
| Nil safety | Safe if called on nil pointer? No | Must guard nil |
| Method set | Called on value OR pointer | Called on pointer only |

**The rule:** If any method on the type *needs* a pointer receiver (for mutation), make *all* methods pointer receivers. Mixing creates confusing method set rules.

### Why it exists
Methods let you attach behavior to data without a class hierarchy. A `time.Time` has methods. A `net.IP` has methods. A custom `Money` type you define can have methods. Same mechanism, no special syntax.

### When to use value vs pointer receivers
- Small immutable value types (e.g., `Point`, `Color`): value receiver.
- Anything with internal state that changes (e.g., `Counter`, `Buffer`, most services): pointer receiver.
- Rule of thumb: if the struct is larger than two or three words, use pointer receiver for performance.

### Common mistakes
1. **Inconsistent receivers**: some methods value, some pointer on the same type. Confuses the method set.
2. **Calling pointer receiver method on a non-addressable value**:
   ```go
   Rectangle{3, 4}.Scale(2) // COMPILE ERROR — can't take address of composite literal
   ```
3. **Nil pointer dereference**: calling a method on a nil pointer that dereferences a field.

**Senior take:** Pointer receivers are the default for structs in production code. Use value receivers only for pure, immutable computation. The performance difference matters at scale.

---

## 2. Method Sets (The Rule Nobody Reads Until They Get Burned)

The **method set** of a type determines which interfaces it satisfies.

```
Type T:   methods with receiver T
Type *T:  methods with receiver T AND *T
```

This means: a value of type `T` can only call methods with value receivers. A value of type `*T` can call both. Interfaces are checked against method sets.

```go
type Sizer interface {
    Size() int
}

type Box struct{ n int }
func (b *Box) Size() int { return b.n } // pointer receiver

var s Sizer = &Box{5}   // ✓ works — *Box has Size()
var s2 Sizer = Box{5}   // ✗ COMPILE ERROR — Box (non-pointer) lacks Size()
```

**Senior take:** This is the source of 80% of "why doesn't my type satisfy this interface" confusion. When in doubt, use `&` (take a pointer).

---

## 3. Interfaces — The Killer Go Feature

### Theory
An interface in Go is a **set of method signatures**. Any type that has those methods *automatically* satisfies the interface. No `implements` keyword. No declaration. The compiler checks structurally.

```go
type Writer interface {
    Write(p []byte) (n int, err error)
}
```

`os.File` satisfies `Writer`. `bytes.Buffer` satisfies `Writer`. Your custom `S3Uploader` satisfies `Writer` if it has that method. None of them say "I implement Writer." They just have the method.

### Why this is revolutionary (vs Java)

In Java, if you want `Dog` to implement `Animal`, you must write `class Dog extends Animal` or `class Dog implements Animal` at definition time. The author of `Dog` must know about `Animal` ahead of time.

In Go: someone writes a `Dog` library. Six months later, *you* define an `Animal` interface. `Dog` satisfies it automatically — **no changes to the Dog library needed**. This decouples producers from consumers. Third-party types satisfy your interfaces. Standard library types satisfy interfaces you define. This is the superpower.

### Small interfaces — the Go ideal

Go's stdlib is full of tiny interfaces:

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}
type Writer interface {
    Write(p []byte) (n int, err error)
}
type Closer interface {
    Close() error
}
type ReadWriter interface {   // composition of two interfaces
    Reader
    Writer
}
```

**The rule:** The bigger the interface, the weaker the abstraction. Single-method interfaces are ideal. `io.Reader`, `io.Writer`, `fmt.Stringer`, `error` — all have one method.

Why? Small interfaces are easy to satisfy, easy to mock, easy to compose. A function that takes `io.Reader` works with files, strings, HTTP bodies, in-memory buffers, network connections — all without knowing any of them.

### When to use interfaces
- When you have multiple concrete implementations that should be interchangeable.
- When you're writing a function that should work with "anything that can X."
- For dependency injection and testability (Day 7).
- When you need to decouple a package from a concrete dependency.

### When NOT to use interfaces
- A single concrete type with no realistic alternatives → don't premature-abstract.
- Performance-critical hot paths: interface calls are indirect (virtual dispatch), slightly slower than direct calls. In tight loops, this matters.
- When it only makes code harder to read for no gain.

**Senior take:** Define interfaces in the **consumer** package, not the producer package. If package `auth` needs to store users, define `type UserStore interface { ... }` in `auth`, not in the storage package. This is a rule almost every Java developer violates when they first write Go.

---

## 4. io.Reader and io.Writer — The Interfaces That Unite Go's I/O

These two interfaces are worth understanding deeply because they're everywhere.

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}
type Writer interface {
    Write(p []byte) (n int, err error)
}
```

`Read` fills the provided buffer and returns how many bytes were written. It returns `io.EOF` when done. This is the universal abstraction for anything you can read from: files, network sockets, HTTP responses, in-memory buffers, gzip streams, TLS connections.

```go
// This function works with ANY reader: file, HTTP body, stdin, buffer...
func countBytes(r io.Reader) (int, error) {
    buf := make([]byte, 4096)
    total := 0
    for {
        n, err := r.Read(buf)
        total += n
        if err == io.EOF {
            return total, nil
        }
        if err != nil {
            return total, err
        }
    }
}
```

This is what "design for interfaces" means. One function, infinite inputs.

---

## 5. The Empty Interface and `any`

```go
// Pre-Go 1.18:
func printAnything(v interface{}) { fmt.Println(v) }

// Go 1.18+:
func printAnything(v any) { fmt.Println(v) }
```

`any` (alias for `interface{}`) holds a value of *any type*. It's Go's escape hatch. Every type satisfies an interface with no methods.

### When to use `any`
- When you genuinely don't know the type at compile time (JSON unmarshaling, logging frameworks, generic containers before Go had generics).
- Testing frameworks (`t.Fatal(args ...any)`).

### When NOT to use `any`
- Everywhere else. `any` loses type safety. The compiler can no longer help you. This is the road to runtime panics.
- **If you're reaching for `any` in new code, ask yourself: can generics (Day 8) solve this better?**

---

## 6. Type Assertions and Type Switches

When you have an interface value and need the concrete type back:

```go
var r io.Reader = os.Stdin

// Single assertion — panics if wrong type
f := r.(*os.File)

// Safe assertion — ok is false instead of panic
f, ok := r.(*os.File)
if ok {
    fmt.Println("stdin is a file:", f.Name())
}
```

Type switch — the idiomatic way to handle multiple concrete types:

```go
func describe(i interface{}) string {
    switch v := i.(type) {
    case int:
        return fmt.Sprintf("integer: %d", v)
    case string:
        return fmt.Sprintf("string: %q (len %d)", v, len(v))
    case error:
        return fmt.Sprintf("error: %v", v)
    case nil:
        return "nil"
    default:
        return fmt.Sprintf("unknown type: %T", v)
    }
}
```

**Senior take:** Frequent type assertions in your code are a design smell. They mean your interfaces are too broad. If you find yourself constantly asserting to get the concrete type, redesign the interface to expose the behavior you need instead.

---

## 7. The Nil Interface Gotcha — One of Go's Top 3 Gotchas

This one burns everyone eventually:

```go
type MyError struct{ msg string }
func (e *MyError) Error() string { return e.msg }

func getError(fail bool) error {
    var err *MyError // nil pointer of type *MyError
    if fail {
        err = &MyError{"something went wrong"}
    }
    return err // BUG: returns a non-nil interface wrapping a nil pointer
}

e := getError(false)
if e != nil {    // This is TRUE! The interface has a type (*MyError) even though the value is nil
    fmt.Println("error:", e) // Prints: error: <nil>
}
```

**Why:** An interface value has two fields internally: (type, value). A nil interface has (nil, nil). But `(*MyError)(nil)` returned as `error` gives you (*MyError, nil) — which is NOT a nil interface.

**Fix:** Return `nil` directly, never return a typed nil:
```go
func getError(fail bool) error {
    if fail {
        return &MyError{"something went wrong"}
    }
    return nil  // ← always return untyped nil for the error interface
}
```

---

## 8. Composition Over Inheritance — the Go Philosophy

Go has **no** inheritance. No `extends`. No `super`. This is intentional.

Instead Go gives you:
1. **Interface composition**: embed interfaces in interfaces
2. **Struct embedding**: embed types in structs (Day 7's topic, preview here)

```go
// Interface composition — build bigger contracts from small ones
type ReadWriter interface {
    io.Reader
    io.Writer
}

type ReadWriteCloser interface {
    io.Reader
    io.Writer
    io.Closer
}
```

```go
// Struct embedding — promotes methods (NOT inheritance, promotion)
type Logger struct {
    prefix string
}
func (l Logger) Log(msg string) { fmt.Printf("[%s] %s\n", l.prefix, msg) }

type Server struct {
    Logger          // embedded — Server "has" Log() promoted to it
    addr   string
}

s := Server{Logger: Logger{prefix: "SERVER"}, addr: ":8080"}
s.Log("starting up")  // calls s.Logger.Log() — promoted, not inherited
```

**Why composition beats inheritance:**
- No fragile base class problem (changing parent breaks all children).
- Behaviors are explicit and discoverable. No need to trace up an inheritance tree.
- Multiple "inheritance" without the diamond problem — just embed multiple types.
- Easy to swap implementations at runtime (via interfaces).

**The Gang of Four famously wrote "favor composition over inheritance" in 1994. Go just made that the only option.**

---

## Expert Thinking Mode — how different levels see interfaces

- **Beginner:** "An interface is a contract that forces my struct to have certain methods."
- **Senior:** "Interfaces let me write functions that work with anything satisfying a behavior, making them testable and composable. I define them in the consumer package."
- **Staff:** "My package boundaries are drawn by interfaces. Package A exports concrete types. Package B defines the interface it needs from A. Packages never import each other — they import the interface. This is how you eliminate import cycles."
- **Architect:** "Interfaces are the seams in my architecture where I can swap implementations (prod vs test, v1 vs v2, SQL vs NoSQL) without touching calling code. Every external dependency — DB, cache, queue, email — is behind an interface. That interface lives in the domain, not the infrastructure layer."

---

## Real-world use

- **Google (net/http):** `http.Handler` is a single-method interface (`ServeHTTP`). Thousands of frameworks, middleware, and servers implement it. One interface connects the entire Go web ecosystem.
- **Stripe:** Payment processors are behind interfaces. `ChargeProcessor` interface lets them swap from their legacy processor to Stripe's own API internally, with tests using mocks. The business logic never changes.
- **Cloudflare:** `io.Reader`/`io.Writer` composability lets their proxy code chain compression, TLS, and logging as pipeline stages — each stage just wraps a reader/writer.
- **Uber (Go kit):** Every service operation goes through interface-typed middleware for tracing, logging, rate-limiting. The core logic is unaware of observability.

---

## Interview Questions

1. What is the difference between a value receiver and a pointer receiver? When would you use each?
2. Explain Go's implicit interface satisfaction. How does it differ from Java's `implements`?
3. What is the method set of type `T` vs `*T`? Give an example where this matters.
4. Why does Go favor small interfaces? What's wrong with a 10-method interface?
5. Explain the nil interface gotcha. What does `(type, value)` mean internally?
6. You have a function `func process(v interface{})`. When is a type assertion safe vs risky? What's the alternative?
7. Why does Go not have inheritance? What problem does this solve in large codebases?

---

## Your tasks for today

Go to `../exercises/`. You'll implement a shape area calculator, an io.Writer middleware, and a type switch dispatcher — plus a production-grade challenge where you design a notification system using interfaces. Try everything before opening `../solutions/`.

## Day 06 companion files

- [Debugging challenge](../debugging/README.md) — the nil-interface gotcha, live.
- [Pitfalls](../PITFALLS.md) — Trap → Why → Fix.
- [Interview questions](../INTERVIEW.md) — with model answers.
- [Notes / cheatsheet](../NOTES.md) — quick reference.
- [Resources](../RESOURCES.md) — curated links.
