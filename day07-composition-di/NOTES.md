# Day 07 — Notes / Cheatsheet

Composition, embedding, and dependency injection — the wiring patterns every Go
service uses. Quick reference; the lesson has the full story.

---

## Embedding syntax — value vs pointer

```go
type Logger struct{ prefix string }
func (l *Logger) Log(m string) { fmt.Printf("[%s] %s\n", l.prefix, m) }

type ServerVal struct { Logger }   // value embed
type ServerPtr struct { *Logger }  // pointer embed
```

| | Value embed `Logger` | Pointer embed `*Logger` |
|---|---|---|
| Zero value | usable (zero Logger) | **nil → panic on call** |
| Copies | per outer instance | shared underlying |
| Init | automatic | must set in constructor |
| Mutation visibility | local copy | shared across owners |
| Use when | small owned helper | shared / injected dependency |

Always initialize a pointer embed: `&ServerPtr{Logger: NewLogger("srv")}`.

---

## Promotion rules

- Embedding promotes the embedded type's **exported methods and fields** to the
  outer type: `outer.Log(x)` == `outer.Logger.Log(x)`.
- Promotion is syntax sugar — the method runs on the **embedded** receiver.
- A promoted method satisfies an interface for the **outer** type (this is how
  decorators auto-satisfy the wrapped interface).
- Promotion only happens at the **shallowest unambiguous** depth.

## Shadowing rule

- If the outer type declares a field/method with the **same name** as a promoted
  one, the **outer wins** and the inner is hidden (reach it via `outer.Embedded.X`).
- **No virtual dispatch:** an embedded method that calls a sibling method binds to
  the *embedded* receiver — it does **not** see the outer "override." (Day's bug.)

## Ambiguous selector

- Two embedded types at the same depth with the same member name → **compile
  error** on the bare selector. Disambiguate: `outer.A.Close()`.

---

## Interface embedding

```go
type Reader interface { Read(p []byte) (int, error) }
type Writer interface { Write(p []byte) (int, error) }

type ReadWriter interface { // requires BOTH method sets
    Reader
    Writer
}
```

Compose small interfaces; the consumer controls the hierarchy.

---

## Constructor injection skeleton

```go
// interface defined in the CONSUMER package
type OrderStore interface {
    Get(id string) (Order, error)
    Save(o Order) error
}

type OrderService struct {       // depends only on interfaces
    store    OrderStore
    notifier Notifier
}

func NewOrderService(store OrderStore, n Notifier) *OrderService { // accept ifaces...
    return &OrderService{store: store, notifier: n}                // ...return struct
}
```

## Accept interfaces, return structs

> Take an **interface** as a parameter (depend on behavior); return a **concrete
> struct** from constructors (callers widen for free). Interface returns only for
> real multi-impl factories.

---

## Functional options skeleton

```go
type config struct {
    addr    string
    timeout time.Duration
}
type Option func(*config)

func WithAddr(a string) Option    { return func(c *config) { c.addr = a } }
func WithTimeout(d time.Duration) Option { return func(c *config) { c.timeout = d } }

func NewServer(opts ...Option) *Server {
    c := &config{addr: ":8080", timeout: 30 * time.Second} // defaults
    for _, o := range opts {
        o(c)
    }
    return &Server{cfg: c}
}

// s := NewServer(WithAddr(":9090"), WithTimeout(10*time.Second))
```

Use for **optional, growing** config. Plain constructor for required params.

---

## Decorator skeleton (embed the interface)

```go
type CachedStore struct {
    Store                  // embed the INTERFACE, not a concrete type
    cache map[string]User
}

func (c *CachedStore) Get(id string) (User, error) {
    if u, ok := c.cache[id]; ok { return u, nil }
    u, err := c.Store.Get(id)  // delegate to wrapped impl
    if err == nil { c.cache[id] = u }
    return u, err
}
// CachedStore satisfies Store (Get overridden, rest promoted) -> decorators stack.
// Trap: override every method whose call path must be enhanced (no virtual dispatch).
```

---

## Key terms

- **Embedding** — placing a type in a struct/interface without a field name so its
  members are promoted; delegation, not inheritance.
- **Promotion** — the compiler exposing an embedded type's methods/fields directly
  on the outer type as syntax sugar.
- **Shadowing** — an outer member with the same name as a promoted one hiding the
  inner; combined with no virtual dispatch, the source of override surprises.
- **Composition root** — the single place (`main.go`/`cmd/`) that constructs
  concrete dependencies and wires them; the only layer that knows concretes.
- **DI (dependency injection)** — passing dependencies in (via constructors)
  instead of creating them inside; in Go it's just functions, no framework.
- **Functional options** — `func(*config)` closures for optional, evolving
  configuration with defaults and non-breaking growth.
- **Decorator** — a type that embeds an interface and overrides a method to add
  behavior (log/cache/retry) while delegating the rest; middleware that stacks.
- **Ports-and-adapters** — hexagonal architecture: business logic defines the
  interfaces (ports), infrastructure provides adapters; dependencies point inward.
