// Day 07 — reference solutions. Try the exercises yourself FIRST.
// Run with: go run main.go
package main

import (
	"errors"
	"fmt"
	"time"
)

// =====================================================================
// EXERCISE 1 — Embedding for delegation
// =====================================================================

type Counter struct {
	Value int
}

func (c *Counter) Increment() { c.Value++ }
func (c *Counter) Reset()     { c.Value = 0 }

type RateLimitedCounter struct {
	Counter
	MaxPerSecond int
	lastReset    time.Time
}

// Increment overrides Counter.Increment — adds rate-limiting logic.
// Calls the embedded Counter.Increment() when within limits.
func (r *RateLimitedCounter) Increment() error {
	if r.Value >= r.MaxPerSecond {
		return errors.New("rate limit exceeded")
	}
	r.Counter.Increment() // explicit delegation to the embedded type
	return nil
}

// =====================================================================
// EXERCISE 2 — Payment service with dependency injection
// =====================================================================

type Payment struct {
	ID, CustomerID, Amount, Status string
}

// Interfaces defined in the consumer (PaymentService's package).
type PaymentStore interface {
	SavePayment(p Payment) error
	GetPayment(id string) (Payment, error)
}

type Notifier interface {
	Notify(to, message string) error
}

// InMemoryPaymentStore — concrete implementation.
type InMemoryPaymentStore struct {
	payments map[string]Payment
}

func NewInMemoryPaymentStore() *InMemoryPaymentStore {
	return &InMemoryPaymentStore{payments: make(map[string]Payment)}
}

func (s *InMemoryPaymentStore) SavePayment(p Payment) error {
	s.payments[p.ID] = p
	return nil
}

func (s *InMemoryPaymentStore) GetPayment(id string) (Payment, error) {
	p, ok := s.payments[id]
	if !ok {
		return Payment{}, fmt.Errorf("payment %s not found", id)
	}
	return p, nil
}

// ConsoleNotifier — concrete implementation.
type ConsoleNotifier struct{}

func (n ConsoleNotifier) Notify(to, message string) error {
	fmt.Printf("  NOTIFY to %s: %s\n", to, message)
	return nil
}

// PaymentService — business logic only, depends on interfaces.
type PaymentService struct {
	store    PaymentStore
	notifier Notifier
}

func NewPaymentService(store PaymentStore, notifier Notifier) *PaymentService {
	return &PaymentService{store: store, notifier: notifier}
}

func (s *PaymentService) Charge(customerID, amount string) (Payment, error) {
	p := Payment{
		ID:         "pay_" + customerID,
		CustomerID: customerID,
		Amount:     amount,
		Status:     "completed",
	}
	if err := s.store.SavePayment(p); err != nil {
		return Payment{}, fmt.Errorf("charge: save payment: %w", err)
	}
	if err := s.notifier.Notify(customerID, "Payment of $"+amount+" received"); err != nil {
		// Non-fatal: notification failure doesn't undo the charge.
		fmt.Printf("  warning: notify failed: %v\n", err)
	}
	return p, nil
}

func (s *PaymentService) GetPayment(id string) (Payment, error) {
	return s.store.GetPayment(id)
}

// Compile-time checks.
var _ PaymentStore = (*InMemoryPaymentStore)(nil)
var _ Notifier = ConsoleNotifier{}

// =====================================================================
// CHALLENGE — Functional options + decorator
// =====================================================================

type ClientConfig struct {
	BaseURL string
	Timeout time.Duration
	Retries int
	Debug   bool
}

type Option func(*ClientConfig)

func WithBaseURL(url string) Option {
	return func(c *ClientConfig) { c.BaseURL = url }
}
func WithTimeout(d time.Duration) Option {
	return func(c *ClientConfig) { c.Timeout = d }
}
func WithRetries(n int) Option {
	return func(c *ClientConfig) { c.Retries = n }
}
func WithDebug(v bool) Option {
	return func(c *ClientConfig) { c.Debug = v }
}

type Doer interface {
	Do(method, path string) (string, error)
}

type HTTPClient struct {
	cfg ClientConfig
}

func NewHTTPClient(opts ...Option) *HTTPClient {
	cfg := ClientConfig{
		BaseURL: "https://api.example.com",
		Timeout: 30 * time.Second,
		Retries: 3,
		Debug:   false,
	}
	for _, o := range opts {
		o(&cfg)
	}
	return &HTTPClient{cfg: cfg}
}

func (c *HTTPClient) Do(method, path string) (string, error) {
	if c.cfg.Debug {
		fmt.Printf("  DEBUG: %s %s%s\n", method, c.cfg.BaseURL, path)
	}
	return fmt.Sprintf("%s %s → 200 OK (simulated)", method, path), nil
}

// RetryingClient wraps any Doer and retries on error.
type RetryingClient struct {
	inner   Doer
	retries int
}

func NewRetryingClient(inner Doer, retries int) *RetryingClient {
	return &RetryingClient{inner: inner, retries: retries}
}

func (r *RetryingClient) Do(method, path string) (string, error) {
	var lastErr error
	for i := 0; i <= r.retries; i++ {
		result, err := r.inner.Do(method, path)
		if err == nil {
			return result, nil
		}
		lastErr = err
		fmt.Printf("  retry %d/%d failed: %v\n", i+1, r.retries, err)
	}
	return "", fmt.Errorf("all %d retries failed: %w", r.retries, lastErr)
}

// Compile-time checks.
var _ Doer = (*HTTPClient)(nil)
var _ Doer = (*RetryingClient)(nil)

// =====================================================================
// main
// =====================================================================

func main() {
	fmt.Println("== Exercise 1: Rate Limited Counter ==")
	rlc := &RateLimitedCounter{MaxPerSecond: 3}
	for i := 0; i < 5; i++ {
		if err := rlc.Increment(); err != nil {
			fmt.Printf("  call %d: error: %v\n", i+1, err)
		} else {
			fmt.Printf("  call %d: value = %d\n", i+1, rlc.Value)
		}
	}
	rlc.Reset() // promoted from Counter
	fmt.Printf("  after Reset: value = %d\n", rlc.Value)

	fmt.Println("\n== Exercise 2: Payment Service ==")
	store := NewInMemoryPaymentStore()
	notifier := ConsoleNotifier{}
	svc := NewPaymentService(store, notifier)

	p1, err := svc.Charge("alice", "49.99")
	if err != nil {
		fmt.Println("  charge error:", err)
	} else {
		fmt.Printf("  charged: %+v\n", p1)
	}

	p2, _ := svc.Charge("bob", "99.00")
	fmt.Printf("  charged: %+v\n", p2)

	retrieved, _ := svc.GetPayment(p1.ID)
	fmt.Printf("  retrieved: %+v\n", retrieved)

	fmt.Println("\n== Challenge: Functional Options + Decorator ==")
	client := NewHTTPClient(
		WithBaseURL("https://staging.api.example.com"),
		WithTimeout(10*time.Second),
		WithDebug(true),
	)
	retrying := NewRetryingClient(client, 3)

	paths := []string{"/users", "/orders", "/health"}
	for _, p := range paths {
		result, err := retrying.Do("GET", p)
		if err != nil {
			fmt.Println("  error:", err)
		} else {
			fmt.Println(" ", result)
		}
	}
}
