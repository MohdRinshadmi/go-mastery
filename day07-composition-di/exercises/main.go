// Day 07 — YOUR exercises. Fill in the TODOs.
//
// Run with:   go run main.go
// Don't peek at ../solutions/ until you've genuinely tried each one.
package main

import (
	"errors"
	"fmt"
	"time"
)

// =====================================================================
// EXERCISE 1 (beginner) — Embedding for delegation
//
// Create a Counter struct with:
//   - Value int
//   - Increment() — adds 1
//   - Reset()      — sets Value to 0
//
// Create a RateLimitedCounter that EMBEDS Counter and adds:
//   - MaxPerSecond int
//   - lastReset    time.Time (private)
//   - Increment()  — overrides the embedded one: checks if Value >= MaxPerSecond;
//                    if so, returns an error "rate limit exceeded";
//                    otherwise calls Counter.Increment() and returns nil.
//
// In main: create RateLimitedCounter{MaxPerSecond: 3}, call Increment 5 times,
// print the value and any errors. Then call Reset() (promoted from Counter) and
// show the value is 0.
// =====================================================================

type Counter struct {
	Value int
}

// TODO: implement Increment() and Reset() on Counter

type RateLimitedCounter struct {
	Counter
	MaxPerSecond int
	lastReset    time.Time
}

// TODO: implement Increment() error on RateLimitedCounter
// (this shadows / overrides Counter.Increment)

// =====================================================================
// EXERCISE 2 (intermediate) — Define a payment service with DI
//
// Define these interfaces (consumer owns them):
//
//   PaymentStore interface:
//     SavePayment(p Payment) error
//     GetPayment(id string) (Payment, error)
//
//   Notifier interface:
//     Notify(to, message string) error
//
// Define the Payment struct:
//   ID, CustomerID, Amount string, Status string
//
// Implement:
//   InMemoryPaymentStore — backed by a map
//   ConsoleNotifier     — just prints "NOTIFY to <to>: <message>"
//
// Implement PaymentService:
//   fields: store PaymentStore, notifier Notifier
//   constructor: NewPaymentService(store, notifier) *PaymentService
//   methods:
//     Charge(customerID, amount string) (Payment, error)
//       — generates an ID ("pay_" + customerID), sets Status="completed"
//       — saves to store
//       — notifies customer ("Payment of $<amount> received")
//       — returns the Payment
//     GetPayment(id string) (Payment, error)
//       — delegates to store
//
// In main: wire up with in-memory store + console notifier, charge two customers.
// =====================================================================

type Payment struct {
	ID, CustomerID, Amount, Status string
}

type PaymentStore interface {
	// TODO: SavePayment and GetPayment
}

type Notifier interface {
	// TODO: Notify
}

type InMemoryPaymentStore struct {
	// TODO: backing map
}

// TODO: implement PaymentStore methods on InMemoryPaymentStore

type ConsoleNotifier struct{}

// TODO: implement Notifier on ConsoleNotifier

type PaymentService struct {
	// TODO: fields
}

// TODO: constructor and methods

// =====================================================================
// CHALLENGE — Functional options + decorator pattern
//
// Build a configurable HTTP client stub:
//
// 1. ClientConfig fields: BaseURL string, Timeout time.Duration, Retries int, Debug bool
//    Defaults: BaseURL="https://api.example.com", Timeout=30s, Retries=3, Debug=false
//
// 2. Define Option type and implement:
//    WithBaseURL(string) Option
//    WithTimeout(time.Duration) Option
//    WithRetries(int) Option
//    WithDebug(bool) Option
//
// 3. HTTPClient struct: cfg ClientConfig
//    NewHTTPClient(opts ...Option) *HTTPClient
//    Do(method, path string) (string, error):
//      — if Debug is true: print "DEBUG: <method> <BaseURL><path>"
//      — always return: "<method> <path> → 200 OK (simulated)", nil
//
// 4. RetryingClient wraps any Doer interface:
//      type Doer interface { Do(method, path string) (string, error) }
//    RetryingClient.Do retries on error up to cfg.Retries times.
//    (For this exercise, the wrapped HTTPClient never errors, so just call through.)
//
// In main: create an HTTPClient with debug=true and custom URL,
// wrap it in RetryingClient, call Do three times with different paths.
// =====================================================================

// TODO: implement ClientConfig, Option, HTTPClient, RetryingClient

// Keep errors imported to avoid "imported and not used" if you don't use it yet.
var _ = errors.New
var _ = time.Second

func main() {
	fmt.Println("== Exercise 1: Rate Limited Counter ==")
	// TODO

	fmt.Println("\n== Exercise 2: Payment Service ==")
	// TODO

	fmt.Println("\n== Challenge: Functional Options + Decorator ==")
	// TODO
}
