# Day 22 Notes — CI/CD for Go (quick reference)

## The deploy loop
```
local edit → fmt/vet/lint (pre-commit) → push
  → CI: lint + test(-race) + build + scan
    → docker build + push (tag = git SHA)
      → deploy → smoke tests → done | rollback
```

## Minimal GitHub Actions CI
```yaml
on: { push: { branches: [main] }, pull_request: { branches: [main] } }
jobs:
  ci:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.26', cache: true }   # cache deps + build
      - uses: golangci/golangci-lint-action@v6
        with: { args: --timeout=5m }
      - run: go test -race -count=1 -coverprofile=coverage.out ./...
      - run: CGO_ENABLED=0 go build ./...
```

## Commands that matter
| Command | Purpose |
|---|---|
| `go vet ./...` | built-in correctness checks |
| `golangci-lint run ./...` | 50+ linters (staticcheck, errcheck, gosec, bodyclose...) |
| `go test -race ./...` | catch data races |
| `go test -count=1 ./...` | bypass test cache (force rerun) |
| `go test -shuffle=on ./...` | surface inter-test coupling |

## go vet highlights
`printf` · `copylocks` · `unreachable` · `unusedresult` · `structtag` · `assign`

## golangci-lint power linters
`staticcheck` (best single) · `errcheck` · `gosec` (security) · `bodyclose`
(leaked HTTP bodies) · `ineffassign` · `unused`

## Flaky-test sources → fixes
| Source | Fix |
|---|---|
| map iteration order | `sort.Strings(keys)` before output |
| `time.Now()` | inject a clock / fixed timestamp |
| goroutine scheduling | `-race`; don't assert on order |
| shared global state | `-shuffle=on`; isolate per-test |

## Gates
- Tag images with **git SHA**, not `latest`.
- Branch protection: require CI checks, require review, no bypass.
- `make ci` runs the *same* commands as CI for local repro.
- Secrets → `${{ secrets.NAME }}` (encrypted, masked); never in YAML.

## Key terms
- **Workflow / job / step** — Actions hierarchy under `.github/workflows/`.
- **`needs:`** — job dependency (build a DAG; fan-in before push).
- **Required status check** — a CI job that must pass to merge.
- **Race detector** — runtime data-race finder (`-race`).
- **Flaky test** — non-deterministic pass/fail on identical code.
- **Immutable tag** — a tag (git SHA / digest) that never moves.
