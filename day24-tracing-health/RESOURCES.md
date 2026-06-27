# Day 24 Resources — Tracing & Health Checks

- **OpenTelemetry Go — getting started / instrumentation**
  https://opentelemetry.io/docs/languages/go/
  Tracer provider setup, spans, exporters, and instrumentation.

- **OpenTelemetry Go — GoDoc**
  https://pkg.go.dev/go.opentelemetry.io/otel
  API reference for `Tracer`, `Span`, attributes, status codes.

- **`otelhttp` instrumentation**
  https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp
  Auto-spans + header propagation for `net/http` handlers and clients.

- **W3C Trace Context spec**
  https://www.w3.org/TR/trace-context/
  The `traceparent`/`tracestate` headers that propagate traces across services.

- **Kubernetes — configure liveness, readiness & startup probes**
  https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/
  The canonical liveness vs readiness semantics and config.

- **`net/http` — `Request.Context` / `ServeMux`**
  https://pkg.go.dev/net/http#Request.Context
  Where the request context (and thus the span) comes from.

- **Jaeger — getting started**
  https://www.jaegertracing.io/docs/latest/getting-started/
  A common OTel backend to visualize trace waterfalls.

- **OpenTelemetry — sampling**
  https://opentelemetry.io/docs/concepts/sampling/
  Head-based vs tail-based sampling strategies.
