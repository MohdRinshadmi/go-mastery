# Challenge 01 — The vanishing column

**Phase 1 · Fundamentals · slices**

## Symptom

We parse a CSV-ish log line into fields, then build two derived rows from it: a "raw" row we keep as-is, and an "enriched" row where we append a computed `status` column. We expect the raw row to stay untouched.

It doesn't. After enriching, the *raw* row has mysteriously gained an extra field — and sometimes the enriched row's first columns get clobbered too. Run it:

```bash
cd bugged
go run .
```

Expected:
```
raw:      [2026-06-27 login user=42]
enriched: [2026-06-27 login user=42 OK]
```

Actual: the raw row is corrupted, or the enriched data leaks back into it.

## Hint

`append` does not always allocate. When the slice has spare capacity, it writes *in place* into the existing backing array. Two slices that share a backing array are not independent — mutating one through `append` can stomp the other. Where did the second slice come from? What's its capacity? (`cap()` is your friend here.)

## How to reproduce

`go run .` in `bugged/`. Add a `fmt.Println(len(fields), cap(fields))` before the append to *see* the spare capacity that makes this misbehave.

---

<details>
<summary><strong>Solution &amp; why</strong></summary>

### Root cause

`enriched := fields` does **not** copy the slice — it copies the slice *header* (pointer, len, cap), so both names point at the **same backing array**. The buggy code also takes a sub-slice (`fields[:len(fields)]`) that still carries the original's full capacity.

When you then `append(enriched, status)` and the backing array has spare capacity (because the slice was created with `make([]string, n, n+4)` or via a low-len/high-cap sub-slice), `append` writes the new element **into the shared array** instead of allocating a fresh one. Now `raw` — which aliases the same array — sees the change. Worse, length/visibility differences mean writes through one slice silently appear (or disappear) in the other.

This is the #1 Go slice gotcha. The rule:

> Two slices that share a backing array are not safe to `append` to independently. `append`'s "maybe it reallocates, maybe it doesn't" behavior makes the bug intermittent and capacity-dependent — the worst kind.

### The fix

Make a genuinely independent copy before appending. The idiomatic, allocation-explicit way:

```go
enriched := make([]string, len(fields), len(fields)+1)
copy(enriched, fields)
enriched = append(enriched, status)
```

Or the modern three-index full-slice expression that forces the next `append` to reallocate:

```go
enriched := append(fields[:len(fields):len(fields)], status) // cap == len, so append MUST allocate
```

`fields[low:high:max]` caps the result's capacity at `max-low`. By setting `max == len`, the new slice has zero spare capacity, so the very next `append` is guaranteed to allocate a new array and leave `fields` alone.

`fixed/main.go` uses the explicit `make`+`copy` form because it reads clearly to a junior — but know the three-index trick; it shows up in code review and interviews.

</details>
