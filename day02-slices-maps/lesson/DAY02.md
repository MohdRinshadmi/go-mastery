# Day 02 — Packages & Visibility, Arrays, Slices, Maps

> Mentor note: Today is where most Go newcomers get burned. Slices look like Python lists, they're not. Maps look like Python dicts / JS objects, they're not. The aliasing trap has caused real production incidents. Read every "Senior take" box and the Common Mistakes sections like they're incident reports — because they will be, if you skip them.

---

## Table of Contents
- [1. Packages & Visibility](#1-packages--visibility)
- [2. Arrays — the foundation you rarely use directly](#2-arrays--the-foundation-you-rarely-use-directly)
- [3. Slices — what you'll actually use every day](#3-slices--what-youll-actually-use-every-day)
- [4. The Slice Aliasing Gotcha](#4-the-slice-aliasing-gotcha)
- [5. make and copy](#5-make-and-copy)
- [6. Maps](#6-maps)
- [Expert Thinking Mode](#expert-thinking-mode)
- [Real-world use](#real-world-use)
- [Interview Questions](#interview-questions)

---

## 1. Packages & Visibility

### Theory
A **package** is the unit of code organization in Go. Every `.go` file starts with `package <name>`. A directory = one package. You import packages by path.

```go
package main   // the special package that produces an executable binary
package utils  // a library package — can be imported by others
```

### Exported vs Unexported — the capitalization rule
Go has **no** `public`, `private`, `protected`, or `internal` keywords. The rule is:
- **Starts with uppercase** → exported (visible outside the package)
- **Starts with lowercase** → unexported (package-private)

```go
// payment/processor.go
package payment

var maxRetries = 3           // unexported — only this package sees it
const Version = "1.0.0"     // exported — anyone who imports this package sees it

type processor struct { ... } // unexported type
type Processor struct { ... } // exported type

func validate(c Card) bool  { ... } // unexported helper
func Charge(c Card) error   { ... } // exported API
```

This applies to: variables, constants, types, functions, struct fields, and interface methods.

### Why it exists
Java has `public`/`private` on every declaration. Go replaces that with a single visual signal. You can scan any `.go` file and immediately know what's part of the public API vs internal detail — uppercase = public contract, lowercase = implementation.

### When to use
- **Exported:** anything that is intentionally part of your package's API surface.
- **Unexported:** everything else. Start unexported, promote to exported when needed. It's much easier to export something later than to un-export it (breaking change).

### Common mistakes
1. Exporting everything because "I might need it" — you create a wide API that you can never shrink.
2. Unexported struct with exported fields — the struct can't be constructed outside the package, but if someone gets a pointer, they can read/write those fields. Usually a bug.
3. **Struct tags need exported fields to work with JSON/encoding packages** (we'll hit this hard on Day 5).

> **Senior take:** Treat exported identifiers as a promise. Every exported name is API. API has consumers. API changes break consumers. Keep your API surface minimal and intentional. At Stripe/Google/etc., code reviews require justification for new exported names.

---

## 2. Arrays — the foundation you rarely use directly

### Theory
Fixed-length sequence of elements of the same type. Length is **part of the type**.

```go
var scores [5]int                   // [0 0 0 0 0] — zero-valued
names := [3]string{"Alice", "Bob", "Carol"}
matrix := [2][3]int{{1,2,3},{4,5,6}} // 2D array
```

### Key properties
- **Value type** — assigning an array copies all elements. Passing to a function copies all elements.
- Length is compile-time constant. `[3]int` and `[4]int` are **different types** — you can't assign one to the other.
- `len(arr)` gives the length. There is no `cap` — arrays are fixed.

```go
a := [3]int{1, 2, 3}
b := a          // full copy — b is independent
b[0] = 99
fmt.Println(a)  // [1 2 3] — a is unchanged
```

### When to use arrays
Rarely. The main cases:
- **Cryptographic hashes:** `[32]byte` is a SHA-256 hash — its fixed size is semantic.
- **Fixed-size buffers** in hot paths where you want stack allocation.
- **Matrices** in certain math/graphics code.

For everything else: use slices.

> **Senior take:** Arrays in Go exist mostly as the backing store for slices. You'll go months without declaring an array directly, but you need to understand them to understand slices.

---

## 3. Slices — what you'll actually use every day

### Theory
A slice is a **view into an underlying array**. It has three components:

```
┌────────────┐
│  pointer   │ → points to element 0 of the underlying array
│  len       │ → number of elements you can see
│  cap       │ → elements from pointer to end of array
└────────────┘
```

This is not just conceptual — the Go runtime represents every slice as this exact 3-word struct. Understanding it explains every behavior below.

### Creation

```go
// Literal — Go allocates an array [5]int behind the scenes
s := []int{10, 20, 30, 40, 50}

// From array — slice of the array (shares memory!)
arr := [5]int{10, 20, 30, 40, 50}
s := arr[1:3]  // [20 30], len=2, cap=4 (from index 1 to end of arr)

// make — allocate a new array of given capacity
s := make([]int, 3)     // len=3, cap=3,  [0 0 0]
s := make([]int, 3, 10) // len=3, cap=10, [0 0 0]

// nil slice — valid zero value
var s []int              // s == nil, len=0, cap=0, safe to append to
```

### len and cap

```go
s := make([]int, 3, 10)
fmt.Println(len(s), cap(s)) // 3 10

// Reslice: you can extend up to cap without allocation
s2 := s[:5]  // len=5, cap=10 — same underlying array
```

### append — the critical behavior

`append(slice, elements...)` returns a **new slice header**. Two cases:

**Case 1 — capacity is sufficient:** appends in place, returns slice with same pointer + len+1.
**Case 2 — capacity is insufficient:** allocates a new, larger array (usually doubles capacity), copies all elements, returns slice pointing to new memory.

```go
s := make([]int, 3, 4)  // len=3, cap=4 — one slot free
s = append(s, 99)       // Case 1: no alloc, same array, len=4, cap=4
s = append(s, 100)      // Case 2: new array allocated, copied, len=5, cap=8
```

**Always assign the result of append:**
```go
// WRONG — you discard the new slice header
append(s, 42)

// RIGHT
s = append(s, 42)
```

### Iterating

```go
for i, v := range s {    // i = index, v = copy of element
    fmt.Println(i, v)
}
for _, v := range s { }  // ignore index
for i := range s { }     // ignore value — just index
```

> **Range gives you a copy of the value.** Modifying `v` does NOT modify the slice. To modify in-place: `s[i] = newValue`.

### Common mistakes — slices

1. **Forgetting that range gives a copy:**
   ```go
   type Point struct{ X, Y int }
   points := []Point{{1,1},{2,2}}
   for _, p := range points {
       p.X = 99 // modifies the copy, not points[i]
   }
   // points is unchanged — classic bug
   ```

2. **Growing with cap-unaware make:**
   ```go
   // If you know you'll add 10000 items:
   s := make([]int, 0, 10000) // pre-allocate — zero len, big cap
   // vs
   var s []int  // 12+ reallocations as it grows
   ```

3. **Three-index slice** (advanced, prevents the aliasing gotcha):
   ```go
   s := data[2:4:4] // s[low:high:max] — cap is limited to max-low
   ```

> **Senior take:** Every time you use `append` or `make`, think: "do I know the final size?" If yes, pre-allocate with `make([]T, 0, n)`. This is one of the highest-impact, lowest-effort performance wins in Go. Stripe and Uber engineers get PR comments about this regularly.

---

## 4. The Slice Aliasing Gotcha

This is the most important section in today's lesson. Skim it and you'll have a production bug inside a month.

### The problem

```go
a := []int{1, 2, 3, 4, 5}
b := a[1:3]  // b = [2, 3], but b shares the underlying array with a!

b[0] = 99    // ALSO modifies a[1] ← this will surprise you
fmt.Println(a) // [1 99 3 4 5]  ← a was changed through b
```

`b` is not a copy. It is a window into `a`'s memory.

### The append trap

```go
a := []int{1, 2, 3, 4, 5}
b := a[:3]  // [1 2 3], cap=5 (cap extends to end of a's array)

b = append(b, 99)  // len < cap, so NO new allocation
                   // writes 99 into a[3] — overwrites a's element!
fmt.Println(a)     // [1 2 3 99 5] ← a was silently mutated
fmt.Println(b)     // [1 2 3 99]
```

When `b` still has capacity (cap > len), `append` writes into the underlying array that `a` also points to. **No new allocation = no new memory = you've just silently modified `a`.**

### When this triggers in real code

```go
func process(data []int) []int {
    result := data[:0]          // "clever" reslice to reuse buffer
    for _, v := range data {
        if v > 0 {
            result = append(result, v)  // writes into data's array!
        }
    }
    return result
}
```

This pattern looks smart (reusing the buffer), but it corrupts `data` for the caller. This exact pattern is a real source of production bugs in Go codebases.

### How to prevent aliasing

**Option 1 — Explicit copy:**
```go
b := make([]int, len(a))
copy(b, a)
// Now b is fully independent
```

**Option 2 — Three-index slice (cap limiting):**
```go
b := a[1:3:3]  // cap of b = 3-1 = 2, same as len
               // any append to b will allocate a new array
               // a is safe
```

**Option 3 — Use append to copy:**
```go
b := append([]int{}, a...)  // idiomatic one-liner clone
```

> **Senior take:** Any time a function returns a slice derived from its input, ask: "does the caller still own the original?" If yes, copy before modifying. If you're building a high-performance path and aliasing is intentional, document it loudly with a comment. Silent aliasing causes incidents.

---

## 5. make and copy

### make — allocate composite types

`make` is only for slices, maps, and channels. It returns an initialized (non-nil) value.

```go
s := make([]int, len, cap)    // slice: len elements pre-zeroed, capacity reserved
m := make(map[string]int)     // map: empty, ready to use
ch := make(chan int, 10)       // buffered channel (Day 9)
```

Why `make` and not `new`? `new([]int)` gives you a pointer to a nil slice — almost never what you want. `make([]int, 0)` gives you an initialized, usable slice.

### copy — safe element transfer

```go
src := []int{1, 2, 3, 4, 5}
dst := make([]int, 3)           // only 3 slots
n := copy(dst, src)             // copies min(len(dst), len(src)) elements
fmt.Println(n, dst)             // 3 [1 2 3]
```

`copy` always copies `min(len(dst), len(src))` elements. It never grows `dst`. It returns the count copied.

**Overlap is safe:** `copy` handles overlapping slices from the same array correctly (left shifts, right shifts).

```go
// Shift elements left by 1 (delete element at index 0)
s := []int{10, 20, 30, 40, 50}
copy(s, s[1:])            // s = [20 30 40 50 50]
s = s[:len(s)-1]          // s = [20 30 40 50]
```

> **Senior take:** Memorize the delete pattern above — it's the idiomatic way to remove an element from a slice without allocating. It modifies in place and is O(n) but allocation-free.

---

## 6. Maps

### Theory
A map is an unordered collection of key→value pairs. Internally it's a hash table. Keys must be comparable (any type that supports `==`): strings, ints, bools, pointers, arrays (not slices).

```go
// Literal
ages := map[string]int{
    "Alice": 30,
    "Bob":   25,
}

// make (preferred when you know approximate size)
ages := make(map[string]int, 100) // hint: ~100 entries

// Zero value is nil — cannot write to nil map
var m map[string]int   // m == nil
m["key"] = 1           // PANIC: assignment to entry in nil map
```

### The comma-ok idiom — always use it for reads

```go
age, ok := ages["Carol"]  // ok = false if key absent
if ok {
    fmt.Println(age)
} else {
    fmt.Println("not found")
}

// Without comma-ok, you get the zero value (0 for int) — silently wrong
age := ages["Carol"] // returns 0 — is Carol 0 years old or not found?
```

The zero value read is a trap: if your values can legitimately be zero (0, "", false), you cannot distinguish "not found" from "found with zero value" without comma-ok.

### Deleting entries

```go
delete(ages, "Bob")  // no-op if key doesn't exist — safe
```

### Iteration — order is randomized deliberately

```go
for k, v := range ages {
    fmt.Println(k, v)  // order is NOT guaranteed between runs
}
```

Go **deliberately randomizes** map iteration order to prevent accidental dependence on ordering. If you need sorted output, extract keys, sort, then iterate:

```go
keys := make([]string, 0, len(ages))
for k := range ages {
    keys = append(keys, k)
}
sort.Strings(keys)
for _, k := range keys {
    fmt.Println(k, ages[k])
}
```

### Maps are reference types

Like slices, assigning a map copies the header (a pointer), not the data:

```go
a := map[string]int{"x": 1}
b := a          // b and a point to the SAME map
b["y"] = 2
fmt.Println(a)  // map[x:1 y:2] — a was modified through b
```

To copy: iterate and copy, or use `maps.Clone` (Go 1.21+).

### Common mistakes — maps

1. **Writing to nil map (runtime panic):**
   ```go
   type Config struct { Options map[string]string }
   var c Config
   c.Options["key"] = "val"  // PANIC — Options is nil
   // Fix: c.Options = make(map[string]string)
   ```

2. **Relying on iteration order** for anything deterministic (tests, CSV output, audit logs).

3. **Concurrent reads+writes without synchronization** → data race → undefined behavior. Use `sync.RWMutex` or `sync.Map` for concurrent access (Day 9).

4. **Using map as a set:** Go has no built-in set. The idiom is `map[T]struct{}`:
   ```go
   seen := make(map[string]struct{})
   seen["hello"] = struct{}{}
   _, exists := seen["hello"]  // comma-ok works perfectly
   ```
   `struct{}` has zero size — the map stores only keys with no memory overhead for values.

### Performance implications

- Map lookups are O(1) amortized but have real overhead: hashing, collision resolution, cache misses.
- For small sets (<10 entries), a linear scan over a slice can be faster than a map due to cache locality.
- Pre-size with `make(map[K]V, n)` when you know approximate entry count. Avoids incremental rehashing.
- Maps are **not concurrency-safe**. The runtime detects concurrent map writes and panics (in Go 1.6+, concurrent reads are safe but concurrent write+read panics too via the race detector).

> **Senior take:** A map with `struct{}` values is the idiomatic Go set. But benchmark before assuming map beats slice for tiny collections — CPU caches matter. In hot paths at Cloudflare/Uber, sorted-slice binary search has beaten small maps in benchmarks.

---

## Expert Thinking Mode — how different levels see slices & maps

- **Beginner:** "A slice is just a list, a map is just a dictionary. I'll use them like Python."
- **Senior:** "A slice is a view into an array. append may or may not allocate. I pre-allocate with `make` when I know the size. I never assume a sub-slice is independent memory."
- **Staff:** "I instrument allocation with `go tool pprof`. I know which hot paths are causing GC pressure from slice grows. I profile before optimizing, but I understand *why* to pre-allocate."
- **Architect:** "Data structure choice is an API decision. A `map[string][]Event` is a different contract than `[]Event` with a sorted key. Changing it later is a migration. Think about access patterns first."

---

## Real-world use

- **Uber's dispatch engine:** Rider→driver assignments use maps keyed by driver ID, slices for sorted queues. They pre-allocate slices with `make([]Ride, 0, expectedBatchSize)` in their hot loop — that one change reduced GC pauses measurably.
- **Cloudflare's DNS resolver:** Uses `map[string]struct{}` as blocklists (millions of entries). They measured that `make(map[string]struct{}, len(entries))` at startup vs growing incrementally saved ~30% of initialization time.
- **Stripe's idempotency layer:** Tracks seen request IDs with a `map[string]time.Time`. The nil-map panic is in their onboarding incident list as "the most common first Go bug our engineers hit."
- **Google's internal RPC:** Protocol buffer slices are always pre-allocated from the decoded message size — they know length upfront, so zero extra allocations.

---

## Interview Questions

1. What are the three internal fields of a slice? What happens to each when you call `append` and there is sufficient capacity vs. insufficient capacity?
2. Explain the aliasing gotcha: when is `b := a[1:3]` dangerous? Show code that silently corrupts `a`.
3. What is the zero value of a map? What is the zero value of a slice? Which panics if you use it, and which is safe?
4. Why does Go randomize map iteration order? What do you do when you need deterministic order?
5. You need a set of strings in Go. Show me the idiomatic implementation and explain the memory design choice.
6. You call `make([]int, 0, 1000)` — what's different from `var s []int` in terms of allocations when you append 1000 items?
7. What does the three-index slice `s[1:3:3]` do differently from `s[1:3]`, and when would you use it?

---

## Your tasks for today

Go to `../exercises/`. There are **3 beginner exercises** + **1 intermediate challenge** with starter files. Fill them in, run them, and tell me when done. I will review each like a production PR.

Don't open `../solutions/` until you've tried. I'll know. 😄
