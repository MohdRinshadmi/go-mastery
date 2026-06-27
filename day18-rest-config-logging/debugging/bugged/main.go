// Command bugged demonstrates a config bug: Load() does NOT fail fast when a
// REQUIRED env var with no safe default (JWT_SECRET) is missing. It returns a
// Config with an empty JWTSecret and a nil error, so the app "starts" — then
// 500s the first time anyone tries to sign a token.
//
// Run it:
//
//	go run .
//
// STDLIB ONLY. To make the failure deterministic (independent of whatever is
// in your real shell), main() simulates a deployment that forgot to set
// JWT_SECRET by unsetting it before calling Load().
package main

import (
	"fmt"
	"os"
)

// Config holds the service configuration, loaded from the environment.
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string // REQUIRED: no safe default — signing tokens needs a real secret.
	LogLevel    string
}

// getenv returns the env var for key, or def when it is unset/empty.
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Load reads configuration from the environment.
//
// BUG: it reads JWT_SECRET with os.Getenv and never validates it. JWT_SECRET
// has no safe default — an empty signing secret is a security hole, not a
// convenience — yet Load() happily returns (cfg, nil) with JWTSecret == "".
// The bad config is discovered at runtime (first request), not at boot.
func Load() (Config, error) {
	cfg := Config{
		Port:        getenv("PORT", "8080"),
		DatabaseURL: getenv("DATABASE_URL", "postgres://localhost:5432/app"),
		JWTSecret:   os.Getenv("JWT_SECRET"), // read but never checked
		LogLevel:    getenv("LOG_LEVEL", "info"),
	}
	// BUG: no `if cfg.JWTSecret == "" { return cfg, error }` guard here.
	return cfg, nil
}

func main() {
	// Simulate a deployment that forgot to set JWT_SECRET. Unsetting it makes
	// the demo deterministic regardless of the real environment.
	os.Unsetenv("JWT_SECRET")

	cfg, err := Load()
	if err != nil {
		fmt.Println("Load() failed fast:", err)
		os.Exit(1)
	}

	// Load() succeeded, so we proudly "start the server" with a broken config.
	fmt.Printf("Load() ok, JWTSecret=%q (BUG: started with no secret!)\n", cfg.JWTSecret)
	fmt.Println("server starting on port", cfg.Port)

	// ...later, the first request tries to sign a token and blows up:
	if cfg.JWTSecret == "" {
		fmt.Println("first request -> 500: cannot sign JWT with empty secret")
	}
}
