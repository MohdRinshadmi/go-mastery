# Day 02 Debugging Challenge — The preview that corrupts the original

A `firstThree` helper returns the first three sensor readings as a "preview",
tagged with a `999` truncation marker. But after you call it, the caller's
original `readings` slice has been **silently mutated** — its 4th element is now
`999` — and a downstream sum comes out wildly wrong.

This is the **slice aliasing / append** trap, the single most important section
of today's lesson.

## Reproduce

```bash
cd bugged
go run .
```

Observed output:

```
preview: [10 20 30 999]
readings: [10 20 30 999 50]   <-- should be [10 20 30 40 50]
sum of readings: 1109         <-- should be 150
```

## Hint

`preview := readings[:3]` — what is `len(preview)`? What is `cap(preview)`? When
`append` has spare capacity, **where** does it write the new element?

<details>
<summary>Solution &amp; why</summary>

`readings[:3]` creates a slice with `len == 3` but `cap == 5`, because the
capacity of a sub-slice extends to the end of the underlying array. `preview`
shares `readings`' backing array.

When you `append(preview, 999)`, Go checks: is there spare capacity?
`cap (5) > len (3)`, yes — so it does **not** allocate a new array. It writes
`999` directly into the next slot of the shared backing array, which is
`readings[3]`. The caller's data is corrupted. The sum becomes
`10+20+30+999+50 = 1109`.

**Fixes (any one works):**

1. **Three-index slice** to cap the capacity at the length, forcing the next
   append to allocate:
   ```go
   preview := readings[:3:3] // len 3, cap 3 -> append must allocate
   ```

2. **Explicit copy** into independent memory (clearest intent):
   ```go
   preview := make([]int, 3)
   copy(preview, readings[:3])
   ```

3. **Append-to-empty clone:**
   ```go
   preview := append([]int{}, readings[:3]...)
   ```

The fixed program uses the copy approach. Now `readings` stays
`[10 20 30 40 50]` and the sum is `150`.

Rule of thumb: any time a function returns a slice derived from its input and may
append to it, decide whether the caller still owns the original. If yes, copy (or
cap-limit) before appending. Silent aliasing causes real production incidents.
</details>
