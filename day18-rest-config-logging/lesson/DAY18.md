# Day 18 — REST API Design, Configuration, Structured Logging (slog)

> Mentor note: Days 16–17 gave you HTTP handlers, middleware, and JWT. Today is the unglamorous stuff that separates a demo from a service you can actually operate: a clean REST contract, configuration done right (12-factor), and **structured logging**. When it's 3am and prod is on fire, the thing that saves you is logs you can search by `request_id` and `user_id` — not `fmt.Println`. Go 1.21 put structured logging in the standard library (`log/slog`); there's no excuse anymore.

---

## 1. REST API design (the parts people get wrong)

- **Resources are nouns, plural**: `/products`, `/orders/{id}`. Not `/getProduct`.
- **HTTP verbs carry the action**: `GET` (read, safe, idempotent), `POST` (create), `PUT` (full replace, idempotent), `PATCH` (partial), `DELETE` (idempotent).
- **Status codes mean things**: `200` OK, `201` Created (+ `Location` header), `204` No Content, `400` bad input, `401` unauthenticated, `403` authenticated-but-forbidden, `404` not found, `409` conflict, `422` validation, `429` rate-limited, `500` server fault, `503` down.
- **Errors have a consistent JSON shape** so clients can parse them:
  ```json
  { "error": { "code": "validation_failed", "message": "price must be > 0", "fields": {"price": "must be > 0"} } }
  ```
- **Versioning**: `/v1/products` or a header. Decide early; it's a contract.

**Senior take:** The status code is part of your API. Returning `200` with `{"error": ...}` in the body (because "the request reached the server") is a classic junior mistake — it breaks every client's error handling and every monitoring tool that alerts on 5xx. Map your error taxonomy (Day 4) to status codes at the HTTP boundary.

### A clean handler shape
```go
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateProductRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
        return
    }
    if err := req.Validate(); err != nil {     // validation -> 422
        writeError(w, http.StatusUnprocessableEntity, "validation_failed", err.Error())
        return
    }
    p, err := h.svc.Create(r.Context(), req)   // pass ctx down!
    if err != nil {
        writeError(w, http.StatusInternalServerError, "internal", "could not create")
        return
    }
    writeJSON(w, http.StatusCreated, p)
}
```
Note: decode → validate → call service (with `r.Context()`) → write. The handler is thin; business logic lives in the service (Day 20).

## 2. Configuration — 12-factor

**Config comes from the environment, not hardcoded constants, not committed files.** The same binary runs in dev/staging/prod, configured by env vars.

```go
type Config struct {
    Port        string
    DatabaseURL string
    JWTSecret   string
    LogLevel    string
}

func Load() (Config, error) {
    cfg := Config{
        Port:        getenv("PORT", "8080"),
        DatabaseURL: os.Getenv("DATABASE_URL"),
        JWTSecret:   os.Getenv("JWT_SECRET"),
        LogLevel:    getenv("LOG_LEVEL", "info"),
    }
    if cfg.JWTSecret == "" {
        return cfg, errors.New("JWT_SECRET is required")  // fail fast at startup
    }
    return cfg, nil
}
func getenv(key, def string) string {
    if v := os.Getenv(key); v != "" { return v }
    return def
}
```

Rules:
- **Secrets never in code or git.** Env vars (or a secret manager). A leaked `JWTSecret` in a commit is a security incident.
- **Fail fast**: validate config at startup and exit if something required is missing. Don't discover a missing DB URL on the first request.
- Libraries like `viper` add files/flags/precedence; for many services plain env + a struct is enough. Don't over-engineer config.

**Senior take:** Config validation at boot is non-negotiable. A service that starts "successfully" then 500s every request because `DATABASE_URL` was empty is worse than one that refuses to start with a clear error. Crash loudly at startup, not silently at runtime.

## 3. Structured logging with `log/slog` (Go 1.21+)

`fmt.Println`/`log.Printf` produce unsearchable text. **Structured logging** emits key-value pairs (usually JSON) you can filter and aggregate.

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
slog.SetDefault(logger)

slog.Info("order created",
    "order_id", order.ID,
    "user_id", userID,
    "amount", order.Total,
)
// -> {"time":"...","level":"INFO","msg":"order created","order_id":"o_1","user_id":"u_9","amount":42}
```

- **Levels**: `Debug`, `Info`, `Warn`, `Error`. Set the threshold via config.
- **Context-scoped loggers**: attach request-scoped fields once and reuse:
  ```go
  reqLog := slog.With("request_id", reqID, "user_id", userID)
  reqLog.Info("handling request")
  reqLog.Error("db failed", "err", err)
  ```
- **`slog.Group`** nests related fields.

### Logging rules that matter in prod
1. **JSON in prod** (machine-parseable for Loki/Elastic/Datadog), text in dev for readability.
2. **One log line per event**, with fields — not five `Println`s.
3. **Never log secrets/PII** (passwords, tokens, full card numbers). Redact.
4. **Log at the right level**: `Error` only for things needing attention (they page someone); routine stuff is `Info`/`Debug`. Crying wolf with `Error` trains everyone to ignore it.
5. **Include correlation IDs** (`request_id`, `trace_id`) so you can stitch one request's logs across services (Day 24 ties this to tracing).
6. Don't log-and-return the same error at every layer (Day 4) — log once, where you handle it.

**Senior take:** Logs are for *machines first, humans second* in a real system. The question isn't "can I read this line" but "can I query all errors for user X in the last hour across 5 services." That requires structure and consistent field names. Standardize them (`user_id` everywhere, never `userId` in one place and `uid` in another).

## Common mistakes
1. `200` with an error body. Use real status codes.
2. Secrets/config hardcoded or committed.
3. `fmt.Println` debugging left in; unstructured logs.
4. Logging PII/secrets.
5. Not passing `r.Context()` into the service/DB layer (loses cancellation + trace).
6. Over-logging `Error` for handled, expected conditions.

## Performance
- `slog` is allocation-conscious; prefer the key-value variadic form, and use `slog.With` once per request rather than rebuilding attributes per line.
- JSON encoding has a cost; at extreme volume, sample debug logs and keep hot-path logging lean.

---

## Expert Thinking Mode — "add logging and config"

- **Beginner:** "`log.Println("got request")` and a hardcoded port."
- **Senior:** "Structured JSON logs with request_id/user_id, levels from config, env-based config validated at startup, correct status codes mapped from error types."
- **Staff:** "Consistent field schema across services so dashboards/alerts work. No PII. Logs correlate with traces and metrics (the three pillars). Config is 12-factor and secrets come from a manager."
- **Architect:** "Observability is a platform concern: log schema, retention, cost, and the contract between services. Error taxonomy → status codes → alerts → SLOs is one continuous chain."

---

## Real-world use

- **Every modern Go service** uses `slog` (or zap/zerolog predating it) emitting JSON to stdout, scraped by the platform.
- **12-factor config** is how containers/K8s inject per-environment settings (Phase 5).
- **Correlation IDs** in logs are what let an on-call engineer trace one failing request across a microservice mesh.

---

## Interview Questions

1. Why return proper HTTP status codes instead of `200` + error body?
2. What does 12-factor say about configuration, and why env vars over config files?
3. Why validate config at startup (fail fast) instead of on first use?
4. What is structured logging and why does it beat `fmt.Println` in production?
5. What should you never log? How do correlation IDs help during an incident?
6. When do you use `Error` vs `Warn` vs `Info`?
7. Why pass `r.Context()` into your service and DB calls?

---

## Your tasks

`../exercises/` has a small REST service skeleton: a `/products` endpoint, a `Config.Load` to finish (env + validation + fail-fast), and slog wiring to complete (JSON handler, request-scoped logger with a request_id, correct status codes + consistent error JSON). Run it, `curl` it, and bring me the log output + your config validation. Reference in `../solutions/`.
