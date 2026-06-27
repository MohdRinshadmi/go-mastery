# Day 21 Debugging — "Works on my laptop, dead in the container"

A Go HTTP service reads its listen address from config and starts up. On a
developer's machine the container logs `listening on ...` and `curl localhost`
works. Shipped to Kubernetes, every readiness/liveness probe fails with
*connection refused*, external traffic never arrives, and the pod gets killed
as unhealthy — even though the process is running fine.

This is a **stdlib-only** simulation: instead of opening a real socket (which
would hang and need a network), we resolve the *effective* bind address exactly
as the server would, then assert whether outside traffic could reach it.

## Symptom

```
$ cd bugged && go run .
listening on 127.0.0.1:8080
UNREACHABLE: server is bound to "127.0.0.1" — probes/external traffic get connection refused
exit status 1
```

The process binds a port and logs success, but nothing outside the container
can connect.

## Reproduce

```bash
cd bugged
go run .            # exits 1: bound to loopback, unreachable
HOST=0.0.0.0 go run .   # only works if the operator happens to override HOST
```

## Hint

<details>
<summary>Hint</summary>

Inside a container, `127.0.0.1` is the container's *own* loopback interface —
a separate network namespace. The kube-proxy probe, the service mesh sidecar,
and other pods all live *outside* that namespace. What address must a server
bind to so that connections from any interface are accepted?

</details>

## Solution & why

<details>
<summary>Solution & why</summary>

The default listen host was `127.0.0.1` (loopback). A server bound to loopback
only accepts connections from *within the same network namespace*. On a laptop
your `curl localhost` shares that namespace, so it works. Inside a container the
probe and external traffic come from outside the namespace, so they get
*connection refused* even though the port is "open".

**Fix:** default the bind host to `0.0.0.0` (all IPv4 interfaces; `::` for all
IPv6). Now connections from any interface are accepted. Operators can still set
`HOST=127.0.0.1` deliberately for loopback-only services (e.g. a metrics port
scraped only by a local sidecar).

```go
host := os.Getenv("HOST")
if host == "" {
    host = "0.0.0.0" // not 127.0.0.1
}
addr := net.JoinHostPort(host, port) // also: build the address robustly
```

Two extra production-grade touches in `fixed/`:
- Use `net.JoinHostPort` instead of `host + ":" + port` — it handles IPv6
  literals (`[::1]:8080`) correctly, which naive concatenation breaks.
- This is the container-specific cousin of the lesson's "env vs flag" config
  story: the value that's right on a laptop is wrong in a container, and only
  config-by-environment surfaces it.

</details>
