# Day 24 Debugging — Liveness that checks the DB (the restart storm)

The service exposes `/healthz` (liveness) and `/readyz` (readiness), and the
author wired **both** to ping the database. It looks thorough. Then the DB has a
five-second blip and the whole fleet goes into `CrashLoopBackOff`: liveness
failed → Kubernetes restarted every pod at once → they all reconnect
simultaneously → the database is buried under a reconnect stampede → it stays
down. A transient hiccup became a full outage.

**Stdlib only** — `httptest` plus a togglable fake dependency, no real DB or
server. Exits promptly.

## Symptom

```
$ cd bugged && go run .
== DB healthy ==
  /healthz -> 200   /readyz -> 200
== DB has a transient blip ==
  /healthz -> 503   /readyz -> 503
=> BUG: liveness FAILED on a DB blip -> orchestrator restarts every pod (restart storm)
```

Liveness returning 503 is the trigger: the orchestrator interprets it as "the
process is broken, restart it."

## Reproduce

```bash
cd bugged
go run .
```

## Hint

<details>
<summary>Hint</summary>

What action does the orchestrator take when **liveness** fails versus when
**readiness** fails? Which of those two is safe to fail during a transient
dependency outage? A DB blip should pull you out of rotation — not restart your
process.

</details>

## Solution & why

<details>
<summary>Solution & why</summary>

Liveness and readiness control different orchestrator actions:

| Probe | Question | On failure |
|---|---|---|
| **Liveness** `/healthz` | "Is the process alive / not deadlocked?" | **Restart** the pod |
| **Readiness** `/readyz` | "Can I serve traffic right now?" | **Stop sending traffic** (no restart) |

The bug put a DB check in **liveness**. So a DB blip made liveness fail, which
restarts the pod. Every replica restarts at once and all reconnect together —
the thundering-herd reconnect finishes off the database. Classic, famous outage
category.

**Fix:** keep liveness dumb (return 200 if the process can run the handler) and
put dependency checks only in readiness:

```go
// liveness: no dependency checks — a DB blip must never restart the process
mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
})
// readiness: check the DB; failing pulls the pod from rotation, no restart
mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
    if err := db.Ping(); err != nil {
        http.Error(w, "db not ready", http.StatusServiceUnavailable); return
    }
    w.WriteHeader(http.StatusOK)
})
```

Now a DB blip yields `/healthz -> 200` (no restart) and `/readyz -> 503` (pulled
from rotation). When the DB recovers, readiness flips back to green and traffic
resumes — no restart storm. Two more production touches: bound the readiness
check with a `context` timeout so a hung dependency doesn't hang the probe, and
keep probes off any auth middleware that could itself fail.

</details>
