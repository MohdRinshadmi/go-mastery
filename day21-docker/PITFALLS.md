# Day 21 Pitfalls — Dockerizing Go

Format: **Trap → Why → Fix**

---

### 1. Binding to `127.0.0.1` inside a container
**Trap:** Your config defaults the listen host to `127.0.0.1` (or `localhost`).
**Why:** Inside a container that's the container's own loopback namespace. The kube probe, the load balancer, and other pods live outside it and get *connection refused* — even though the port is "open" and the logs say `listening`.
**Fix:** Bind to `0.0.0.0` (all interfaces) in containers. Use `net.JoinHostPort` to build the address so IPv6 works. Only bind loopback deliberately (e.g. a metrics port scraped by a local sidecar).

---

### 2. Forgetting `CGO_ENABLED=0` with `scratch`/`distroless/static`
**Trap:** `go build` links against glibc; the binary crashes at startup in a scratch image.
**Why:** `scratch` and `distroless/static` have no C runtime / dynamic linker. A CGO-enabled binary needs glibc that isn't there → `exec format error` or "no such file or directory" on a binary that's clearly present.
**Fix:** `CGO_ENABLED=0 go build ...` to produce a fully static binary. If you genuinely need CGO (e.g. some DB drivers), use a glibc base like `distroless/base` or Debian-slim.

---

### 3. `COPY . .` before `go mod download`
**Trap:** You copy all source, then download deps.
**Why:** Any source change busts the Docker layer cache for the `go mod download` step, so every code edit re-downloads every dependency. Builds crawl.
**Fix:** Copy `go.mod`/`go.sum` first, `RUN go mod download`, *then* `COPY . .`. Deps only re-download when the manifest changes.

---

### 4. Running as root in the final image
**Trap:** `scratch` and most bases run as UID 0 by default.
**Why:** A container escape or RCE starts with root privileges — larger blast radius.
**Fix:** Add a non-root `USER`. `distroless/static:nonroot` gives you a ready-made `nonroot` user; with scratch, copy a crafted `/etc/passwd` or set `USER 65532:65532`.

---

### 5. No CA certificates → HTTPS calls fail from scratch
**Trap:** Your service makes outbound HTTPS calls and gets `x509: certificate signed by unknown authority`.
**Why:** `scratch` has no root CA bundle. Go's TLS stack has nothing to verify against.
**Fix:** `COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/` — or use `distroless/static`, which bundles CA certs.

---

### 6. No healthcheck / `depends_on: service_healthy` in Compose
**Trap:** Your API container starts before Postgres can accept connections and crashes on first query.
**Why:** `depends_on` without a condition only waits for the container to *start*, not to be *ready*. Postgres takes a few seconds to accept connections.
**Fix:** Add a `healthcheck` (e.g. `pg_isready`) on the DB and `depends_on: { postgres: { condition: service_healthy } }` on the app.

---

### 7. Using `latest` / no `.dockerignore`
**Trap:** `postgres:latest` silently upgrades and breaks your schema; missing `.dockerignore` ships `.git` and fixtures into the build context.
**Why:** `latest` is mutable — non-reproducible builds. The build context is tarred and sent to the daemon before any step runs; a big `.git` adds hundreds of MB per build.
**Fix:** Pin image tags (`postgres:16-alpine`). Write `.dockerignore` (`.git`, `*_test.go`, `vendor/`, `*.md`, secrets) *before* the Dockerfile.
