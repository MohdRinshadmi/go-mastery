// Day 06 — reference solutions. Try the exercises yourself FIRST.
// Run with: go run main.go
package main

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

// =====================================================================
// EXERCISE 1 + 2 — Shape interface + Stringer
// =====================================================================

type Shape interface {
	Area() float64
	Perimeter() float64
}

type Circle struct{ Radius float64 }

func (c Circle) Area() float64      { return math.Pi * c.Radius * c.Radius }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.Radius }
func (c Circle) String() string     { return fmt.Sprintf("Circle(r=%.2f)", c.Radius) }

type Rectangle struct{ Width, Height float64 }

func (r Rectangle) Area() float64      { return r.Width * r.Height }
func (r Rectangle) Perimeter() float64 { return 2 * (r.Width + r.Height) }
func (r Rectangle) String() string {
	return fmt.Sprintf("Rectangle(%.2f x %.2f)", r.Width, r.Height)
}

func totalArea(shapes []Shape) float64 {
	var total float64
	for _, s := range shapes {
		total += s.Area()
	}
	return total
}

// =====================================================================
// EXERCISE 3 — UpperWriter
// =====================================================================

type UpperWriter struct {
	w io.Writer
}

func NewUpperWriter(w io.Writer) *UpperWriter {
	return &UpperWriter{w: w}
}

func (uw *UpperWriter) Write(p []byte) (int, error) {
	upper := strings.ToUpper(string(p))
	return uw.w.Write([]byte(upper))
}

// Compile-time check: *UpperWriter satisfies io.Writer.
var _ io.Writer = (*UpperWriter)(nil)

// =====================================================================
// CHALLENGE — Notification system
// =====================================================================

type Notifier interface {
	Send(to, subject, body string) error
}

type EmailNotifier struct{}

func (e EmailNotifier) Send(to, subject, body string) error {
	fmt.Printf("  EMAIL to %s | %s: %s\n", to, subject, body)
	return nil
}

type SMSNotifier struct{}

func (s SMSNotifier) Send(to, subject, body string) error {
	// SMS ignores subject
	fmt.Printf("  SMS to %s | %s\n", to, body)
	return nil
}

// MultiNotifier sends to all notifiers and collects errors.
type MultiNotifier struct {
	notifiers []Notifier
}

func NewMultiNotifier(ns ...Notifier) *MultiNotifier {
	return &MultiNotifier{notifiers: ns}
}

func (m *MultiNotifier) Send(to, subject, body string) error {
	var errs []string
	for _, n := range m.notifiers {
		if err := n.Send(to, subject, body); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// LoggingNotifier wraps any Notifier and logs before/after.
type LoggingNotifier struct {
	inner Notifier
}

func NewLoggingNotifier(n Notifier) *LoggingNotifier {
	return &LoggingNotifier{inner: n}
}

func (l *LoggingNotifier) Send(to, subject, body string) error {
	fmt.Printf("  [NOTIFY] sending to %s\n", to)
	err := l.inner.Send(to, subject, body)
	fmt.Printf("  [NOTIFY] done (err=%v)\n", err)
	return err
}

// Compile-time checks: all types satisfy Notifier.
var _ Notifier = EmailNotifier{}
var _ Notifier = SMSNotifier{}
var _ Notifier = (*MultiNotifier)(nil)
var _ Notifier = (*LoggingNotifier)(nil)

// =====================================================================
// main
// =====================================================================

func main() {
	fmt.Println("== Exercise 1: Shape interface & totalArea ==")
	shapes := []Shape{
		Circle{Radius: 3},
		Rectangle{Width: 4, Height: 5},
		Circle{Radius: 1.5},
	}
	for _, s := range shapes {
		fmt.Printf("  area=%.3f  perimeter=%.3f\n", s.Area(), s.Perimeter())
	}
	fmt.Printf("  totalArea = %.3f\n", totalArea(shapes))

	fmt.Println("\n== Exercise 2: Stringer ==")
	fmt.Printf("  %v\n", Circle{Radius: 3})
	fmt.Printf("  %v\n", Rectangle{Width: 4, Height: 2})

	fmt.Println("\n== Exercise 3: UpperWriter ==")
	uw := NewUpperWriter(os.Stdout)
	fmt.Fprintln(uw, "  hello world from UpperWriter")
	fmt.Fprintln(uw, "  go interfaces are awesome")

	var buf strings.Builder
	ubuf := NewUpperWriter(&buf)
	fmt.Fprintln(ubuf, "buffered text")
	fmt.Printf("  buffer contents: %q\n", buf.String())

	fmt.Println("\n== Challenge: Notification system ==")
	multi := NewMultiNotifier(EmailNotifier{}, SMSNotifier{})
	logged := NewLoggingNotifier(multi)
	if err := logged.Send("alice@example.com", "Hello", "This is a test."); err != nil {
		fmt.Println("  send error:", err)
	}
}
