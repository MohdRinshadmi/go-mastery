# Day 09 — Pitfalls (Testing &amp; Mocking)

Seven traps that turn a green test suite into false confidence. Format:
**Trap → Why → Fix.**

---

### 1. The table test that never reads `tc`

**Trap:** You loop `for _, tc := range tests` and inside the subtest you assert
against a hard-coded constant (`got != "pass"`, `got != ""`) or call the
function with literal args instead of `tc.input` / `tc.want`.

**Why it bites:** Every case "passes" regardless of the data. The table *looks*
exhaustive — six named subtests in the output — but it's testing nothing
case-specific. Real bugs (off-by-ones at boundaries) ride straight through. See
`debugging/` for a worked example.

**Fix:** The subtest body must reference both `tc.input` (fed to the function)
and `tc.want` (compared to the result): `if got != tc.want { ... }`. If neither
appears, the table is decoration. Write one case you expect to fail and confirm
it goes red before trusting the rest.

---

### 2. Shared mutable state across subtests

**Trap:** A package-level variable, a shared map, or a struct field mutated
inside one subtest and read by another.

**Why it bites:** Subtests can run in any order, and with `t.Parallel()` they
run concurrently. State leaks make case B pass only because case A ran first —
or fail intermittently under `-race`. The result is flaky tests, which erode
trust until the team starts ignoring red.

**Fix:** Construct fresh inputs/fakes *inside* each subtest. No package-level
mutable test state. Use `t.Cleanup` to reset anything genuinely shared. Run
`go test -race ./...` to surface accidental sharing.

---

### 3. Tests that depend on execution order

**Trap:** `TestCreate` inserts a row, `TestRead` assumes it's there. Or tests
rely on alphabetical function ordering.

**Why it bites:** `go test` makes no ordering guarantee across packages, caches
per-package results, and may parallelize. Reorder a file, add a case, run a
single test with `-run`, and the chain breaks. CI fails on a change unrelated to
the broken test.

**Fix:** Each test is fully self-contained: it sets up its own world and tears
it down with `t.Cleanup`. You should be able to run any single test in
isolation (`go test -run TestRead`) and have it pass.

---

### 4. Testing the implementation instead of the behavior

**Trap:** Asserting that an internal method was called N times, or pinning exact
private struct layout, rather than checking the observable output/contract.

**Why it bites:** The test breaks on every refactor even when behavior is
unchanged. People stop refactoring to avoid touching tests, or they update the
test to match the new internals without re-checking correctness — the test
stops being a spec.

**Fix:** Assert on outputs, returned errors, and externally visible effects.
Mock at the *boundary* (the narrow interface), not on internal call counts. Ask:
"if I rewrote the internals but kept the contract, should this test still pass?"
If not, it's testing the wrong thing.

---

### 5. Real network / DB / clock in a "unit" test

**Trap:** The unit test makes an HTTP call, hits a real database, or reads
`time.Now()` directly.

**Why it bites:** Slow, flaky, and non-hermetic — fails on a plane, in CI
without network, or at midnight when a date rolls over. Time-based logic
(`expires after 24h`) is untestable without waiting. Flaky tests are an outage
of the dev pipeline.

**Fix:** Inject dependencies behind small interfaces (`Store`, `Clock`,
`Charger`) and pass hand-written fakes in tests. For time, inject a `now func()
time.Time` or a `Clock` interface and return a fixed instant. Keep real
network/DB for explicitly-labeled integration tests gated by `-short`.

---

### 6. Forgetting `t.Helper()` in a test helper

**Trap:** You extract `assertOK(t, got, want)` but omit `t.Helper()` inside it.

**Why it bites:** When the assertion fails, the reported file:line points at the
helper's internals, not the calling test case. With a table test you can't tell
*which case* failed, so debugging is a hunt.

**Fix:** Call `t.Helper()` as the first line of every helper that takes a
`*testing.T` (or `*testing.B`). Failures then report the caller's line. Also use
`t.Fatal` inside helpers sparingly — it only stops the goroutine it runs on.

---

### 7. Chasing 100% coverage

**Trap:** Treating the coverage number as the goal and writing tests for trivial
getters, generated code, and unreachable branches to hit 100%.

**Why it bites:** Coverage measures lines *executed*, not behavior *verified* —
a false-green test (trap 1) can run a line and assert nothing, inflating the
number while testing nothing. Chasing the last few percent wastes effort on
low-value code and breeds brittle tests.

**Fix:** Target the logic and edge cases — boundaries, error paths, the tricky
branches. Use `go test -coverprofile=c.out && go tool cover -html=c.out` to find
*untested logic*, not to worship a percentage. 80% of meaningful coverage beats
100% of line-touching.

---

### Bonus — loop variable capture (pre-Go 1.22 history)

**Trap:** In older Go, `t.Run(tc.name, func(t *testing.T) {...})` with
`t.Parallel()` inside captured the loop variable `tc` by reference; by the time
the parallel subtest ran, the loop had advanced and every subtest saw the last
case.

**Why it mattered:** All parallel cases tested the same (final) row — another
silent false-green. The classic fix was `tc := tc` to shadow per-iteration.

**Fix:** Go 1.22 changed loop-variable semantics so each iteration gets a fresh
`tc` — the `tc := tc` dance is no longer needed. But know the history: you'll see
it in older codebases, and it explains why parallel table tests once needed it.
