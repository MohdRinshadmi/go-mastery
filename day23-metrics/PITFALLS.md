# Day 23 Pitfalls — Metrics & Prometheus

Format: **Trap → Why → Fix**

---

### 1. High-cardinality labels
**Trap:** Labeling a metric with `user_id`, `email`, request ID, or the raw URL (`/orders/12345`).
**Why:** Every unique label combination is a *new time series*. Unbounded values create millions of series and OOM Prometheus — the metrics equivalent of a memory leak. It's a famous outage category.
**Fix:** Use bounded labels only — route *templates* (`/orders/{id}`), method, status class. Normalize paths before recording. If you must capture an ID, it goes in a log/trace, not a label.

---

### 2. Unsynchronized metric updates
**Trap:** Incrementing a counter map from many request goroutines without a lock.
**Why:** Concurrent map writes panic (`concurrent map writes`) and `count++` loses updates under contention → undercounted, untrustworthy dashboards.
**Fix:** Use a thread-safe metric (the prometheus client uses atomics internally; a hand-rolled one needs a mutex). Run `go test -race` to catch it. (This is the Day 23 debugging exercise.)

---

### 3. Graphing the raw counter instead of `rate()`
**Trap:** Dashboard plots `http_requests_total` directly.
**Why:** Counters only go up and **reset to 0 on restart/redeploy**. The raw line is meaningless and dips to zero on every deploy.
**Fix:** Always graph `rate(http_requests_total[5m])` — requests/sec, restart-safe.

---

### 4. Averages instead of percentiles
**Trap:** Alerting/dashboarding on mean latency.
**Why:** Averages hide the tail. A p50 of 20ms can coexist with a p99 of 3s — the slow requests your users actually feel are invisible.
**Fix:** Use a histogram and `histogram_quantile(0.99, rate(..._bucket[5m]))`. SLOs are stated in percentiles, not averages.

---

### 5. Alerting on causes, not symptoms
**Trap:** Paging on "CPU > 80%" or "memory high".
**Why:** High CPU may be perfectly fine (efficient use). Cause-based alerts create fatigue and miss real user pain.
**Fix:** Page on **symptoms** — error rate, latency SLO burn, "no requests for 2m". Keep resource metrics on dashboards for diagnosis, not paging.

---

### 6. Forgetting to register collectors / expose `/metrics`
**Trap:** You create metrics but never `MustRegister` them, or never mount `promhttp.Handler()`.
**Why:** Unregistered metrics don't appear in the scrape; no `/metrics` endpoint means Prometheus has nothing to scrape — silent blind spot.
**Fix:** Register at init (`prometheus.MustRegister(...)`) and `mux.Handle("/metrics", promhttp.Handler())`. Verify with `curl /metrics`.

---

### 7. Instrumenting everything → noise
**Trap:** A metric on every function and branch.
**Why:** Thousands of low-value series cost memory and bury the signal.
**Fix:** Start with **RED per endpoint** (Rate, Errors, Duration). Add more only when a question demands it.
