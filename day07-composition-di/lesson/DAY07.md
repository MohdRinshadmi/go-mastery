# Day 07 — Struct Embedding, Composition, and Dependency Injection

> Mentor note: This is the day your architecture either clicks or breaks. Every senior Go dev has a moment where they realise: "I don't need inheritance, I need composition + interfaces." Today we build that intuition from scratch, and then show you the pattern every production Go service uses for wiring up dependencies — constructor injection.

---

## 0. Why This Day Matters

When Java developers come to Go, they immediately try to recreate class hierarchies with embedding. That's the wrong mental model. Go embedding is not inheritance — it's **delegation with promoted syntax sugar**. Once you internalize that, the whole design space opens up.

Dependency Injection (DI) is the second big concept. In Java, DI typically means a framework (Spring). In Go, DI is just functions: you pass dependencies in through constructors. No reflection, no runtime magic, no annotations. This sounds boring — it's actually liberating.

---

## 1. Struct Embedding — Not Inheritance

### Theory
Embedding a type inside a struct *promotes* its methods and fields to the outer type.

```go
type Logger struct {
    prefix string
}

func (l *Logger) Log(msg string) {
    fmt.Printf("[%s] %s\n", l.prefix, msg)
}

type Server struct {
    *Logger             // embedded pointer — Server "has" Log()
    addr string
}
```

You can call `server.Log("started")` because `Log` is promoted from the embedded `*Logger`. But this is **syntactic sugar**. Internally:
- `server.Log("x")` is exactly `server.Logger.Log("x")`
- `Server` does NOT extend `Logger`. There is no "is-a" relationship.
- A `Server` is not a `Logger`. You cannot pass a `Server` where a `Logger` is expected.
- The promoted method operates on the embedded field, not the outer struct.

### Embedding value vs embedding pointer

| | `type Server struct { Logger }` | `type Server struct { *Logger }` |
|---|---|---|
| Initialization | Logger zero value | Must set to non-nil pointer |
| Mutation | Copy of Logger per Server | Shared Logger pointer |
| Nil risk | None | Panics if nil on method call |
| When to use | Simple, owned behavior | Shared/injectable dependency |

**Embed a value** when the embedded type is a small helper owned by the outer struct.
**Embed a pointer** when the embedded type is a shared dependency or needs to be injected.

### Why NOT to embed for "reuse"

```go
// ANTIPATTERN — don't do this
type Animal struct {
    Name string
}
func (a Animal) Breathe() { fmt.Println(a.Name, "breathes") }

type Dog struct {
    Animal            // attempting "inheritance" for Breathe()
    Breed string
}
```

The problem: every change to `Animal` affects `Dog`. If you add a method to `Animal` that conflicts with an interface, `Dog` breaks. More importantly: `Dog` is not an `Animal` in Go's type system — you can't pass a `Dog` where an `Animal` is expected.

**Correct mental model:** Embed when you want to *delegate* a specific responsibility (logging, timing, locking) to a specialized type. Don't embed to say "Dog is a kind of Animal."

---

## 2. Interface Embedding — Composing Contracts

```go
type Reader interface {
    Read(p []byte) (int, error)
}
type Writer interface {
    Write(p []byte) (int, error)
}
type Closer interface {
    Close() error
}

// Interface embedding: ReadWriter requires BOTH methods.
type ReadWriter interface {
    Reader
    Writer
}

// Can embed as many as needed.
type ReadWriteCloser interface {
    Reader
    Writer
    Closer
}
```

Why this matters: you write small, single-purpose interfaces, then compose them where needed. The concrete types don't change. The interface hierarchy is in the consumer's control.

---

## 3. Dependency Injection in Go

### Theory
DI means: **don't create your dependencies inside a function — accept them as parameters**.

```go
// BAD — hard-coded dependency inside
func ProcessOrder(orderID string) error {
    db := sql.Open("postgres", os.Getenv("DATABASE_URL")) // ← hard to test
    // ...
}

// GOOD — dependency injected, easy to swap in tests
func ProcessOrder(store OrderStore, orderID string) error {
    order, err := store.Get(orderID)
    // ...
}
```

### The Go DI pattern: constructor injection

No frameworks. Just functions:

```go
// 1. Define the interface in the CONSUMER package
type OrderStore interface {
    Get(id string) (Order, error)
    Save(o Order) error
}

// 2. Accept interface, return concrete struct
type OrderService struct {
    store    OrderStore
    notifier Notifier
    logger   *slog.Logger
}

func NewOrderService(store OrderStore, notifier Notifier, logger *slog.Logger) *OrderService {
    return &OrderService{
        store:    store,
        notifier: notifier,
        logger:   logger,
    }
}

// 3. Business logic is pure — no hard-coded dependencies
func (s *OrderService) Process(id string) error {
    order, err := s.store.Get(id)
    if err != nil {
        return fmt.Errorf("process order %s: %w", id, err)
    }
    // ... process ...
    return s.notifier.Send(order.CustomerEmail, "Order confirmed", "...")
}
```

### Accept interfaces, return structs — the Go mantra

- **Accept interfaces** in function parameters: you write against the behavior, not the implementation. This enables testability and swappability.
- **Return concrete types** (structs, not interfaces) from constructors: callers can always widen a concrete type to an interface they need. Going the other direction (narrowing) requires a type assertion. Returning structs keeps code simple and avoids unnecessary abstraction.

```go
// GOOD: returns concrete *UserService
func NewUserService(store UserStore) *UserService { ... }

// LESS GOOD: returns interface — locks the API to this contract,
// harder to add methods without breaking callers
func NewUserService(store UserStore) UserServicer { ... }
```

**Exception:** if you have multiple implementations of the constructor (factory pattern), returning an interface makes sense. But default to concrete.

---

## 4. Clean Architecture Implications

In a real service, dependency injection gives you layers:

```
cmd/main.go          ← wires everything together ("composition root")
    ↓
internal/service/    ← business logic, depends only on interfaces
    ↓
internal/store/      ← concrete implementations (Postgres, in-memory)
internal/notifier/   ← concrete implementations (SMTP, Twilio)
```

`service/` imports only interfaces it defines or stdlib. It never imports `store/` or `notifier/` directly. This means:
- Business logic is testable in isolation with mock implementations.
- You can swap the database without touching service logic.
- Circular imports become impossible.

```
// service/order.go defines:
type OrderStore interface { Get(id string) (Order, error) }

// store/postgres.go implements it:
type PostgresOrderStore struct { db *sql.DB }
func (p *PostgresOrderStore) Get(id string) (Order, error) { ... }

// cmd/main.go wires it:
store := store.NewPostgresOrderStore(db)
svc   := service.NewOrderService(store, notifier)
```

No framework magic. Plain Go. Full testability.

---

## 5. Functional Options Pattern

When constructors get many optional parameters, the idiomatic Go approach is **functional options**:

```go
type ServerConfig struct {
    addr    string
    timeout time.Duration
    maxConn int
}

type Option func(*ServerConfig)

func WithAddr(addr string) Option {
    return func(c *ServerConfig) { c.addr = addr }
}
func WithTimeout(d time.Duration) Option {
    return func(c *ServerConfig) { c.timeout = d }
}
func WithMaxConn(n int) Option {
    return func(c *ServerConfig) { c.maxConn = n }
}

func NewServer(opts ...Option) *Server {
    cfg := &ServerConfig{
        addr:    ":8080",      // sensible defaults
        timeout: 30 * time.Second,
        maxConn: 100,
    }
    for _, o := range opts {
        o(cfg)
    }
    return &Server{cfg: cfg}
}

// Call site: only override what you need
s := NewServer(
    WithAddr(":9090"),
    WithTimeout(10 * time.Second),
)
```

Why this beats a Config struct parameter:
- Zero-value Config doesn't mean "use defaults" — it means all fields zero, which might be invalid.
- Functional options let you add new options without breaking existing callers.
- Options self-document at the call site.

---

## 6. Embedding for Decoration / Middleware

The most useful application of embedding in production is decorating/wrapping:

```go
// Base behavior: concrete implementation
type DBUserStore struct { db *sql.DB }
func (s *DBUserStore) Get(id string) (User, error) { /* SQL query */ }

// Decorator: wraps any UserStore and adds caching
type CachedUserStore struct {
    UserStore              // embed the interface — delegates to whatever is underneath
    cache map[string]User
    mu    sync.RWMutex
}

func (c *CachedUserStore) Get(id string) (User, error) {
    c.mu.RLock()
    if u, ok := c.cache[id]; ok {
        c.mu.RUnlock()
        return u, nil
    }
    c.mu.RUnlock()
    u, err := c.UserStore.Get(id) // delegate to embedded interface
    if err == nil {
        c.mu.Lock()
        c.cache[id] = u
        c.mu.Unlock()
    }
    return u, err
}
```

`CachedUserStore` embeds the `UserStore` *interface* (not a concrete struct). This means it can wrap any implementation: `DBUserStore`, `InMemoryUserStore`, another `CachedUserStore`. This is the decorator pattern in idiomatic Go.

---

## Common mistakes

1. **Treating embedding as inheritance.** `type Dog struct { Animal }` does NOT make `Dog` an `Animal`. You cannot pass a `Dog` where the type system expects an `Animal`, and there is no `super`. Embedding promotes methods; it does not create an "is-a" relationship. Reach for embedding to *delegate* a responsibility, not to model a taxonomy.
2. **Method/field shadowing surprises.** If the outer struct declares a field or method with the same name as a promoted one, the outer wins and the promotion is silently hidden — `s.Name` reads the outer `Name`, never the embedded one. Worse, a promoted method that calls another promoted method calls the *embedded* type's version, not your override (Go has no virtual dispatch). When you "override" a method on an embedded type, the embedded type's other methods still call the original.
3. **Embedding a pointer and forgetting to initialize it.** `type Server struct { *Logger }` leaves `Logger` as a nil `*Logger`. The first promoted `server.Log(...)` call then panics with a nil dereference. Either embed by value or set the pointer in your constructor.
4. **Returning interfaces from constructors by default.** `func NewUserService(...) UserServicer` locks your API to that contract and forces every new method through the interface. Accept interfaces, *return concrete structs* — callers can always widen a struct to whatever interface they need.
5. **Creating dependencies inside business logic.** Calling `sql.Open` / `http.DefaultClient` / `time.Now` directly inside a service method makes it untestable and couples it to one implementation. Inject them through the constructor so tests can pass fakes.
6. **Config struct instead of functional options for optional params.** A `Config{}` passed by value can't distinguish "zero means default" from "zero means off," and adding a field is a breaking change for positional callers. Use functional options for genuinely optional, growing configuration.
7. **Defining the interface in the producer package.** Putting `OrderStore` in the `store` package forces `service` to import `store` and invites import cycles. Define the interface in the *consumer* (`service`) package; the concrete type satisfies it implicitly.

---

## Expert Thinking Mode

- **Beginner:** "I embed types so I don't have to copy-paste methods."
- **Senior:** "I embed small, focused helper types to compose behavior. I use constructor injection to make every service testable without mocks at the binary level."
- **Staff:** "My composition root (`main.go`) is the only place that knows about concrete types. Everything else operates on interfaces. Swapping an implementation is a one-line change in `main.go`."
- **Architect:** "The shape of the dependency graph IS the architecture. Go interfaces let me make the dependency graph flow from business logic toward infrastructure — not the reverse. This is ports-and-adapters (hexagonal architecture) without a framework."

---

## Real-world use

- **Stripe's Go services:** Every external dependency (Stripe API, DB, cache) is behind an interface defined in the service layer. Tests inject in-memory fakes. The service code never changes for new infrastructure.
- **Google's internal services:** Functional options are the standard for configuring gRPC servers, HTTP clients, and service options. See `google.golang.org/grpc` — the entire API is functional options.
- **Uber's Go monorepo:** Clean architecture with `service/`, `store/`, `transport/` packages. Each layer defines the interfaces it needs. Zero circular imports enforced by a custom linter.
- **HashiCorp (Terraform, Vault):** Plugin interfaces — every provider implements a Go interface. The core never imports provider code. Pure interface-based extensibility.

---

## Interview Questions

1. Explain struct embedding in Go. Is it inheritance? What's the difference?
2. What does "accept interfaces, return structs" mean? Why?
3. How would you test a service that depends on a database, without hitting an actual DB?
4. What is the "composition root" in a Go application?
5. What problem does the functional options pattern solve? When would you use it?
6. A `CachedStore` wraps any `Store` (interface). How do you implement this without generics?
7. How do you prevent circular imports in a layered Go architecture?

---

## Your tasks for today

Go to `../exercises/`. Build a payment service with injected storage and notification dependencies. You'll also implement the functional options pattern for a configurable client. Try everything before opening `../solutions/`.

## Day 07 companion files

- [Debugging challenge](../debugging/README.md) — embedding shadowing & no virtual dispatch.
- [Pitfalls](../PITFALLS.md) — Trap → Why → Fix.
- [Interview questions](../INTERVIEW.md) — with model answers.
- [Notes / cheatsheet](../NOTES.md) — quick reference.
- [Resources](../RESOURCES.md) — curated links.
