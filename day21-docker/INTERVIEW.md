# Day 21 Interview Questions — Dockerizing Go

Lesson questions plus a few extras. Answers in `<details>`.

---

### 1. What does `CGO_ENABLED=0` do and why is it required for `FROM scratch`?

<details>
<summary>Answer</summary>

It disables cgo, so the Go toolchain produces a **fully static** binary with no
dynamic links to glibc. `scratch` (and `distroless/static`) have no C runtime or
dynamic linker, so a CGO-enabled binary would fail at startup. With
`CGO_ENABLED=0` the binary carries everything it needs and runs on an empty
image.
</details>

---

### 2. Difference between `distroless/static` and `scratch`? When choose each?

<details>
<summary>Answer</summary>

`scratch` is literally empty — no files at all. `distroless/static` adds the
minimum useful extras: CA certs, timezone data, `/etc/passwd` with a `nonroot`
user. Neither has a shell or package manager. Use **distroless/static as the
default** (avoids the CA-cert and non-root footguns); reach for **scratch** when
you've measured and need the last few MB, or in hardened environments where even
distroless is too much.
</details>

---

### 3. Why does the order of `COPY` and `RUN` matter for layer caching?

<details>
<summary>Answer</summary>

Docker caches each instruction as a layer keyed by its inputs. If you `COPY . .`
before `go mod download`, any source change invalidates the cache for the
download step, re-downloading all deps. Copy `go.mod`/`go.sum` and run
`go mod download` first, then copy source — deps are re-fetched only when the
manifest changes.
</details>

---

### 4. What does `-ldflags="-s -w"` do?

<details>
<summary>Answer</summary>

`-s` strips the symbol table; `-w` strips DWARF debug info. Together they shrink
the binary roughly 30–40%. The trade-off: stack traces lose symbol detail and
you can't run a debugger against it — fine for production images, where you keep
an unstripped build artifact for debugging if needed.
</details>

---

### 5. What is `depends_on: condition: service_healthy` in Compose and why does it matter?

<details>
<summary>Answer</summary>

Plain `depends_on` only waits for the dependency container to *start*, not to be
*ready to serve*. `condition: service_healthy` waits until the dependency's
`healthcheck` passes. Without it, your app starts before Postgres accepts
connections and crashes on first query.
</details>

---

### 6. What belongs in `.dockerignore` and what's the cost of omitting it?

<details>
<summary>Answer</summary>

`.git`, test fixtures, `vendor/`, `*.md`, local `.env`/secrets, build output,
IDE folders. The build context is tarred and sent to the Docker daemon before
any build step runs; a large `.git/` or fixtures can add hundreds of MB to every
build and can leak secrets into the image.
</details>

---

### 7. How do you inject git commit / version into a Go binary at build time?

<details>
<summary>Answer</summary>

Declare package-level vars (`var Version, GitCommit, BuildTime = "dev", ...`),
pass `ARG`s into the Dockerfile, and inject them with linker flags:
`-ldflags="-X main.Version=$VERSION -X main.GitCommit=$GIT_COMMIT"`. Then
`GET /version` returns real deploy metadata so you always know what's running.
</details>

---

### 8. (Extra) Why is `127.0.0.1` the wrong default bind address in a container?

<details>
<summary>Answer</summary>

`127.0.0.1` is the container's own loopback in its network namespace. Probes,
the load balancer, and other pods are *outside* that namespace and get
connection refused — even though the process logs "listening". Bind to `0.0.0.0`
so any interface accepts connections. (This is the Day 21 debugging exercise.)
</details>

---

### 9. (Extra) Your scratch-based service can't make HTTPS calls. Why?

<details>
<summary>Answer</summary>

No CA certificate bundle in the image, so Go's TLS verification has no roots and
returns `x509: certificate signed by unknown authority`. Copy
`/etc/ssl/certs/ca-certificates.crt` from the builder, or use
`distroless/static`, which includes CA certs.
</details>

---

### 10. (Extra) Why does image size matter beyond aesthetics?

<details>
<summary>Answer</summary>

In an autoscaling cluster, every new node must *pull* the image before a pod can
start. A 1.1 GB image can take 60+ seconds to pull; a 15 MB image takes ~2s.
Under load with pod churn, that gap is lost requests and slower recovery. Smaller
images also mean less attack surface and faster CI.
</details>

---

### 11. (Extra) Multi-stage build — what problem does it solve?

<details>
<summary>Answer</summary>

It lets you build in one image (with the full Go toolchain) and ship in a
different, minimal image that contains only the binary. You get the build
environment's tools without shipping them — small, secure final image from a
single Dockerfile, no fragile copy-between-images scripts.
</details>
