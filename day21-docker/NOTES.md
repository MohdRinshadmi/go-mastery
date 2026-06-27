# Day 21 Notes — Dockerizing Go (quick reference)

## Multi-stage build skeleton
```dockerfile
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /server /server
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]
```

## Build flags cheat sheet
| Flag | Effect |
|---|---|
| `CGO_ENABLED=0` | fully static binary (needed for scratch/distroless-static) |
| `GOOS=linux GOARCH=amd64` | cross-compile target |
| `-ldflags="-s -w"` | strip symbols + DWARF, ~30–40% smaller |
| `-ldflags="-X main.Version=$V"` | inject version/commit at build time |

## Base image choice
| Base | Size overhead | CA certs | Non-root | Shell | Use |
|---|---|---|---|---|---|
| `scratch` | 0 | manual copy | manual | no | advanced/hardened |
| `distroless/static` | ~3 MB | included | `nonroot` built-in | no | **default** |
| `alpine` | ~8 MB | `apk add` | manual | `sh` | dev/debug |

## Container config rules
- Bind `0.0.0.0`, not `127.0.0.1` (probes/LB are outside the netns).
- Read config from **env vars** (12-factor) so the same image runs in any env.
- Build address with `net.JoinHostPort(host, port)` (IPv6-safe).

## Compose essentials
- `depends_on: { db: { condition: service_healthy } }` + a `healthcheck`.
- Named `volumes:` for DB data so `down` doesn't wipe it.
- Pin image tags (`postgres:16-alpine`), never `latest`.
- Secrets via `.env`/Docker secrets, never committed.

## .dockerignore starter
```
.git
*_test.go
vendor/
*.md
*.env
bin/
.idea/
.vscode/
```

## Key terms
- **Multi-stage build** — build in a fat image, ship from a minimal one.
- **scratch** — the empty image (no OS).
- **distroless** — minimal Google base: certs + tzdata + nonroot, no shell.
- **Build context** — the files tarred and sent to the daemon for a build.
- **Layer cache** — Docker reuses unchanged instruction layers; order matters.
- **ldflags / `-X`** — linker injection of variables at build time.
- **Loopback (127.0.0.1)** — interface reachable only within the same netns.
