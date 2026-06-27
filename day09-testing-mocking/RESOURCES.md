# Day 09 — Resources (Testing &amp; Mocking)

Curated, real links. Each with a one-line "why."

- **[`testing` package — pkg.go.dev](https://pkg.go.dev/testing)**
  The canonical reference: `*testing.T`, `t.Run`, `t.Helper`, `t.Cleanup`,
  `t.Parallel`, `testing.Short` — read the source of truth, not blog summaries.

- **[Using Subtests and Sub-benchmarks — Go blog](https://go.dev/blog/subtests)**
  The official explanation of `t.Run`: naming, isolation, filtering, and
  parallel subtests — the backbone of table-driven tests.

- **[Prefer table driven tests — Dave Cheney](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)**
  The clearest argument for the Go table-test idiom, with the standard
  `name/input/want` shape reviewers expect.

- **[TableDrivenTests — Go wiki](https://go.dev/wiki/TableDrivenTests)**
  The community reference page for the pattern, including parallel-subtest
  notes and the historical loop-variable caveat.

- **[Test fixtures in Go — Dave Cheney](https://dave.cheney.net/2016/05/10/test-fixtures-in-go)**
  How to manage setup data and golden files cleanly with `testdata/` and
  helpers — fixtures without leaking state across tests.

- **[The cover story — Go blog](https://go.dev/blog/cover)**
  How `go test -cover`, coverage profiles, and `go tool cover -html` work; use
  it to find untested logic rather than chase a percentage.

- **[stretchr/testify — GitHub](https://github.com/stretchr/testify)**
  The near-universal assertion/mock library (`assert`, `require`, `mock`).
  Optional sugar over the standard `testing` package — know it because most
  teams use it. (Not used in the Day 09 debugging module — std lib only.)
