# Debugging Challenge — Config That Starts Broken

A service `Config.Load()` reads its environment, returns no error, and the app
boots cleanly. Then the *first* request that needs to sign a token 500s with an
empty signing secret. The config was broken at startup — `Load()` just never
said so. This is the signature config gotcha of Day 18: **fail at runtime
instead of failing fast at boot.**

## Symptom

`JWT_SECRET` is not set in the environment (a deployment forgot it). Yet:

- `Load()` returns `(cfg, nil)` — no error.
- The server "starts successfully" on port 8080.
- The first request that signs a JWT blows up with a 500, because
  `cfg.JWTSecret == ""`.

A service that starts and then fails every request is worse than one that
refuses to start with a clear message.

## Repro

Bugged (wrong — silent success):

```sh
cd /Users/ioss/Documents/StudyProjects/GO/day18-rest-config-logging/debugging/bugged
go run .
```

Expected (buggy) output:

```
Load() ok, JWTSecret="" (BUG: started with no secret!)
server starting on port 8080
first request -> 500: cannot sign JWT with empty secret
```

(exit code `0` — the process thinks everything is fine.)

Fixed (correct — fails fast at boot):

```sh
cd /Users/ioss/Documents/StudyProjects/GO/day18-rest-config-logging/debugging/fixed
go run .
```

Expected (correct) output:

```
Load() failed fast: JWT_SECRET is required
```

(exit code `1` — the process refuses to start.)

Both programs `os.Unsetenv("JWT_SECRET")` inside `main()` so the demo is
deterministic regardless of what is in your real shell.

## Hint

Look at how each field is loaded in `Load()`. `PORT`, `DATABASE_URL`, and
`LOG_LEVEL` all flow through `getenv(key, default)` — they have *safe defaults*.
What default does `JWT_SECRET` get? Should a signing secret have one at all?
Where is the value validated before `Load()` returns `nil`?

<details>
<summary>Solution &amp; why</summary>

There are two distinct kinds of configuration value, and they must be handled
differently:

- **Optional, with a safe default** — `PORT` (→ `8080`), `LOG_LEVEL` (→ `info`).
  A `getenv(key, default)` helper is exactly right here.
- **Required, with no safe default** — `JWT_SECRET`. There is no sensible
  fallback: an empty signing secret is a security hole, not a convenience. It
  *must* be present, and the only correct response to its absence is to refuse
  to start.

The bug treats the required value as if it were optional-ish: it reads it but
never validates it, so a missing secret slides through as an empty string.

```go
// BUG: read but never checked — empty secret escapes as a valid config.
func Load() (Config, error) {
	cfg := Config{
		Port:        getenv("PORT", "8080"),
		DatabaseURL: getenv("DATABASE_URL", "postgres://localhost:5432/app"),
		JWTSecret:   os.Getenv("JWT_SECRET"), // <- no guard after this
		LogLevel:    getenv("LOG_LEVEL", "info"),
	}
	return cfg, nil // returns nil even when JWTSecret == ""
}
```

The fix is to **fail fast**: validate every required-with-no-default variable
at the end of `Load()` and return an error if it is missing. `main` then exits
non-zero, and the bad deploy is caught at boot — in the logs, in the crash
loop, in the readiness probe — not by a user hitting a 500.

```go
// FIX: validate required config at boot; crash loudly, not silently.
func Load() (Config, error) {
	cfg := Config{ /* ...same... */ }
	if cfg.JWTSecret == "" {
		return cfg, errors.New("JWT_SECRET is required")
	}
	return cfg, nil
}
```

**Why fail-fast beats fail-at-runtime:**

- A startup error is *one* loud event at deploy time, with a clear message.
- A runtime error is *N* failed requests scattered across logs, far from the
  cause, often after the deploy looked green.
- Orchestrators (K8s, systemd) treat a non-zero exit / failing liveness probe
  as "don't route traffic here" — which is exactly what you want.

**Rules of thumb:**

- Sort config into *optional-with-default* vs *required-no-default*. Route the
  first through `getenv(key, def)`; **validate** the second explicitly.
- A required secret must NEVER be given a default. A default secret is a
  committed secret waiting to happen.
- Validate all required config at startup and exit non-zero if anything is
  missing. Crash loudly at boot, not silently at runtime.
- `go vet` will not catch this — "missing validation" is a logic gap, not a
  type error. Treat it as a code-review checklist item.

</details>
