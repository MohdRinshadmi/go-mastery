// Day 06 walkthrough — methods, interfaces, composition.
// Run with: go run main.go
//
// Read top to bottom. Each section maps to the lesson.
// Try changing things: swap value/pointer receivers, break interface satisfaction,
// trigger the nil interface gotcha. Breaking it is the lesson.
package main

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

// ==========================================================================
// SECTION 1 — Value vs Pointer receivers
// ==========================================================================

type Rectangle struct {
	Width, Height float64
}

// Value receiver: gets a copy. Pure computation, no mutation.
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

// Value receiver: also fine for immutable derived values.
func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// Pointer receiver: mutates the struct — needs the original.
func (r *Rectangle) Scale(factor float64) {
	r.Width *= factor
	r.Height *= factor
}

// String() with value receiver — implements fmt.Stringer interface.
// fmt.Printf("%v") calls this automatically.
func (r Rectangle) String() string {
	return fmt.Sprintf("Rectangle(%.1f x %.1f)", r.Width, r.Height)
}

// ==========================================================================
// SECTION 2 — Interfaces (implicit satisfaction, no "implements")
// ==========================================================================

// Shape is a small, focused interface. Any type with these two methods satisfies it.
// Note: defined here in the consumer, not forced on the producers.
type Shape interface {
	Area() float64
	Perimeter() float64
}

// Circle — completely separate type, never mentions Shape anywhere.
type Circle struct {
	Radius float64
}

func (c Circle) Area() float64      { return math.Pi * c.Radius * c.Radius }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.Radius }
func (c Circle) String() string {
	return fmt.Sprintf("Circle(r=%.1f)", c.Radius)
}

// Triangle — another producer. Still never says "implements Shape."
type Triangle struct {
	A, B, C float64 // side lengths
}

func (t Triangle) Perimeter() float64 { return t.A + t.B + t.C }
func (t Triangle) Area() float64 {
	// Heron's formula
	s := t.Perimeter() / 2
	return math.Sqrt(s * (s - t.A) * (s - t.B) * (s - t.C))
}

// printShapeInfo accepts any Shape — works with Rectangle, Circle, Triangle,
// and anything you add next year without changing this function.
func printShapeInfo(s Shape) {
	fmt.Printf("  %-30T  area=%.3f  perimeter=%.3f\n", s, s.Area(), s.Perimeter())
}

// ==========================================================================
// SECTION 3 — io.Reader / io.Writer: the stdlib interfaces that unite I/O
// ==========================================================================

// PrefixWriter wraps any io.Writer and prepends a prefix to every write.
// This is the decorator / middleware pattern: satisfies the same interface
// it wraps — so it can be chained indefinitely.
type PrefixWriter struct {
	prefix string
	w      io.Writer
}

func NewPrefixWriter(prefix string, w io.Writer) *PrefixWriter {
	return &PrefixWriter{prefix: prefix, w: w}
}

// Write implements io.Writer — PrefixWriter can go anywhere an io.Writer is accepted.
func (pw *PrefixWriter) Write(p []byte) (int, error) {
	// Prepend prefix to each line
	lines := strings.Split(string(p), "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = pw.prefix + line
		}
	}
	prefixed := strings.Join(lines, "\n")
	return pw.w.Write([]byte(prefixed))
}

// countBytes reads from any io.Reader — file, buffer, HTTP body, stdin...
func countBytes(r io.Reader) (int, error) {
	buf := make([]byte, 256)
	total := 0
	for {
		n, err := r.Read(buf)
		total += n
		if err == io.EOF {
			return total, nil
		}
		if err != nil {
			return total, err
		}
	}
}

// ==========================================================================
// SECTION 4 — Empty interface / any + type switches
// ==========================================================================

// describe uses a type switch — idiomatic when you must handle concrete types.
// In new code, prefer generics (Day 8) over type switches on basic types.
func describe(i any) string {
	switch v := i.(type) {
	case nil:
		return "nil"
	case int:
		return fmt.Sprintf("int(%d)", v)
	case float64:
		return fmt.Sprintf("float64(%.2f)", v)
	case string:
		return fmt.Sprintf("string(%q, len=%d)", v, len(v))
	case bool:
		return fmt.Sprintf("bool(%t)", v)
	case []int:
		return fmt.Sprintf("[]int(len=%d, cap=%d)", len(v), cap(v))
	case Shape:
		return fmt.Sprintf("Shape: area=%.2f", v.Area())
	case error:
		return fmt.Sprintf("error: %v", v)
	default:
		return fmt.Sprintf("unknown: %T = %v", v, v)
	}
}

// ==========================================================================
// SECTION 5 — The nil interface gotcha (read carefully)
// ==========================================================================

type AppError struct {
	Code    int
	Message string
}

func (e *AppError) Error() string {
	return fmt.Sprintf("error %d: %s", e.Code, e.Message)
}

// WRONG: returns a typed nil — interface is NOT nil even though *AppError is nil.
func getBadError(fail bool) error {
	var err *AppError // typed nil pointer
	if fail {
		err = &AppError{500, "internal server error"}
	}
	return err // interface = (*AppError, nil) ← NOT a nil interface!
}

// RIGHT: return nil directly, never return a typed nil pointer as an interface.
func getGoodError(fail bool) error {
	if fail {
		return &AppError{500, "internal server error"}
	}
	return nil // interface = (nil, nil) ← truly nil
}

// ==========================================================================
// SECTION 6 — Interface composition + the Stringer interface
// ==========================================================================

// fmt.Stringer is defined in fmt package as:
//   type Stringer interface { String() string }
//
// Rectangle and Circle both have String() — they both satisfy fmt.Stringer.
// fmt.Printf calls String() automatically when you use %v or %s.

// Named type on primitive — methods on non-struct types.
type Celsius float64
type Fahrenheit float64

func (c Celsius) String() string    { return fmt.Sprintf("%.1f°C", float64(c)) }
func (f Fahrenheit) String() string { return fmt.Sprintf("%.1f°F", float64(f)) }

func (c Celsius) ToFahrenheit() Fahrenheit {
	return Fahrenheit(c*9/5 + 32)
}

// ==========================================================================
// main
// ==========================================================================

func main() {
	// --- Section 1: Methods ---
	fmt.Println("== 1. Value vs Pointer Receivers ==")
	r := Rectangle{Width: 5, Height: 3}
	fmt.Printf("  %v  area=%.1f  perimeter=%.1f\n", r, r.Area(), r.Perimeter())
	r.Scale(2)
	fmt.Printf("  after Scale(2): %v\n", r)

	// Go auto-takes address when you call pointer receiver on addressable value.
	// r.Scale(2) is shorthand for (&r).Scale(2).

	// --- Section 2: Interface satisfaction ---
	fmt.Println("\n== 2. Interface Satisfaction (implicit) ==")
	shapes := []Shape{
		Rectangle{Width: 4, Height: 6},
		Circle{Radius: 3},
		Triangle{A: 3, B: 4, C: 5},
	}
	for _, s := range shapes {
		printShapeInfo(s)
	}

	// --- Section 3: io.Reader / io.Writer ---
	fmt.Println("\n== 3. io.Reader / io.Writer ==")

	// PrefixWriter decorates os.Stdout — same interface, extra behavior.
	pw := NewPrefixWriter("  [LOG] ", os.Stdout)
	fmt.Fprintln(pw, "server started")
	fmt.Fprintln(pw, "listening on :8080")

	// countBytes works with bytes.NewReader, strings.NewReader, os.File, etc.
	content := "Hello, interface!\nSecond line.\n"
	n, _ := countBytes(strings.NewReader(content))
	fmt.Printf("  countBytes(strings.NewReader) = %d bytes\n", n)

	var buf bytes.Buffer
	buf.WriteString("buffered content here")
	n, _ = countBytes(&buf)
	fmt.Printf("  countBytes(bytes.Buffer) = %d bytes\n", n)

	// --- Section 4: Type switch ---
	fmt.Println("\n== 4. Type Switch ==")
	vals := []any{42, 3.14, "hello", true, []int{1, 2, 3}, Circle{Radius: 2}, nil}
	for _, v := range vals {
		fmt.Printf("  describe(%v) → %s\n", v, describe(v))
	}

	// --- Section 5: nil interface gotcha ---
	fmt.Println("\n== 5. Nil Interface Gotcha ==")
	badErr := getBadError(false)
	if badErr != nil {
		// This branch RUNS even though we passed fail=false!
		fmt.Printf("  BAD: getBadError(false) != nil — gotcha! type=%T value=%v\n", badErr, badErr)
	}

	goodErr := getGoodError(false)
	if goodErr == nil {
		fmt.Println("  GOOD: getGoodError(false) == nil — correct!")
	}

	goodErr2 := getGoodError(true)
	if goodErr2 != nil {
		fmt.Println("  GOOD: getGoodError(true) != nil:", goodErr2)
	}

	// --- Section 6: Named types with methods ---
	fmt.Println("\n== 6. Named Type Methods + Stringer ==")
	temp := Celsius(37)
	fmt.Printf("  Body temperature: %v = %v\n", temp, temp.ToFahrenheit())

	// fmt.Stringer in action — %v calls String() automatically
	temps := []Celsius{0, 20, 37, 100}
	for _, t := range temps {
		fmt.Printf("  %v → %v\n", t, t.ToFahrenheit())
	}

	// --- Bonus: safe type assertion ---
	fmt.Println("\n== Bonus: Type Assertion ==")
	var s Shape = Circle{Radius: 5}
	if c, ok := s.(Circle); ok {
		fmt.Printf("  s is a Circle with radius %.1f\n", c.Radius)
	}
	if _, ok := s.(Rectangle); !ok {
		fmt.Println("  s is NOT a Rectangle — safe assertion returned ok=false")
	}
}
