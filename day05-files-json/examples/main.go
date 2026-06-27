// Day 05 walkthrough — files, io.Reader/Writer, JSON.
// Run: go run main.go
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type Product struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Price float64  `json:"price"`
	Tags  []string `json:"tags,omitempty"`
	sku   string   // unexported -> never serialized
}

// Idiomatic: accept io.Reader, not a filename. Works with files, HTTP
// bodies, or (in this demo) a strings.Reader. Testable without disk.
func countLines(r io.Reader) (int, error) {
	scanner := bufio.NewScanner(r)
	n := 0
	for scanner.Scan() {
		n++
	}
	return n, scanner.Err()
}

func main() {
	fmt.Println("== io.Reader: countLines from a string ==")
	text := "line one\nline two\nline three\n"
	n, _ := countLines(strings.NewReader(text)) // no file needed
	fmt.Printf("  counted %d lines\n", n)

	fmt.Println("== JSON Marshal (note omitempty + unexported sku) ==")
	p := Product{ID: "p1", Name: "Keyboard", Price: 49.99, sku: "secret"}
	b, _ := json.MarshalIndent(p, "  ", "  ")
	fmt.Printf("  %s\n", b) // no "tags", no "sku"

	fmt.Println("== JSON Unmarshal (note &target) ==")
	raw := `{"id":"p2","name":"Mouse","price":19.5,"tags":["wireless"]}`
	var p2 Product
	if err := json.Unmarshal([]byte(raw), &p2); err != nil {
		fmt.Println("  error:", err)
	}
	fmt.Printf("  decoded: %+v\n", p2)

	fmt.Println("== Streaming: Decoder/Encoder over io ==")
	// Decode from any Reader:
	var p3 Product
	_ = json.NewDecoder(strings.NewReader(`{"id":"p3","name":"Cable","price":5}`)).Decode(&p3)
	// Encode straight to any Writer (here a bytes.Buffer):
	var buf bytes.Buffer
	_ = json.NewEncoder(&buf).Encode(p3)
	fmt.Printf("  round-tripped: %s", buf.String())

	fmt.Println("== io.Copy: any Reader -> any Writer ==")
	var out bytes.Buffer
	_, _ = io.Copy(&out, strings.NewReader("streamed bytes")) // 1 line streaming copy
	fmt.Printf("  copied: %q\n", out.String())
}
