# Day 23 Interview Questions — Metrics & Prometheus

Lesson questions plus extras. Answers in `<details>`.

---

### 1. Logs vs metrics vs traces — what's each best at?

<details>
<summary>Answer</summary>

**Logs**: discrete, high-detail events about one request ("what did request X
do?"). **Metrics**: cheap numeric aggregates over time across all requests
("what's the p99 and error rate right now?") — what alerts fire on. **Traces**:
one request's path across services ("where did the time go?"). You need all
three; correlate them by IDs.
</details>

---

### 2. Counter vs gauge vs histogram — an example of each.

<details>
<summary>Answer</summary>

**Counter** — monotonic, only up: `http_requests_total`, `errors_total` (graph
its `rate()`). **Gauge** — up and down: in-flight requests, queue depth, memory
in use. **Histogram** — bucketed samples for distributions: request duration,
which lets you compute percentiles (p50/p95/p99).
</details>

---

### 3. What is the RED method?

<details>
<summary>Answer</summary>

Per endpoint/service track **R**ate (requests/sec), **E**rrors (failed
requests/sec), **D**uration (latency distribution → percentiles). RED on your
HTTP handlers answers "is my service healthy?" for ~90% of cases. (USE —
Utilization/Saturation/Errors — is the resource-side counterpart.)
</details>

---

### 4. Why percentiles (p99) over averages for latency?

<details>
<summary>Answer</summary>

Averages hide the tail. A handful of multi-second requests barely move the mean
but ruin the experience for those users. p99/p95 expose exactly that tail, which
is what SLOs are written against. "Averages lie; percentiles tell the truth."
</details>

---

### 5. What is metric cardinality and how can labels take down Prometheus?

<details>
<summary>Answer</summary>

Cardinality is the number of distinct time series, and each unique *label-value
combination* is a new series. An unbounded label (user ID, raw URL, error string)
spawns millions of series, exhausting Prometheus memory and crashing it. Keep
labels bounded — templates, methods, status classes.
</details>

---

### 6. Why graph `rate(counter)` instead of the raw counter?

<details>
<summary>Answer</summary>

Counters only increase and reset to 0 on process restart/redeploy, so the raw
value is meaningless and dips on every deploy. `rate()` computes per-second
change over a window and handles resets correctly, giving you a stable
requests/sec (or errors/sec) signal.
</details>

---

### 7. Should you alert on high CPU? Why prefer symptom-based alerts?

<details>
<summary>Answer</summary>

Generally no. High CPU can be healthy (efficient use) and is a *cause*, not a
symptom — paging on it causes fatigue and false alarms. Alert on what hurts users
(elevated error rate, latency SLO breach, traffic dropping to zero) and keep
resource metrics for diagnosis on dashboards.
</details>

---

### 8. (Extra) Histogram vs summary — which and why?

<details>
<summary>Answer</summary>

Prefer **histogram**. It exposes raw buckets, so Prometheus can aggregate across
instances and compute quantiles server-side (`histogram_quantile`). A **summary**
computes quantiles client-side per instance and can't be meaningfully aggregated
across replicas. Use summaries only when you need exact client-side quantiles for
a single instance.
</details>

---

### 9. (Extra) Why is a counter increment a concurrency concern?

<details>
<summary>Answer</summary>

If you implement metrics over a shared map without synchronization, concurrent
request goroutines panic on concurrent map writes and lose increments
(`count++` is read-modify-write). The real prometheus client guards this with
atomics; a hand-rolled one needs a mutex. `go test -race` catches the bug. (Day
23 debugging exercise.)
</details>

---

### 10. (Extra) Write the PromQL for error rate and p99 latency.

<details>
<summary>Answer</summary>

Error rate: `sum(rate(http_requests_total{status=~"5.."}[5m])) /
sum(rate(http_requests_total[5m]))`. p99 latency:
`histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))`.
</details>

---

### 11. (Extra) What's a cardinality budget?

<details>
<summary>Answer</summary>

A deliberate limit on how many series a metric (or service) may produce,
enforced by capping label values and reviewing new labels. It prevents the slow
creep of high-cardinality labels that eventually OOMs the monitoring system, and
makes "added `user_id` as a label" a caught-in-review mistake instead of an
outage.
</details>
