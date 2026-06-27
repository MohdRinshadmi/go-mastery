// Day 07 walkthrough — struct embedding, composition, dependency injection.
// Run with: go run main.go
//
// Read top to bottom. Notice: no class hierarchy anywhere.
// All wiring happens in main(). Business logic never imports infrastructure.
package main

import (
	"fmt"
	"strings"
	"time"
)

// ==========================================================================
// SECTION 1 — Struct Embedding: delegation, not inheritance
// ==========================================================================

// Logger is a small, focused helper. Knows nothing about what uses it.
type Logger struct {
	prefix string
}

func NewLogger(prefix string) *Logger { return &Logger{prefix: prefix} }

func (l *Logger) Log(msg string) {
	fmt.Printf("[%s] %s %s\n", l.prefix, time.Now().Format("15:04:05"), msg)
}
func (l *Logger) Logf(format string, args ...any) {
	l.Log(fmt.Sprintf(format, args...))
}

// Timer is another small helper for measuring latency.
type Timer struct {
	start time.Time
}

func (t *Timer) Start()              { t.start = time.Now() }
func (t *Timer) Elapsed() time.Duration { return time.Since(t.start) }

// HTTPHandler embeds both Logger and Timer by pointer.
// It delegates logging/timing, not inherits it.
// Crucially: HTTPHandler is NOT a Logger and NOT a Timer.
type HTTPHandler struct {
	*Logger
	*Timer
	route string
}

func NewHTTPHandler(route string) *HTTPHandler {
	return &HTTPHandler{
		Logger: NewLogger("HTTP"),
		Timer:  &Timer{},
		route:  route,
	}
}

func (h *HTTPHandler) Handle(method, path string) {
	h.Timer.Start()
	h.Logf("→ %s %s", method, path) // h.Logger.Logf via promotion
	// ... handle request ...
	h.Logf("← %s %s took %v", method, path, h.Elapsed())
}

// ==========================================================================
// SECTION 2 — Interface embedding: composing small contracts
// ==========================================================================

// Each interface has one job — maximum reusability.
type Getter interface {
	Get(id string) (string, bool)
}
type Setter interface {
	Set(id, value string)
}
type Deleter interface {
	Delete(id string)
}

// Compose bigger contracts from small ones.
type Store interface {
	Getter
	Setter
	Deleter
}

// InMemoryStore implements Store.
type InMemoryStore struct {
	data map[string]string
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{data: make(map[string]string)}
}

func (s *InMemoryStore) Get(id string) (string, bool) {
	v, ok := s.data[id]
	return v, ok
}
func (s *InMemoryStore) Set(id, value string) { s.data[id] = value }
func (s *InMemoryStore) Delete(id string)     { delete(s.data, id) }

// ==========================================================================
// SECTION 3 — Dependency Injection: constructor injection pattern
// ==========================================================================

// Define interfaces in the CONSUMER (service) layer.
// UserStore is defined here, not in any "storage" package.
type UserStore interface {
	GetUser(id string) (User, error)
	SaveUser(u User) error
}

type Emailer interface {
	SendEmail(to, subject, body string) error
}

type User struct {
	ID    string
	Name  string
	Email string
}

// UserService: business logic. Depends ONLY on interfaces.
// No concrete types, no direct DB/email imports.
type UserService struct {
	store   UserStore
	emailer Emailer
	logger  *Logger
}

// NewUserService: constructor injection. Takes interfaces, returns concrete struct.
func NewUserService(store UserStore, emailer Emailer, logger *Logger) *UserService {
	return &UserService{store: store, emailer: emailer, logger: logger}
}

func (s *UserService) Register(id, name, email string) error {
	u := User{ID: id, Name: name, Email: email}
	if err := s.store.SaveUser(u); err != nil {
		return fmt.Errorf("register user %s: %w", id, err)
	}
	if err := s.emailer.SendEmail(email, "Welcome!", "Thanks for joining, "+name); err != nil {
		// Non-fatal: log it but don't fail registration
		s.logger.Logf("warning: welcome email failed for %s: %v", id, err)
	}
	s.logger.Logf("registered user %s (%s)", id, name)
	return nil
}

func (s *UserService) GetUser(id string) (User, error) {
	u, err := s.store.GetUser(id)
	if err != nil {
		return User{}, fmt.Errorf("get user %s: %w", id, err)
	}
	return u, nil
}

// ==========================================================================
// SECTION 4 — Fake/stub implementations (what tests and main() use)
// ==========================================================================

// FakeUserStore: in-memory implementation for development/testing.
type FakeUserStore struct {
	users map[string]User
}

func NewFakeUserStore() *FakeUserStore {
	return &FakeUserStore{users: make(map[string]User)}
}

func (f *FakeUserStore) GetUser(id string) (User, error) {
	u, ok := f.users[id]
	if !ok {
		return User{}, fmt.Errorf("user %s not found", id)
	}
	return u, nil
}

func (f *FakeUserStore) SaveUser(u User) error {
	f.users[u.ID] = u
	return nil
}

// FakeEmailer: captures sent emails (great for assertions in tests).
type FakeEmailer struct {
	Sent []struct{ To, Subject, Body string }
}

func (f *FakeEmailer) SendEmail(to, subject, body string) error {
	f.Sent = append(f.Sent, struct{ To, Subject, Body string }{to, subject, body})
	fmt.Printf("  [EMAIL] to=%s subject=%q\n", to, subject)
	return nil
}

// ==========================================================================
// SECTION 5 — Functional Options: clean constructors for many options
// ==========================================================================

type ClientConfig struct {
	baseURL    string
	timeout    time.Duration
	retries    int
	userAgent  string
	apiKey     string
}

// Option is a function that modifies a ClientConfig.
// This is the functional options pattern.
type Option func(*ClientConfig)

func WithBaseURL(url string) Option {
	return func(c *ClientConfig) { c.baseURL = url }
}
func WithTimeout(d time.Duration) Option {
	return func(c *ClientConfig) { c.timeout = d }
}
func WithRetries(n int) Option {
	return func(c *ClientConfig) { c.retries = n }
}
func WithUserAgent(ua string) Option {
	return func(c *ClientConfig) { c.userAgent = ua }
}
func WithAPIKey(key string) Option {
	return func(c *ClientConfig) { c.apiKey = key }
}

type APIClient struct {
	cfg ClientConfig
}

// NewAPIClient: sensible defaults, overridden by options.
func NewAPIClient(opts ...Option) *APIClient {
	cfg := ClientConfig{
		baseURL:   "https://api.example.com",
		timeout:   30 * time.Second,
		retries:   3,
		userAgent: "go-client/1.0",
	}
	for _, o := range opts {
		o(&cfg)
	}
	return &APIClient{cfg: cfg}
}

func (c *APIClient) String() string {
	key := "none"
	if c.cfg.apiKey != "" {
		key = c.cfg.apiKey[:4] + "****"
	}
	return fmt.Sprintf("APIClient{url=%s timeout=%v retries=%d ua=%q key=%s}",
		c.cfg.baseURL, c.cfg.timeout, c.cfg.retries, c.cfg.userAgent, key)
}

// ==========================================================================
// SECTION 6 — Decorator via interface embedding
// ==========================================================================

// LoggingStore wraps any Store and logs every operation.
// Embeds the Store interface — delegates unknown methods to it automatically.
type LoggingStore struct {
	Store          // embedded interface — delegates all methods
	logger *Logger
}

func NewLoggingStore(s Store, logger *Logger) *LoggingStore {
	return &LoggingStore{Store: s, logger: logger}
}

// Override only Get to add logging; Set and Delete are delegated automatically.
func (l *LoggingStore) Get(id string) (string, bool) {
	v, ok := l.Store.Get(id)
	l.logger.Logf("Get(%q) → (%q, %v)", id, v, ok)
	return v, ok
}

func (l *LoggingStore) Set(id, value string) {
	l.Store.Set(id, value)
	l.logger.Logf("Set(%q, %q)", id, value)
}

// ==========================================================================
// main — the "composition root": only place that knows about concrete types
// ==========================================================================

func main() {
	// --- Section 1: Struct embedding ---
	fmt.Println("== 1. Struct Embedding ==")
	h := NewHTTPHandler("/api/users")
	h.Handle("GET", "/api/users/42")

	// --- Section 2: Interface composition ---
	fmt.Println("\n== 2. Interface Composition ==")
	store := NewInMemoryStore()
	store.Set("k1", "hello")
	store.Set("k2", "world")
	if v, ok := store.Get("k1"); ok {
		fmt.Println("  Get k1:", v)
	}
	store.Delete("k1")
	if _, ok := store.Get("k1"); !ok {
		fmt.Println("  k1 deleted")
	}

	// --- Section 3 & 4: DI with fakes ---
	fmt.Println("\n== 3. Dependency Injection ==")
	fakeStore := NewFakeUserStore()
	fakeEmailer := &FakeEmailer{}
	logger := NewLogger("SVC")

	svc := NewUserService(fakeStore, fakeEmailer, logger)

	if err := svc.Register("u1", "Alice", "alice@example.com"); err != nil {
		fmt.Println("  error:", err)
	}

	u, err := svc.GetUser("u1")
	if err != nil {
		fmt.Println("  error:", err)
	} else {
		fmt.Printf("  retrieved: %+v\n", u)
	}

	fmt.Printf("  emails sent: %d\n", len(fakeEmailer.Sent))

	_, err = svc.GetUser("u999")
	fmt.Printf("  get u999: err=%v\n", err)

	// --- Section 5: Functional options ---
	fmt.Println("\n== 4. Functional Options ==")

	// Default client — no options
	c1 := NewAPIClient()
	fmt.Println(" ", c1)

	// Customised client — only override what differs
	c2 := NewAPIClient(
		WithBaseURL("https://api.staging.example.com"),
		WithTimeout(10*time.Second),
		WithAPIKey("sk_live_abcdef12345"),
	)
	fmt.Println(" ", c2)

	// --- Section 6: Decorator ---
	fmt.Println("\n== 5. Logging Decorator ==")
	baseStore := NewInMemoryStore()
	loggedStore := NewLoggingStore(baseStore, NewLogger("STORE"))
	loggedStore.Set("session:abc", "user_id=42")
	val, _ := loggedStore.Get("session:abc")
	fmt.Println("  value:", val)

	// --- Verify interface satisfaction at compile time ---
	var _ UserStore = (*FakeUserStore)(nil)
	var _ Emailer = (*FakeEmailer)(nil)
	var _ Store = (*InMemoryStore)(nil)
	var _ Store = (*LoggingStore)(nil)
	_ = strings.ToUpper // avoid unused import; strings used elsewhere conceptually
}
