# Interview Prep — Consolidated Q&A Bank

> Mentor note: Your 30 daily lessons each ended with a handful of interview questions, but scattered across 30 files with no answers, they're useless under pressure. This folder is your **single-source quiz bank**. I've pulled every interview question from the curriculum, grouped them by phase, added the classics that *will* come up in a senior/staff Go loop, and written a crisp model answer for each — the kind of answer I'd want to hear across the table, not a textbook dump.

## How to use this bank

1. **Self-quiz first.** Each answer is hidden inside a `<details>` block. Read the question, say your answer out loud (or write it), *then* expand to check. If you peek first, you're not studying — you're reading.
2. **Answer like an engineer, not a flashcard.** The model answers are 2–6 sentences with a code snippet where the code is the clearest argument. In a real interview, lead with the one-sentence claim, then justify it. Brevity reads as mastery.
3. **Chase the "why".** Every answer here ties back to *why Go made a design choice* or *what breaks in production if you get it wrong*. That's the layer that separates "knows syntax" from "senior".
4. **Run the code.** Where a snippet illustrates a gotcha (slice aliasing, nil interface, loop-var capture), paste it into a scratch `main.go` and watch it misbehave. Surprise cements memory.
5. **Loop weak spots.** Mark questions you fumbled and re-quiz a day later. Spaced repetition beats re-reading.

## Topics map

| File | Phase | Days | Covers |
|------|-------|------|--------|
| [`01-fundamentals.md`](01-fundamentals.md) | 1 — Go Fundamentals | 01–05 | zero values, `defer`/`iota`/shadowing, slice internals & append growth, maps & sets, structs & pointers, receivers, `range` semantics, error values vs exceptions, `%w`/`errors.Is`/`As`, files & JSON, `io.Reader` design |
| [`02-core-engineering.md`](02-core-engineering.md) | 2 — Core Engineering | 06–10 | methods & method sets, implicit interfaces, small-interface design, nil-interface gotcha, type assertions/switches, embedding & composition, DI & "accept interfaces, return structs", functional options, generics & constraints, table-driven tests & mocking, `-race`, benchmarks/pprof/escape analysis |
| [`03-concurrency.md`](03-concurrency.md) | 3 — Concurrency | 11–15 | goroutines vs threads, GMP scheduler, channels (buffered/unbuffered/closed/directional), `select` & deadlock, goroutine leaks, `WaitGroup`/`Once`/`Mutex`/`RWMutex`/atomic, `context` cancellation, the Go memory model, worker pools, fan-out/fan-in, `errgroup`, pipelines |
| [`04-backend.md`](04-backend.md) | 4 — Backend Development | 16–20 | `net/http` `Handler`/`HandlerFunc`, 1.22 routing, server timeouts, middleware, JWT & `alg:none`, RBAC, config (12-factor, fail-fast), structured logging (`slog`), repository pattern, parameterized queries, connection pools, clean architecture & dependency direction |
| [`05-production.md`](05-production.md) | 5 — Production Engineering | 21–25 | Docker (`scratch`/distroless, `CGO_ENABLED=0`, layer caching, `.dockerignore`), CI/CD (`go vet`, `golangci-lint`, SHA tags, secrets), observability (logs/metrics/traces, RED, histograms, cardinality), liveness vs readiness, rate limiting (token bucket), graceful shutdown |
| [`06-advanced.md`](06-advanced.md) | 6 — Advanced Go | 26–30 | gRPC vs REST, HTTP/2 & streaming, protobuf evolution, gRPC status codes & interceptors, Kafka vs RabbitMQ, partitions & consumer groups, at-least-once & idempotency, DLQ, caching (cache-aside, stampede, penetration), CAP & consistency models, `GOGC`/`GOMEMLIMIT`/`sync.Pool`, job queues, outbox & sagas, end-to-end request tracing |

## A word before you walk in

The Go interview rewards *precision about tradeoffs*. Nobody senior is impressed that you can recite "goroutines are cheap" — they want to know *how* cheap (a few KB stack, grown on demand), *who* schedules them (the runtime, GMP, not the OS), and *what goes wrong* (leaks, unsynchronized shared state, the nil-interface trap). Aim every answer at the failure mode. That's the voice this whole program has been training.
