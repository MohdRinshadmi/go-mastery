# Day 02 Notes — Quick Reference

## Packages & visibility
```go
package main   // produces an executable
package utils  // library package
```
- **Uppercase** identifier = exported (public API). **lowercase** = unexported.
- Applies to vars, consts, types, funcs, struct fields, interface methods.

## Arrays (value type, fixed length)
```go
var a [5]int                 // [0 0 0 0 0]
b := [3]string{"a", "b", "c"}
c := a                       // full copy — independent
```
- Length is part of the type: `[3]int` != `[4]int`.
- Rarely used directly; mostly the backing store for slices.

## Slices (3-word header: pointer, len, cap)
```go
s := []int{10, 20, 30}        // literal
s := make([]int, 3)           // len 3, cap 3
s := make([]int, 0, 100)      // len 0, cap 100 (pre-allocated)
var s []int                   // nil, but safe to append
sub := arr[1:3]               // shares arr's memory!
```

### append
```go
s = append(s, 42)             // ALWAYS assign the result
// cap > len -> writes in place (may alias!)
// cap == len -> allocates new array, copies
```

### Avoid aliasing
```go
b := a[1:3:3]                 // 3-index slice: cap == len, next append allocates
b := append([]int{}, a...)    // idiomatic clone
b := make([]int, len(a)); copy(b, a)
```

### copy & delete-in-place
```go
n := copy(dst, src)           // copies min(len(dst), len(src))
copy(s, s[1:]); s = s[:len(s)-1]  // delete element 0, allocation-free
```

### range copies the value
```go
for i, v := range s { }       // v is a COPY
s[i] = newValue               // to modify in place, index in
```

## Maps (reference type, hash table)
```go
m := map[string]int{"a": 1}
m := make(map[string]int, 100)   // pre-size hint
var m map[string]int             // nil — read OK, WRITE PANICS
```

### comma-ok (always for reads)
```go
v, ok := m["k"]               // ok == false if absent
```

### delete / iterate / set
```go
delete(m, "k")                // safe even if absent
for k, v := range m { }       // order RANDOMIZED — sort keys if you need order
set := map[string]struct{}{}  // idiomatic set; struct{} is zero-size
set["x"] = struct{}{}
```

### make vs new
- `make` — slices, maps, channels: returns an initialized, usable value.
- `new(T)` — returns `*T` to a zeroed value; rarely what you want for slices/maps.

## Key terms
- **Slice header** — {pointer, len, cap}; what a slice value actually is.
- **Aliasing** — two slices sharing one backing array.
- **Three-index slice** — `s[low:high:max]`, caps the capacity.
- **Comma-ok** — `v, ok := m[k]`, distinguishes absent from zero.
- **Reference type** — slices and maps copy a header/pointer, not the data.
- **Exported / unexported** — visibility by uppercase / lowercase first letter.
