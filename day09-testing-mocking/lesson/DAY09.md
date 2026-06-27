# Day 09 — Testing, Table-Driven Tests, Mocking

> Mentor note: In Go, testing is not a framework you bolt on — it ships in the toolchain (`go test`). There is no JUnit, no pytest, no separate runner. Files ending `_test.go`, functions starting `Test`, and `go test ./...` is the whole story. Because it's so frictionless, the cultural expectation on a Go team is: **untested code does not merge.** Today you learn the idioms that make Go tests fast to write and trustworthy.

---

## 1. The mechanics

- Test files live next to the code: `payment.go` → `payment_test.go`, same package.
- Test functions: `func TestXxx(t *testing.T)`.
- Run: `go test ./...` (all packages), `go test -v` (verbose), `go test -run TestName` (filter).

```go
// math.go
package mathx
func Add(a, b int) int { return a + b }

// math_test.go
package mathx
import "testing"

func TestAdd(t *testing.T) {
    got := Add(2, 3)
    if got != 5 {
        t.Errorf("Add(2,3) = %d; want 5", got) // t.Errorf: fail, keep going
    }
}
```

- `t.Errorf` — record failure, continue. `t.Fatalf` — fail and stop this test immediately (use when continuing makes no sense, e.g. a nil you'd deref next).

## 2. Table-driven tests — THE Go idiom

Don't write 10 test functions. Write one with a table of cases:

```go
func TestDivide(t *testing.T) {
    tests := []struct {
        name    string
        a, b    float64
        want    float64
        wantErr bool
    }{
        {"simple", 10, 2, 5, false},
        {"by zero", 1, 0, 0, true},
        {"negative", -6, 3, -2, false},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {   // subtest: isolated, named in output
            got, err := Divide(tc.a, tc.b)
            if (err != nil) != tc.wantErr {
                t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
            }
            if !tc.wantErr && got != tc.want {
                t.Errorf("got %v, want %v", got, tc.want)
            }
        })
    }
}
```

Why it's the standard: adding a case is one line. `t.Run` gives each case its own name in output (`TestDivide/by_zero`), isolation, and the ability to run one case with `-run TestDivide/by_zero`.

**Senior take:** When I review a Go PR, the first thing I look for is a table test on the new function. The shape — `name, inputs, want, wantErr` — is so standard that reviewers read it instantly. Match it.

## 3. Mocking — via interfaces, not magic

Go has no built-in mock framework reflection magic (and rarely needs one). You mock by **defining a small interface and providing a fake implementation**. This is *why* "accept interfaces" (Day 6/7) matters: it's what makes code testable.

```go
// The dependency your code needs, as a narrow interface:
type Charger interface {
    Charge(amount int) (string, error)
}

// Production code depends on the interface, not a concrete gateway:
func Checkout(c Charger, amount int) (string, error) {
    if amount <= 0 {
        return "", errors.New("amount must be positive")
    }
    return c.Charge(amount)
}

// In the test, a hand-written fake — full control, no network:
type fakeCharger struct {
    wantErr bool
}
func (f fakeCharger) Charge(amount int) (string, error) {
    if f.wantErr {
        return "", errors.New("gateway down")
    }
    return "txn_123", nil
}
```

Now `Checkout(fakeCharger{}, 100)` tests your logic with zero external dependencies.

### When to use a mocking library
For big interfaces, hand-writing fakes is tedious. Tools like `testify/mock` or `mockgen` generate them. But: prefer **small interfaces** that are trivial to fake by hand. If you need a 20-method mock, your interface is too big — that's a design smell, not a tooling problem.

## 4. testify — the one external lib most teams use

`go get github.com/stretchr/testify`. It gives readable assertions:

```go
import "github.com/stretchr/testify/assert"

assert.Equal(t, 5, Add(2, 3))
assert.NoError(t, err)
assert.ErrorIs(t, err, ErrNotFound)
```

It's optional sugar — std `testing` is complete on its own — but it's near-universal in industry, so know it.

## 5. Coverage & helpers

- `go test -cover` → coverage %. `go test -coverprofile=c.out && go tool cover -html=c.out` → visual.
- `t.Helper()` inside a test helper makes failures report the caller's line, not the helper's.
- `t.Cleanup(fn)` registers teardown (better than defer for shared setup).
- Don't chase 100% coverage — cover the logic and edge cases, not trivial getters.

## Common mistakes
1. One giant test with no `t.Run` — when it fails you don't know which case.
2. Tests that depend on each other or on order — each test must be independent. `go test` may run them in any order / parallel.
3. Testing the implementation, not behavior — asserting internal calls instead of outputs makes refactors painful.
4. Real network/DB/time in unit tests — inject interfaces and a fake clock instead. Flaky tests erode trust until everyone ignores them.
5. Forgetting `tc := tc` capture in parallel subtests on older Go (fixed by the Go 1.22 loopvar change, but know the history).

## Performance / workflow
- `go test` caches results — unchanged packages don't re-run. `go clean -testcache` to force.
- `go test -race ./...` runs the race detector (Phase 3) — make it part of CI.
- `go test -short` lets you skip slow tests in quick loops via `if testing.Short() { t.Skip() }`.

---

## Expert Thinking Mode — "test this function"

- **Beginner:** "Call it, check it returns the right value."
- **Senior:** "Table of cases incl. error paths and boundaries. Inject dependencies via interfaces so there's no network. Test behavior, not internals."
- **Staff:** "What's the contract this package promises? Tests *are* that contract's executable spec. Race + coverage in CI. Fakes model real failure modes (timeouts, partial writes)."
- **Architect:** "Test strategy across the org: unit (fast, isolated) vs integration (real deps, fewer) vs e2e (rare). The test pyramid keeps CI fast and signal high. Flaky tests are an outage of the dev pipeline."

---

## Real-world use

- **Every Go shop** gates merges on `go test ./...` + `-race` in CI.
- **Stripe/Uber:** narrow interfaces (`Charger`, `Store`) with hand-written or generated fakes; integration tests hit real Postgres in Docker (Phase 4/5).
- **Table tests** are the default review-expected shape across the entire ecosystem incl. the Go standard library itself — go read `strings/strings_test.go`.

---

## Interview Questions

1. How does Go's testing differ from JUnit/pytest? What's the file/function convention?
2. What is a table-driven test and why is it the Go idiom? What does `t.Run` add?
3. `t.Error` vs `t.Fatal` — when each?
4. How do you mock a dependency in Go without a mocking framework? How does interface design enable it?
5. Why must tests be independent of order? What runs them in parallel?
6. What does `go test -race` do and why run it in CI?
7. Is 100% coverage a good goal? Why or why not?

---

## Your tasks

`../exercises/` has a `wallet.go` with a `Charger` interface and a `Withdraw`/`Checkout` function, plus a `wallet_test.go` skeleton. Write: (1) a table-driven test for the pure logic, (2) a hand-written fake `Charger` to test the checkout path including the gateway-failure case. Run `go test -v ./...` until green. Reference in `../solutions/`.
