# Day 03 Interview Questions — Structs, Pointers, Control Flow

---

**1. Difference between a value receiver and a pointer receiver? When do you choose each?**

<details><summary>Answer</summary>

A **value receiver** gets a copy of the struct; it cannot mutate the original and
is suited to small, read-only types. A **pointer receiver** gets the address, so
it can mutate the original and avoids copying large structs. Choose pointer when:
you need mutation, the struct is large, or other methods on the type already use
pointer receivers (consistency). Choose value for small immutable value-semantic
types.
</details>

---

**2. Can you return the address of a local variable in Go? Why is it safe (unlike C)?**

<details><summary>Answer</summary>

Yes. The compiler does **escape analysis**: if a local's address escapes the
function (e.g. you return `&u`), it's allocated on the heap instead of the stack,
and the garbage collector keeps it alive as long as anything references it. In C
the stack frame is reclaimed on return so the pointer dangles; in Go it's safe and
idiomatic.
</details>

---

**3. Explain the Go 1.22 loop variable fix — old behavior, plus a bug and the workaround.**

<details><summary>Answer</summary>

Pre-1.22, the loop variable was a **single** variable reused across iterations, so
closures/goroutines captured a shared variable that held its final value:

```go
for i := 0; i < 3; i++ {
    funcs[i] = func() { fmt.Println(i) } // pre-1.22 prints 3,3,3
}
```
The workaround was to shadow per-iteration: `i := i`. In **Go 1.22+** each
iteration gets its own variable, so it correctly prints 0,1,2 and the workaround
is no longer needed.
</details>

---

**4. What does embedding do? Is it inheritance? What can and can't you do with it?**

<details><summary>Answer</summary>

Embedding places a type inside a struct with no field name, **promoting** its
fields and methods to the outer struct so you can call them directly. It is
**composition / forwarding, not inheritance**: the outer type is *not* a subtype,
so you can't pass a `Dog` where an `Animal` is expected. For polymorphism you use
interfaces. Name conflicts are resolved in favour of the outer type.
</details>

---

**5. Given `switch v := i.(type)`, what is `v`'s type in each branch, and what is a type switch for?**

<details><summary>Answer</summary>

In each `case` branch, `v` has the **specific type of that case** (e.g. in
`case int:` it's an `int`; in `case string:` a `string`). In the `default` branch
it keeps the original interface type. A type switch lets you branch on the dynamic
type stored in an interface value — common when handling heterogeneous data.
</details>

---

**6. What does `break outer` do? When would you use a labeled break over a boolean flag?**

<details><summary>Answer</summary>

A labeled `break outer` exits the loop tagged `outer:`, not just the innermost
loop/switch. Use it to break out of nested loops cleanly — it's clearer and less
error-prone than threading a `found` boolean through both loops and checking it at
each level.
</details>

---

**7. Why does `for _, v := range s { v = 99 }` not modify the slice? How do you modify elements in place?**

<details><summary>Answer</summary>

`v` is a fresh **copy** of each element, so assigning to it changes only the copy.
To modify the slice, index into it: `for i := range s { s[i] = 99 }`.
</details>

---

**8. Why is `var t time.Time` a footgun for "not set"?**

<details><summary>Answer</summary>

The zero value of `time.Time` is year 0001 (`0001-01-01 00:00:00 UTC`), **not**
"now" or "invalid." Code that treats the zero `time.Time` as "unset" works by
accident but reads confusingly. For an explicit "not set" use a pointer
`*time.Time` and check for `nil`.
</details>

---

**9. Why does `Counter{}.Inc()` fail to compile when `Inc` has a pointer receiver, but `c.Inc()` works?**

<details><summary>Answer</summary>

A pointer-receiver call needs the receiver's address. `c` is an **addressable**
variable, so Go rewrites `c.Inc()` to `(&c).Inc()`. `Counter{}` is a temporary
(non-addressable), so its address can't be taken and the call is a compile error.
The same applies to map values, which aren't addressable.
</details>

---

**10. What does Go's escape analysis decide, and how can you inspect it?**

<details><summary>Answer</summary>

Escape analysis decides whether a value lives on the **stack** (stays local) or
the **heap** (its address escapes the function). Stack allocation is cheaper and
GC-free. You can inspect the compiler's decisions with
`go build -gcflags='-m'`, which prints "escapes to heap" / "does not escape"
notes. This is why "always use pointers for speed" is a myth — pointers can force
heap allocation and cache-missing pointer chasing.
</details>

---

**11. Why are some structs not comparable with `==`?**

<details><summary>Answer</summary>

A struct is comparable only if **all** its fields are comparable. Fields of type
slice, map, or func are not comparable (only against `nil`), so any struct
containing one can't be compared with `==` and can't be used as a map key.
Structs of only comparable fields compare field-by-field.
</details>
