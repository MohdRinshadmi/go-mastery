# Day 18 — REST, Config & Logging Interview Questions

The lesson's seven, plus four deeper ones. Each answer is what a senior would
actually say in the room.

---

### 1. Why return proper HTTP status codes instead of `200` + an error body?

<details>
<summary>Answer</summary>

The status code is part of the API contract. Every HTTP client library, proxy,
load balancer, retry policy, and monitoring tool reads the status to decide what
happened — success vs client error vs server error. If you return `200` with
`{"error": ...}` in the body:

- Clients must parse the body to learn the call failed; generic error handling
  (`if !res.ok`) never triggers.
- Monitoring/alerting that watches 4xx/5xx rates sees a healthy `200` — your
  outage is invisible on the dashboard.
- Retry/circuit-breaker logic keyed on 5xx won't fire.
- Caches may cache the "successful" error response.

Map your error taxonomy to codes at the HTTP boundary: `400` malformed input,
`401` unauthenticated, `403` forbidden, `404` not found, `409` conflict, `422`
validation failed, `429` rate limited, `500` server fault, `503` down. Write the
status *before* the body.

</details>

---

### 2. What does 12-factor say about configuration, and why env vars over config files?

<details>
<summary>Answer</summary>

12-factor says config — anything that varies between deploys (DB URLs, secrets,
ports, feature flags) — lives in the **environment**, strictly separated from
code. The *same built artifact* runs in dev/staging/prod, configured only by env
vars.

Env vars over committed config files because:

- **No secrets in git.** A committed `config.prod.yaml` with a DB password is a
  leak waiting to happen.
- **Language/OS agnostic and easy to inject.** Containers, K8s, and CI all set
  env vars natively; no file templating or per-env build needed.
- **One artifact, many environments.** You don't rebuild to change a port; you
  change an env var and restart.

The nuance: env vars are great for *secrets and per-deploy values*. Large
structured config (routing tables, lengthy lists) can still live in files —
just not the secrets, and ideally not committed.

</details>

---

### 3. Why validate config at startup (fail fast) instead of on first use?

<details>
<summary>Answer</summary>

Because a service that starts "successfully" and then 500s on every request is
worse than one that refuses to start. Validating at boot turns a missing
`DATABASE_URL` into **one loud, clear error at deploy time** instead of N
scattered runtime failures far from the cause.

Operationally, fail-fast plugs into the platform: a non-zero exit or a failing
readiness probe tells K8s/systemd "don't route traffic here," so the bad deploy
never takes traffic and (with rollout policies) gets rolled back automatically.
Discovering the problem on the first user request means the deploy looked green,
traffic shifted, and *then* it broke.

Concretely: validate all required-with-no-default values in `Load()`, return an
error, and `log.Fatal` / `os.Exit(1)` in `main`.

</details>

---

### 4. What is structured logging and why does it beat `fmt.Println` in production?

<details>
<summary>Answer</summary>

Structured logging emits each event as a set of **key/value pairs** (usually
JSON) rather than a prose string: `{"level":"INFO","msg":"order
created","user_id":"u_9","order_id":"o_1"}`. Go's `log/slog` (1.21+) does this in
the standard library.

It beats `fmt.Println` because logs in a real system are read by **machines
first**: shipped to Loki/Elastic/Datadog, then filtered, aggregated, and
alerted on. With structure you can run "all `level=ERROR` for `user_id=X` in the
last hour across 5 services" — a query that's impossible against free-form text.
You also get consistent fields, levels you can threshold by, and one searchable
line per event instead of five `Println`s.

</details>

---

### 5. What should you never log? How do correlation IDs help during an incident?

<details>
<summary>Answer</summary>

**Never log** secrets or PII: passwords, raw tokens/API keys, `Authorization`
headers, full card numbers, and personal data you don't have a reason to retain.
Logs are shipped to third parties, kept for months, and read by many people — so
a token in a log line is a credential leak and a compliance breach. Log stable
identifiers (`user_id`) and redact the rest.

**Correlation IDs** (`request_id`, `trace_id`) are a unique value attached to
every log line for a single request. During an incident they let you stitch one
request's logs together — across handlers, the service layer, the DB, and other
microservices it called. Instead of guessing which of 10,000 interleaved lines
belong to the failing request, you filter by its `request_id` and see the whole
path. Generate it (or accept an inbound one) at the edge, put it on the
request-scoped logger with `slog.With`, and propagate it downstream.

</details>

---

### 6. When do you use `Error` vs `Warn` vs `Info`?

<details>
<summary>Answer</summary>

- **`Error`** — something is broken and needs human attention; this is what
  pages someone or drives alerts. A DB that's down, a panic recovered, a write
  that failed and lost data. If a `404` for a missing record logs `Error`,
  you'll page on-call for normal traffic.
- **`Warn`** — unusual or degraded but handled; worth noticing, not worth waking
  anyone. A retry that succeeded on the 2nd attempt, a deprecated endpoint hit,
  approaching a quota.
- **`Info`** — routine, expected business events: request served, order created,
  job finished. The default narrative of a healthy system.
- (`Debug` — verbose developer detail, off in prod or sampled.)

The rule: `Error` must be *actionable*. Crying wolf with `Error` for handled,
expected conditions trains everyone to ignore the channel, so real errors get
missed. Set the threshold via config (`LOG_LEVEL`).

</details>

---

### 7. Why pass `r.Context()` into your service and DB calls?

<details>
<summary>Answer</summary>

`r.Context()` carries the request's **deadline, cancellation signal, and
request-scoped values** (like a trace/request ID). Threading it down means:

- **Cancellation propagates.** If the client disconnects or the request deadline
  fires, `ctx.Done()` closes and a `QueryContext`/HTTP call honoring `ctx` is
  aborted — you stop burning a connection and CPU on a result nobody will read.
- **Deadlines are enforced** end to end, so one slow dependency can't hang a
  request forever.
- **Trace context flows**, so downstream spans/logs link back to the request.

Using `context.Background()` in a handler detaches all of that: queries outlive
their request and trace correlation is lost. Rule: handlers get `ctx` from
`r.Context()` and pass it as the first argument to every service/DB call.

</details>

---

### 8. (Deeper) PUT vs PATCH vs POST — and what does idempotency mean here?

<details>
<summary>Answer</summary>

- **POST** — create a new resource (or a non-idempotent action). `POST /orders`
  twice creates *two* orders. Not idempotent. Typically returns `201 Created`
  with a `Location` header.
- **PUT** — full replace of a resource at a known URI. `PUT /products/42` with
  the complete representation. **Idempotent**: sending the same body N times
  leaves the resource in the same final state as sending it once.
- **PATCH** — partial update: send only the fields to change. Idempotency
  depends on the patch — a "set price=10" patch is idempotent; an "increment
  stock by 1" patch is not.

**Idempotent** means *repeating the request has the same effect as making it
once*. It matters for retries: a client (or proxy) can safely retry a `PUT`,
`GET`, or `DELETE` after a timeout without fear of duplicating an effect — but
retrying a `POST` may double-charge a card, which is why payment APIs add an
*idempotency key*. `GET`/`PUT`/`DELETE` are idempotent; `POST`/`PATCH`
generally are not.

</details>

---

### 9. (Deeper) What do 401, 403, and 422 mean precisely, and when do they differ?

<details>
<summary>Answer</summary>

- **401 Unauthorized** — *authentication* failed or is missing. The server
  doesn't know who you are: no/invalid/expired credentials. "Who are you? Log
  in." Spec-wise it should include a `WWW-Authenticate` header. Retrying with
  valid credentials can succeed.
- **403 Forbidden** — *authorization* failed. The server knows who you are
  (authenticated) but you're not allowed to do this — wrong role, not the
  resource owner. "I know you, and no." Re-authenticating won't help.
- **422 Unprocessable Entity** — the request was well-formed (valid JSON) and
  you're allowed, but the *content* fails business/validation rules: `price`
  must be > 0, email already taken. Distinct from `400`, which is for syntax you
  couldn't even parse.

Mnemonic: `401` = not logged in, `403` = logged in but not permitted, `422` =
logged in, permitted, but the data is invalid. (`400` = couldn't parse it at
all.)

</details>

---

### 10. (Deeper) How does `slog.With` reduce allocations? Handler vs logger?

<details>
<summary>Answer</summary>

`slog.With(args...)` returns a **new `*slog.Logger`** whose attributes are
pre-resolved and stored once. Every log call on that logger reuses them instead
of re-parsing and re-allocating the same `request_id`/`user_id` attrs on every
line. So for a request that logs 10 times, you build the common attrs *once* (in
`With`) rather than 10 times — fewer allocations on the hot path. The key/value
variadic form (`"k", v`) is also cheaper than building `slog.Attr` values
yourself in many cases.

Logger vs handler: the **`slog.Handler`** is the backend — it formats records
and writes them (JSONHandler, TextHandler, or a custom one), and owns the
output, level threshold, and `ReplaceAttr`. The **`*slog.Logger`** is the
front-end you call (`Info`, `With`, ...); it builds a `Record` and hands it to
its handler. You configure cross-cutting policy (format, level, redaction) once
in the handler; you create per-request loggers cheaply with `With`. Note
`With`'s pre-resolution is a property of the handler implementation, which the
stdlib JSON/Text handlers provide.

</details>

---

### 11. (Deeper) JSON handler vs text handler — which in dev, which in prod?

<details>
<summary>Answer</summary>

- **`slog.NewJSONHandler`** emits one JSON object per line. Use it in **prod**:
  it's machine-parseable, so Loki/Elastic/Datadog can index every field and you
  can query/aggregate/alert. Hard to read by eye, but in prod machines read logs
  first.
- **`slog.NewTextHandler`** emits `key=value` pairs in a line. Use it in **dev**:
  it's far easier for a human to scan at the terminal while iterating.

Pick the handler from config (e.g. `LOG_FORMAT` / environment), keep the same
`*slog.Logger` call sites, and you get readable local logs and structured prod
logs from identical code. Either way, set the level threshold from config and
keep the field schema (`user_id`, `request_id`) consistent across both.

</details>
