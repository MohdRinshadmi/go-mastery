# Day 21 Resources — Dockerizing Go

- **Docker official — Go language guide / best practices**
  https://docs.docker.com/language/golang/build-images/
  Multi-stage builds, layer caching, and image-size guidance for Go.

- **Google distroless images**
  https://github.com/GoogleContainerTools/distroless
  The `static`, `base`, and `nonroot` variants and when to use each.

- **`docker build` / multi-stage builds reference**
  https://docs.docker.com/build/building/multi-stage/
  The canonical explanation of `FROM ... AS` and `COPY --from`.

- **Docker Compose file reference**
  https://docs.docker.com/reference/compose-file/
  `depends_on`, `healthcheck`, `volumes`, `condition: service_healthy`.

- **`.dockerignore` reference**
  https://docs.docker.com/build/concepts/context/#dockerignore-files
  Build context and ignore syntax.

- **Go cmd/link flags (`-s`, `-w`, `-X`)**
  https://pkg.go.dev/cmd/link
  What the linker flags do, including symbol stripping and variable injection.

- **Go `net.JoinHostPort`**
  https://pkg.go.dev/net#JoinHostPort
  IPv6-safe host:port construction (used in the fixed debugging exercise).

- **12-Factor App — Config**
  https://12factor.net/config
  Why container config comes from the environment, not baked-in defaults.
