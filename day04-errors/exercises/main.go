// Day 04 — YOUR exercises. Fill in the TODOs. Run: go run main.go
// Mentor will review like a PR. Don't peek at ../solutions until you try.
package main

import "fmt"

// =====================================================================
// EXERCISE 1 (beginner) — Sentinel error
// Define ErrEmptyInput. Write firstRune(s string) (rune, error) that
// returns ErrEmptyInput when s is "". Caller in main should detect it
// with errors.Is (add the "errors" import).
// =====================================================================

// TODO: var ErrEmptyInput = ...

func firstRune(s string) (rune, error) {
	// TODO
	return 0, nil
}

// =====================================================================
// EXERCISE 2 (beginner) — Wrapping with %w
// Write parsePort(s string) (int, error). Use strconv.Atoi; if it fails,
// wrap the error with context: "parsePort %q: %w". Also reject ports
// outside 1..65535 with a plain fmt.Errorf (no %w needed there).
// =====================================================================

func parsePort(s string) (int, error) {
	// TODO: import "strconv"
	return 0, nil
}

// =====================================================================
// EXERCISE 3 (beginner) — Custom error type
// Define type RangeError struct{ Value, Min, Max int } with an Error()
// method. Write checkRange(v, min, max int) error returning *RangeError
// when out of range. In main, extract it with errors.As and print Value.
// =====================================================================

// TODO: type RangeError ...

func checkRange(v, min, max int) error {
	// TODO
	return nil
}

// =====================================================================
// CHALLENGE (intermediate) — config loader with branching errors
//
// type Config struct{ Port int; Host string }
// var ErrMissing = errors.New("config key missing")
//
// loadConfig(raw map[string]string) (Config, error):
//   - "host" missing  -> wrap ErrMissing: "loadConfig host: %w"
//   - "port" missing  -> wrap ErrMissing similarly
//   - "port" not a valid 1..65535 int -> return a *RangeError (reuse ex.3)
//     wrapped with context via %w
// In main, call it with 3 inputs (valid, missing host, bad port) and use
// errors.Is(err, ErrMissing) and errors.As(&RangeError) to branch and
// print a different message for each failure category.
// =====================================================================

type Config struct {
	Port int
	Host string
}

func loadConfig(raw map[string]string) (Config, error) {
	// TODO
	return Config{}, nil
}

func main() {
	fmt.Println("== Exercise 1 ==")
	// TODO: call firstRune("") and firstRune("hi"), handle errors.Is

	fmt.Println("== Exercise 2 ==")
	// TODO: parsePort("8080"), parsePort("abc"), parsePort("99999")

	fmt.Println("== Exercise 3 ==")
	// TODO: checkRange(200, 0, 100), extract *RangeError via errors.As

	fmt.Println("== Challenge ==")
	// TODO: 3 loadConfig calls, branch with errors.Is / errors.As
}
