# Day 10 Pitfalls — Benchmarking & Profiling

Benchmarks are easy to write and easy to lie with. Each trap below is a way a
real engineer has shipped a wrong number. Read them as **Trap → Why → Fix**.

---

### 1. The compiler eliminates un-sunk work (the fake ns/op)

**Trap:** Your benchmark calls a pure function and discards the result; it
reports ~0.3 ns/op.

**Why:** A pure, inlinable function whose return value nobody reads is *dead
code*. The optimizer deletes it, so you time an empty loop. The number is
unrelated to the function and won't change when the function gets heavier.

**Fix:** Assign the result to a **package-level sink** variable
(`sink = work()`), or use the modern `for b.Loop()` form (Go 1.24+), which is
built to keep the loop body alive. A local `_ =` is not reliably enough. See the
[debugging challenge](debugging/README.md).

---

### 2. Counting setup time (missing `b.ResetTimer`)

**Trap:** You build a 1M-element fixture inside the benchmark, then loop. Your
ns/op is dominated by the one-time setup, not the operation under test.

**Why:** The timer starts before your loop. Everything between the start of the
function and the loop is folded into the measurement and divided across
iterations, inflating ns/op.

**Fix:** Do the expensive setup first, then call `b.ResetTimer()` immediately
before the `for` loop so only the measured work is timed. (Use `b.StopTimer()` /
`b.StartTimer()` if you must do per-iteration setup.)

---

### 3. Trusting one run

**Trap:** You run the benchmark once, see a 4% improvement, and declare victory.

**Why:** CPU frequency scaling, background processes, and thermal state make a
single run noisy. A 4% "win" is well within run-to-run variance and may be pure
noise — or even a regression.

**Fix:** Run `-count=10` for both old and new, and compare with **benchstat**
(`benchstat old.txt new.txt`). It reports the mean, the variance (±), and a
p-value telling you whether the difference is statistically real.

---

### 4. Optimizing without profiling

**Trap:** You eyeball the code, decide "this loop looks slow," and rewrite it.
The service is no faster.

**Why:** Intuition about hotspots is almost always wrong. The time is usually
somewhere unglamorous (allocation, a map probe, JSON, a lock) that you didn't
suspect.

**Fix:** Follow the loop: **benchmark** to confirm a problem → `pprof top` to
find the hot function → `pprof list Func` to find the hot line → fix → re-bench
to prove the win. Never skip straight to the rewrite.

---

### 5. Ignoring allocs/op

**Trap:** You report only ns/op and call the code "fast." Under production load
it has surprise tail-latency spikes.

**Why:** Allocations drive GC pressure, and GC pauses are where a Go service
loses p99 latency under concurrency. A benchmark with no `-benchmem` hides half
the story.

**Fix:** Always run with `-benchmem` (or `b.ReportAllocs()`) and watch **B/op**
and **allocs/op**, not just ns/op. Driving allocs/op toward zero (pre-sizing
slices, `strings.Builder`, keeping values on the stack) is often the real win.

---

### 6. Micro-optimizing the cold path

**Trap:** You spend a day shaving 5 ns off a function — that runs once at
startup.

**Why:** 90% of a service's time is in ~10% of the code. Effort spent off the
hot path buys nothing and usually *costs* readability, making future bugs more
likely.

**Fix:** Profile to find the hot 10% and optimize only that. On the cold path,
prefer the clearest code. A better algorithm (O(n) over O(n²)) dwarfs any
constant-factor micro-trick anyway.

---

### 7. Comparing benchmarks across machines / thermal states

**Trap:** You benchmark "before" on your laptop on battery, "after" plugged in
(or on a different machine / a busy CI runner), and compare the numbers.

**Why:** ns/op depends on CPU model, clock speed, power profile, thermal
throttling, and noisy neighbors. Cross-machine or cross-state numbers are not
comparable — you're measuring the environment, not the code.

**Fix:** Compare old vs new on the **same machine, same power/thermal state,
back-to-back**, ideally with the machine otherwise idle. Use benchstat on those
paired runs. Treat absolute ns/op as machine-specific; trust the *ratio*.
