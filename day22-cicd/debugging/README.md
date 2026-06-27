# Day 22 Debugging — The flaky test that's green locally, red in CI

A unit test for `summarize()` (which builds a `k=v,k=v` string from a map of
feature flags) passes on every developer's laptop. Then CI fails it
*intermittently* — re-running the job makes it green again. Engineers learn to
"just hit retry", trust in the pipeline collapses, and a real regression
eventually slips through behind the noise.

This is the canonical **flaky test**: the code depends on Go's **randomized map
iteration order**.

## Symptom

```
$ cd bugged && go test -count=20 .
--- FAIL: TestSummarize (0.00s)
    main_test.go:18: summarize() = "region=eu,beta=on,cache=off", want "beta=on,cache=off,region=eu"
FAIL
```

It will not fail every time — that is exactly what makes it insidious. Run it a
few times with `-count=20` and you will eventually see a failure.

## Reproduce

```bash
cd bugged
go run .                 # shows summarize() yields MULTIPLE distinct outputs
go test -count=20 .      # flaps: passes sometimes, fails sometimes
```

`go run .` calls `summarize()` 1000 times and counts distinct outputs — proof
the function is non-deterministic without waiting for CI to flake.

## Hint

<details>
<summary>Hint</summary>

Go deliberately randomizes the order of `for k := range myMap`. The test asserts
one specific ordering. What must you do to the keys before joining them so the
output is the same on every run, on every machine?

</details>

## Solution & why

<details>
<summary>Solution & why</summary>

`summarize()` ranged the map directly:

```go
for k, v := range flags { parts = append(parts, k+"="+v) }
```

The Go runtime **randomizes map iteration order on purpose** (since Go 1.0) so
that code can't accidentally come to depend on a particular order. So the output
string's order is random, and an exact-match assertion is a coin flip.

**Fix:** collect the keys, `sort.Strings(keys)`, then iterate in sorted order.
The function is now deterministic and the test is stable:

```go
keys := make([]string, 0, len(flags))
for k := range flags { keys = append(keys, k) }
sort.Strings(keys)
for _, k := range keys { parts = append(parts, k+"="+flags[k]) }
```

**The general CI lesson:** flaky tests almost always come from a hidden source
of non-determinism — map order, `time.Now()`, goroutine scheduling, network,
or shared global state between tests. Pin them all down:
- Map order → sort keys.
- Time → inject a clock or use a fixed timestamp.
- Concurrency → run `go test -race`, don't assert on scheduling.
- Order-independence → run `go test -shuffle=on` in CI to catch tests that leak
  state into each other.

A flaky test should be treated as a Severity-1 bug, not retried away.

</details>
