# Day 09 — Notes / Cheatsheet (Testing &amp; Mocking)

Standard-library `testing` only. Quick reference for the toolchain and idioms.

---

## File &amp; function conventions

| Thing            | Rule                                                         |
| ---------------- | ----------------------------------------------------------- |
| Test file        | ends in `_test.go`, next to the code (`foo.go`→`foo_test.go`)|
| Test func        | `func TestXxx(t *testing.T)` — `Test` prefix, capital next  |
| Same package     | `package foo` — can test unexported symbols (white-box)     |
| External package | `package foo_test` — tests only the public API (black-box)  |
| Benchmark        | `func BenchmarkXxx(b *testing.B)`                            |
| Example          | `func ExampleXxx()` with `// Output:` comment (also tested)  |

---

## Table-driven test skeleton

```go
func TestClassify(t *testing.T) {
    tests := []struct {
        name  string
        input int
        want  string
    }{
        {"low fail", 10, "fail"},
        {"boundary", 70, "distinction"},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            got := Classify(tc.input)
            if got != tc.want {              // assert against tc.want — never a constant
                t.Errorf("Classify(%d) = %q; want %q", tc.input, got, tc.want)
            }
        })
    }
}
```

With an error column:

```go
got, err := Divide(tc.a, tc.b)
if (err != nil) != tc.wantErr {
    t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
}
if !tc.wantErr && got != tc.want {
    t.Errorf("got %v, want %v", got, tc.want)
}
```

---

## `t.Run` — subtests

```go
t.Run(tc.name, func(t *testing.T) { ... })
```

- Names each case in output: `TestClassify/boundary`.
- Run one case: `go test -run 'TestClassify/boundary'`.
- A failure in one subtest doesn't abort the others (unless it calls `t.Fatal`).

---

## `t.Error` vs `t.Fatal`

| Call               | Effect                                  | Use when                          |
| ------------------ | --------------------------------------- | --------------------------------- |
| `t.Error` / `Errorf` | mark failed, **continue**             | independent assertions            |
| `t.Fatal` / `Fatalf` | mark failed, **stop this test** (Goexit) | setup failed, or nil before deref |

`t.Fatal` only stops the goroutine it runs in — don't call it from a spawned goroutine.

---

## Hand-written fake skeleton (mocking via small interface)

```go
// Production: depend on a narrow interface.
type Charger interface {
    Charge(amount int) (string, error)
}

func Checkout(c Charger, amount int) (string, error) {
    if amount <= 0 { return "", errors.New("amount must be positive") }
    return c.Charge(amount)
}

// Test: hand-written fake — full control, no network.
type fakeCharger struct {
    wantErr bool
    gotAmt  int        // optional: record args for assertions
}
func (f *fakeCharger) Charge(amount int) (string, error) {
    f.gotAmt = amount
    if f.wantErr { return "", errors.New("gateway down") }
    return "txn_123", nil
}
```

Test both paths: `Checkout(&fakeCharger{}, 100)` and `Checkout(&fakeCharger{wantErr: true}, 100)`.

---

## `t.Helper()` and `t.Cleanup()`

```go
func assertEqual(t *testing.T, got, want string) {
    t.Helper()                 // failures report the CALLER's line, not here
    if got != want { t.Errorf("got %q, want %q", got, want) }
}

func setup(t *testing.T) *Server {
    s := NewServer()
    t.Cleanup(func() { s.Close() })   // runs after the test (and subtests) finish
    return s
}
```

- `t.Helper()` — first line of any helper taking `*testing.T`; fixes failure line numbers.
- `t.Cleanup(fn)` — register teardown; runs LIFO, better than `defer` for shared setup.

---

## Faking time

```go
// Inject the clock; never call time.Now() in testable logic.
type Clock interface { Now() time.Time }
type fixedClock struct{ t time.Time }
func (c fixedClock) Now() time.Time { return c.t }
```

Or simpler: a `now func() time.Time` field. (Go 1.24+: `testing/synctest` for
concurrent time-based code.)

---

## Coverage commands

```bash
go test -cover ./...                          # print coverage %
go test -coverprofile=c.out ./...             # write profile
go tool cover -func=c.out                     # per-function breakdown
go tool cover -html=c.out                     # open visual report in browser
go test -coverpkg=./... ./...                 # count coverage across packages
```

Goal: cover the *logic* and edge cases — not chase 100%.

---

## Run / filter / race / short

```bash
go test ./...                       # run all packages
go test -v ./...                    # verbose (per-test output)
go test -run TestClassify ./...     # regex filter on test name
go test -run 'TestClassify/boundary'  # one subtest
go test -race ./...                 # data-race detector (use in CI)
go test -short ./...                # skip tests guarded by testing.Short()
go test -count=1 ./...              # disable result cache (force re-run)
go clean -testcache                 # clear the cache entirely
go test -parallel 4 ./...           # cap concurrent t.Parallel() tests
go vet ./...                        # static checks (run alongside tests)
```

`-short` guard inside a slow test:

```go
if testing.Short() { t.Skip("skipping slow test in -short mode") }
```

---

## Key terms

- **Table-driven test** — one test function looping over a slice of case structs
  (`name`, inputs, `want`); the Go-idiomatic way to cover many cases. Assert
  against `tc.want`, never a constant.
- **Subtest** — a test launched with `t.Run(name, fn)`; named in output,
  isolated, individually runnable with `-run`.
- **Fake / Stub / Mock** — test doubles. *Stub* returns canned values; *fake* is
  a lightweight working implementation (in-memory store); *mock* asserts it was
  called as expected (tests implementation — use sparingly in Go).
- **Test double** — umbrella term for any stand-in (stub, fake, mock, spy) you
  swap in for a real dependency.
- **`t.Helper()`** — marks a function as a helper so failures report the caller's
  line, not the helper's.
- **Coverage** — percentage of code lines executed by tests; a guide to find
  untested logic, not a goal to maximize.
- **Race detector** (`-race`) — runtime instrumentation that flags concurrent
  unsynchronized memory access; a CI gate for concurrency bugs.
