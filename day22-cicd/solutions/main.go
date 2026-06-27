// Day 22 — solutions: all 5 lint issues fixed.
// Compare with exercises/main.go to see each fix.
package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
)

// FIX 1: Removed the unused global variable.
// If you actually need it, give it a purpose. If you don't, delete it.

func fetchURL(url string) error {
	// FIX 2: Close the response body. Without this, the underlying TCP
	// connection is not returned to the pool and you leak connections.
	// The `bodyclose` linter catches exactly this pattern.
	resp, err := http.Get(url) //nolint:noctx // solution file
	if err != nil {
		return err
	}
	defer resp.Body.Close() // <- the fix

	fmt.Println("status:", resp.StatusCode)
	return nil
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	// FIX 3: Removed the unused `unused` variable.
	// Go refuses to compile unused local variables — but the `_ = unused`
	// trick suppresses the compile error while leaving dead code.
	// The right fix is to just delete the variable.
	result := string(data)
	return result, nil
}

func divide(a, b int) int {
	// FIX 4: Guard against division by zero.
	if b == 0 {
		return 0
	}
	return a / b
}

// FIX 5: Now the function actually returns an error when name is empty,
// so the (error) return type is meaningful and callers who check it
// get real value.
func greet(name string) error {
	if name == "" {
		return errors.New("greet: name must not be empty")
	}
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

	if err := greet("Majeed"); err != nil {
		fmt.Println("greet error:", err)
	}
}
