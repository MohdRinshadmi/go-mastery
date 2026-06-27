# Resources — Curated Go Learning Hub

A working engineer's bookmark file. Everything here is something I'd actually open during a real workday — not a link dump. Skim it once now so you know what exists, then come back when a specific phase needs it. The phase map at the bottom tells you *which* of these to reach for and *when*.

A few ground rules from the mentor chair:
- **Read the docs before the blog post.** `pkg.go.dev` and the spec answer 80% of questions faster than any tutorial.
- **One tool at a time.** Don't install all of Section 4 on day one. Add each as the bootcamp calls for it.
- **Recency matters.** Go moves. Prefer release notes and `go.dev` over a 2018 Stack Overflow answer.

---

## 1. Official & Canonical

The source of truth. When a blog post and the spec disagree, the spec wins.

| Resource | What it is | URL |
|----------|-----------|-----|
| **A Tour of Go** | Interactive, in-browser intro to the language. Your warm-up. | https://go.dev/tour/ |
| **Effective Go** | How idiomatic Go is actually written. Read it twice. | https://go.dev/doc/effective_go |
| **The Go Programming Language Specification** | The precise rules. Terse but authoritative — your tie-breaker. | https://go.dev/ref/spec |
| **Standard library docs** | The std lib reference. Live here. | https://pkg.go.dev/std |
| **pkg.go.dev** | Docs + searchable index for every public module/package. | https://pkg.go.dev/ |
| **The Go Blog** | Official deep-dives (generics, errors, slices internals, etc.). | https://go.dev/blog/ |
| **golang.org/x** | Official-but-not-std extended packages (`x/sync/errgroup`, `x/time/rate`, `x/tools`). | https://pkg.go.dev/golang.org/x |
| **Go release notes** | What changed each version. Skim every release. | https://go.dev/doc/devel/release |
| **Go Memory Model** | The rules for what concurrent reads/writes are guaranteed. Read before Phase 3. | https://go.dev/ref/mem |
| **Go Modules reference** | The full `go.mod` / versioning / `go mod` story. | https://go.dev/ref/mod |
| **Frequently Asked Questions (FAQ)** | "Why does Go do X?" — answered by the team. | https://go.dev/doc/faq |

---

## 2. Books

You don't need all of these. Pick one as your primary, keep *100 Go Mistakes* as a reference.

- **The Go Programming Language** — Donovan & Kernighan ("the K&R of Go"). The canonical book. Dense, precise, timeless. *For:* anyone who wants the rigorous foundation. (https://www.gopl.io/)
- **Learning Go** — Jon Bodner (O'Reilly), 2nd ed. The best *modern* intro — covers generics and current idioms. *For:* your day-to-day primary text alongside this bootcamp.
- **100 Go Mistakes and How to Avoid Them** — Teiva Harsanyi (Manning). 100 concrete traps with fixes. *For:* leveling up from "it works" to "it's correct" — read a few mistakes per day. (https://100go.co/ has a free companion site)
- **Concurrency in Go** — Katherine Cox-Buday (O'Reilly). The deep, careful treatment of channels, patterns, and the memory model. *For:* Phase 3, when goroutines stop being magic.
- **Let's Go** & **Let's Go Further** — Alex Edwards. Build a real web app with the std lib, then harden it (auth, JSON APIs, deployment). *For:* Phase 4–5; the most practical backend Go books available. (https://lets-go.alexedwards.net/)

---

## 3. Talks & Videos

Watch these on a slow afternoon. They shape how you *think*, not just what you type.

- **Concurrency is not Parallelism** — Rob Pike. The single most important talk for understanding Go's concurrency model. *Watch before Phase 3.* https://go.dev/blog/waza-talk (video + transcript)
- **Go Proverbs** — Rob Pike (GopherCon 2015). Short, quotable design wisdom ("Don't communicate by sharing memory; share memory by communicating"). https://go-proverbs.github.io/
- **GopherCon talks** — The official channel. Highlights: "Understanding Channels" (Kavya Joshi), "Aggregate Programming" patterns, runtime/GC deep-dives. https://www.youtube.com/@GopherAcademy
- **Dave Cheney's writing & talks** — Practical Go, "SOLID Go Design", error-handling essays. Some of the best applied Go thinking anywhere. https://dave.cheney.net/
- **Justforfunc (Francesc Campoy)** — Hands-on episodes building and profiling real Go. Great for seeing the workflow. https://www.youtube.com/c/justforfunc

---

## 4. Tools

Your toolbelt. Install each one *when the bootcamp reaches it* (noted in the phase map), not all at once.

| Tool | What it does | Install |
|------|-------------|---------|
| **gopls** | The official language server (autocomplete, refactor, errors in-editor). Usually auto-installed by your editor. | `go install golang.org/x/tools/gopls@latest` |
| **golangci-lint** | The meta-linter — runs dozens of linters in one pass. The standard in CI. | `brew install golangci-lint` (or see https://golangci-lint.run/welcome/install/) |
| **staticcheck** | High-signal static analysis (subset is inside golangci-lint, but great standalone). | `go install honnef.co/go/tools/cmd/staticcheck@latest` |
| **delve (dlv)** | The Go debugger. Breakpoints, stepping, goroutine inspection. | `go install github.com/go-delve/delve/cmd/dlv@latest` |
| **pprof** | CPU/heap/block/mutex profiling. Built into the toolchain. | Built in: `go tool pprof` (+ `net/http/pprof`) |
| **race detector** | Finds data races at runtime. Non-negotiable for concurrent code. | Built in: `go test -race` / `go run -race` |
| **benchstat** | Compares benchmark results with statistical confidence. | `go install golang.org/x/perf/cmd/benchstat@latest` |
| **govulncheck** | Scans your deps (and call graph) for known vulnerabilities. | `go install golang.org/x/vuln/cmd/govulncheck@latest` |
| **air** | Live reload for web servers — recompiles on save. | `go install github.com/air-verse/air@latest` |

> Built-in friends you already have: `go vet`, `go test -cover`, `gofmt`/`go fmt`, `go mod tidy`. Use them constantly.

---

## 5. Style & Idioms

Go has unusually strong consensus on style. Internalize these and your PRs get shorter review cycles.

- **Effective Go** — the foundation (also listed in §1). https://go.dev/doc/effective_go
- **Go Code Review Comments** — the community checklist reviewers actually cite. Short and worth memorizing. https://go.dev/wiki/CodeReviewComments
- **Google Go Style Guide** — comprehensive, opinionated, used internally at Google. https://google.github.io/styleguide/go/
- **Uber Go Style Guide** — practical, example-heavy, widely adopted in industry. https://github.com/uber-go/guide/blob/master/style.md
- **Go Doc Comments** — how to write docs the tooling renders well. https://go.dev/doc/comment

---

## 6. Practice

Reading isn't learning — reps are. Use these *between* bootcamp days to stay sharp.

- **Exercism — Go track** — Free exercises with real human mentorship. The best structured practice. https://exercism.org/tracks/go
- **Go by Example** — Annotated, runnable snippets for nearly every language feature. Your quick-reference. https://gobyexample.com/
- **Gophercises** — Jon Calhoun's free project-based exercises (CLI tools, parsers, etc.). https://gophercises.com/
- **Advent of Code** — December puzzles; doing them in Go forces clean problem-solving. https://adventofcode.com/
- **LeetCode (in Go)** — For interview prep specifically. Set your language to Go and grind patterns. https://leetcode.com/

---

## 7. Staying Current

Go ships ~twice a year and the ecosystem moves with it. Wire up a couple of these so news finds you.

- **Golang Weekly** — The newsletter of record. One email a week, high signal. https://golangweekly.com/
- **r/golang** — Active subreddit; good for "is this a good library?" sanity checks. https://www.reddit.com/r/golang/
- **Gophers Slack** — The big community chat (`#general`, `#newbies`, topic channels). Invite: https://invite.slack.golangbridge.org/
- **Dave Cheney** — Applied Go, performance, API design. https://dave.cheney.net/
- **Eli Bendersky** — Deep technical posts; excellent on Go internals and the runtime. https://eli.thegreenplace.net/
- **Ardan Labs / Bill Kennedy** — Mechanical-sympathy-focused Go (memory, scheduling, performance). https://www.ardanlabs.com/blog/
- **Go release blog posts** — Each major release gets a write-up on https://go.dev/blog/ — read the one for whatever version you're on.

---

## 8. Map to This Bootcamp's Phases

Don't open everything at once. For each phase, here's what to keep on the second monitor.

| Phase (days) | Focus | Reach for these |
|--------------|-------|-----------------|
| **1 — Fundamentals** (01–05) | Modules, types, errors, files, JSON | A Tour of Go (§1) · Go by Example (§6) · *Learning Go* (§2) · Effective Go (§5) |
| **2 — Core Engineering** (06–10) | Interfaces, generics, testing, profiling | The Go Blog on generics (§1) · `go test`/benchstat/pprof (§4) · *100 Go Mistakes* (§2) |
| **3 — Concurrency** (11–15) | Goroutines, channels, context, pools | "Concurrency is not Parallelism" (§3) · Go Memory Model (§1) · *Concurrency in Go* (§2) · `-race` detector (§4) |
| **4 — Backend** (16–20) | net/http, middleware, Postgres/Redis, clean arch | *Let's Go* / *Let's Go Further* (§2) · `air` live reload (§4) · `golang.org/x` (`x/time/rate`) (§1) |
| **5 — Production** (21–25) | Docker, CI/CD, metrics, tracing, shutdown | golangci-lint + govulncheck in CI (§4) · Google/Uber style guides (§5) · Ardan Labs perf blog (§7) |
| **6 — Advanced** (26–30) | gRPC, Kafka, caching, distributed systems | pkg.go.dev for protobuf/grpc-go (§1) · *Concurrency in Go* (§2) · Eli Bendersky internals (§7) · GopherCon distributed-systems talks (§3) |

---

*Tip:* Bookmark this file. When you finish Day 30, Sections 6 and 7 are how you keep growing after the bootcamp ends.
