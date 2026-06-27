# Debugging — The Decorator That Forgets to Log

A `LoggingStore` wraps any `Store` and "overrides" `Get` to log every read. It
works for direct `Get` calls — but `GetAll` reads slip through unlogged. This is
the day's signature trap: **Go has no virtual dispatch**, so a promoted method
binds to the embedded type, not your override.

## Symptom

The override on `Get` adds a `[log]` line. Calling `store.Get("1")` logs fine.
But `store.GetAll([...])` returns the right data with **no log lines at all** for
the gets it performs internally. Logging silently disappears on the GetAll path.

## Repro

Bugged (override skipped for GetAll — only 1 log line):

```
cd /Users/ioss/Documents/StudyProjects/GO/day07-composition-di/debugging/bugged && go run .
```

Fixed (every get logged — 3 log lines):

```
cd /Users/ioss/Documents/StudyProjects/GO/day07-composition-di/debugging/fixed && go run .
```

Both compile and run. The bug is a **logic** bug, not a compile error.

## Hint

Where does `GetAll` actually run? `LoggingStore` doesn't declare `GetAll`, so the
one you call is *promoted* from the embedded `Store` interface — meaning it
executes inside `baseStore`, with `baseStore`'s receiver. When `baseStore.GetAll`
calls `b.Get`, which `Get` is that? Trace the receiver, not the wrapper.

<details>
<summary>Solution & why</summary>

**Root cause: Go has no virtual dispatch (no "super", no overriding).**

`LoggingStore` embeds the `Store` interface and defines its own `Get`. That
shadows the promoted `Get` — so `store.Get("1")` runs `LoggingStore.Get` and
logs. Good so far.

But `LoggingStore` does **not** define `GetAll`. So `store.GetAll(...)` resolves
to the `GetAll` *promoted* from the embedded interface, which is
`baseStore.GetAll`. Inside that method the receiver is `*baseStore`, and its body
calls `b.Get(id)`. That selector is statically bound to `baseStore.Get` at
compile time. There is no mechanism for it to "dispatch up" to the
`LoggingStore.Get` override sitting outside it — the embedded type has no
knowledge that it was wrapped. In an inheritance language with virtual methods,
`this.Get` inside `GetAll` would resolve to the most-derived override at runtime.
Go does not do that. Method selection on a concrete receiver is fixed.

Result: the GetAll fan-out runs `baseStore.Get` directly, the override is never
reached, and the log lines vanish.

**The fix:** give the decorator its own `GetAll` that routes each id through the
decorator's `Get`:

```go
func (l *LoggingStore) GetAll(ids []string) ([]string, error) {
    out := make([]string, 0, len(ids))
    for _, id := range ids {
        v, err := l.Get(id) // l.Get => LoggingStore.Get (logged)
        if err != nil {
            return nil, err
        }
        out = append(out, v)
    }
    return out, nil
}
```

Now `GetAll` runs on `*LoggingStore`, and `l.Get` selects the logged override.
All three gets are logged.

**Takeaway for decorators:** if a method on the wrapped type calls *other*
methods of the same type, embedding-and-overriding one method is not enough — the
internal calls won't see your override. Either override every method that
participates in the call chain, or restructure so the cross-method calls go
through the decorator. This is exactly why "favor composition" still requires you
to think about *which receiver* a call binds to.
</details>
