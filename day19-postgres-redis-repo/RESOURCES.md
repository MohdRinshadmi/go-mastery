# Day 19 — Postgres, Repository & Redis Resources

Curated, real links. Read the docs for the library you actually use, not blog
summaries of it.

- **`database/sql` package** — the stdlib generic SQL interface (drivers, rows,
  `QueryRowContext`, `ErrNoRows`): https://pkg.go.dev/database/sql
- **pgx** — the de-facto Postgres driver/toolkit for Go (`pgxpool`,
  `pgx.ErrNoRows`, `QueryRow`/`Scan`): https://github.com/jackc/pgx
- **go-redis** — the standard Redis client for Go (`Set` with TTL, `Get`,
  `redis.Nil` on miss): https://github.com/redis/go-redis
- **`errors` package** — `errors.Is` / `errors.As` / `%w` wrapping, the basis of
  sentinel-error handling: https://pkg.go.dev/errors
- **golang-migrate** — versioned, up/down SQL migrations run in CI/CD:
  https://github.com/golang-migrate/migrate
- **Alex Edwards — Let's Go Further** — the canonical Go practitioner's guide to
  structuring DB access, the repository/model layer, and clean error mapping:
  https://lets-go-further.alexedwards.net/
- **Cache-aside pattern (AWS ElastiCache "Caching strategies")** — lazy loading
  vs write-through, TTL, and invalidation tradeoffs:
  https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/Strategies.html
- **Redis client-side caching / patterns docs** — official guidance on caching
  with Redis: https://redis.io/docs/latest/develop/use/patterns/
