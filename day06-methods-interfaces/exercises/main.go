// Day 06 — YOUR exercises. Fill in the TODOs.
//
// Run with:   go run main.go
// I will review this like a production PR. Write clean, idiomatic Go.
// Don't peek at ../solutions/ until you've genuinely tried each one.
package main

import (
	"fmt"
	"io"
	"math"
	"strings"
)

// =====================================================================
// EXERCISE 1 (beginner) — Shape interface
//
// Define a Shape interface with two methods: Area() float64 and Perimeter() float64.
// Implement it for Circle and Rectangle.
// Write a function: totalArea(shapes []Shape) float64
// In main: create a slice of shapes and print the total area.
// =====================================================================

type Shape interface {
	// TODO: add Area() and Perimeter() methods
}

type Circle struct {
	Radius float64
}

// TODO: implement Area and Perimeter for Circle (use math.Pi, math.Sqrt as needed)
func (c Circle) Area() float64      { return 0 }
func (c Circle) Perimeter() float64 { return 0 }

type Rectangle struct {
	Width, Height float64
}

// TODO: implement Area and Perimeter for Rectangle
func (r Rectangle) Area() float64      { return 0 }
func (r Rectangle) Perimeter() float64 { return 0 }

func totalArea(shapes []Shape) float64 {
	// TODO: sum and return the area of all shapes
	return 0
}

// =====================================================================
// EXERCISE 2 (beginner) — fmt.Stringer
//
// Add a String() method to both Circle and Rectangle so they print nicely.
// Circle should print: "Circle(r=3.00)"
// Rectangle should print: "Rectangle(4.00 x 2.00)"
// In main: fmt.Printf("%v\n", ...) each shape and confirm the output.
// =====================================================================

// TODO: add String() methods to Circle and Rectangle above (edit them in place).

// =====================================================================
// EXERCISE 3 (intermediate) — io.Writer middleware
//
// Implement a UpperWriter that wraps any io.Writer.
// Every call to Write should uppercase the bytes before writing to the underlying writer.
// Use strings.ToUpper on the string representation.
//
// In main: wrap os.Stdout with UpperWriter and fmt.Fprintln several strings through it.
// Then also wrap a *bytes.Buffer and verify the contents.
// =====================================================================

type UpperWriter struct {
	// TODO: add fields (hint: the underlying io.Writer)
}

// TODO: implement NewUpperWriter(w io.Writer) *UpperWriter

func (uw *UpperWriter) Write(p []byte) (int, error) {
	// TODO: uppercase p, write to underlying writer
	return 0, nil
}

// =====================================================================
// CHALLENGE (advanced) — Notification system with interface composition
//
// Design a notification system:
//
// 1. Define a Notifier interface:
//      Send(to, subject, body string) error
//
// 2. Implement two concrete notifiers:
//    - EmailNotifier: prints "EMAIL to <to> | <subject>: <body>"
//    - SMSNotifier: prints "SMS to <to> | <body>" (subject ignored)
//
// 3. Implement a MultiNotifier that holds []Notifier and sends to all of them.
//    If any Send fails, it collects ALL errors and returns them joined with "; ".
//    (Hint: use strings.Builder or fmt.Errorf)
//
// 4. Implement a LoggingNotifier that wraps any Notifier and logs:
//    "[NOTIFY] sending to <to>" before calling the underlying Send,
//    then "[NOTIFY] done (err=<nil or error>)" after.
//
// 5. In main:
//    - Create a MultiNotifier containing email + SMS.
//    - Wrap it with LoggingNotifier.
//    - Call Send("alice@example.com", "Hello", "This is a test.")
//    - The output should show all four log/send lines.
// =====================================================================

type Notifier interface {
	// TODO: define Send
}

type EmailNotifier struct{}
type SMSNotifier struct{}
type MultiNotifier struct {
	// TODO: add fields
}
type LoggingNotifier struct {
	// TODO: add fields
}

// TODO: implement Send for each type

// Keep these here so the file compiles with stubs. Remove when implemented.
var _ = math.Pi
var _ = strings.ToUpper
var _ io.Writer = (*UpperWriter)(nil) // compile-time interface check

func main() {
	fmt.Println("== Exercise 1: Shape interface & totalArea ==")
	// TODO: create []Shape{Circle{3}, Rectangle{4, 5}}, print each shape and total area

	fmt.Println("\n== Exercise 2: Stringer ==")
	// TODO: fmt.Printf("%v\n", ...) each shape

	fmt.Println("\n== Exercise 3: UpperWriter ==")
	// TODO: wrap os.Stdout, write strings through it

	fmt.Println("\n== Challenge: Notification system ==")
	// TODO: build multi+logging notifier chain and send a notification
}
