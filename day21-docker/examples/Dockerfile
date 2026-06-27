# =============================================================================
# Day 21 — Production-grade multi-stage Dockerfile
# =============================================================================
#
# Build:
#   docker build \
#     --build-arg VERSION=v1.0.0 \
#     --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
#     --build-arg BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
#     -t shop-api:v1.0.0 .
#
# Run:
#   docker run -p 8080:8080 shop-api:v1.0.0
#
# Inspect size:
#   docker image ls shop-api

# =============================================================================
# Stage 1 — Builder
# =============================================================================
# We use the official Go image on Alpine (small Linux) only for building.
# Alpine adds ~8 MB but gives us a shell, apk, and ca-certificates. That is
# all we need here — the compiler handles everything else.
FROM golang:1.22-alpine AS builder

# Install CA certificates so go get / go mod download works over HTTPS.
RUN apk --no-cache add ca-certificates

# Create a non-root user for the runtime stage.
# We create it here because scratch/distroless have no useradd command.
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /build

# ---- Dependency layer (cached separately from source) ----
# Copy only the module files first. Docker caches this layer. It is only
# invalidated when go.mod or go.sum changes — NOT on every code edit.
# This is the single most important Docker performance pattern for Go.
COPY go.mod go.sum ./
RUN go mod download

# ---- Source layer ----
COPY . .

# Build arguments — injected by CI at build time.
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_TIME=unknown

# Build the binary.
# CGO_ENABLED=0  — fully static binary, no glibc dependency.
# GOOS/GOARCH    — explicit cross-compile target (important on Apple Silicon).
# -ldflags:
#   -s   strip symbol table  }  together cut binary size ~30%
#   -w   strip DWARF info    }
#   -X   inject version variables at link time
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-s -w \
    -X main.Version=${VERSION} \
    -X main.GitCommit=${GIT_COMMIT} \
    -X main.BuildTime=${BUILD_TIME}" \
  -o /server ./cmd/server

# =============================================================================
# Stage 2 — Runtime (distroless)
# =============================================================================
# gcr.io/distroless/static:nonroot contains:
#   - CA certificates (so HTTPS calls work)
#   - /etc/passwd with a nonroot user (uid 65532)
#   - Timezone data
#   - Nothing else — no shell, no package manager, no bash
#
# Why distroless over scratch?
#   scratch is smaller but you must manually copy CA certs and create users.
#   distroless/static gives you those for free. It is the right default.
FROM gcr.io/distroless/static:nonroot

# Copy the compiled binary from the builder stage.
# No source, no Go toolchain, no apk — only the binary.
COPY --from=builder /server /server

# Run as non-root. Best practice: never run production containers as root.
USER nonroot:nonroot

# Document the port (does not actually publish it).
EXPOSE 8080

# Use exec form (array) not shell form. Exec form receives signals directly,
# so SIGTERM/SIGINT for graceful shutdown works (Day 25 topic).
ENTRYPOINT ["/server"]
