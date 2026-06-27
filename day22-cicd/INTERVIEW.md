# Day 22 Interview Questions — CI/CD for Go

Lesson questions plus extras. Answers in `<details>`.

---

### 1. What does `go vet` catch that the compiler does not?

<details>
<summary>Answer</summary>

Bugs that are valid syntax but wrong semantics: `Printf` format/arg mismatches,
copying a `sync.Mutex` by value (`copylocks`), unreachable code, calling a
pure function and discarding the result (`unusedresult`), malformed struct tags,
self-assignment. The compiler accepts these; `go vet` flags them.
</details>

---

### 2. What is the `bodyclose` linter and what bug does it prevent?

<details>
<summary>Answer</summary>

It flags HTTP response bodies that are never closed after `http.Get`/`Do`. If you
don't `resp.Body.Close()`, the connection isn't returned to the pool and you leak
file descriptors / connections until they time out — the #1 Go HTTP gotcha under
load.
</details>

---

### 3. Why tag Docker images with the git SHA instead of `latest`?

<details>
<summary>Answer</summary>

`latest` is mutable — you can't tell what's running or roll back deterministically.
The git SHA is immutable: `ghcr.io/org/app:a1b2c3d` always points at exactly one
build, so you know the deployed commit and can pin a previous one to roll back.
</details>

---

### 4. What does `-race` do in `go test`, and the trade-off?

<details>
<summary>Answer</summary>

It enables the race detector, which instruments memory accesses and reports data
races (concurrent unsynchronized read/write of the same memory) at runtime. The
trade-off is 2–10× slower execution and more memory — worth it in CI, since it
only flags code paths that actually run. A production race is far costlier.
</details>

---

### 5. Explain `needs:` (job dependencies) in GitHub Actions.

<details>
<summary>Answer</summary>

`needs:` makes one job wait for another to succeed before it runs, creating a
DAG. You'd gate `docker-push` on `lint` and `test` passing so you never push an
image built from un-linted, failing code. Independent jobs (e.g. lint and test)
can run in parallel for speed; the push fans them back in with `needs: [lint, test]`.
</details>

---

### 6. What is a branch protection rule and how does it enforce CI gates?

<details>
<summary>Answer</summary>

A repo setting that constrains pushes/merges to a branch: require specific status
checks (your CI jobs) to pass, require PR review, and disallow bypass. It means
no one — not even an admin — can merge code that fails CI. The pipeline becomes
the gate, not trust in individuals.
</details>

---

### 7. How do you securely pass credentials to a GitHub Actions workflow?

<details>
<summary>Answer</summary>

Store them as **encrypted secrets** (repo/org/environment level) and reference
them as `${{ secrets.NAME }}`. They're encrypted at rest, injected at runtime,
and masked in logs. Don't put them in the YAML, env defaults, or
non-secret variables. Prefer short-lived OIDC tokens over long-lived static
credentials where the registry/cloud supports it.
</details>

---

### 8. (Extra) What is a flaky test and how do you fix one?

<details>
<summary>Answer</summary>

A test that passes and fails non-deterministically on the same code. Causes:
map iteration order, `time.Now()`, goroutine scheduling, shared global state
between tests, network/IO. Fix the *source* of non-determinism — sort keys,
inject a clock, run `-race`, run `go test -shuffle=on` to surface inter-test
coupling. Never "fix" it by retrying; treat it as a real bug.
</details>

---

### 9. (Extra) Why should `make ci` exist?

<details>
<summary>Answer</summary>

So the exact commands CI runs (lint, test, build) can be run locally with one
command. Engineers reproduce CI failures in seconds instead of pushing repeatedly
to debug. If CI does things `make ci` can't, CI is "too magical" and failures
become un-debuggable.
</details>

---

### 10. (Extra) What's the difference between `go vet` and `golangci-lint`?

<details>
<summary>Answer</summary>

`go vet` is the built-in analyzer with a focused set of correctness checks.
`golangci-lint` is a fast aggregator that runs 50+ linters in parallel —
including `govet`, plus `staticcheck`, `errcheck`, `gosec`, `bodyclose`,
`ineffassign`, etc. — configured via `.golangci.yml`. You run both; vet is the
floor, golangci-lint is the ceiling.
</details>

---

### 11. (Extra) Why keep CI under ~5 minutes?

<details>
<summary>Answer</summary>

Long CI makes engineers batch many changes into one push to avoid waiting, which
makes each merge bigger and riskier and harder to bisect. Fast CI keeps changes
small and feedback tight. You get there with caching, parallel jobs, and testing
only the affected subgraph in large repos.
</details>
