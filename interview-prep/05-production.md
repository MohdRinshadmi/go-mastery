# Phase 5 — Production Engineering (Days 21–25)

Docker, CI/CD, metrics/tracing/health, rate limiting & graceful shutdown. Self-quiz: answer aloud, then expand.

---

### 1. What does `CGO_ENABLED=0` do, and why is it required for `FROM scratch`?

<details><summary>Answer</summary>

`CGO_ENABLED=0` forces a **pure-Go, statically linked** binary with no dependency on the system C library (`libc`) or dynamic loader. `FROM scratch` is an **empty** image — no shell, no libc, nothing — so a CGO-linked binary that needs `libc.so` would fail to start there. Disabling CGO (and `GOOS=linux`) yields a self-contained binary that runs in an empty image. Watch out: pure-Go also swaps in a native DNS resolver, and you still need to add CA certs yourself if you make HTTPS calls.
</details>

---

### 2. `distroless/static` vs `scratch` — when each?

<details><summary>Answer</summary>

`scratch` is truly empty: smallest possible image, but **you** must add CA certificates, timezone data, and `/etc/passwd` if needed, and there's no shell to debug with. `distroless/static` is nearly as small but ships **CA certs, tzdata, and a nonroot user** out of the box, so HTTPS and TLS "just work" and you get a non-root default. Choose `scratch` for absolute minimalism on a binary that makes no TLS calls; choose `distroless/static` for almost everything else — it's the pragmatic default.
</details>

---

### 3. Why does `COPY`/`RUN` order matter for layer caching, and what does `-ldflags="-s -w"` do?

<details><summary>Answer</summary>

Docker caches each instruction as a layer and **invalidates everything after the first change**. So copy `go.mod`/`go.sum` and run `go mod download` **before** copying source code: dependencies (which change rarely) get cached, and editing application code doesn't re-download the whole module graph. `-ldflags="-s -w"` strips the **symbol table (`-s`) and DWARF debug info (`-w`)** from the binary, shrinking it noticeably — fine for production where you don't debug with symbols in-image (you keep them separately for symbolication if needed).
</details>

---

### 4. What should always be in `.dockerignore`, and what's the cost of omitting it?

<details><summary>Answer</summary>

`.git`, `node_modules`, build artifacts, local env/secret files, test data, and anything not needed for the build. Without it, the **entire build context** (potentially a huge `.git` history and local junk) is sent to the daemon — slowing every build, busting cache, bloating images, and risking **secrets leaking** into a layer. It's the Docker analog of `.gitignore` and is a cheap, high-leverage hygiene win.
</details>

---

### 5. What is `depends_on: condition: service_healthy` in Compose, and how do you inject version info at build time?

<details><summary>Answer</summary>

Plain `depends_on` only waits for a container to **start**, not to be **ready** — your app may connect to a DB that's still initializing. `condition: service_healthy` makes Compose wait until the dependency's **healthcheck passes** before starting the dependent service, eliminating "connection refused" races on boot. You inject build-time info via **`-ldflags="-X main.version=$GIT_SHA"`**, which sets a package variable at link time, so the binary can report its exact build/commit at runtime (`/version` endpoint, logs).
</details>

---

### 6. What does `go vet` catch that the compiler doesn't, and what's `bodyclose`?

<details><summary>Answer</summary>

`go vet` catches **suspicious-but-compilable** code: `Printf` format/arg mismatches, copying locks by value, unreachable code, struct-tag typos, lost context cancels, etc. — bugs the type checker permits. `bodyclose` (a `golangci-lint` linter) flags an **unclosed `http.Response.Body`**, which leaks the connection (and its goroutine/file descriptor) and prevents connection reuse — a slow resource leak that eventually exhausts the server. Both are static analysis layered on top of compilation.
</details>

---

### 7. Why tag images with the git SHA instead of `latest`? How do you pass secrets to a GitHub Actions workflow?

<details><summary>Answer</summary>

`latest` is **mutable and ambiguous** — two deploys can pull different bytes, rollbacks are impossible, and you can't tell which commit is running. An **immutable git-SHA tag** ties each image to an exact commit, making deploys reproducible, rollbacks trivial (re-deploy the prior SHA), and provenance clear. Secrets go in **GitHub Actions encrypted secrets** (repo/org/environment level), referenced as `${{ secrets.REGISTRY_PASSWORD }}` — never hard-coded, never echoed to logs (they're masked), and scoped with environment protection rules for prod.
</details>

---

### 8. What's a branch protection rule, and why depend jobs in a workflow?

<details><summary>Answer</summary>

A **branch protection rule** enforces gates on a branch (usually `main`): required passing status checks (lint/test/build), required reviews, no force-push, up-to-date-before-merge. It makes CI a **hard merge gate** instead of advisory, so broken code can't land. **Job dependencies** (`needs:`) order the pipeline — e.g., only build the image and deploy *after* lint and tests pass — so you fail fast and never ship an artifact built from code that didn't pass quality checks.
</details>

---

### 9. Logs vs metrics vs traces — what's each best at?

<details><summary>Answer</summary>

**Logs** = discrete, high-detail events ("what exactly happened on this request") — best for debugging specifics and post-incident forensics. **Metrics** = cheap numeric aggregates over time ("how many requests, how fast, error rate") — best for dashboards and alerting at scale. **Traces** = the causal path of one request across services with per-span timing ("where did the latency go") — best for diagnosing distributed bottlenecks. The "three pillars": metrics tell you *something's wrong*, traces tell you *where*, logs tell you *why*.
</details>

---

### 10. Counter vs gauge vs histogram — give an example of each. Why graph `rate(counter)`?

<details><summary>Answer</summary>

**Counter** = monotonically increasing total (`http_requests_total`) — only goes up (resets on restart). **Gauge** = a value that goes up and down (`goroutines_active`, `queue_depth`, memory in use). **Histogram** = bucketed distribution of observations (`http_request_duration_seconds`) — lets you compute percentiles. You graph **`rate(counter[5m])`** rather than the raw counter because the raw value is a meaningless ever-growing line (and resets to 0 on restart); the **per-second rate** is the actual signal — requests/sec, errors/sec — and it's restart-safe.
</details>

---

### 11. What is the RED method? Why percentiles (p99) over averages? What is cardinality and how do labels take down Prometheus?

<details><summary>Answer</summary>

**RED** = per endpoint, measure **Rate** (requests/sec), **Errors** (failed/sec), **Duration** (latency distribution) — the three numbers that describe request-driven service health. **Percentiles over averages** because the average hides the tail: a p99 of 2s means 1% of users wait 2 seconds even if the mean is 50ms — and the average is easily skewed by outliers in both directions. **Cardinality** is the number of unique label-combinations; each unique combo is a separate time series. Putting **unbounded values** (user IDs, raw URLs with IDs, request IDs) in labels explodes cardinality and **OOMs Prometheus** — keep labels low-cardinality (method, route *template*, status class).
</details>

---

### 12. Should you alert on high CPU? Why prefer symptom-based alerts?

<details><summary>Answer</summary>

Generally no — high CPU is a **cause**, not a symptom, and a service can be perfectly healthy while pegged (efficient utilization) or unhealthy at low CPU. Alert on **user-visible symptoms** — elevated error rate, p99 latency breaching SLO, request rate cratering, queue backing up — because those are what actually hurt users and warrant waking someone. Cause metrics (CPU, memory, GC) are for **investigation** once a symptom alert fires, not for paging. Symptom-based alerting cuts noise and false pages.
</details>

---

### 13. Trace vs span vs trace ID? How does a trace span multiple services?

<details><summary>Answer</summary>

A **trace** is the whole request's journey; a **span** is one unit of work within it (an HTTP handler, a DB call) with a start/end and attributes; the **trace ID** is the shared identifier stitching all spans of one request together (each span also has a span ID and parent ID forming a tree). A trace spans services via **context propagation**: the caller injects the trace context into outgoing request headers (the W3C **`traceparent`** header for HTTP), and the callee extracts it and continues the same trace — so spans across services link into one tree.
</details>

---

### 14. Why does every function taking `ctx` matter for tracing, and how do you correlate a log line with its trace?

<details><summary>Answer</summary>

Tracing rides on `context`: the active span lives in the `ctx`, so a function that **drops `ctx`** (or starts a fresh `Background()`) **breaks the trace** — its work shows up as an orphan span or not at all, leaving a hole in the latency picture. Threading `ctx` everywhere keeps the span tree intact. You correlate logs with traces by **logging the trace ID** (and span ID) as fields on every log line; then in your observability tooling you jump from a slow span straight to its logs, or from an error log to the full trace.
</details>

---

### 15. Liveness vs readiness — what does each control, and why must liveness NOT check the DB?

<details><summary>Answer</summary>

**Liveness (`/healthz`)** answers "is this process alive or wedged?" — failing it makes the orchestrator **restart** the pod. **Readiness (`/readyz`)** answers "can I serve traffic right now?" — failing it makes the orchestrator **stop routing** traffic to the pod without killing it. **Liveness must not check the database**: if the DB blips, every replica's liveness fails at once and Kubernetes **restarts your entire fleet** — turning a recoverable dependency hiccup into a self-inflicted total outage. Liveness checks only the process itself; dependency checks belong in readiness.
</details>

---

### 16. Why sample traces, what's tail-based sampling, and why a timeout on the readiness dependency check?

<details><summary>Answer</summary>

You **sample** because tracing every request at high traffic is prohibitively expensive in storage and overhead, and most traces are boring. **Head-based** sampling decides at the start (cheap but blind — may drop the rare slow/error trace). **Tail-based** sampling buffers the whole trace and decides *after* seeing the outcome, so you can **keep all errors and slow traces** and sample the fast successes — far better signal at the cost of buffering complexity. The readiness dependency check needs a **timeout** so a hung dependency doesn't make the readiness probe itself hang (which the orchestrator may misread); you want a fast, bounded "ready/not-ready" answer.
</details>

---

### 17. Explain the token bucket algorithm. What do `rate` and `burst` control?

<details><summary>Answer</summary>

A bucket holds up to `burst` tokens and refills at `rate` tokens/sec; each request consumes one token — if a token is available it proceeds, otherwise it's rejected (or waits). **`rate`** sets the sustained allowed throughput (long-run requests/sec); **`burst`** sets how many requests can spike through instantly when the bucket is full (the buffer for short bursts). Go's `golang.org/x/time/rate` implements exactly this (`rate.NewLimiter(r, b)`). It's preferred over fixed windows because it smooths bursts without the boundary-spike problem of window counters.
</details>

---

### 18. Why is a global in-memory rate limiter wrong across replicas, and what's the fix? What status/headers for a limited response?

<details><summary>Answer</summary>

An in-memory limiter is **per-process**, so with N replicas behind a load balancer each enforces its own limit and the **effective limit is N× the intended one** — the global cap is silently violated, and which replica you hit changes your quota. The fix is a **shared/distributed limiter** backed by Redis (atomic counters or a Lua token-bucket script) so all replicas share one budget. A rate-limited response returns **`429 Too Many Requests`** with **`Retry-After`** (seconds to wait) and ideally `RateLimit-Limit`/`RateLimit-Remaining`/`RateLimit-Reset` headers so clients can back off intelligently.
</details>

---

### 19. Walk through graceful shutdown with `signal.NotifyContext` + `srv.Shutdown`. Why flip readiness *before* draining?

<details><summary>Answer</summary>

`ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)` gives a context cancelled on a shutdown signal. On `<-ctx.Done()`, you flip **readiness to "not ready"** first (so the load balancer stops sending *new* requests), then call `srv.Shutdown(shutdownCtx)`, which **stops accepting new connections and waits for in-flight requests to finish** within a deadline before returning. You flip readiness first because there's lag before the LB notices a pod is going away; draining before that flip means new requests keep arriving and either get dropped or extend the drain — flip, let traffic stop, *then* drain the stragglers.
</details>

---

### 20. Why must the shutdown deadline be shorter than the orchestrator grace period? What if `http.Server{}` has no timeouts?

<details><summary>Answer</summary>

Kubernetes sends `SIGTERM`, waits `terminationGracePeriodSeconds`, then `SIGKILL`s the pod. If your `srv.Shutdown` deadline is **longer** than that grace period, the orchestrator hard-kills you mid-drain — **dropping in-flight requests** — defeating the whole point. Keep the shutdown deadline comfortably *under* the grace period so your clean drain finishes first. A `http.Server{}` with **no timeouts** is a Slowloris liability (slow clients pin connections forever) *and* makes graceful shutdown unbounded — a single slow request can stall the drain past the grace period; sane `Read/Write/Idle` timeouts bound both.
</details>
