// Day 01 walkthrough — run with: go run main.go
//
// Read top to bottom. Each section maps to the lesson. Run it, then change
// things and re-run to see what breaks. Breaking it is how you learn.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// ---- 3. Constants & iota -------------------------------------------------

// Untyped constant — flexible, inlined by the compiler, zero runtime cost.
const MaxRetries = 3

// iota enum: this is how Go does enums. Each line increments iota.
type OrderStatus int

const (
	Pending OrderStatus = iota // 0
	Paid                       // 1
	Shipped                    // 2
	Delivered                  // 3
)

// Giving the enum a String() method makes it print nicely (preview of Day 6 methods).
func (s OrderStatus) String() string {
	switch s {
	case Pending:
		return "PENDING"
	case Paid:
		return "PAID"
	case Shipped:
		return "SHIPPED"
	case Delivered:
		return "DELIVERED"
	default:
		return "UNKNOWN"
	}
}

// ---- 4. Functions --------------------------------------------------------

// The canonical Go signature: (result, error). You'll write this forever.
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

// Variadic: nums is a []int inside the function.
func sum(nums ...int) int {
	total := 0
	for _, n := range nums { // range over slice; _ ignores the index
		total += n
	}
	return total
}

// Closure: returns a function that "remembers" count.
func counter() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

// defer demo: cleanup that always runs, in LIFO order.
func deferDemo() {
	defer fmt.Println("  defer A (printed last)")
	defer fmt.Println("  defer B (printed second)")
	fmt.Println("  body C (printed first)")
}

func main() {
	// ---- 2. Variables & zero values ----
	fmt.Println("== Variables & Zero Values ==")
	var count int      // zero value 0 — immediately safe
	var name string    // zero value "" — empty, not nil
	var active bool    // zero value false
	city := "Bangalore" // short declaration, inferred string
	fmt.Printf("  count=%d name=%q active=%t city=%s\n", count, name, active, city)

	// ---- 3. Constants ----
	fmt.Println("== Constants & iota ==")
	fmt.Printf("  MaxRetries=%d\n", MaxRetries)
	status := Paid
	fmt.Printf("  order status value=%d name=%s\n", status, status)

	// ---- 4. Functions ----
	fmt.Println("== Functions ==")

	// Multiple returns + immediate error handling — the Go heartbeat.
	if result, err := divide(10, 2); err != nil {
		fmt.Println("  error:", err)
	} else {
		fmt.Printf("  10 / 2 = %.1f\n", result)
	}
	if _, err := divide(10, 0); err != nil {
		fmt.Println("  expected error:", err)
	}

	fmt.Printf("  sum(1,2,3,4) = %d\n", sum(1, 2, 3, 4))

	c := counter()
	fmt.Printf("  counter: %d %d %d\n", c(), c(), c())

	fmt.Println("== defer (LIFO) ==")
	deferDemo()

	// Tiny real-world taste: build a CSV line with strings.Builder
	// (its zero value is ready to use — idiomatic Go API design).
	fmt.Println("== strings.Builder (zero value ready) ==")
	var sb strings.Builder
	for i, v := range []string{"id", "name", "status"} {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(v)
	}
	fmt.Println("  header:", sb.String())

	// Exit cleanly. os.Exit skips defers — shown here only to mention that gotcha.
	_ = os.Stdout
}
