# Day 10 Resources — Benchmarking & Profiling

Curated, real links. Start with the testing docs and the Go blog's pprof post.

- **[testing package — Benchmarks](https://pkg.go.dev/testing#hdr-Benchmarks)**
  The authoritative reference for `BenchmarkXxx`, `b.N`, `b.Loop`,
  `b.ResetTimer`, `b.ReportAllocs`, and the `-bench`/`-benchmem` flags.

- **[Profiling Go Programs (Go blog)](https://go.dev/blog/pprof)**
  The canonical worked example of using `pprof` to find and fix a real hotspot —
  the benchmark → profile → fix → re-benchmark loop in action.

- **[Diagnostics (go.dev/doc/diagnostics)](https://go.dev/doc/diagnostics)**
  The official overview of Go's profiling, tracing, and debugging tools and when
  to reach for each — a map of the whole diagnostics toolbox.

- **[Dave Cheney — High Performance Go Workshop](https://dave.cheney.net/high-performance-go-workshop/dotgo-paris.html)**
  The best end-to-end tour of benchmarking, pprof, escape analysis, inlining,
  and avoiding allocations, with hands-on exercises. Read this in full.

- **[benchstat docs](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)**
  How to install and read benchstat: comparing runs, the ± variation, and the
  p-value that tells you whether a change is real or noise.

- **[Go 1.24 release notes — testing.B.Loop](https://go.dev/doc/go1.24#testing)**
  Where the modern `for b.Loop()` benchmark form was introduced, explaining how
  it prevents dead-code elimination and excludes setup automatically.

- **[Allocation efficiency / escape analysis in Go (segment.com engineering)](https://segment.com/blog/allocation-efficiency-in-high-performance-go-services/)**
  A practical deep dive on reading `go build -gcflags=-m`, understanding what
  escapes to the heap, and cutting allocations in hot paths.
