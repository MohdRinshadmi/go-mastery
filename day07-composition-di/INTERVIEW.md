# Day 07 — Interview Questions

The first 7 are the lesson's questions; the last 3 push deeper. Answers are how a
senior would actually respond — short, opinionated, with the "why."

---

### 1. Explain struct embedding in Go. Is it inheritance? What's the difference?

<details>
<summary>Answer</summary>

Embedding puts a type inside a struct without a field name; its exported methods
and fields get **promoted** to the outer type as syntax sugar. `outer.Method()`
is rewritten to `outer.Embedded.Method()`.

It is **not** inheritance. There's no "is-a" relationship, no subtyping, no
`super`, and no virtual dispatch. `Dog{Animal}` does not let you pass a `Dog`
where an `Animal` is expected — they're distinct types. Promoted methods still
run on the *embedded* receiver, so an embedded type's internal calls never see an
outer "override." The mental model is **delegation with promoted syntax**, not a
class hierarchy. The only polymorphism in Go is interface satisfaction.
</details>

---

### 2. What does "accept interfaces, return structs" mean? Why?

<details>
<summary>Answer</summary>

**Accept interfaces** in parameters: depend on the behavior you need, not a
concrete type. That makes functions testable (pass a fake) and swappable (pass a
different impl).

**Return structs** from constructors: a caller can always widen a concrete struct
to whatever interface it needs — that direction is free. Returning an interface
instead locks your API to that contract, forces every new method through it
(breaking implementers), and hides struct-only helpers. So default to concrete
return types; return an interface only for a real factory with multiple
implementations.
</details>

---

### 3. How would you test a service that depends on a database, without hitting an actual DB?

<details>
<summary>Answer</summary>

Define a narrow interface in the service package for exactly the DB operations
the service uses (`Get`, `Save`), inject it through the constructor, and in tests
pass an in-memory fake that satisfies it. The service never imports the real
store; `main.go` wires the Postgres implementation in production and the test
wires the fake. No mocking framework needed — a struct with a map is usually
enough. Keep the interface small (defined by the consumer) so the fake is trivial.
</details>

---

### 4. What is the "composition root" in a Go application?

<details>
<summary>Answer</summary>

The single place — usually `main.go` / `cmd/` — where you construct all the
concrete dependencies and wire them together. It's the *only* layer that knows
about concrete types (Postgres, SMTP, Redis). Everything below it operates on
interfaces. Because all construction is centralized, swapping an implementation
is a one-line change at the root, and the dependency graph (which *is* your
architecture) is visible in one file. Keep business logic out of it; it just
assembles.
</details>

---

### 5. What problem does the functional options pattern solve? When would you use it?

<details>
<summary>Answer</summary>

It handles constructors with many **optional** parameters that will grow over
time. A `Config` struct passed by value can't tell "zero means default" from
"zero means off," and adding a field breaks positional callers. Functional
options (`WithTimeout(d) Option`) let you set sensible defaults, override only
what you need, self-document at the call site, and add new options without
breaking anyone. Use it for genuinely optional, evolving config (servers,
clients). Don't use it for 1-2 required params — that's just a plain constructor.
</details>

---

### 6. A `CachedStore` wraps any `Store` (interface). How do you implement this without generics?

<details>
<summary>Answer</summary>

Embed the **interface** (not a concrete type) in the decorator and override the
method you want to enhance, delegating to the embedded interface for the real
work:

```go
type CachedStore struct {
    Store                  // embedded interface
    cache map[string]User
}
func (c *CachedStore) Get(id string) (User, error) {
    if u, ok := c.cache[id]; ok { return u, nil }
    u, err := c.Store.Get(id)  // delegate
    if err == nil { c.cache[id] = u }
    return u, err
}
```

Because the embedded field is the interface, `CachedStore` wraps *any*
implementation and itself satisfies `Store`, so decorators stack. Watch the
no-virtual-dispatch trap: any other `Store` method that internally calls `Get`
won't hit your cache unless you override it too.
</details>

---

### 7. How do you prevent circular imports in a layered Go architecture?

<details>
<summary>Answer</summary>

Define interfaces in the **consumer** package, not the producer. The `service`
layer declares the `OrderStore` interface it needs; the `store` package's
concrete type satisfies it implicitly (structural typing), and `service` never
imports `store`. Dependencies then flow one direction — infrastructure depends on
the abstractions business logic declares. The composition root imports everything
and wires it. This is ports-and-adapters, and it makes cycles structurally
impossible.
</details>

---

### 8. Embedding by value vs by pointer — when do you use each?

<details>
<summary>Answer</summary>

Embed by **value** when the embedded type is a small, owned helper with no shared
state — it's initialized to its zero value automatically, each outer instance
gets its own copy, and there's no nil risk. Embed a **pointer** when the embedded
type is a shared or injected dependency, is expensive to copy, or carries state
that multiple owners must see (e.g. a `*sync.Mutex` or a shared `*Logger`).

The catch with pointers: the zero value is `nil`, so a promoted method call
panics unless you set it in the constructor. Also, embedding `sync.Mutex` by
value is correct (you want a per-struct lock and you never copy the struct after
use); embedding it by pointer is usually a smell. Rule of thumb: value for owned
behavior, pointer for shared/injected dependencies — and always initialize
pointers.
</details>

---

### 9. How does the decorator/middleware pattern work with an embedded interface?

<details>
<summary>Answer</summary>

You declare a struct that **embeds the interface** and stores the wrapped value in
that embedded field. By embedding the interface, the decorator automatically
satisfies it (all methods are promoted from the wrapped value), so you only have
to write the *one* method you want to enhance — log, cache, retry, time — and
delegate the rest. Inside the overridden method you call `d.Inner.Method()` to do
the real work plus your added behavior.

Because the field type is the interface, decorators compose: you can wrap a wrap
of a wrap (`Logging(Caching(Retrying(real)))`). This is how Go middleware chains
work without inheritance. The trap: it only enhances the methods you override —
sibling methods on the inner type that call the enhanced method won't route
through your decorator (no virtual dispatch), so override every method on the call
path that matters.
</details>

---

### 10. What is an ambiguous selector / the "diamond" in embedding?

<details>
<summary>Answer</summary>

If two embedded types at the **same depth** both expose a member with the same
name (say `Close`), the outer type can't promote either — calling `x.Close()` is a
**compile error: ambiguous selector**. Promotion only happens for a name that is
unambiguous at the shallowest depth where it appears.

The subtle "diamond" case: if the same method is reachable via two paths but one
is *shallower*, the shallower one wins silently — which can hide a bug. Resolve
ambiguity explicitly with `x.A.Close()` / `x.B.Close()`, or declare an outer
`Close()` that decides the combined behavior. Go deliberately makes the ambiguous
case a compile error rather than picking arbitrarily — it forces you to be
explicit.
</details>
