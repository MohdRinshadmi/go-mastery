# Day 21 — Dockerizing Go: Multi-Stage Builds, Distroless, Docker Compose

> Mentor note: Shipping code that works on your laptop but breaks in production is amateur hour. Today you'll learn to package Go binaries into tiny, secure, reproducible containers. This is the first step from "developer" to "platform engineer." Every concept here has a direct line to how Cloudflare, Uber, and Google ship Go at scale.

---

## 0. Why containers for Go?

Go produces a **single static binary**. No JVM, no node_modules, no interpreter. This is its superpower for containers — the resulting image can be as small as 10 MB vs a Java Spring Boot image that weighs 300 MB+.

But naive Docker usage throws away that advantage. You'll see devs ship a `golang:1.22` image (1.1 GB!) with a 10 MB binary in it. That's shipping a full compiler + toolchain just to run one executable. Today we fix that.

---

## 1. Multi-Stage Builds

### Theory
Docker multi-stage builds let you use one image to **build** and a completely different (smaller) image to **run**. The final image contains only what the runtime needs.

### Why it exists
Before multi-stage builds (Docker 17.05+), teams either:
1. Shipped the builder image (huge, full of attack surface)
2. Maintained two separate Dockerfiles and a `build.sh` that copied artifacts between them (fragile)

Multi-stage builds collapse this into one `Dockerfile`.

### When to use
- Always. Any production Docker image for Go should be multi-stage.

### When NOT to use
- Development-only images where you want live reload tools (then a single-stage `golang:alpine` dev image is fine).
- When your binary has CGO dependencies that need a glibc runtime (you'll need a Debian/Alpine base, not scratch).

### The pattern

```dockerfile
# ---- Stage 1: Builder ----
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download   # cached layer — only re-runs when deps change

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /server ./cmd/server

# ---- Stage 2: Runtime ----
FROM scratch
COPY --from=builder /server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
```

**What the flags mean:**
- `CGO_ENABLED=0` — disable C-extensions. Produces a fully static binary.
- `GOOS=linux GOARCH=amd64` — cross-compile explicitly (even if you're on Mac/ARM).
- `-ldflags="-s -w"` — strip debug symbols (`-s`) and DWARF tables (`-w`). Cuts binary size 30-40%.
- `scratch` — the empty Docker image. No OS at all. Just your binary.

### Beginner example (run it):
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /hello .

FROM scratch
COPY --from=builder /hello /hello
ENTRYPOINT ["/hello"]
```

### Production-grade example (what you actually ship):
See `examples/Dockerfile` — adds non-root user, CA certificates, build args for versioning.

### Common mistakes
1. **Forgetting `CGO_ENABLED=0`** — your binary links to glibc, which doesn't exist in scratch. Container crashes at runtime with `exec format error` or a dynamic linker error.
2. **Not downloading dependencies before copying source** — you lose the Docker layer cache. Every code change re-downloads all deps.
3. **Running as root** — scratch images run as root by default. Add a `USER` directive.
4. **Not copying CA certs** — your Go binary won't be able to make HTTPS calls from scratch because there are no root certificates. Copy `/etc/ssl/certs/ca-certificates.crt` from the builder stage.

### Performance implications
| Image type | Typical size | Cold start |
|---|---|---|
| `golang:1.22` (naive) | 1.1 GB | Slow pull |
| `golang:1.22-alpine` builder | 300 MB | Moderate |
| `distroless/static` | ~3 MB + binary | Fast pull |
| `scratch` | binary size only (5-20 MB) | Fastest pull |

**Senior take:** Size isn't vanity. In a Kubernetes cluster autoscaling under load, pulling a 1.1 GB image on a new node takes 60+ seconds. Pulling a 15 MB image takes 2 seconds. At 100 rps with a pod crash, that difference costs you real money in lost requests.

---

## 2. Distroless vs. Scratch

### Theory
`scratch` is literally nothing. `gcr.io/distroless/static` (Google's distroless) adds just enough to be useful: `/etc/passwd` (so you can have a non-root user), timezone data, CA certs. No shell, no package manager, no bash.

### When to use which

| | `scratch` | `distroless/static` | `alpine` |
|---|---|---|---|
| Size | Smallest | ~3MB overhead | ~8MB overhead |
| CA certs | Manual copy needed | Included | `apk add ca-certificates` |
| Non-root user | Manual | Built-in `nonroot` | Manual |
| Debug shell | None | None | `sh` available |
| Production recommendation | Advanced users | **Default recommendation** | Development/debugging |

### Beginner example
```dockerfile
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /server /server
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]
```

### Production-grade example
See `examples/Dockerfile` — uses distroless with nonroot user, version labels.

**Senior take:** Use `distroless/static` as your default. It's almost as small as scratch but avoids the footguns (CA certs, user creation). Reach for `scratch` when you've measured and need the last few MB, or in security-hardened environments where even distroless is too permissive.

---

## 3. .dockerignore

### Theory
`.dockerignore` prevents files from being sent to the Docker build context. Without it, your entire project (including `.git`, test fixtures, local secrets) gets sent to the daemon.

### Why it matters
The Docker build context is tarred and sent over a socket before a single build step runs. A missing `.dockerignore` in a repo with a large `.git/` folder or test fixtures can add 500 MB to every `docker build`.

### Production .dockerignore
```
.git
.gitignore
.env
*.env
docker-compose*.yml
Makefile
*.md
**_test.go
vendor/
.idea/
.vscode/
coverage.out
bin/
```

**Senior take:** Always create `.dockerignore` before you even write the `Dockerfile`. It's a 2-minute habit that saves your team minutes per build.

---

## 4. Docker Compose: App + Postgres + Redis

### Theory
Docker Compose is a tool for defining multi-container applications in a single YAML file. It handles networking (containers talk by service name), volumes (persistent data), dependency ordering, and environment variables.

### Why it exists
In production, your Go API talks to a database, a cache, maybe a queue. Running all three manually with `docker run` flags is tedious and error-prone. Compose defines the whole system declaratively.

### When to use
- Local development: spin up the full stack with `docker compose up`.
- Integration tests in CI: `docker compose -f docker-compose.test.yml up --abort-on-container-exit`.

### When NOT to use
- Production at scale: use Kubernetes (or ECS/Fargate). Compose doesn't do service discovery, horizontal scaling, or health-based routing.

### Production-grade compose file
```yaml
services:
  api:
    build: .
    ports: ["8080:8080"]
    environment:
      DATABASE_URL: postgres://user:pass@postgres:5432/shop?sslmode=disable
      REDIS_URL: redis://redis:6379
    depends_on:
      postgres: { condition: service_healthy }
      redis: { condition: service_started }

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
      POSTGRES_DB: shop
    volumes: [pgdata:/var/lib/postgresql/data]
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U user"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    volumes: [redisdata:/data]

volumes:
  pgdata:
  redisdata:
```

### Common mistakes
1. **No healthcheck on postgres** — your app container starts before Postgres is ready to accept connections. Always use `depends_on: condition: service_healthy` with a healthcheck.
2. **Hardcoding secrets in compose file** — use `.env` files or Docker secrets. Never commit passwords.
3. **No volumes for DB data** — every `docker compose down` wipes your local DB. Named volumes persist.
4. **Using `latest` image tags** — pin versions (`postgres:16-alpine`). `latest` will silently upgrade and break your schema.

**Senior take:** `docker compose up -d && docker compose logs -f api` is your development loop. If this doesn't work reliably in under 60 seconds on a fresh clone, your onboarding story is broken.

---

## 5. Build arguments and versioning

### Theory
`ARG` in Dockerfile + `-ldflags` lets you inject the git commit, version, and build time directly into your binary.

```dockerfile
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_TIME=unknown

RUN go build \
  -ldflags="-s -w \
    -X main.Version=${VERSION} \
    -X main.GitCommit=${GIT_COMMIT} \
    -X main.BuildTime=${BUILD_TIME}" \
  -o /server ./cmd/server
```

```go
var (
    Version   = "dev"
    GitCommit = "unknown"
    BuildTime = "unknown"
)
```

Now `GET /version` returns real deploy metadata. You'll never wonder "which version is running in prod?" again.

**Senior take:** Every production binary should know its own version. Debugging a production incident and realizing you don't know what code is deployed is a terrible feeling. Build it in from day one.

---

## Expert Thinking Mode

- **Beginner:** "Docker runs my app on a server."
- **Senior:** "Docker gives me reproducible, portable environments. My multi-stage build produces a minimal image to reduce attack surface and deployment time."
- **Staff:** "Image size directly affects autoscaling latency. I own the build pipeline. I ensure images are signed, scanned for CVEs, and pushed to a private registry with immutable tags (sha256 digest, not `latest`)."
- **Architect:** "The container strategy is part of the platform contract. Every service team gets a base image with standard observability, secrets injection, and security hardening baked in. They inherit it — they don't configure it per service."

---

## Real-world use

- **Cloudflare:** Go services shipped as distroless images, run in a Kubernetes cluster on their own edge network. Image size is a first-class concern because pods start on new PoPs constantly.
- **Uber:** Go monorepo with Bazel + Docker. Hundreds of services share base images. `--ldflags` version injection is standard.
- **Google:** The inventors of distroless. Every Go service at Google runs in a minimal container without a shell. The security stance is: if there's no shell, an attacker who gets code execution can't easily pivot.
- **Stripe:** Multi-stage builds enforced by a shared Makefile. `make docker-build` is the only way to build the production image — no improvising.

---

## Interview Questions

1. What does `CGO_ENABLED=0` do and why is it required for `FROM scratch`?
2. What is the difference between `distroless/static` and `scratch`? When would you choose each?
3. Why does the order of `COPY` and `RUN` instructions matter for Docker layer caching?
4. What does `-ldflags="-s -w"` do to your binary?
5. What is the `depends_on: condition: service_healthy` pattern in Compose and why does it matter?
6. What should always be in `.dockerignore`? What's the performance consequence of omitting it?
7. How do you inject git commit/version info into a Go binary at build time?

---

## Your tasks for today

Go to `../exercises/`. You have:
1. A minimal HTTP service to containerize (write the Dockerfile and compose file)
2. A Dockerfile debugging exercise (find the 4 mistakes)
3. A compose challenge (add Redis to an existing compose file)

Build, run, verify the containers start and the HTTP endpoints respond. I'll review your Dockerfile like a production PR — I'll check for root users, missing CA certs, layer ordering, and image size.

---

## Day 21 companion files

Self-contained study material for this day (in the day folder root):

- [Debugging exercise](../debugging/README.md) — "works on my laptop, dead in the container": a service that binds `127.0.0.1` and is unreachable from outside the container ([bugged](../debugging/bugged/main.go) vs [fixed](../debugging/fixed/main.go)).
- [PITFALLS.md](../PITFALLS.md) — 7 Docker/Go traps as Trap → Why → Fix.
- [INTERVIEW.md](../INTERVIEW.md) — interview Q&A with model answers.
- [NOTES.md](../NOTES.md) — quick reference + key terms.
- [RESOURCES.md](../RESOURCES.md) — curated links for Day 21.
