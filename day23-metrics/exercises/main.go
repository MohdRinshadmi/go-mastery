// Day 23 — YOUR exercises. Add Prometheus instrumentation.
// Run: go run .  then curl localhost:8080/metrics
package main

import (
	"net/http"
)

// =====================================================================
// TASK 1 — declare metrics (register in init):
//   httpRequests: CounterVec name "http_requests_total" labels method/path/status
//   httpDuration: HistogramVec name "http_request_duration_seconds" labels method/path
// (import github.com/prometheus/client_golang/prometheus)
// =====================================================================

// TODO: var ( httpRequests = ...; httpDuration = ... )
// TODO: func init() { prometheus.MustRegister(...) }

// =====================================================================
// TASK 2 — a statusRecorder to capture the response status code.
// =====================================================================

// TODO: type statusRecorder struct { http.ResponseWriter; status int }
// TODO: WriteHeader override

// =====================================================================
// TASK 3 — metricsMiddleware(routeLabel, next) that times the request and
// records the counter + histogram. Use a BOUNDED routeLabel, not r.URL.Path.
// =====================================================================

func metricsMiddleware(routeLabel string, next http.HandlerFunc) http.HandlerFunc {
	// TODO
	return next
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", metricsMiddleware("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello\n"))
	}))
	// TASK 4: mux.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", mux)
}

// PromQL to write in your notes:
//   error rate:  rate(http_requests_total{status=~"5.."}[5m])
//   p99 latency: histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))
