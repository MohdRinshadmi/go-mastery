# Day 03 — Structs, Pointers, and Control Flow

> Mentor note: Go's type system is deceptively simple — but the value-vs-pointer decision is where senior Go engineers still pause and think. Control flow in Go is also where many "but I know Python/Java" assumptions burn people. Read the Go 1.22 loop variable section carefully — this changed real production code.

---

## Table of Contents
- [1. Structs](#1-structs)
- [2. Embedding / Anonymous Fields](#2-embedding--anonymous-fields)
- [3. Struct Tags (intro)](#3-struct-tags-intro)
- [4. Pointers](#4-pointers)
- [5. Value vs Pointer Receivers](#5-value-vs-pointer-receivers)
- [6. Control Flow](#6-control-flow)
- [7. The Go 1.22 Loop Variable Fix](#7-the-go-122-loop-variable-fix)
- [Expert Thinking Mode](#expert-thinking-mode)
- [Real-world use](#real-world-use)
- [Interview Questions](#interview-questions)

---

## 1. Structs

### Theory
A struct is a composite type that groups named fields. In Go, structs are the primary mechanism for creating domain types — there are no classes, no inheritance.

```go
type User struct {
    ID        int
    Name      string
    Email     string
    CreatedAt time.Time
}
```

### Initialization

```go
// Named fields (preferred — order-independent, readable, robust to field additions)
u := User{
    ID:    1,
    Name:  "Alice",
    Email: "alice@example.com",
}

// Positional (fragile — breaks silently if fields are reordered)
u := User{1, "Alice", "alice@example.com", time.Now()}

// Zero value — all fields zeroed (int→0, string→"", time.Time→zero time)
var u User

// Pointer to new struct
u := &User{Name: "Alice"} // u is *User
```

### Struct equality and comparison
Structs are comparable if all their fields are comparable. Two structs are equal if all fields are equal:

```go
a := User{ID: 1, Name: "Alice"}
b := User{ID: 1, Name: "Alice"}
fmt.Println(a == b) // true — field-by-field comparison
```

Structs containing slices, maps, or functions are **not** comparable with `==`.

### Anonymous structs — when to use
Inline one-off types: test cases, config shapes you'll never reuse:

```go
testCases := []struct {
    input    string
    expected int
}{
    {"hello", 5},
    {"hi", 2},
}
for _, tc := range testCases {
    got := len(tc.input)
    if got != tc.expected {
        fmt.Printf("FAIL: %q → got %d want %d\n", tc.input, got, tc.expected)
    }
}
```

### Common mistakes — structs
1. **Positional initialization:** `User{1, "Alice", "alice@example.com"}` — a new field added in the middle silently re-assigns all subsequent values. Always use named fields in production code.
2. **Large structs passed by value:** every function call copies all fields. For structs with more than a few small fields, use pointers (see section 5).
3. **Forgetting zero value semantics for time.Time:** `var t time.Time` is year 0001, not "now" or "invalid." Use a pointer `*time.Time` and nil for "not set."

> **Senior take:** Define structs as value types, but think carefully about size. A struct with two int64s = 16 bytes. A struct with a slice + map + string + time.Time = ~80 bytes. Passing the latter by value through a hot loop costs real CPU. Profile first, but know the mental model.

---

## 2. Embedding / Anonymous Fields

### Theory
Go has no inheritance, but it has **composition via embedding**. When you embed a type in a struct, its fields and methods are promoted to the outer struct.

```go
type Animal struct {
    Name string
}

func (a Animal) Speak() string {
    return a.Name + " makes a sound"
}

type Dog struct {
    Animal        // embedded — no field name, just the type
    Breed  string
}

d := Dog{
    Animal: Animal{Name: "Rex"},
    Breed:  "Labrador",
}
fmt.Println(d.Name)    // promoted from Animal — no need for d.Animal.Name
fmt.Println(d.Speak()) // promoted method — d.Speak() calls Animal.Speak()
fmt.Println(d.Breed)   // Dog's own field
```

### Embedding is not inheritance
This is critical: embedding is **field/method forwarding via composition**, not subtyping. A `Dog` is not an `Animal` in Go's type system — you cannot pass a `Dog` where an `Animal` is expected (unlike Java/Python). Interfaces handle polymorphism (Day 6).

### Embedding with interfaces (preview)
The most powerful use of embedding in production Go is embedding interfaces in structs, or embedding a concrete type in a test double to forward most methods. We'll cover this in depth on Day 6.

### Name conflicts
If the outer struct and the embedded type both define a field/method with the same name, the outer struct's definition wins (shadows the inner):

```go
type Base struct { ID int }
type Child struct {
    Base
    ID string // Child.ID shadows Base.ID
}
c := Child{}
c.ID        // accesses Child.ID (string)
c.Base.ID   // explicit path to Base.ID (int)
```

> **Senior take:** Embedding is how Go stdlib achieves "mix-in" behavior. `http.ResponseWriter` embeds an `io.Writer`. `sync.RWMutex` embedded in a struct gives it thread-safe read/write methods. Understand this and you'll read stdlib source code fluidly.

---

## 3. Struct Tags (intro)

### Theory
Struct tags are string literals attached to struct fields, read at runtime via reflection. They're the primary way Go communicates metadata to encoding/decoding libraries.

```go
type Product struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Price float64 `json:"price,omitempty"` // omitted if zero
}
```

The tag format is a raw string literal: `` `key:"value"` ``. Multiple keys are space-separated:

```go
type DBRecord struct {
    ID   int    `json:"id" db:"id" validate:"required"`
    Name string `json:"name" db:"name" validate:"min=1,max=100"`
}
```

### Why it exists
Go has no annotations (Java) or decorators (Python). Tags are the opted-in metadata mechanism. The `encoding/json`, `database/sql` wrappers, and validation libraries all read them.

### Common mistakes — tags
1. **Unexported fields are ignored by encoding packages** — your tag does nothing if the field is lowercase.
2. **Typos in tags are silently ignored** — `json:"naem"` will serialize as `naem`, not `name`. No compile error.
3. **Tag format is rigid:** `json: "name"` (space after colon) → broken. It must be `json:"name"` with no space.

We'll use struct tags extensively on Day 5 (JSON). That's where they really click.

---

## 4. Pointers

### Theory
A pointer stores the memory address of another value. In Go:
- `&x` → takes the address of `x`, gives you `*T`
- `*p` → dereferences pointer `p`, gives you the value

```go
x := 42
p := &x      // p is *int, points to x
fmt.Println(*p)  // 42 — dereference
*p = 99
fmt.Println(x)   // 99 — modified x through the pointer
```

### No pointer arithmetic
Unlike C/C++, Go **does not allow** pointer arithmetic. You cannot do `p++` to advance to the next memory address. This is intentional — it prevents an entire class of memory safety bugs. The garbage collector needs to track all pointer movement; arbitrary arithmetic would break that.

### nil pointer dereference — the most common Go panic

```go
var p *User    // p is nil (zero value of *User is nil)
fmt.Println(p.Name) // PANIC: nil pointer dereference
```

Guard always:
```go
if p == nil {
    return errors.New("user pointer is nil")
}
fmt.Println(p.Name) // safe
```

### When Go is pass-by-value

Everything in Go is passed by value. When you pass a pointer, the pointer value (the address) is copied — but both the caller and callee point to the same underlying data.

```go
func increment(n int) {
    n++ // modifies local copy — caller's n is unchanged
}

func incrementPtr(n *int) {
    *n++ // dereferences and modifies the original value
}

x := 5
increment(x)
fmt.Println(x)    // 5 — unchanged

incrementPtr(&x)
fmt.Println(x)    // 6 — changed
```

This is not "pass by reference" in the Java/C++ sense — the pointer address itself is copied, but the data it points to is shared.

### new vs & composite literal

```go
p := new(User)    // allocates User, zero-initialized, returns *User
q := &User{}      // same result — more idiomatic for structs in practice
r := &User{ID:1}  // with field initialization
```

`new(T)` is rarely used for structs. `&T{...}` is idiomatic.

### Heap vs stack (the Go compiler decides)
Unlike C, you don't choose heap vs stack. The Go compiler does **escape analysis**:
- If a value's address is taken and escapes the function (returned, stored in heap object, etc.) → heap allocation.
- If it stays local → stack allocation.

```go
func newUser() *User {
    u := User{ID: 1}  // compiler sees this escapes — heap-allocated
    return &u         // this is safe in Go (unlike C!) — GC manages lifetime
}
```

This is safe and idiomatic. You would never return the address of a local variable in C, but in Go it's fine — the garbage collector ensures the object lives as long as anything points to it.

> **Senior take:** Don't cargo-cult "always use pointers for performance." For small structs (1-3 fields, < 32 bytes), passing by value is often FASTER because it avoids heap allocation and is cache-friendly. Pointer chasing (dereferencing pointers to heap objects) causes cache misses. Profile with `go tool pprof` before optimizing.

---

## 5. Value vs Pointer Receivers

Methods can have value or pointer receivers. This is the decision you'll make on every struct in Go.

```go
type Counter struct {
    n int
}

// Value receiver — gets a COPY. Cannot mutate the original.
func (c Counter) Value() int {
    return c.n
}

// Pointer receiver — gets a pointer. CAN mutate the original.
func (c *Counter) Increment() {
    c.n++
}
```

### The rules

**Use a pointer receiver when:**
1. The method needs to mutate the receiver.
2. The struct is large (copying is expensive).
3. There are other pointer receiver methods on the same type — **be consistent**.

**Use a value receiver when:**
1. The method is read-only AND the struct is small.
2. The type is a basic value type (like a `type Celsius float64`).
3. You deliberately want to prevent mutation (immutability by convention).

**The consistency rule is the most important:** if any method on a type has a pointer receiver, all methods should have pointer receivers. Mixing causes subtle bugs with interfaces.

### Automatic addressing

Go automatically takes the address when calling a pointer-receiver method on an addressable value:

```go
c := Counter{}
c.Increment()  // Go auto-converts to (&c).Increment() — works
```

But not on non-addressable values:
```go
Counter{}.Increment()  // COMPILE ERROR — can't take address of temporary
```

### Common mistake: value receiver on map/slice fields

```go
type Cache struct {
    data map[string]int
}

// Value receiver — c is a copy of the Cache struct header
// BUT: the map inside is a reference type — c.data and the original
// both point to the SAME underlying map.
func (c Cache) Set(k string, v int) {
    c.data[k] = v  // This DOES mutate the original map! Confusing!
}
```

A value receiver on a struct containing a map still mutates the map — because the map itself is a reference. This is a common source of confusion.

> **Senior take:** When in doubt for non-trivial types, use pointer receivers. The Go standard library types (`http.Request`, `bytes.Buffer`, `sync.Mutex`) all use pointer receivers. When you see `var mu sync.Mutex` and call `mu.Lock()`, Go is calling `(&mu).Lock()` under the hood.

---

## 6. Control Flow

### if — statement, not expression

```go
// Basic
if x > 0 {
    fmt.Println("positive")
} else if x < 0 {
    fmt.Println("negative")
} else {
    fmt.Println("zero")
}

// With initialization statement — extremely common in Go
if err := doSomething(); err != nil {
    return err
}
// err is scoped to the if block — doesn't pollute the outer scope
```

The init-statement form (`if x := f(); ...`) is idiomatic for error handling and reduces variable scope. You'll see it constantly in Go codebases.

### switch — more powerful than you think

```go
// Expression switch — no fallthrough by default (unlike C/Java!)
switch status {
case "pending":
    process()
case "done", "cancelled":  // multiple values per case
    archive()
default:
    log.Printf("unknown status: %s", status)
}

// Switch with initialization statement
switch v := getValue(); {
case v < 0:
    fmt.Println("negative")
case v == 0:
    fmt.Println("zero")
default:
    fmt.Println("positive")
}

// Type switch — extremely useful with interfaces
func describe(i interface{}) {
    switch v := i.(type) {
    case int:
        fmt.Printf("int: %d\n", v)
    case string:
        fmt.Printf("string: %q\n", v)
    case bool:
        fmt.Printf("bool: %t\n", v)
    default:
        fmt.Printf("unknown: %T\n", v)
    }
}
```

**Fallthrough:** Go switch does NOT fall through by default. Use `fallthrough` keyword explicitly if you need it (rare).

### for — the only loop

Go has ONE loop keyword: `for`. It replaces `while`, `do-while`, and `for`.

```go
// Classic C-style
for i := 0; i < 10; i++ { }

// While-style (condition only)
for n > 0 {
    n /= 2
}

// Infinite loop
for {
    // runs forever — use break to exit
}

// Range over slice
for i, v := range slice { }

// Range over map
for k, v := range myMap { }

// Range over string — iterates RUNES (Unicode code points), not bytes!
for i, r := range "héllo" {
    fmt.Printf("  index=%d rune=%c\n", i, r)
    // i is byte offset, r is rune (int32) — not necessarily sequential indices
}

// Range over channel (Day 9)
for v := range ch { }
```

### range over string — the Unicode gotcha

```go
s := "héllo" // 'é' is 2 bytes in UTF-8
for i, r := range s {
    fmt.Printf("i=%d r=%c\n", i, r)
    // Output: i=0 h, i=1 é, i=3 l, i=4 l, i=5 o
    // index 2 is skipped — 'é' occupies bytes 1-2
}
```

If you need byte access: `for i := 0; i < len(s); i++` — `len(s)` gives byte count, not rune count.
For rune count: `utf8.RuneCountInString(s)`.

### break and continue with labels

Labels let you break/continue an outer loop — something you can't do in most languages without a flag variable:

```go
outer:
    for i := 0; i < 3; i++ {
        for j := 0; j < 3; j++ {
            if i == 1 && j == 1 {
                break outer  // exits the OUTER loop, not just the inner
            }
            fmt.Println(i, j)
        }
    }
```

This is cleaner than a `found` boolean flag propagated through nested loops.

### Common mistakes — control flow

1. **Switch fallthrough confusion:** coming from Java/C, expecting Go to fall through by default. It doesn't.
2. **range loop copy:** `for _, v := range bigSlice` — `v` is a copy. Modifying `v` doesn't modify the slice.
3. **Range over nil slice is safe:** `for _, v := range nil { }` iterates zero times, no panic.
4. **break inside a select/switch inside a loop:** `break` only breaks the innermost select/switch, NOT the outer for loop. Use a label to break the for.

---

## 7. The Go 1.22 Loop Variable Fix

This is the most impactful Go behavioral change in years. Pre-1.22, it caused countless bugs. In Go 1.22+, it's fixed — but you MUST understand the old behavior to read pre-1.22 code and understand why patterns exist.

### The old bug (pre-Go 1.22, go.mod: `go 1.21` or earlier)

```go
funcs := make([]func(), 3)
for i := 0; i < 3; i++ {
    funcs[i] = func() { fmt.Println(i) } // captures i by reference
}
for _, f := range funcs {
    f()  // prints: 3, 3, 3 — NOT 0, 1, 2
         // because all closures share the SAME i, which is 3 after the loop
}
```

The loop variable `i` is a **single variable** reused each iteration. Closures captured its address, so by the time they run, `i == 3`.

The classic range version:
```go
items := []string{"a", "b", "c"}
funcs := make([]func(), len(items))
for i, v := range items {
    funcs[i] = func() { fmt.Println(v) }  // pre-1.22: all print "c"
}
```

### The fix (workaround in pre-1.22 code)

```go
for i, v := range items {
    i, v := i, v  // shadow with new variable per iteration
    funcs[i] = func() { fmt.Println(v) }  // each closure captures its own v
}
```

### Go 1.22 behavior

In Go 1.22+ (with `go 1.22` or later in `go.mod`), **each loop iteration gets its own variable**. The bug is gone. The workaround (`i, v := i, v`) is no longer needed.

```go
// Go 1.22+ (go.mod: go 1.22)
for i, v := range items {
    funcs[i] = func() { fmt.Println(v) }  // correctly prints a, b, c
}
```

**Why you still need to know this:**
1. You'll read pre-1.22 codebases.
2. The workaround pattern (`v := v`) still appears in code and you need to recognize it.
3. Goroutines with loop variables are the most common form (Day 9).

> **Senior take:** The loop variable change is in the top 3 most-complained-about Go bugs historically. Rob Pike has said in public it was a design mistake. Go 1.22 fixed it, but the old behavior is deeply embedded in Go folklore. Every senior Go engineer has been burned by it at least once.

---

## Expert Thinking Mode — value vs pointer

- **Beginner:** "I'll use pointers everywhere because that's what looks like C/Java references."
- **Senior:** "I use pointer receivers for mutation and large structs. I use value receivers for small immutable types. I apply the consistency rule. I check escape analysis with `go build -gcflags='-m'` when performance matters."
- **Staff:** "I design types so mutation is explicit. If a type is semantically a value (a `Color`, a `UUID`), value receiver. If it's a resource with lifecycle (`*DB`, `*Server`), pointer. The type name signals the semantic."
- **Architect:** "Pointer vs value is an API contract. A value type says: 'you can copy me, compare me, put me in maps.' A pointer type says: 'I have identity, I must be initialized, copy is not meaningful.' This contract must be documented and consistent across your codebase."

---

## Real-world use

- **Stripe's payment types:** `Amount`, `Currency`, `Timestamp` are value types (copyable, comparable, usable as map keys). `*PaymentIntent` is a pointer — it has lifecycle, it's large, and it's mutable.
- **Google's protobuf (generated Go code):** All generated message types use pointer receivers. The generator makes this decision for you — but knowing why helps you understand the generated code.
- **Uber's Go style guide:** Explicitly states "use pointer receivers consistently on a given type." It's a lint rule in their CI.
- **Go 1.22 loop fix:** Multiple large Go codebases (including CockroachDB and Kubernetes) had actual bugs from this. The Kubernetes fix involved hundreds of files.
- **Struct tags:** Every Go JSON API, ORM (GORM), and validation library (go-playground/validator) reads struct tags. Understanding them is prerequisite to Day 5.

---

## Interview Questions

1. What is the difference between a value receiver and a pointer receiver? When do you choose each?
2. Can you return the address of a local variable in Go? Why is this safe (unlike C)?
3. Explain the Go 1.22 loop variable fix. What was the old behavior? Show code that demonstrates the bug and the workaround.
4. What does embedding do in Go? Is it inheritance? What can and can't you do with it?
5. Given `switch v := i.(type)`, what is `v`'s type in each case branch? What is a type switch used for?
6. What does `break outer` do? When would you use a labeled break over a boolean flag?
7. Why does `for _, v := range mySlice { v = 99 }` not modify the slice? How do you modify elements in-place?

---

## Your tasks for today

Go to `../exercises/`. There are **3 beginner exercises** + **1 intermediate challenge** with starter files. Fill them in, run them, and tell me when done. I will review each like a production PR.

Don't open `../solutions/` until you've tried. I'll know. 😄
