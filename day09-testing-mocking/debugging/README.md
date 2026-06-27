# Debugging challenge — the false-green table test

> A test suite is **green**, the reviewer approves, the code ships — and a real
> bug rides along untouched. This is the most dangerous kind of test failure:
> the one that never fails.

## Symptom

`go test ./...` passes (exit 0) in `bugged/`, but `Classify(70)` returns
`"pass"` when the spec says it must return `"distinction"`. The test suite is
green yet a real bug ships to production.

The function contract:

```
score <  40        -> "fail"
40 <= score <  70  -> "pass"
score >= 70        -> "distinction"
```

`bugged/classify.go` has an **off-by-one**: it uses `score > 70` instead of
`score >= 70`, so the boundary value `70` is mislabeled. There is a six-case
table test that *should* catch this. It doesn't.

## Repro

Both directories are GREEN. That is the point — only one of them is actually
testing anything.

```bash
# Bugged: PASSES (exit 0) even though Classify(70) is wrong.
cd /Users/ioss/Documents/StudyProjects/GO/day09-testing-mocking/debugging/bugged && go test ./...
# ok  	day09-dbg/bugged

# Fixed: PASSES (exit 0) AND actually validates behavior.
cd /Users/ioss/Documents/StudyProjects/GO/day09-testing-mocking/debugging/fixed && go test ./...
# ok  	day09-dbg/fixed
```

Why is the green in `bugged/` dangerous? Look at the assertion inside the
subtest:

```go
got := Classify(tc.input)
if got == "" {                       // <-- asserts almost nothing
    t.Errorf("Classify(%d) returned empty string", tc.input)
}
```

The table carries a `want` column for every case, but the assertion never
reads it. `Classify` always returns a non-empty string, so the check can never
fire. The "boundary distinction" case (`input: 70, want: "distinction"`) passes
not because the answer is right, but because the test isn't looking.

## Proof the bug is real

Change the single assertion line in `bugged/classify_test.go` from
`if got == ""` to `if got != tc.want`, run `go test ./...`, and it turns RED:

```
--- FAIL: TestClassify/boundary_distinction (0.00s)
    classify_test.go:41: Classify(70) = "pass"; want "distinction"
```

That one boundary case is the bug the false-green test was hiding.

## Hint

Count the assertions in `bugged/classify_test.go`. The table has a `want`
column on every row — where is it actually used? A table-driven test that never
compares `got` to `tc.want` isn't testing the table; it's testing that the
function compiles. The "want" data is decoration.

<details>
<summary>Solution &amp; why</summary>

**Two defects, one root cause.** The function has an off-by-one, and the test
is written so it can never catch *any* wrong answer.

**1. The function bug (`classify.go`):**

```go
case score > 70:   // BUG: 70 falls through to the default "pass"
    return "distinction"
```

Fix:

```go
case score >= 70:  // 70 is a distinction, per the spec
    return "distinction"
```

**2. The test bug (`classify_test.go`) — the real lesson:**

```go
// BUGGED: collects tc.want but never asserts it.
if got == "" {
    t.Errorf("Classify(%d) returned empty string", tc.input)
}
```

```go
// FIXED: assert got against the per-case expected value from the table.
if got != tc.want {
    t.Errorf("Classify(%d) = %q; want %q", tc.input, got, tc.want)
}
```

A table-driven test earns its keep only when each subtest **asserts `got`
against `tc.want`** (and `tc.input` is actually fed to the function). If the
assertion is hard-coded — `got != "pass"`, `got != ""`, or it ignores `tc`
entirely — the table is theater: it *looks* thorough (six named cases in the
output) while testing nothing case-specific.

How to never ship this:

- Make the assertion reference `tc.want` (and the call reference `tc.input`).
  If neither appears in the subtest body, the test is suspect.
- Write a case you *expect to fail*, run it, watch it go red, then fix the code.
  A test you've never seen fail might be a test that can't fail.
- Check coverage of the boundary specifically — `go test -run
  TestClassify/boundary_distinction -v` should exercise the exact `70` path.

The corrected `fixed/` package does both: the off-by-one is `>= 70`, and the
test asserts `got != tc.want`, so the boundary case passes for the *right*
reason.

</details>
