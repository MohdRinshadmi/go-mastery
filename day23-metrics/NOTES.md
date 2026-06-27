# Day 23 Notes — Metrics & Prometheus (quick reference)

## Three pillars of observability
- **Logs** — discrete events, one request, high detail.
- **Metrics** — cheap numeric aggregates over all requests; alerts fire here.
- **Traces** — one request's path across services.

## Four metric types
| Type | Behavior | Example | Graph as |
|---|---|---|---|
| Counter | only up | `http_requests_total` | `rate()` |
| Gauge | up & down | in-flight reqs, queue depth | raw value |
| Histogram | bucketed samples | request duration | `histogram_quantile` |
| Summary | client-side quantiles | (prefer histogram) | quantile series |

## RED method (per endpoint)
- **R**ate — requests/sec.
- **E**rrors — failed requests/sec.
- **D**uration — latency distribution → percentiles.

## Instrumentation sketch (client_golang)
```go
httpRequests := prometheus.NewCounterVec(
    prometheus.CounterOpts{Name: "http_requests_total"},
    []string{"method", "path", "status"})       // BOUNDED labels only
prometheus.MustRegister(httpRequests, httpDuration)
mux.Handle("/metrics", promhttp.Handler())
httpRequests.WithLabelValues(m, templatePath(p), code).Inc()
```

## PromQL you must know
| Question | Query |
|---|---|
| req/sec | `rate(http_requests_total[5m])` |
| error rate | `rate(http_requests_total{status=~"5.."}[5m])` |
| p99 latency | `histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))` |

## Cardinality rules
- Each unique label combo = one time series.
- NEVER label with: user ID, email, request ID, raw URL, error message.
- DO label with: route template (`/orders/{id}`), method, status class.
- Concurrency: counter updates must be thread-safe (atomics/mutex) — `go test -race`.

## Alerting
- Alert on **symptoms** (errors, latency SLO burn, no-traffic), not **causes** (CPU).
- Counters reset on restart → always `rate()`.

## Key terms
- **Cardinality** — count of distinct time series.
- **RED / USE** — service-level / resource-level metric methods.
- **Histogram bucket (`_bucket`, `le`)** — cumulative count ≤ upper bound.
- **`histogram_quantile`** — percentile from histogram buckets.
- **Scrape** — Prometheus pulling `/metrics` on an interval.
- **SLO / error budget** — reliability target and allowed failure.
