# Day 18 — REST, Config & Logging Pitfalls (Trap → Why → Fix)

Six traps that turn a working demo into an unoperatable service. Each is
**Trap → Why → Fix**.

---

## 1. Returning `200 OK` with an error body

**Trap.**

```go
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if err := req.Validate(); err != nil {
		// status stays 200 (default); error hidden in the body
		json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}
}
```

**Why.** The status code is part of your API. `200` means success — every HTTP
client library, proxy, retry policy, and monitoring tool reads it that way. A
`200` with `{"error": ...}` means clients must parse the body to know if the
call worked, your dashboards never see the failure (they alert on 4xx/5xx), and
retries don't fire. It's the classic junior mistake.

**Fix.** Map your error taxonomy to real status codes at the HTTP boundary, and
write the code explicitly *before* the body.

```go
func writeError(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status) // set the status FIRST
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": code, "message": msg},
	})
}

writeError(w, http.StatusUnprocessableEntity, "validation_failed", err.Error())
```

---

## 2. Not failing fast on missing required config

**Trap.**

```go
func Load() Config {
	return Config{
		DatabaseURL: os.Getenv("DATABASE_URL"), // empty? nobody notices
		JWTSecret:   os.Getenv("JWT_SECRET"),
	}
}
```

**Why.** The service starts "successfully" and then 500s on the first request
that touches the DB or signs a token. The failure shows up far from its cause —
as scattered runtime errors instead of one clear boot error — after the deploy
already looked green.

**Fix.** Validate required config at startup and return an error; exit non-zero.

```go
func Load() (Config, error) {
	cfg := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
	}
	if cfg.DatabaseURL == "" {
		return cfg, errors.New("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return cfg, errors.New("JWT_SECRET is required")
	}
	return cfg, nil
}
// in main: if err != nil { log.Fatal(err) }  // crash loudly at boot
```

---

## 3. Giving a required secret a default (or hardcoding it)

**Trap.**

```go
JWTSecret: getenv("JWT_SECRET", "dev-secret-change-me"), // BAD default
// ...or worse:
const jwtSecret = "s3cr3t-hardcoded" // committed to git
```

**Why.** A default secret is a committed secret. If `JWT_SECRET` is unset in
prod, the service silently signs tokens with `dev-secret-change-me` — which is
in your public repo, so anyone can forge a valid token. A leaked signing secret
is a security incident, not a config quirk.

**Fix.** Secrets are *required-with-no-default*. Read them with plain
`os.Getenv`, never `getenv(key, def)`, and validate at boot (see #2). Use
`getenv(key, def)` only for genuinely optional values like `PORT` or
`LOG_LEVEL`.

```go
JWTSecret: os.Getenv("JWT_SECRET"), // no default; validated to be non-empty
Port:      getenv("PORT", "8080"),  // safe default — fine
```

---

## 4. Unstructured logs (`fmt.Println` / `log.Printf`) in prod

**Trap.**

```go
fmt.Println("user", userID, "created order", orderID, "for", amount)
log.Printf("got request %s", r.URL.Path)
```

**Why.** Plain text is unsearchable. At 3am you need "all errors for user X in
the last hour across 5 services" — that requires key/value fields a log platform
(Loki/Elastic/Datadog) can filter and aggregate. A line of prose can't be
queried, joined, or alerted on.

**Fix.** Use `log/slog` with a JSON handler; emit one line per event with
fields.

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))
slog.SetDefault(logger)

slog.Info("order created", "user_id", userID, "order_id", orderID, "amount", amount)
// {"time":"...","level":"INFO","msg":"order created","user_id":"u_9",...}
```

---

## 5. Logging secrets / PII

**Trap.**

```go
slog.Info("login attempt", "email", email, "password", password, "token", jwt)
slog.Debug("request", "authorization", r.Header.Get("Authorization"))
```

**Why.** Logs get shipped to third-party platforms, retained for months, and
read by lots of people. Passwords, tokens, full card numbers, and personal data
in logs are a compliance breach and a credential leak. Logs are not a safe place
for secrets.

**Fix.** Never log credentials or raw PII. Log stable identifiers instead, and
redact anything sensitive.

```go
slog.Info("login attempt", "user_id", userID, "result", "ok") // no email/password
// if you must reference a token, log a non-reversible hint, never the value:
slog.Debug("auth", "token_present", jwt != "")
```

---

## 6. Not passing `r.Context()` into the service/DB layer

**Trap.**

```go
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.Get(context.Background(), id) // wrong context!
	// ...
}
```

**Why.** `context.Background()` is detached from the request. When the client
disconnects or the request deadline fires, your DB query keeps running — wasting
a connection and CPU on a result nobody will read. You also lose any
request-scoped values (trace IDs, deadlines) carried on `r.Context()`.

**Fix.** Thread `r.Context()` from the handler all the way down to the
service and DB calls.

```go
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.Get(r.Context(), id) // cancellation + deadline + trace flow down
	// ...
}

func (s *Service) Get(ctx context.Context, id string) (Product, error) {
	return s.db.QueryRowContext(ctx, "SELECT ... WHERE id = $1", id) // honors ctx
}
```
