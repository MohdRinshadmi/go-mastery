# Elite GoLang Mastery — 30 Day Program

A production-focused Go bootcamp. Taught like a senior engineer mentoring a junior on a real team — not a university course.

## How this works

1. Each day lives in its own folder: `dayNN-topic/`.
2. You **do not advance** until you pass the day's exercises and quiz (target: 7/10+).
3. Every lesson follows the same loop:
   - Concept → Why it exists → When to use / when NOT to → Examples (beginner + production) → Common mistakes → Performance → Exercises → Real-world → Interview Qs.
4. You write code in `exercises/`. I review it like a production PR. Reference answers live in `solutions/` (don't peek until you've tried).

## Folder layout per day

```
dayNN-topic/
  lesson/        # the written lesson (markdown)
  examples/      # runnable Go I walk you through
  exercises/     # YOUR code goes here
  solutions/     # reference answers (try first!)
  debugging/     # find & fix the bug: bugged/ (broken) + fixed/ (verified)
  PITFALLS.md    # this day's gotchas: Trap → Why → Fix
  INTERVIEW.md   # this day's interview Q&A (answers in collapsible blocks)
  NOTES.md       # this day's quick-reference cheatsheet + key terms
  RESOURCES.md   # curated links for this day's topic
```

Each lesson embeds: Theory · Why it exists · When to use / when NOT (tradeoffs) ·
Examples · Common mistakes (pitfalls) · Performance · Expert Thinking Mode (mental
models) · Real-world use · Interview Questions — and links to the five companion
files above at the bottom. **Every one of the 30 days now carries this full set.**

## Companion tracks (use alongside the days)

These top-level folders cut *across* all 30 days — pull them in whenever you need them:

| Track | What it's for |
|-------|---------------|
| [`debugging-challenges/`](debugging-challenges/) | "Find & fix the bug" exercises (one per phase): slice aliasing, nil interface, data race, context leak, ticker leak, unbounded queue. Each has a `bugged/` and a verified `fixed/`. |
| [`interview-prep/`](interview-prep/) | Consolidated Q&A bank (~129 questions) with model answers in collapsible blocks — one file per phase, for self-quizzing. |
| [`notes/`](notes/) | Quick reference: `cheatsheet.md`, `glossary.md` (~50 terms), and a consolidated `pitfalls.md` (Trap → Why → Fix). |
| [`resources/`](resources/) | Curated external resources — official docs, books, talks, tools, style guides — mapped to each phase. |
| [`COVERAGE.md`](COVERAGE.md) | The full coverage map: every mastery topic and where it lives. |

## Roadmap

| Phase | Days  | Focus |
|-------|-------|-------|
| 1 | 01–05 | Go Fundamentals |
| 2 | 06–10 | Core Engineering (interfaces, generics, testing, profiling) |
| 3 | 11–15 | Concurrency |
| 4 | 16–20 | Backend Development (E-Commerce API) |
| 5 | 21–25 | Production Engineering (Docker, CI/CD, observability) |
| 6 | 26–30 | Advanced Go (gRPC, Kafka, distributed systems) |

## Status

**All 30 days are built and every Go file is compile-verified** (88 modules, 0 build/test failures). Work them **in order** — the "don't advance until you pass the exercises + quiz" rule is about *your learning*, not about whether the material exists. Bring each day's `exercises/` to your mentor for a PR-style review before moving on.

> Tip: pre-existing folders may use slightly different names than the table below (e.g. `day02-slices-maps`). Run `ls` to see them.

## Curriculum (all delivered)

**Phase 1 — Fundamentals**
- [ ] Day 01 — `day01-go-fundamentals` — Install, Modules, Variables, Constants, Functions
- [ ] Day 02 — `day02-slices-maps` — Packages, Arrays, Slices (aliasing), Maps
- [ ] Day 03 — `day03-structs-pointers` — Structs, Pointers, Control Flow
- [ ] Day 04 — `day04-errors` — Error handling (sentinels, %w, custom types, panic/recover)
- [ ] Day 05 — `day05-files-json` — Files, io.Reader/Writer, JSON + capstone CLI

**Phase 2 — Core Engineering**
- [ ] Day 06 — `day06-methods-interfaces` — Methods, Interfaces
- [ ] Day 07 — `day07-composition-di` — Composition, Dependency Injection
- [ ] Day 08 — `day08-generics` — Generics
- [ ] Day 09 — `day09-testing-mocking` — Testing, table-driven tests, mocking
- [ ] Day 10 — `day10-benchmark-profiling` — Benchmarking, profiling + capstone

**Phase 3 — Concurrency**
- [ ] Day 11 — `day11-goroutines-channels` — Goroutines, Channels
- [ ] Day 12 — `day12-select-sync` — Buffered channels, select, WaitGroups
- [ ] Day 13 — `day13-context-mutex` — Context, Mutexes, race detector
- [ ] Day 14 — `day14-worker-pools` — Worker pools, fan-out/in, errgroup
- [ ] Day 15 — `day15-pipelines` — Pipelines + capstone (URL checker, pipeline)

**Phase 4 — Backend**
- [ ] Day 16 — `day16-net-http` — net/http, Gin
- [ ] Day 17 — `day17-middleware-jwt` — Middleware, JWT, validation
- [ ] Day 18 — `day18-rest-config-logging` — REST, config, slog
- [ ] Day 19 — `day19-postgres-redis-repo` — Postgres, repository pattern, Redis
- [ ] Day 20 — `day20-ecommerce-architecture` — Clean architecture E-Commerce capstone

**Phase 5 — Production**
- [ ] Day 21 — `day21-docker` — Docker, Docker Compose
- [ ] Day 22 — `day22-cicd` — CI/CD, GitHub Actions
- [ ] Day 23 — `day23-metrics` — Logging, metrics, Prometheus
- [ ] Day 24 — `day24-tracing-health` — Tracing (OTel), health checks
- [ ] Day 25 — `day25-ratelimit-shutdown` — Rate limiting, graceful shutdown

**Phase 6 — Advanced**
- [ ] Day 26 — `day26-grpc-protobuf` — gRPC, Protocol Buffers
- [ ] Day 27 — `day27-messaging` — Kafka, RabbitMQ, event-driven
- [ ] Day 28 — `day28-redis-caching` — Caching patterns, distributed systems
- [ ] Day 29 — `day29-performance-jobqueue` — Performance + distributed job queue
- [ ] Day 30 — `day30-final-evaluation` — Microservices platform + **Final Exam**

## A note on external infra (Days 19, 21, 26, 27, 28)

Every runnable demo works **offline** using in-memory implementations. The real
Postgres / Redis / Kafka / RabbitMQ / OTel code is included as `*_reference.go`
files (build-tagged `//go:build ignore`) plus `docker-compose.yml` where relevant —
present and documented, but not required to compile the offline demos. Remove the
build tag and `docker compose up` to run the real thing.

## Running code

```bash
cd day01-go-fundamentals/examples
go run 01_hello.go
```

Scores and weaknesses get logged in `PROGRESS.md`.
