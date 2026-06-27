# Day 23 — Metrics, Monitoring & Prometheus

> Mentor note: Logs tell you what happened in *one* request. **Metrics** tell you what's happening across *all* requests, right now: request rate, error rate, latency percentiles, resource use. When you're on call, metrics are the dashboard you stare at and the thing that pages you at 3am. The industry-standard stack in the Go world is **Prometheus** (scrapes & stores metrics) + **Grafana** (dashboards). Today you instrument a service so it exposes `/metrics` that Prometheus can scrape.

---

## 1. The three pillars of observability
- **Logs** (Day 18): discrete events, high detail, one request. "What did request X do?"
- **Metrics** (today): numeric aggregates over time, cheap, system-wide. "What's the p99 latency and error rate?"
- **Traces** (Day 24): one request's path across services. "Where did the time go?"

You need all three. Metrics are the cheapest and the first line of defense — they're what alerts fire on.

## 2. The four metric types
- **Counter** — only goes up (requests total, errors total). You graph its *rate*.
- **Gauge** — goes up and down (in-flight requests, queue depth, memory).
- **Histogram** — samples bucketed into ranges (request duration). Lets you compute **percentiles** (p50/p95/p99) — the single most important latency signal. Averages lie; percentiles tell the truth about tail latency.
- **Summary** — like histogram but computes quantiles client-side; prefer Histogram in most cases (aggregatable across instances).

## 3. The RED method (what to measure for a service)
For every endpoint/service, track:
- **R**ate — requests per second.
- **E**rrors — failed requests per second.
- **D**uration — latency distribution (histogram → percentiles).

(For resources, the USE method: Utilization, Saturation, Errors.) RED on your HTTP handlers covers 90% of "is my service healthy?"

## 4. Instrumenting with the Prometheus Go client

`go get github.com/prometheus/client_golang/prometheus github.com/prometheus/client_golang/prometheus/promhttp`

```go
var (
    httpRequests = prometheus.NewCounterVec(
        prometheus.CounterOpts{Name: "http_requests_total", Help: "Total HTTP requests"},
        []string{"method", "path", "status"},   // labels = dimensions you can slice by
    )
    httpDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "Request duration",
            Buckets: prometheus.DefBuckets, // .005 .01 ... 10s
        },
        []string{"method", "path"},
    )
)

func init() {
    prometheus.MustRegister(httpRequests, httpDuration)
}

// middleware records RED metrics for every request
func metricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        rec := &statusRecorder{ResponseWriter: w, status: 200}
        next.ServeHTTP(rec, r)
        dur := time.Since(start).Seconds()
        httpRequests.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(rec.status)).Inc()
        httpDuration.WithLabelValues(r.Method, r.URL.Path).Observe(dur)
    })
}

// expose for Prometheus to scrape:
mux.Handle("/metrics", promhttp.Handler())
```

Prometheus is configured to **scrape** `http://yourservice/metrics` every N seconds and store the time series. Grafana queries Prometheus with **PromQL**:
- `rate(http_requests_total[5m])` — requests/sec.
- `rate(http_requests_total{status=~"5.."}[5m])` — error rate.
- `histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))` — p99 latency.

### Labels — power and danger
Labels (`method`, `path`, `status`) let you slice metrics. But **every unique label combination is a new time series** — high-cardinality labels (user ID, request ID, raw URLs with IDs in them) explode memory and can take Prometheus down. Use bounded labels: route *templates* (`/orders/{id}`), not actual paths (`/orders/12345`).

**Senior take:** Cardinality is the #1 way people break Prometheus. Never put unbounded values (user IDs, emails, full URLs, error messages) in labels. The label set must be small and bounded. This is the metrics equivalent of a memory leak.

## 5. Alerting
Prometheus Alertmanager fires alerts on PromQL conditions: "error rate > 1% for 5m", "p99 > 500ms", "no requests for 2m (service down)". Alerts should be **symptom-based** (users are seeing errors) not cause-based (CPU is high — which may be fine). Page on what hurts users; everything else is a dashboard.

## Common mistakes
1. High-cardinality labels → Prometheus OOM.
2. Measuring averages instead of percentiles — averages hide the tail where users suffer.
3. Counters reset on restart — always graph `rate()`, never the raw counter.
4. Alerting on causes (CPU, memory) instead of symptoms (errors, latency) → alert fatigue.
5. No `/metrics` endpoint or forgetting to register collectors (`MustRegister`).
6. Instrumenting everything → noise. Start with RED per endpoint.

## Performance
- Counters/gauges are near-free (atomic adds). Histograms cost a bit more (bucketing) but are fine at high QPS.
- The `/metrics` scrape serializes all series — keep cardinality sane and it's cheap.

---

## Expert Thinking Mode — "is my service healthy?"

- **Beginner:** "I'll check the logs / SSH in and look at CPU."
- **Senior:** "RED metrics per endpoint, p99 latency, error rate on a dashboard, alerts on symptoms. `/metrics` scraped by Prometheus, visualized in Grafana."
- **Staff:** "SLOs (e.g. 99.9% of requests < 300ms) with error budgets; alerts tied to budget burn rate. Cardinality budget for labels. Metrics consistent across services for fleet-wide dashboards."
- **Architect:** "Observability is a platform: standardized metric names, RED/USE conventions, SLO framework, cost of cardinality and retention. Metrics drive autoscaling and capacity planning."

---

## Real-world use

- **Prometheus + Grafana** is the default Go observability stack (Cloudflare, Uber, and most cloud-native shops; it's a CNCF project born partly from Go infra).
- **RED dashboards** per service are standard; **histogram_quantile** for p99 is the canonical latency query.
- **Cardinality incidents** (someone added `user_id` as a label) are a well-known class of outage.

---

## Interview Questions

1. Logs vs metrics vs traces — what's each best at?
2. Counter vs gauge vs histogram — give an example metric for each.
3. What is the RED method? What three things do you measure per endpoint?
4. Why percentiles (p99) over averages for latency?
5. What is metric cardinality and how can labels take down Prometheus?
6. Why always graph `rate(counter)` instead of the raw counter?
7. Should you alert on high CPU? Why prefer symptom-based alerts?

---

## Your tasks

`../exercises/` has an HTTP service missing its instrumentation. Add: (1) a `http_requests_total` counter with `method/path/status` labels, (2) a `http_request_duration_seconds` histogram, (3) a middleware that records both, and (4) the `/metrics` endpoint. Run it, hit some routes, and `curl /metrics` to see your series. Then write the PromQL for error-rate and p99. Reference + a Grafana panel note in `../solutions/`.
