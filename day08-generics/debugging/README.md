# Debugging Challenge — The Bare-Zero Generic Lookup

A generic map-lookup helper that compiles, runs, and quietly lies to you.

## Symptom

A scoreboard reports a player who **never played** as having "scored 0 points" —
identical to a player who genuinely scored 0. The lookup helper has no way to tell
"key missing" from "key present with the zero value."

## Repro

Bugged (wrong / ambiguous output):

```sh
cd /Users/ioss/Documents/StudyProjects/GO/day08-generics/debugging/bugged && go run .
```

Fixed (correct output — "not found"):

```sh
cd /Users/ioss/Documents/StudyProjects/GO/day08-generics/debugging/fixed && go run .
```

## Hint

Look at the return type of `Get[T any](m map[string]T, key string) T`. What does
`m[key]` give you when `key` isn't in the map? A plain map index has a *second*
return value you can ask for — but this generic signature throws it away. The zero
value of `T` is not a sentinel: `0`, `""`, and `nil` are all legitimate stored values.

<details>
<summary>Solution & why</summary>

**The bug.** `Get` returns a bare `T`:

```go
func Get[T any](m map[string]T, key string) T {
	return m[key] // missing key -> zero value of T, no signal
}
```

When `key` is absent, Go's map indexing returns the **zero value** of the element
type — `0` for `int`, `""` for `string`, `nil` for pointers/slices/maps. That zero
value is *ambiguous*: for an `int` scoreboard, `0` is also a perfectly real score.
So a missing player ("Zoe") and a real zero-scoring player ("Dana") come back
identical, and the caller reports both as "scored 0 points."

This is the generics version of a classic trap. The single-return map index hides
the same information; the standard Go idiom is the **comma-ok** form
(`v, ok := m[key]`), and your generic wrapper has to *thread that bool through*.

**The fix.** Return `(T, bool)`:

```go
func Get[T any](m map[string]T, key string) (T, bool) {
	v, ok := m[key]
	return v, ok
}
```

Now the caller can branch on `ok`:

```go
score, ok := Get(scores, name)
if !ok {
	// genuinely missing
}
```

**Takeaway.** With `[T any]` you cannot invent a "missing" sentinel value — every
possible value of `T` might be a legitimate stored value, and the zero value is the
worst candidate of all because it shows up the most. Any generic lookup, cache get,
pop, or "find" must return `(T, bool)` (or `(T, error)`), never a bare `T`. Thread
the comma-ok through the signature rather than guessing from the zero value.

</details>
