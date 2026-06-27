# Day 07 — Pitfalls (Trap → Why → Fix)

The day's lesson already lists a **Common mistakes** section. This file goes
deeper on the traps that actually bite in code review and production. Each is
**Trap → Why → Fix**.

---

## 1. Embedding is not inheritance

**Trap:** You write `type Dog struct { Animal }` expecting `Dog` to *be an*
`Animal`, then try to pass a `Dog` to a function taking `Animal`.

**Why:** Embedding promotes the embedded type's methods and fields to the outer
type as syntax sugar — `dog.Breathe()` is rewritten to `dog.Animal.Breathe()`.
It creates **no "is-a" relationship**. `Dog` and `Animal` are distinct types;
there is no subtyping, no `super`, no upcast. The only "is-a" in Go is *interface
satisfaction*.

**Fix:** Embed to *delegate a responsibility* (logging, locking, timing), not to
model a taxonomy. If you need polymorphism, define an interface and let both
types satisfy it.

---

## 2. Method/field shadowing & no virtual dispatch

**Trap:** You "override" a method on an embedded type, but other methods of that
embedded type still call the original — your override is silently bypassed. (This
is the debugging exercise.)

**Why:** When the outer struct declares a member with the same name as a promoted
one, the outer **shadows** the inner — `s.Name` reads the outer field. But Go has
**no virtual dispatch**: a method on the embedded type that calls a sibling
method binds that call to the *embedded* receiver at compile time. Your outer
override is invisible to it.

**Fix:** Don't rely on overriding to retrofit behavior into the embedded type's
internal call chains. Override every method that participates, or restructure so
cross-method calls flow through the wrapper. See `debugging/`.

---

## 3. Nil embedded pointer panic

**Trap:** `type Server struct { *Logger }` then `srv.Log("up")` panics with a nil
pointer dereference.

**Why:** Embedding a *pointer* leaves it as a nil `*Logger` in the zero value.
The promoted call `srv.Log(...)` is `srv.Logger.Log(...)`, which dereferences a
nil pointer the moment the method touches the receiver.

**Fix:** Either embed by **value** (`type Server struct { Logger }`) when the
helper is owned, or **always set the pointer in your constructor**
(`&Server{Logger: NewLogger(...)}`). Never hand out a struct with an
uninitialized embedded pointer.

---

## 4. Returning an interface from a constructor

**Trap:** `func NewUserService(...) UserServicer` — returning an interface "to be
flexible."

**Why:** It locks your public API to that contract. Every new method must be
added to the interface (a breaking change for implementers), callers can't reach
struct-only helpers, and you've added abstraction with no caller asking for it.
It also obscures what the function actually returns.

**Fix:** **Accept interfaces, return structs.** Return the concrete `*UserService`.
Callers widen a struct to whatever interface they need for free; narrowing an
interface back to a struct needs a type assertion. Default to concrete; return an
interface only for a genuine factory with multiple implementations.

---

## 5. Creating a dependency inside business logic

**Trap:** A service method calls `sql.Open(...)`, `http.DefaultClient.Do(...)`, or
`time.Now()` directly.

**Why:** The dependency is now hard-coded. You can't substitute a fake in tests,
you can't swap the implementation, and the method secretly reaches into global
state. Tests become integration tests that need a real DB or network.

**Fix:** **Inject** every dependency through the constructor — store, HTTP client,
clock — as an interface field. Business logic operates only on what it was given.
The composition root (`main.go`) is the one place that constructs concretes.

---

## 6. Ambiguous embedding (same method from two embedded types)

**Trap:** You embed two types that each have a `Close()` method, then call
`x.Close()` — **compile error: ambiguous selector x.Close**.

**Why:** Promotion only happens when the name is unambiguous at the *shallowest*
depth. Two embedded types at the same depth both offering `Close` means Go can't
pick one, so it refuses to promote either. (The same diamond can resolve
silently if one path is shallower — a different hazard.)

**Fix:** Disambiguate explicitly: `x.A.Close()` or `x.B.Close()`. Or declare an
outer `Close()` that decides what the combined type should do. Don't leave the
caller guessing.

---

## 7. Interface defined in the producer package → import cycle

**Trap:** You define `OrderStore` in the `store` package (where the concrete lives)
and import it from `service`. Later `store` needs a type from `service`, and the
build fails with an **import cycle**.

**Why:** Putting the interface next to its implementation forces every consumer to
import the producer package, coupling layers in the wrong direction and creating
cycles the moment the dependency is bidirectional.

**Fix:** Define the interface in the **consumer** package (`service`) — it owns the
contract it needs. The concrete type in `store` satisfies it implicitly (Go
interfaces are structural). Dependencies now flow one way: infrastructure depends
on the abstractions the business layer declares. That's ports-and-adapters.
