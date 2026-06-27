// Day 23 walkthrough — Prometheus RED metrics. Run: go run .
// Then: curl localhost:8080/hello ; curl localhost:8080/metrics | grep http_
package main

import (
	"log/slog"
	"math/rand"
	"net/http"
	"os"
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
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
	inFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{Name: "http_in_flight_requests", Help: "In-flight requests"},
	)
)

func init() {
	prometheus.MustRegister(httpRequests, httpDuration, inFlight)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// metricsMiddleware records RED metrics for every request.
func metricsMiddleware(routeLabel string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		inFlight.Inc()
		defer inFlight.Dec()
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next(rec, r)
		// Use a bounded route label, NOT r.URL.Path (cardinality!)
		httpRequests.WithLabelValues(r.Method, routeLabel, strconv.Itoa(rec.status)).Inc()
		httpDuration.WithLabelValues(r.Method, routeLabel).Observe(time.Since(start).Seconds())
	}
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	mux := http.NewServeMux()

	mux.HandleFunc("/hello", metricsMiddleware("/hello", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Duration(rand.Intn(30)) * time.Millisecond) // simulate work
		w.Write([]byte("hello\n"))
	}))
	mux.HandleFunc("/maybe-error", metricsMiddleware("/maybe-error", func(w http.ResponseWriter, r *http.Request) {
		if rand.Intn(2) == 0 {
			http.Error(w, "boom", http.StatusInternalServerError)
			return
		}
		w.Write([]byte("ok\n"))
	}))
	mux.Handle("/metrics", promhttp.Handler())

	port := "8080"
	slog.Info("listening", "port", port, "metrics", "/metrics")
	http.ListenAndServe(":"+port, mux)
}
