# Day 23 Resources — Metrics & Prometheus

- **Prometheus `client_golang` — instrumenting a Go app**
  https://prometheus.io/docs/guides/go-application/
  Counters, histograms, registering collectors, exposing `/metrics`.

- **`client_golang` GoDoc**
  https://pkg.go.dev/github.com/prometheus/client_golang/prometheus
  API reference for `CounterVec`, `HistogramVec`, `MustRegister`, `promhttp`.

- **Prometheus — metric types**
  https://prometheus.io/docs/concepts/metric_types/
  Counter vs gauge vs histogram vs summary, with guidance.

- **Prometheus — naming & labels best practices**
  https://prometheus.io/docs/practices/naming/
  Naming conventions and bounded-label discipline.

- **Prometheus — histograms and quantiles**
  https://prometheus.io/docs/practices/histograms/
  Buckets, `histogram_quantile`, and why histograms aggregate.

- **The RED Method (Tom Wilkie / Grafana)**
  https://grafana.com/blog/2018/08/02/the-red-method-how-to-instrument-your-services/
  Rate, Errors, Duration per service.

- **Google SRE Book — Monitoring & the four golden signals**
  https://sre.google/sre-book/monitoring-distributed-systems/
  Symptom-based alerting and signal selection.

- **Prometheus Alerting — best practices**
  https://prometheus.io/docs/practices/alerting/
  Symptom vs cause alerting; alert design.
