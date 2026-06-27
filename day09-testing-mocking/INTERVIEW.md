# Day 09 — Interview Questions (Testing &amp; Mocking)

Ten questions with model answers. The first seven are the lesson's; the last
three go a level deeper. Answer out loud before expanding.

---

### 1. How does Go's testing differ from JUnit/pytest? What's the file/function convention?

<details>
<summary>Answer</summary>

Testing is built into the toolchain — `go test` is the runner; there's no
separate framework, no annotations, no XML config. Convention *is* the
framework:

- Test files end in `_test.go` and live next to the code, usually in the same
  package (`payment.go` → `payment_test.go`).
- Test functions are `func TestXxx(t *testing.T)` (exported name, `Test`
  prefix, capital next letter).
- `go test ./...` discovers and runs everything; results are cached per package.

There are no assertion macros in the standard library — you write plain `if`
checks and call `t.Errorf`. Because it's frictionless and ships in the box, the
cultural norm on Go teams is that untested code doesn't merge.

</details>

---

### 2. What is a table-driven test and why is it the Go idiom? What does `t.Run` add?

<details>
<summary>Answer</summary>

A table-driven test is one test function with a slice of case structs
(`name`, inputs, `want`, often `wantErr`), looped over with `t.Run` creating a
subtest per case:

```go
for _, tc := range tests {
    t.Run(tc.name, func(t *testing.T) {
        got, err := Fn(tc.input)
        ...
        if got != tc.want { t.Errorf("got %v, want %v", got, tc.want) }
    })
}
```

It's the idiom because adding a case is one line, the shape is instantly
readable to any Go reviewer, and `t.Run` gives each case its own name in output
(`TestFn/by_zero`), isolation, and the ability to run one case via
`-run TestFn/by_zero`. The critical discipline: assert against `tc.want`, not a
constant — otherwise the table is theater (see the Day 09 debugging challenge).

</details>

---

### 3. `t.Error` vs `t.Fatal` — when each?

<details>
<summary>Answer</summary>

- `t.Error` / `t.Errorf`: record the failure and **keep going**. Use it for
  independent assertions where the rest of the test still yields useful
  information — e.g. checking several fields of a result.
- `t.Fatal` / `t.Fatalf`: record the failure and **stop this test immediately**
  (it calls `runtime.Goexit`). Use it when continuing makes no sense or is
  unsafe — e.g. a setup step failed, or you got a `nil` you'd dereference on the
  next line.

Rule of thumb: `Fatal` for "can't proceed" (setup, guards before a deref),
`Error` for "this assertion is wrong but the next is still meaningful."
`t.Fatal` only stops the goroutine it runs in, so don't call it from a helper
goroutine.

</details>

---

### 4. How do you mock a dependency in Go without a mocking framework? How does interface design enable it?

<details>
<summary>Answer</summary>

You define a **small interface** for the dependency and write a fake that
implements it. Production code accepts the interface; the test passes the fake:

```go
type Charger interface { Charge(amount int) (string, error) }

func Checkout(c Charger, amount int) (string, error) { ... }

type fakeCharger struct{ wantErr bool }
func (f fakeCharger) Charge(amount int) (string, error) {
    if f.wantErr { return "", errors.New("gateway down") }
    return "txn_123", nil
}
```

This is *why* "accept interfaces, return structs" matters — narrow interfaces
are what make code testable without network or DB. Hand-written fakes give full
control over success and failure modes. If you'd need a 20-method mock, the
interface is too big — that's a design smell, not a tooling gap.

</details>

---

### 5. Why must tests be independent of order? What runs them in parallel?

<details>
<summary>Answer</summary>

`go test` gives no guarantee about ordering across packages, caches results per
package, and can run packages concurrently; within a package, subtests that call
`t.Parallel()` run concurrently with each other. If `TestB` relies on state left
by `TestA`, then reordering, filtering with `-run`, or parallelism breaks it —
producing flaky failures unrelated to real bugs.

Independence means each test builds its own inputs and fakes and tears them down
with `t.Cleanup`, so any test passes when run alone (`go test -run TestB`).
`t.Parallel()` opts a test into concurrent execution; `-parallel N` caps how
many run at once.

</details>

---

### 6. What does `go test -race` do and why run it in CI?

<details>
<summary>Answer</summary>

`-race` enables the race detector: an instrumented build that watches memory
accesses at runtime and reports when two goroutines touch the same location
concurrently without synchronization, with one of them writing. It catches data
races that are invisible in normal runs and may only manifest under production
load.

Run it in CI because races are timing-dependent and often pass locally yet
corrupt data or crash in prod. It's a runtime detector — it only flags races on
code paths your tests actually exercise — so it pairs with good coverage. It
adds CPU/memory overhead (slower, more memory), which is why it's a CI gate
rather than every local run.

</details>

---

### 7. Is 100% coverage a good goal? Why or why not?

<details>
<summary>Answer</summary>

No, not as a target in itself. Coverage measures which lines *executed*, not
whether behavior was *verified* — a false-green test can run a line and assert
nothing, so 100% coverage can still hide bugs. Chasing the last few percent
drives tests for trivial getters and unreachable branches, which is wasted
effort and brittle.

The useful goal is covering the *logic*: boundaries, error paths, tricky
branches. Use `-coverprofile` + `go tool cover -html` to find untested logic,
not to worship a number. Meaningful 80% beats line-touching 100%.

</details>

---

### 8. What does `t.Parallel()` actually do, and what do you have to watch for?

<details>
<summary>Answer</summary>

Calling `t.Parallel()` signals that a test (or subtest) can run concurrently
with other parallel tests. The runner pauses it, lets the serial tests in the
package finish, then resumes all parallel ones together (capped by
`-parallel N`, default `GOMAXPROCS`). It speeds up I/O-bound or independent
suites.

What to watch for:

- **Shared mutable state** becomes a real data race — construct everything per
  subtest; run `-race`.
- **Pre-Go 1.22 loop capture:** parallel table subtests captured the loop var by
  reference, so every case saw the last row. Old fix: `tc := tc`. Go 1.22 made
  each iteration's variable fresh, so this is no longer needed — but you'll see
  it in older code.
- Ordering of output and execution is non-deterministic — don't depend on it.

</details>

---

### 9. How do you fake time in tests? What's a test double — stub vs mock vs fake?

<details>
<summary>Answer</summary>

**Faking time:** never call `time.Now()` directly in logic you want to test.
Inject it — either a `now func() time.Time` field or a small `Clock` interface
(`Now() time.Time`). In tests, return a fixed instant so time-dependent behavior
("token expires after 24h") is deterministic and instant. The standard library's
`testing/synctest` (Go 1.24+) helps test concurrent, time-based code with a
controlled fake clock.

**Test doubles** (Meszaros' taxonomy):

- **Stub** — returns canned answers to calls; no logic. "Charge always returns
  txn_123."
- **Fake** — a real, lightweight working implementation. An in-memory map
  standing in for a database.
- **Mock** — a double with *expectations*: it asserts it was called with certain
  args a certain number of times, and fails the test otherwise.

Go culture leans on stubs and fakes (hand-written, via small interfaces) and is
wary of mocks, since asserting on calls tests implementation rather than
behavior.

</details>

---

### 10. Table-driven test vs subtest — what's the relationship? And what is testify?

<details>
<summary>Answer</summary>

They're orthogonal but usually combined. A **subtest** is any test launched via
`t.Run(name, fn)` — it gets its own name in output, isolation, and `-run`
filtering. A **table-driven test** is the pattern of looping over a slice of
case structs. You don't *need* subtests for a table (you could loop with bare
`if` checks), but pairing them is idiomatic: `t.Run(tc.name, ...)` per row gives
each case a name, lets one failing case not abort the others, and lets you run a
single case. So: table = the data-driven structure; subtests = how each row is
isolated and labeled.

**testify** (`github.com/stretchr/testify`) is the near-universal external
assertion library: `assert.Equal(t, want, got)`, `require.NoError(t, err)`,
`assert.ErrorIs(...)`. `assert` continues on failure (like `t.Error`),
`require` stops (like `t.Fatal`). It's optional sugar — the standard `testing`
package is complete — but know it because most teams use it. (The Day 09
debugging module stays standard-library only.)

</details>
