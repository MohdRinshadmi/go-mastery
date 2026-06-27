# Day 02 Pitfalls — Packages, Slices, Maps

**Trap → Why it bites → Fix.**

---

### 1. Slice aliasing: a sub-slice shares memory

**Trap**
```go
a := []int{1, 2, 3, 4, 5}
b := a[1:3]
b[0] = 99
fmt.Println(a) // [1 99 3 4 5] — a changed through b
```

**Why it bites** `b` is a *view* into `a`'s backing array, not a copy. Writing
through `b` writes through `a`.

**Fix** Copy when you need independence: `b := append([]int{}, a[1:3]...)` or
`make` + `copy`.

---

### 2. `append` with spare capacity overwrites the neighbour

**Trap**
```go
a := []int{1, 2, 3, 4, 5}
b := a[:3]            // len 3, cap 5
b = append(b, 99)    // no alloc — writes into a[3]!
fmt.Println(a)       // [1 2 3 99 5]
```

**Why it bites** When `cap > len`, `append` reuses the shared backing array
instead of allocating. The "new" element lands on top of the original owner's
data.

**Fix** Cap-limit with a three-index slice `a[:3:3]`, or copy first. If aliasing
is intentional for performance, comment it loudly.

---

### 3. Forgetting to assign the result of `append`

**Trap**
```go
append(s, 42)   // result discarded; s unchanged (or stale header)
```

**Why it bites** `append` returns a *new slice header* (possibly new pointer,
new len/cap). Ignoring the return throws that away.

**Fix** Always `s = append(s, 42)`.

---

### 4. `range` gives you a copy of each element

**Trap**
```go
type Point struct{ X, Y int }
pts := []Point{{1, 1}, {2, 2}}
for _, p := range pts {
    p.X = 99 // modifies the copy, not pts[i]
}
// pts unchanged
```

**Why it bites** The loop variable is a fresh copy each iteration; mutating it
doesn't touch the slice.

**Fix** Index in: `for i := range pts { pts[i].X = 99 }`.

---

### 5. Writing to a `nil` map panics

**Trap**
```go
var m map[string]int
m["k"] = 1 // panic: assignment to entry in nil map
```

**Why it bites** The zero value of a map is `nil`. Reads from nil maps are safe
(return zero), but writes panic. Easy to hit via a struct field you forgot to
initialize.

**Fix** `m := make(map[string]int)` (or a literal) before writing.

---

### 6. Reading a map without comma-ok hides "not found"

**Trap**
```go
age := ages["Carol"] // 0 — is Carol 0, or absent?
```

**Why it bites** A missing key returns the value type's zero. If zero is a valid
value, you can't tell "absent" from "present and zero."

**Fix** Use comma-ok: `age, ok := ages["Carol"]; if !ok { ... }`.

---

### 7. Depending on map iteration order

**Trap**
```go
for k, v := range m { write(k, v) } // order differs every run
```

**Why it bites** Go **deliberately randomizes** map iteration order. Tests, CSV
output, or audit logs that assume an order will flake.

**Fix** Collect keys into a slice, `sort` them, then iterate the sorted keys.

---

### 8. Maps are reference types — assignment shares the data

**Trap**
```go
a := map[string]int{"x": 1}
b := a
b["y"] = 2
fmt.Println(a) // map[x:1 y:2] — a changed too
```

**Why it bites** Assigning a map copies the header (a pointer), not the entries.

**Fix** To copy, iterate and copy entries, or use `maps.Clone` (Go 1.21+).
