// Day 22 — YOUR exercises.
//
// EXERCISE 1: This file has 5 issues that golangci-lint would catch.
// Find and fix them without running the linter first (think like a linter).
// Then verify with: go vet ./...
// After: install golangci-lint and run: golangci-lint run ./...
//
// EXERCISE 2: Write a .golangci.yml in this directory that:
//   - enables errcheck, staticcheck, gosec, bodyclose
//   - sets a 3-minute timeout
//   - disables gochecknoglobals (we need package-level vars here)
//   - excludes _test.go from errcheck
//
// EXERCISE 3: Write a Makefile with targets: vet, test, build, ci
//   - ci target runs vet, test, build in sequence
//   - test uses -race flag
//   - build injects VERSION from git or "dev"
//
// EXERCISE 4 (CHALLENGE): Write .github/workflows/ci.yml that:
//   - triggers on push to main and pull_request to main
//   - has a lint job, a test job, and a build job
//   - test job depends on lint job (not build)
//   - build job depends on test job
//   - uses actions/setup-go@v5 with cache: true
//   - runs go test -race ./...
//
// Fill in the TODOs below to fix the 5 lint issues.

package main

import (
	"fmt"
	"net/http"
	"os"
)

// ISSUE 1: This global is never used outside this file and has no purpose.
// A linter would flag this as "unused" or "ineffectual".
var unusedGlobal = "I am never used"

func fetchURL(url string) error {
	// ISSUE 2: http.Get returns (*Response, error). The response body
	// must be closed or we leak connections. What linter catches this?
	// TODO: fix this
	resp, err := http.Get(url) //nolint:noctx // exercise file
	if err != nil {
		return err
	}
	fmt.Println("status:", resp.StatusCode)
	return nil
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	// ISSUE 3: We declared 'result' but immediately returned it — that's fine.
	// But look at this unused variable:
	unused := "this is never used" // TODO: remove this
	_ = unused
	result := string(data)
	return result, nil
}

func divide(a, b int) int {
	// ISSUE 4: This ignores the possibility of division by zero.
	// Not a linter issue per se, but write it defensively.
	// For this exercise: add a guard and return 0 if b == 0.
	// TODO: add guard
	return a / b
}

// ISSUE 5: This function signature has a problem — it returns error but
// the body never actually returns a non-nil error. The errcheck linter
// would flag callers who ignore this return value.
// For this exercise: actually USE the error return meaningfully.
// TODO: rewrite to return an error when name is empty.
func greet(name string) error {
	fmt.Println("Hello,", name)
	return nil
}

func main() {
	if err := fetchURL("https://example.com"); err != nil {
		fmt.Println("fetch error:", err)
	}

	if content, err := readFile("nonexistent.txt"); err != nil {
		fmt.Println("read error:", err)
	} else {
		fmt.Println("content:", content)
	}

	fmt.Println("10 / 2 =", divide(10, 2))
	fmt.Println("10 / 0 =", divide(10, 0))

	if err := greet(""); err != nil {
		fmt.Println("greet error:", err)
	}
}
