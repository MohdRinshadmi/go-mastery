# Debugging Challenges

> Mentor note: Writing new code is the easy half of the job. The half that pays your salary is staring at code that *looks* correct, runs in prod, and quietly does the wrong thing — then finding out *why*. This track trains the muscle the 30-day curriculum doesn't: reading a bug report, reproducing it, forming a hypothesis, and proving the fix.

Each challenge is a small, realistic program with a single planted bug. Your job:

1. Read the `README.md` — it gives you the **symptom**, a **hint**, and **how to reproduce**. It does *not* hand you the answer up front.
2. Open `bugged/` and run it. See the wrong output / race / hang for yourself.
3. Form a hypothesis. Fix it in a copy.
4. Only then expand the `<details>` "Solution & why" at the bottom of the challenge README, and compare against `fixed/`.

**Rules of engagement (the same ones that matter on a real team):**

- Reproduce *before* you theorize. A bug you can't reproduce, you can't fix.
- For concurrency bugs, `go run -race .` is not optional — it's your first move. The race detector finds in 200ms what code review misses for weeks.
- The smallest diff that fixes the root cause beats a big rewrite. We review root cause, not band-aids.

## The challenges

| #  | Phase | Topic | Bug class |
|----|-------|-------|-----------|
| 01 | 1 — Fundamentals | [Slice aliasing](challenge-01-slice-aliasing/) | `append` shares backing array; a "safe" copy isn't |
| 02 | 2 — Core | [Nil interface](challenge-02-nil-interface/) | typed-nil stored in an interface is `!= nil` |
| 03 | 3 — Concurrency | [Data race](challenge-03-data-race/) | unsynchronized shared counter under goroutines |
| 04 | 4 — Backend | [Context leak](challenge-04-context-leak/) | request context dropped; cancellation never propagates |
| 05 | 5 — Production | [Ticker leak](challenge-05-ticker-leak/) | `time.Ticker` / goroutine leaked on every request |
| 06 | 6 — Advanced | [Unbounded queue](challenge-06-unbounded-queue/) | unbuffered/ungated job intake grows without limit |

## How to run a challenge

```bash
cd challenge-03-data-race/bugged
go run -race .     # watch it fail

cd ../fixed
go vet ./... && go build ./... && go run -race .   # clean
```

Work them in order — they roughly track the six phases of the main program. Don't peek at `fixed/` until you've made the bug reproduce and written down *why* it happens.
