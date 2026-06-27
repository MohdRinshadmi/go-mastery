# Day 06 Pitfalls — Methods & Interfaces

Each entry: **Trap → Why → Fix**.

---

### 1. A nil interface is not always `nil`

**Trap:** A function returns `error`, you returned a nil `*MyError` variable, and
the caller's `if err != nil` is `true` even though "nothing went wrong".

**Why:** An interface value is a two-word pair `(type, value)`. Boxing a nil
concrete pointer gives `(*MyError, nil)` — the type word is set, so the interface
is not equal to `nil`. Only `(nil, nil)` compares equal to `nil`.

**Fix:** Return the untyped `nil` literal directly on the success path. Never
return a concrete typed nil into an interface:
```go
func validate() error {
    if bad { return &MyError{...} }
    return nil // not: return someNilTypedPointer
}
```

---

### 2. Value doesn't satisfy an interface whose method has a pointer receiver

**Trap:** `var s Sizer = Box{5}` fails to compile with "Box does not implement
Sizer (Size method has pointer receiver)", even though `Size()` clearly exists.

**Why:** The method set of `Box` (value) includes only value-receiver methods.
Pointer-receiver methods belong to `*Box` only. So `Box` does not satisfy
`Sizer`, but `*Box` does.

**Fix:** Pass a pointer: `var s Sizer = &Box{5}`. When in doubt, take the address.

---

### 3. Calling a pointer-receiver method on a non-addressable value

**Trap:** `Counter{}.Inc()` or `m["k"].Inc()` won't compile: "cannot call
pointer method on ... / cannot take the address of ...".

**Why:** Calling a pointer-receiver method on a value requires Go to take its
address (`(&v).Inc()`). That auto-address only works on **addressable** values.
Composite literals and map-index expressions are not addressable.

**Fix:** Store in an addressable variable first, or use a pointer:
```go
c := Counter{}; c.Inc()        // ok: c is addressable
v := m["k"]; v.Inc(); m["k"] = v // map values: copy out, mutate, write back
```

---

### 4. Mixing value and pointer receivers on the same type

**Trap:** Some methods use `(t T)`, others use `(t *T)`. Now `T` satisfies some
interfaces and `*T` satisfies others, and copies silently lose mutations.

**Why:** Mixed receivers split the method set in confusing ways and make it
ambiguous whether a method mutates the original or a copy.

**Fix:** Pick one receiver kind per type. If any method needs a pointer receiver
(mutation, large struct), make **all** methods pointer receivers for consistency.

---

### 5. A type assertion without comma-ok panics

**Trap:** `f := r.(*os.File)` panics with "interface conversion" when `r` holds a
different concrete type, crashing the program.

**Why:** The single-result form of a type assertion panics on mismatch. It only
belongs where you are certain of the dynamic type.

**Fix:** Use the comma-ok form and handle the failure:
```go
if f, ok := r.(*os.File); ok {
    use(f)
}
```
Use a type switch when you need to branch on several concrete types.

---

### 6. Over-broad interfaces and constant type assertions

**Trap:** You take an `any` (or a fat 10-method interface), then immediately
type-assert/type-switch inside to recover the real behavior.

**Why:** Frequent assertions mean the interface doesn't express the behavior you
actually need. You've pushed type checking from compile time to run time —
exactly backwards. Big interfaces are weak abstractions.

**Fix:** Shrink the interface to the methods the consumer calls (often one
method). Define it in the **consumer** package. Let the compiler check
satisfaction instead of asserting at run time.

---

### 7. Forgetting `Read` returns `n > 0` together with `io.EOF`

**Trap:** Reading from an `io.Reader`, you check `err == io.EOF` and discard the
buffer, dropping the final bytes.

**Why:** `Read` may return `n > 0` **and** a non-nil error (including `io.EOF`) in
the same call. The error does not mean "no data this time".

**Fix:** Always process the `n` bytes first, then inspect the error:
```go
n, err := r.Read(buf)
process(buf[:n])
if err == io.EOF { break }
if err != nil { return err }
```
