// Command fixed shows the corrected config loader: Load() validates every
// REQUIRED variable that has no safe default and fails fast at startup. A
// missing JWT_SECRET now returns an error, so the process refuses to start
// instead of crashing on the first request.
//
// Run it:
//
//	go run .
//
// STDLIB ONLY. main() unsets JWT_SECRET to simulate a deployment that forgot
// it, then prints how Load() reacts.
package main

import (
	"errors"
	"fmt"
	"os"
)

// Config holds the service configuration, loaded from the environment.
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string // REQUIRED: no safe default — must be present at boot.
	LogLevel    string
}

// getenv returns the env var for key, or def when it is unset/empty.
// Use this only for values that HAVE a safe default. Required secrets must
// never go through getenv-with-default.
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Load reads configuration from the environment and validates it. Required
// variables with no safe default are checked here so the service crashes
// loudly at startup rather than silently 500ing later.
func Load() (Config, error) {
	cfg := Config{
		Port:        getenv("PORT", "8080"),
		DatabaseURL: getenv("DATABASE_URL", "postgres://localhost:5432/app"),
		JWTSecret:   os.Getenv("JWT_SECRET"), // required: no default
		LogLevel:    getenv("LOG_LEVEL", "info"),
	}
	// FIX: fail fast. A required secret with no safe default is validated at
	// boot. If it is missing, return an error so main can exit non-zero.
	if cfg.JWTSecret == "" {
		return cfg, errors.New("JWT_SECRET is required")
	}
	return cfg, nil
}

func main() {
	// Simulate a deployment that forgot to set JWT_SECRET.
	os.Unsetenv("JWT_SECRET")

	cfg, err := Load()
	if err != nil {
		fmt.Println("Load() failed fast:", err)
		os.Exit(1)
	}

	// We only reach here with a valid config.
	fmt.Printf("Load() ok, JWTSecret set (len=%d), server starting on port %s\n",
		len(cfg.JWTSecret), cfg.Port)
}
