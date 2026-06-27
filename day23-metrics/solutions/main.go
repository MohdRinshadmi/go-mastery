// Day 23 — reference solution. Run: go run .
package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "http_requests_total", Help: "Total HTTP requests"},
		[]string{"method", "path", "status"},
	)
	httpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func init() { prometheus.MustRegister(httpRequests, httpDuration) }

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func metricsMiddleware(routeLabel string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next(rec, r)
		httpRequests.WithLabelValues(r.Method, routeLabel, strconv.Itoa(rec.status)).Inc()
		httpDuration.WithLabelValues(r.Method, routeLabel).Observe(time.Since(start).Seconds())
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", metricsMiddleware("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello\n"))
	}))
	mux.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", mux)
}
