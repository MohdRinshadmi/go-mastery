# Day 18 — REST, Config & Logging Resources

Curated, real links. Start with the slog docs and the 12-factor pages.

## Structured logging (slog)

- **`log/slog` package docs** — the standard library reference: handlers,
  levels, `With`, `Group`, `HandlerOptions`.
  https://pkg.go.dev/log/slog
- **"Structured Logging with slog" (Go blog)** — the official introduction and
  design rationale by Jonathan Amsterdam.
  https://go.dev/blog/slog
- **"Logging in Go with slog: The Ultimate Guide" (Better Stack)** — practical,
  thorough guide: JSON vs text, levels, context, custom handlers, redaction.
  https://betterstack.com/community/guides/logging/logging-in-go/

## Configuration (12-factor)

- **12-Factor — Config** — config in the environment, strictly separated from
  code.
  https://12factor.net/config
- **12-Factor — Logs** — treat logs as event streams to stdout; let the platform
  route them.
  https://12factor.net/logs

## REST & HTTP

- **MDN — HTTP response status codes** — authoritative reference for every code
  and its precise meaning.
  https://developer.mozilla.org/en-US/docs/Web/HTTP/Status

## Going further

- **Alex Edwards — "Let's Go Further"** — building production-grade JSON APIs in
  Go: REST design, error JSON, config, structured logging, end to end.
  https://lets-go-further.alexedwards.net/
