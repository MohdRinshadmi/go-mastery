# Day 25 Debugging — `Close()` vs `Shutdown()`: dropped requests on every deploy

The service handles `SIGTERM` on deploy by calling `srv.Close()`. It looks like
a clean shutdown. But every deploy produces a little burst of 502s in the
dashboards, and nobody can reproduce it locally because there's "nothing wrong
with the code." The cause: `Close()` **immediately severs all active
connections** — any request that was mid-flight (e.g. 100ms from returning) gets
a broken connection. The correct call is `srv.Shutdown(ctx)`, which **drains**
in-flight requests first.

**Stdlib only.** We start a real server on a random localhost port, fire one
slow in-flight request, trigger shutdown mid-request, and observe the outcome.
Exits promptly — no waiting on a real SIGTERM, no hanging server.

## Symptom

```
$ cd bugged && go run -race .
in-flight request FAILED: Get "http://127.0.0.1:.../work": EOF
=> BUG: srv.Close() dropped an in-flight request (502/EOF on every deploy)
```

## Reproduce

```bash
cd bugged
go run -race .
```

## Hint

<details>
<summary>Hint</summary>

`http.Server` has two ways to stop: one closes connections immediately, the
other stops accepting new connections and waits for active ones to finish within
a deadline. Which one gives you a zero-downtime deploy? And once you switch,
make sure `main` actually waits for the drain to return.

</details>

## Solution & why

<details>
<summary>Solution & why</summary>

`srv.Close()` closes the listener **and all active connections immediately**.
It does not wait for handlers to finish, so an in-flight request is severed —
the client gets an EOF / 502. Multiply that across every deploy and you have a
recurring, unexplained error spike.

**Fix:** use `srv.Shutdown(ctx)`:

```go
shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
if err := srv.Shutdown(shutdownCtx); err != nil {
    log.Println("forced shutdown:", err)
}
```

`Shutdown` (1) stops accepting new connections, (2) waits for in-flight requests
to finish up to the ctx deadline, then (3) returns. The in-flight `/work`
request now completes with 200 instead of being dropped.

Two things that are part of the same bug class:

- **Wait for the drain.** Calling `Shutdown` and then immediately `os.Exit` (or
  letting `main` return) kills the drain you just started. Block until
  `Shutdown` returns.
- **Deadline < grace period.** The shutdown deadline must be shorter than the
  orchestrator's `terminationGracePeriodSeconds` (default 30s on K8s) or you get
  SIGKILLed mid-drain — the worst of both worlds. And flip readiness to "not
  ready" *before* draining so the load balancer stops sending new traffic a beat
  early.

</details>
