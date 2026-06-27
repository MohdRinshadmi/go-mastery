# Day 22 Pitfalls — CI/CD for Go

Format: **Trap → Why → Fix**

---

### 1. Tests that depend on map iteration order
**Trap:** A test asserts an exact string built by ranging a map; it's green locally, red intermittently in CI.
**Why:** Go randomizes map iteration order on purpose. The output order is non-deterministic, so an exact-match assertion is a coin flip — and CI eventually loses it.
**Fix:** Sort keys before producing observable output (`sort.Strings(keys)`). More generally, eliminate every source of non-determinism: map order, time, scheduling, shared state. (This is the Day 22 debugging exercise.)

---

### 2. Skipping `-race` in CI because it's slow
**Trap:** `go test ./...` without `-race` to keep the pipeline fast.
**Why:** The race detector is the only reliable way to catch data races, and it only flags code that runs under it. A race in production is far more expensive than a 2–10× slower test job.
**Fix:** `go test -race -count=1 ./...` in CI. Use `-count=1` to defeat the test cache when you want a true rerun.

---

### 3. Tagging images `:latest` only
**Trap:** Deploy `ghcr.io/org/app:latest`.
**Why:** `latest` is mutable — you can't tell which commit is live or roll back deterministically. Two deploys an hour apart can both be "latest" and differ.
**Fix:** Tag with the immutable git SHA (`:${{ github.sha }}`). Keep `latest` as a convenience pointer only, never as the deploy reference.

---

### 4. No module/build cache in the workflow
**Trap:** `setup-go` without `cache: true`; every run re-downloads deps and recompiles from scratch.
**Why:** Cold builds make CI minutes-long. Slow CI pushes engineers to batch changes, which makes each merge riskier.
**Fix:** `actions/setup-go@v5` with `cache: true` (caches the module + build cache keyed on `go.sum`).

---

### 5. CI checks that don't actually block merge
**Trap:** Lint and test jobs run, but branch protection doesn't *require* them.
**Why:** A decorative check lets broken code merge to `main` anyway — the gate isn't a gate.
**Fix:** In branch protection, mark the CI jobs as **required status checks**, require PR review, and disallow bypass (even for admins).

---

### 6. CI that can't be reproduced locally
**Trap:** CI has implicit setup, magic env vars, or steps no developer can run.
**Why:** When CI fails, no one can reproduce it, so failures get retried until green — masking real problems.
**Fix:** Put the exact pipeline behind `make ci` (lint + test + build). If `make ci` is green locally, CI should be green too.

---

### 7. Letting `golangci-lint` / `go vet` warnings accumulate
**Trap:** "We'll clean up the lint later."
**Why:** Once there are 200 findings, no one reads new ones — real bugs (`bodyclose`, `errcheck`) hide among the noise.
**Fix:** Gate the build on a clean `golangci-lint run`. Use `//nolint` with a reason for the rare justified exception; keep the baseline at zero.
