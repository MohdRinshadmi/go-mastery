# Day 02 Interview Questions — Packages, Slices, Maps

---

**1. What are the three internal fields of a slice? What happens to each on `append` with sufficient vs insufficient capacity?**

<details><summary>Answer</summary>

A slice is a 3-word header: **pointer** (to element 0 of the backing array),
**len** (visible elements), **cap** (elements from the pointer to the end of the
array). On `append` with **sufficient** capacity (`cap > len`): no allocation —
the element is written into the existing array, the returned slice has the same
pointer and `len+1`. With **insufficient** capacity: a new, larger array is
allocated (typically ~2×), all elements are copied, and the returned slice points
to the new memory with a new cap.
</details>

---

**2. Explain the aliasing gotcha: when is `b := a[1:3]` dangerous? Show code that silently corrupts `a`.**

<details><summary>Answer</summary>

`b` shares `a`'s backing array, so writing through `b` writes through `a`, and an
`append` to `b` with spare capacity overwrites `a`'s elements:

```go
a := []int{1, 2, 3, 4, 5}
b := a[:3]        // len 3, cap 5
b = append(b, 99) // no alloc -> writes into a[3]
// a is now [1 2 3 99 5]
```
Fix by copying (`append([]int{}, a[:3]...)`) or cap-limiting (`a[:3:3]`).
</details>

---

**3. What is the zero value of a map vs a slice? Which panics on use, which is safe?**

<details><summary>Answer</summary>

Both zero values are `nil`. A `nil` slice is fully safe: `len`/`cap` are 0, you
can range it (zero iterations) and `append` to it. A `nil` map is safe to
**read** (returns the zero value) but **panics on write** — you must `make` it
first.
</details>

---

**4. Why does Go randomize map iteration order? What do you do when you need deterministic order?**

<details><summary>Answer</summary>

Randomization is deliberate: it stops code from accidentally depending on an order
the runtime never promised, which would break when the implementation changes. For
deterministic output, extract the keys into a slice, `sort` them, then iterate the
sorted slice.
</details>

---

**5. You need a set of strings in Go. Show the idiomatic implementation and explain the memory choice.**

<details><summary>Answer</summary>

```go
seen := make(map[string]struct{})
seen["hello"] = struct{}{}
_, exists := seen["hello"]
```
The value type is `struct{}`, the empty struct, which occupies **zero bytes**. The
map stores only keys, so there's no per-entry value overhead — the idiomatic Go
"set."
</details>

---

**6. `make([]int, 0, 1000)` vs `var s []int` — what's the difference when appending 1000 items?**

<details><summary>Answer</summary>

`make([]int, 0, 1000)` pre-allocates capacity for 1000 elements, so appending
1000 items causes **zero** reallocations/copies. `var s []int` starts at cap 0 and
grows by repeated doubling, triggering ~10+ reallocations and copies. Pre-sizing
when you know the count is a cheap, high-impact performance win.
</details>

---

**7. What does the three-index slice `s[1:3:3]` do differently from `s[1:3]`, and when would you use it?**

<details><summary>Answer</summary>

`s[1:3:3]` is `s[low:high:max]`: it sets `cap = max-low = 2`, equal to the length,
whereas `s[1:3]` leaves capacity extending to the end of the backing array. With
cap == len, the next `append` is **forced to allocate** a new array, so you can't
accidentally overwrite the original's tail. Use it to hand out a sub-slice safely
when the caller might append to it.
</details>

---

**8. What's the difference between exported and unexported identifiers, and how is it expressed?**

<details><summary>Answer</summary>

Visibility is by **capitalization**, not keywords: an identifier starting with an
uppercase letter is **exported** (visible to importers of the package); lowercase
is **unexported** (package-private). It applies to variables, constants, types,
functions, struct fields, and interface methods. Start unexported and promote to
exported only when something is genuinely part of your API.
</details>

---

**9. Why does `copy(dst, src)` sometimes copy fewer elements than `src` has?**

<details><summary>Answer</summary>

`copy` copies `min(len(dst), len(src))` elements and never grows `dst`. If `dst`
is shorter, only `len(dst)` elements are copied; it returns the count actually
copied. To copy everything you must size `dst` to at least `len(src)`.
</details>

---

**10. Why can't a slice be a map key, but an array can?**

<details><summary>Answer</summary>

Map keys must be **comparable** with `==`. Arrays are comparable when their
element type is (it's an element-by-element comparison on a fixed length). Slices
are **not** comparable (only `== nil` is allowed), because comparing variable-length
views with shared backing arrays has no well-defined value semantics — so slices
can't be keys.
</details>
