# Coverage Map — does this program cover everything to master Go?

Short answer: **yes** — across the 30 days plus the four companion tracks. This file
is the index of *where* each mastery dimension lives, so nothing is "covered but
invisible."

## The mastery checklist

| Dimension | Status | Where it lives |
|-----------|--------|----------------|
| **Theory / concepts** | ✅ | Embedded in every `dayNN/lesson/DAYNN.md` (Theory + Why-it-exists sections) |
| **Examples** | ✅ | `dayNN/examples/` — runnable Go, beginner + production |
| **Exercises** | ✅ | `dayNN/exercises/` — your code; reviewed PR-style |
| **Solutions** | ✅ | `dayNN/solutions/` — reference answers |
| **Tradeoffs** | ✅ | "When to use / when NOT" in every lesson |
| **Mental models** | ✅ | "Expert Thinking Mode" (Beginner→Senior→Staff→Architect) in every lesson |
| **Real-world problems** | ✅ | "Real-world use" in every lesson |
| **Interview questions** | ✅ | Per-lesson sections **+** consolidated bank with answers → [`interview-prep/`](interview-prep/) |
| **Pitfalls** | ✅ | Per-lesson "Common mistakes" **+** consolidated → [`notes/pitfalls.md`](notes/pitfalls.md) |
| **Debugging challenges** | ✅ | [`debugging-challenges/`](debugging-challenges/) — added: one broken→fixed program per phase |
| **Notes / quick reference** | ✅ | [`notes/`](notes/) — cheatsheet, glossary (~50 terms), pitfalls |
| **Resources** | ✅ | [`resources/`](resources/) — docs, books, talks, tools, mapped to phases |
| **Mini projects / capstones** | ✅ | Phase capstones inside the days: Day 05 (CLI), Day 10 (profiling), Day 15 (URL checker/pipeline), Day 20 (E-commerce clean architecture), Day 29 (job queue), Day 30 (microservices platform + final exam) |

## Topic coverage by Go domain

- **Language core:** vars/const/iota, functions, slices/maps (+ aliasing), structs/pointers,
  control flow, the Go 1.22 loop-var fix — Days 01–03.
- **Errors:** sentinels, `%w` wrapping, `errors.Is/As`, custom types, panic/recover, defer — Day 04.
- **I/O & encoding:** files, `io.Reader/Writer`, JSON — Day 05.
- **Types & abstraction:** methods, method sets, interfaces, nil-interface gotcha, composition,
  DI, functional options, generics & constraints — Days 06–08.
- **Quality:** table-driven tests, mocking via interfaces, testify, coverage, benchmarks,
  pprof, escape analysis — Days 09–10.
- **Concurrency:** goroutines, channels, select, WaitGroup/Once, context, mutex/RWMutex/atomic,
  the memory model, worker pools, fan-out/in, pipelines, errgroup, race detector — Days 11–15.
- **Backend:** net/http, Gin, middleware, JWT, validation, REST, config, slog, Postgres, Redis,
  repository pattern, clean architecture — Days 16–20.
- **Production:** Docker, CI/CD, metrics/Prometheus, tracing/OTel, health checks, rate limiting,
  graceful shutdown — Days 21–25.
- **Advanced/distributed:** gRPC/protobuf, Kafka/RabbitMQ, caching patterns, performance tuning
  (GOGC/GOMEMLIMIT/sync.Pool), distributed job queue, microservices — Days 26–30.

## What was added in this pass

The 30 days were already complete. These cross-cutting tracks were the genuine gaps and
are now built and verified:

1. **`debugging-challenges/`** — 6 challenges (one per phase), each `bugged/` + verified `fixed/`.
2. **`interview-prep/`** — ~129 Q&A with model answers, organized by phase.
3. **`notes/`** — cheatsheet, glossary, consolidated pitfalls.
4. **`resources/`** — curated external learning hub mapped to phases.
