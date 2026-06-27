# Day 18 — REST, Config & Logging Cheatsheet

Quick reference. Standard library only (`net/http`, `encoding/json`, `os`,
`errors`, `log/slog`).

---

## HTTP status codes that matter

| Code | Name                  | Use when                                            |
|------|-----------------------|-----------------------------------------------------|
| 200  | OK                    | Successful GET / general success with a body        |
| 201  | Created               | POST created a resource (+ `Location` header)       |
| 204  | No Content            | Success, no body (e.g. DELETE)                       |
| 400  | Bad Request           | Malformed input you can't even parse (bad JSON)     |
| 401  | Unauthorized          | Not authenticated — missing/invalid credentials     |
| 403  | Forbidden             | Authenticated but not permitted                     |
| 404  | Not Found             | Resource doesn't exist                              |
| 409  | Conflict              | State conflict (duplicate, version mismatch)        |
| 422  | Unprocessable Entity  | Well-formed but fails validation/business rules     |
| 429  | Too Many Requests     | Rate limited                                        |
| 500  | Internal Server Error | Server bug/fault                                    |
| 503  | Service Unavailable   | Down / overloaded / dependency unavailable          |

Rule: the status code IS the API. Never `200` + an error body.

---

## Consistent error JSON shape

```json
{ "error": { "code": "validation_failed", "message": "price must be > 0", "fields": { "price": "must be > 0" } } }
```

```go
func writeError(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status) // status FIRST, then body
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": code, "message": msg},
	})
}
```

---

## Thin handler shape: decode → validate → service(ctx) → write

```go
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if err := req.Validate(); err != nil { // validation -> 422
		writeError(w, http.StatusUnprocessableEntity, "validation_failed", err.Error())
		return
	}
	p, err := h.svc.Create(r.Context(), req) // pass ctx down!
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "could not create")
		return
	}
	writeJSON(w, http.StatusCreated, p)
}
```

Handler stays thin: parse + map errors to codes. Business logic lives in the
service, which always takes `ctx` as its first argument.

---

## 12-factor config: getenv-with-default + fail-fast

```go
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string // required, NO safe default
	LogLevel    string
}

func Load() (Config, error) {
	cfg := Config{
		Port:        getenv("PORT", "8080"),     // optional + default
		DatabaseURL: os.Getenv("DATABASE_URL"),  // required
		JWTSecret:   os.Getenv("JWT_SECRET"),    // required, no default
		LogLevel:    getenv("LOG_LEVEL", "info"),// optional + default
	}
	if cfg.DatabaseURL == "" {
		return cfg, errors.New("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" { // fail fast at startup
		return cfg, errors.New("JWT_SECRET is required")
	}
	return cfg, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// in main:
cfg, err := Load()
if err != nil {
	log.Fatal(err) // crash loudly at boot, not silently at runtime
}
```

Precedence: a set, non-empty env var wins; otherwise the default. Required
secrets go through plain `os.Getenv` (never a default) and are validated.

---

## slog: JSON handler + request-scoped logger

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo, // threshold from config in real code
}))
slog.SetDefault(logger)

slog.Info("order created", "user_id", userID, "order_id", order.ID, "amount", order.Total)
// {"time":"...","level":"INFO","msg":"order created","user_id":"u_9",...}
```

Request-scoped logger — bind common fields once with `slog.With`, reuse:

```go
reqLog := slog.With("request_id", reqID, "user_id", userID)
reqLog.Info("handling request")
reqLog.Error("db failed", "err", err)
```

JSON handler in prod (machine-parseable); `slog.NewTextHandler` in dev
(readable). Pick by config.

---

## Logging rules

1. JSON in prod, text in dev.
2. One log line per event, with fields — not five `Println`s.
3. **Never** log secrets/PII (passwords, tokens, `Authorization`, card numbers). Redact.
4. Right level: `Error` only for actionable/pageable things; routine = `Info`/`Debug`.
5. Include correlation IDs (`request_id`, `trace_id`) to stitch a request's logs.
6. Log once, where you handle the error — don't log-and-return at every layer.
7. Consistent field names: `user_id` everywhere, never `userId` in one place and `uid` in another.

---

## Key terms

- **REST** — resource-oriented HTTP API: plural noun URIs (`/products/{id}`), verbs carry the action, status codes carry the outcome.
- **Idempotent** — repeating the request has the same effect as making it once; `GET`/`PUT`/`DELETE` are, `POST`/`PATCH` generally are not. Matters for safe retries.
- **12-factor** — methodology that keeps config (esp. secrets) in the environment, so one built artifact runs in every environment.
- **Fail-fast** — validate required config at startup and refuse to start (non-zero exit) if anything is missing, instead of failing on the first request.
- **Structured logging** — emitting events as key/value pairs (usually JSON) you can filter, aggregate, and alert on — not prose.
- **slog.Handler** — the logging backend that formats records and writes them (JSON/Text/custom); owns output, level, and redaction.
- **Correlation / request_id** — a unique ID attached to every log line of one request so its logs can be stitched together across layers and services.
- **slog.With** — returns a child logger with pre-resolved attributes bound once, reused on every call (fewer allocations than re-adding them per line).
