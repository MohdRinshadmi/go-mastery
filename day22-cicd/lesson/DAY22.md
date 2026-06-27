# Day 22 — CI/CD: GitHub Actions, Makefile, go vet, golangci-lint

> Mentor note: CI/CD is not a DevOps concern — it's an engineering discipline. Every production outage I've witnessed from a "simple change" happened because no automated gate caught the regression before it hit prod. Today you build the pipeline that protects your team at 2 AM when you're not watching.

---

## 0. The production deployment loop

Here's the lifecycle of code in a healthy Go shop:

```
local edit
  → go fmt / go vet / golangci-lint (pre-commit)
    → git push
      → CI: lint + test + build + security scan
        → docker build + push to registry
          → deploy (Kubernetes/ECS rollout)
            → smoke tests
              → done (or rollback)
```

Today covers everything from "git push" through "docker push." Deployment mechanics (Kubernetes, ECS) are ops territory; what we're building is the gate before you ever get there.

---

## 1. GitHub Actions — The Basics

### Theory
GitHub Actions is a CI/CD platform built into GitHub. You define **workflows** in YAML files under `.github/workflows/`. A workflow runs on events (push, PR, schedule) and executes a series of **jobs** made of **steps**.

### Why it exists
CI servers used to be a dedicated machine (Jenkins, TeamCity) that you maintained separately. Actions runs on GitHub's infrastructure — no servers to manage, first 2000 minutes/month free on public repos.

### Core concepts

```
Workflow (.github/workflows/ci.yml)
  └─ triggered by: push, pull_request, schedule, workflow_dispatch
      └─ Job (lint, test, build, docker)
           └─ Steps (actions/checkout, setup-go, go test, docker/build-push-action)
```

**Key YAML fields:**
```yaml
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest      # GitHub-hosted runner
    steps:
      - uses: actions/checkout@v4            # checks out your code
      - uses: actions/setup-go@v5            # installs Go
        with:
          go-version: '1.22'
          cache: true                         # caches go module downloads
      - run: go build ./...
```

### When to use
- Every project that uses GitHub, period. Free, integrated, no extra setup.

### When NOT to use
- Organizations with compliance requirements that mandate on-prem CI (use Jenkins/GitLab self-hosted).
- Extremely complex build graphs with thousands of jobs (consider Bazel + remote execution).
- Secrets that cannot leave your network (use self-hosted runners).

---

## 2. The CI Pipeline: lint → test → build → docker push

### Step 1: Lint with golangci-lint

`go vet` is the built-in static analyzer — it catches real bugs (suspicious `Printf` calls, unreachable code, copying mutexes). `golangci-lint` runs 50+ linters in parallel, including `go vet`, `staticcheck`, `errcheck`, `gosec`, and more.

```yaml
- name: golangci-lint
  uses: golangci/golangci-lint-action@v6
  with:
    version: v1.59
    args: --timeout=5m
```

You configure which linters run in `.golangci.yml`:

```yaml
linters:
  enable:
    - errcheck      # finds unchecked errors
    - staticcheck   # the best single Go linter
    - gosec         # security issues
    - govet         # all go vet checks
    - ineffassign   # assigned but never used
    - unused        # unused code
  disable:
    - depguard      # dependency restrictions (not needed for small teams)

issues:
  exclude-rules:
    - path: _test\.go
      linters: [errcheck]   # test files can ignore errors in helpers
```

### Step 2: Test with coverage

```yaml
- name: Test
  run: go test -race -coverprofile=coverage.out -covermode=atomic ./...

- name: Upload coverage
  uses: codecov/codecov-action@v4
  with:
    file: ./coverage.out
```

`-race` enables Go's race detector — it catches data races at runtime. Always run `-race` in CI even if it slows tests by 2-10x. A race condition in production is far more expensive.

### Step 3: Build for multiple platforms

```yaml
- name: Build
  run: |
    CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -o bin/server-linux-amd64 ./cmd/server
    CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -o bin/server-linux-arm64 ./cmd/server
    CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -o bin/server-darwin-arm64 ./cmd/server
```

### Step 4: Docker build and push

```yaml
- name: Build and push Docker image
  uses: docker/build-push-action@v5
  with:
    context: .
    push: ${{ github.event_name != 'pull_request' }}
    tags: |
      ghcr.io/${{ github.repository }}:latest
      ghcr.io/${{ github.repository }}:${{ github.sha }}
    build-args: |
      VERSION=${{ github.ref_name }}
      GIT_COMMIT=${{ github.sha }}
      BUILD_TIME=${{ github.event.head_commit.timestamp }}
```

**Senior take:** Tag images with the git SHA, not just `latest`. `latest` is a lie in production — you never know what it points to. `ghcr.io/org/app:a1b2c3d` is immutable. If a deploy breaks, you know exactly which commit is running and can roll back by pinning the previous SHA.

---

## 3. The Makefile

### Theory
`make` is the classic build automation tool. In Go projects, it's a **local developer interface** — it wraps common commands so you type `make test` instead of `go test -race -count=1 -timeout 60s ./...`.

### Why it exists
You want the same commands used by developers locally to be the same commands used by CI. If `make ci` runs the full CI pipeline locally, developers catch failures before pushing.

### Production Makefile patterns

```makefile
.PHONY: all lint test build docker-build ci

# Default: run everything
all: lint test build

# Lint: go vet + golangci-lint
lint:
    go vet ./...
    golangci-lint run ./...

# Test: with race detector and coverage
test:
    go test -race -coverprofile=coverage.out -covermode=atomic ./...
    go tool cover -func=coverage.out

# Build: with version injection
build:
    CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/server ./cmd/server

# ci: exactly what CI runs (use this to debug CI failures locally)
ci: lint test build

# Clean
clean:
    rm -rf bin/ coverage.out
```

**Senior take:** `make ci` should be a thing. When CI fails, developers run `make ci` locally to reproduce it. If you can't reproduce CI locally in under 5 minutes, your CI is too magical — it has environment-specific behavior that will bite you.

---

## 4. go vet — Understanding What It Catches

`go vet` catches real bugs, not style issues. Key checks:

| Check | What it catches |
|---|---|
| `printf` | `fmt.Sprintf("%d", "string")` — format/arg type mismatch |
| `copylocks` | Copying a mutex by value (causes data races) |
| `unreachable` | Code after `return`/`panic` that never runs |
| `unusedresult` | Calling `strings.Replace` without using the result |
| `shadow` | Variable shadowing (with `-vet=shadow`) |
| `structtag` | Malformed struct tags like `json:"name,omitempty,extra"` |
| `assign` | `x = x` (self-assignment, almost always a bug) |

Run it always: `go vet ./...`

---

## 5. golangci-lint — The Power Linters

Beyond `go vet`, these are the linters that catch real production bugs:

**`staticcheck`** — the most valuable single linter. Finds: deprecated API usage, nil pointer dereferences that `go vet` misses, unreachable code, unnecessary type conversions.

**`errcheck`** — every function that returns an error should have that error checked. This is the most common Go bug in codebases that skip linting.

**`gosec`** — security linter. Finds: SQL injection risks, weak crypto (MD5/SHA1 for passwords), hardcoded credentials, insecure HTTP (HTTP instead of HTTPS), path traversal.

**`gocritic`** — opinionated style + performance. Finds: inefficient string conversions, redundant `else` after `return`, `append` in a loop that could pre-allocate.

**`bodyclose`** — `resp.Body.Close()` after `http.Get()`. If you forget it, you leak connections until they timeout. This is the #1 Go HTTP gotcha.

```bash
# Install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1

# Run
golangci-lint run ./...

# With specific linters
golangci-lint run --enable=gosec,errcheck ./...
```

---

## 6. Branch strategy and merge gates

### The protected main branch pattern

```
feature/* → develop → main
              ↑              ↑
         PR gates       Release tags
         (CI must pass)  (docker push + deploy)
```

In GitHub Settings → Branches → Branch protection rules:
- ✅ Require status checks to pass (your CI jobs)
- ✅ Require pull request reviews
- ✅ Do not allow bypassing above settings

**This means:** No one — not even the repo owner — can push broken code to `main`. The CI pipeline is the gate, not trust in developers.

---

## 7. Secrets management in CI

```yaml
env:
  DOCKER_USERNAME: ${{ secrets.DOCKER_USERNAME }}
  DOCKER_PASSWORD: ${{ secrets.DOCKER_PASSWORD }}
```

GitHub Actions secrets are encrypted at rest and injected at runtime — they never appear in logs. Configure them in: Repository → Settings → Secrets and variables → Actions.

**What to store as secrets:** registry credentials, API keys, signing keys, database URLs for integration tests.
**What NOT to store as secrets:** versions, feature flags, non-sensitive configuration.

**Senior take:** Rotation policy. Every secret should have an owner and a rotation schedule. Secrets that never rotate are breach multipliers. When you set a secret in CI, immediately ask: "When does this rotate, and who owns that process?"

---

## Common mistakes

1. **Flaky tests treated as noise.** A test that depends on map iteration order, `time.Now()`, goroutine scheduling, or shared global state passes locally and fails intermittently in CI. Engineers learn to "just retry" — and a real regression eventually hides in the noise. Pin down the non-determinism (sort keys, inject a clock, run `-race` and `-shuffle=on`) and treat a flake as a Sev-1 bug, not a retry.
2. **Not running `-race` in CI.** The race detector only catches a data race if the racy code path actually executes under it. Skipping `go test -race` because "it's slow" means real concurrency bugs ship — far more expensive than a 2–10× slower test job.
3. **CI that can't be reproduced locally.** If CI does things no developer can run (magic env vars, implicit setup), failures become un-debuggable. Make `make ci` run the exact same lint/test/build so "it works on my machine" is impossible.
4. **Tagging images `latest` only.** `latest` is mutable — you can't tell what's deployed or roll back deterministically. Always tag with the immutable git SHA (`ghcr.io/org/app:a1b2c3d`).
5. **No caching → slow pipeline.** Forgetting `cache: true` on `setup-go` (or not caching the module/build cache) re-downloads and recompiles everything every run. Slow CI makes engineers batch changes, which makes each deploy riskier.
6. **Lint/test failures that don't block merge.** Status checks that aren't *required* by branch protection are decorative — broken code still merges. Make the CI jobs required checks and disallow bypass.

---

## Expert Thinking Mode

- **Beginner:** "CI runs my tests automatically. Convenient."
- **Senior:** "CI is the canonical definition of 'done.' Code isn't done until it's green in CI. I own the pipeline as much as I own the code."
- **Staff:** "CI gate quality directly predicts deployment frequency. If CI takes 45 minutes, engineers batch changes and that makes each deploy riskier. I invest in keeping CI under 5 minutes through parallelism, caching, and right-sizing."
- **Architect:** "The CI/CD pipeline is a platform product. It has customers (developers) and SLAs (p95 < 5 min, 99.9% availability). When CI is flaky, the whole org slows down. I treat flaky tests as Severity 1 incidents."

---

## Real-world use

- **Cloudflare:** Every Go PR goes through lint + test + build + docker push + canary deploy. If a canary has elevated error rate for 15 minutes, the rollout stops automatically.
- **Uber:** Go monorepo with Bazel. CI tests only the subgraph affected by your change — they don't run all 10,000 tests, only the ~200 tests for your change. Result: CI in under 3 minutes on a 200-engineer Go team.
- **Stripe:** `make ci` runs locally in a Docker container that exactly mirrors CI. "It works on my machine" is impossible by design.
- **GitHub itself:** Uses GitHub Actions to deploy GitHub. The pipeline is a first-class engineering product with an oncall rotation.

---

## Interview Questions

1. What does `go vet` catch that the compiler does not?
2. What is the `golangci-lint` linter `bodyclose` and what production bug does it prevent?
3. Why should Docker images be tagged with git SHA instead of `latest`?
4. What does the `-race` flag do in `go test`? What's the trade-off?
5. Explain `depends_on` in GitHub Actions — when and why would you want jobs to depend on each other?
6. What is a "branch protection rule" and how does it enforce CI quality gates?
7. How do you securely pass credentials (registry password, DB URL) to a GitHub Actions workflow?

---

## Your tasks for today

Go to `../exercises/`. You have:
1. A `.golangci.yml` to write for a given set of requirements
2. A buggy Go file to run golangci-lint on — fix all findings
3. A GitHub Actions workflow to complete (fill in the TODOs)
4. A Makefile to write

These exercises are about understanding the pipeline, not just syntax. For each one, explain in a comment *why* each choice matters.

---

## Day 22 companion files

Self-contained study material for this day (in the day folder root):

- [Debugging exercise](../debugging/README.md) — the flaky test: code that depends on Go's randomized map iteration order, green locally and red in CI ([bugged](../debugging/bugged/main.go) vs [fixed](../debugging/fixed/main.go)).
- [PITFALLS.md](../PITFALLS.md) — 7 CI/CD traps as Trap → Why → Fix.
- [INTERVIEW.md](../INTERVIEW.md) — interview Q&A with model answers.
- [NOTES.md](../NOTES.md) — quick reference + key terms.
- [RESOURCES.md](../RESOURCES.md) — curated links for Day 22.
