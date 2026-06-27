//go:build ignore

// REFERENCE ONLY — the real Postgres + Redis implementations.
// Build-tagged `ignore` so `go build ./...` here doesn't require the drivers.
// To use for real: remove the build tag, then
//   go get github.com/jackc/pgx/v5 github.com/redis/go-redis/v9
//   docker compose up -d   (starts Postgres + Redis)
//   DATABASE_URL=postgres://app:secret@localhost:5432/shop REDIS_ADDR=localhost:6379 go run .
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// ---- PostgresUserRepo ----------------------------------------------------
type PostgresUserRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepo(ctx context.Context, dsn string) (*PostgresUserRepo, error) {
	pool, err := pgxpool.New(ctx, dsn) // connection POOL, not a single conn
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	return &PostgresUserRepo{pool: pool}, nil
}

func (r *PostgresUserRepo) Create(ctx context.Context, u User) error {
	// parameterized query ($1..$3) — never string-concatenate user input
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, email, name) VALUES ($1, $2, $3)`,
		u.ID, u.Email, u.Name)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *PostgresUserRepo) GetByID(ctx context.Context, id string) (User, error) {
	var u User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, name FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.Email, &u.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrUserNotFound // map driver error -> DOMAIN error
		}
		return User{}, fmt.Errorf("get user %s: %w", id, err)
	}
	return u, nil
}

// ---- RedisCachingUserRepo: cache-aside in front of any UserRepository ----
type RedisCachingUserRepo struct {
	inner UserRepository
	rdb   *redis.Client
	ttl   time.Duration
}

func NewRedisCachingUserRepo(inner UserRepository, addr string, ttl time.Duration) *RedisCachingUserRepo {
	return &RedisCachingUserRepo{
		inner: inner,
		rdb:   redis.NewClient(&redis.Options{Addr: addr}),
		ttl:   ttl,
	}
}

func (r *RedisCachingUserRepo) key(id string) string { return "user:" + id }

func (r *RedisCachingUserRepo) GetByID(ctx context.Context, id string) (User, error) {
	// 1. try cache
	if b, err := r.rdb.Get(ctx, r.key(id)).Bytes(); err == nil {
		var u User
		if json.Unmarshal(b, &u) == nil {
			return u, nil // cache hit
		}
	} else if !errors.Is(err, redis.Nil) {
		// real Redis error (not a miss) — log + fall through to DB
	}
	// 2. miss -> DB
	u, err := r.inner.GetByID(ctx, id)
	if err != nil {
		return User{}, err
	}
	// 3. populate cache with TTL
	if b, err := json.Marshal(u); err == nil {
		r.rdb.Set(ctx, r.key(id), b, r.ttl)
	}
	return u, nil
}

func (r *RedisCachingUserRepo) Create(ctx context.Context, u User) error {
	if err := r.inner.Create(ctx, u); err != nil {
		return err
	}
	r.rdb.Del(ctx, r.key(u.ID)) // invalidate on write
	return nil
}

// Schema (put in a migration file 0001_init.up.sql):
//   CREATE TABLE users (
//       id    TEXT PRIMARY KEY,
//       email TEXT UNIQUE NOT NULL,
//       name  TEXT NOT NULL
//   );
