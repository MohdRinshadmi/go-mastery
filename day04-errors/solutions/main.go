// Day 04 — reference solutions. Try first! Run: go run main.go
package main

import (
	"errors"
	"fmt"
	"strconv"
)

// Exercise 1
var ErrEmptyInput = errors.New("empty input")

func firstRune(s string) (rune, error) {
	if s == "" {
		return 0, ErrEmptyInput
	}
	for _, r := range s { // range over string yields runes
		return r, nil
	}
	return 0, ErrEmptyInput
}

// Exercise 2
func parsePort(s string) (int, error) {
	p, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("parsePort %q: %w", s, err)
	}
	if p < 1 || p > 65535 {
		return 0, fmt.Errorf("parsePort: %d out of range 1..65535", p)
	}
	return p, nil
}

// Exercise 3 (reused in challenge)
type RangeError struct {
	Value, Min, Max int
}

func (e *RangeError) Error() string {
	return fmt.Sprintf("%d not in range [%d, %d]", e.Value, e.Min, e.Max)
}

func checkRange(v, min, max int) error {
	if v < min || v > max {
		return &RangeError{Value: v, Min: min, Max: max}
	}
	return nil
}

// Challenge
type Config struct {
	Port int
	Host string
}

var ErrMissing = errors.New("config key missing")

func loadConfig(raw map[string]string) (Config, error) {
	host, ok := raw["host"]
	if !ok {
		return Config{}, fmt.Errorf("loadConfig host: %w", ErrMissing)
	}
	portStr, ok := raw["port"]
	if !ok {
		return Config{}, fmt.Errorf("loadConfig port: %w", ErrMissing)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return Config{}, fmt.Errorf("loadConfig port: %w", err)
	}
	if re := checkRange(port, 1, 65535); re != nil {
		return Config{}, fmt.Errorf("loadConfig port: %w", re)
	}
	return Config{Port: port, Host: host}, nil
}

func main() {
	fmt.Println("== Exercise 1 ==")
	for _, s := range []string{"", "hi"} {
		r, err := firstRune(s)
		if errors.Is(err, ErrEmptyInput) {
			fmt.Printf("  %q -> empty\n", s)
		} else {
			fmt.Printf("  %q -> %c\n", s, r)
		}
	}

	fmt.Println("== Exercise 2 ==")
	for _, s := range []string{"8080", "abc", "99999"} {
		if p, err := parsePort(s); err != nil {
			fmt.Println("  err:", err)
		} else {
			fmt.Println("  port:", p)
		}
	}

	fmt.Println("== Exercise 3 ==")
	err := checkRange(200, 0, 100)
	var re *RangeError
	if errors.As(err, &re) {
		fmt.Printf("  out of range, value=%d\n", re.Value)
	}

	fmt.Println("== Challenge ==")
	inputs := []map[string]string{
		{"host": "localhost", "port": "8080"},
		{"port": "8080"},
		{"host": "localhost", "port": "70000"},
	}
	for i, in := range inputs {
		cfg, err := loadConfig(in)
		switch {
		case err == nil:
			fmt.Printf("  [%d] ok: %+v\n", i, cfg)
		case errors.Is(err, ErrMissing):
			fmt.Printf("  [%d] missing key: %v\n", i, err)
		default:
			var re *RangeError
			if errors.As(err, &re) {
				fmt.Printf("  [%d] bad port value %d: %v\n", i, re.Value, err)
			} else {
				fmt.Printf("  [%d] other: %v\n", i, err)
			}
		}
	}
}
