# Day 25 Resources — Rate Limiting & Graceful Shutdown

- **`golang.org/x/time/rate` — GoDoc**
  https://pkg.go.dev/golang.org/x/time/rate
  Token-bucket `Limiter`, `Allow`, `Wait`, `Reserve`, `Limit`, `Burst`.

- **`net/http` — `Server.Shutdown`**
  https://pkg.go.dev/net/http#Server.Shutdown
  Graceful drain semantics and the `ErrServerClosed` contract.

- **`net/http` — `Server` (timeout fields)**
  https://pkg.go.dev/net/http#Server
  `ReadTimeout`, `WriteTimeout`, `IdleTimeout`, `ReadHeaderTimeout`,
  `MaxHeaderBytes`.

- **`os/signal` — `NotifyContext`**
  https://pkg.go.dev/os/signal#NotifyContext
  Turning SIGINT/SIGTERM into a cancellable context.

- **Cloudflare blog — exposing Go on the internet (timeouts)**
  https://blog.cloudflare.com/exposing-go-on-the-internet/
  Why the default `http.Server` timeouts are dangerous; slow-loris.

- **`net/http` — `MaxBytesReader`**
  https://pkg.go.dev/net/http#MaxBytesReader
  Bounding request body size against giant-payload DoS.

- **Kubernetes — pod termination & `terminationGracePeriodSeconds`**
  https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination
  The SIGTERM→grace→SIGKILL sequence your shutdown deadline must beat.

- **MDN — HTTP 429 Too Many Requests**
  https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/429
  The status code and `Retry-After` header for rate limiting.
