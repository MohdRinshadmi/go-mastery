# Debugging Challenge — The Lost Status Code

A handler means to return `404 Not Found`, calls `w.WriteHeader(http.StatusNotFound)`,
and yet the client gets `200 OK`. The code compiles, runs, and serves the right
*body* — but the wrong status line. This is the signature `net/http` gotcha of
Day 16: the **header/body ordering contract**.

## Symptom

`getProduct` for a missing product is supposed to respond `404`. Instead the
recorded status is `200`, and a live server logs `http: superfluous
response.WriteHeader call`. The body text ("not found") is correct, which is
exactly what makes the bug sneaky — it *looks* right in the browser body but
breaks every client that checks the status code.

## Repro

Bugged (wrong output):

```sh
cd /Users/ioss/Documents/StudyProjects/GO/day16-net-http/debugging/bugged
go run .
```

Expected (buggy) output:

```
=== bugged ===
intended status=404, got status=200
```

Fixed (correct output):

```sh
cd /Users/ioss/Documents/StudyProjects/GO/day16-net-http/debugging/fixed
go run .
```

Expected (correct) output:

```
=== fixed ===
intended status=404, got status=404
```

The demonstration is deterministic — no live port. We drive the handler with
`httptest.NewRecorder()` + `httptest.NewRequest(...)` and read back
`rec.Code`. (`fixed` also ships a `main_test.go`; run `go test ./...`.)

## Hint

Read the order of calls inside the `id == "missing"` branch. Which executes
first: the body `Write`/`Fprintf`, or `WriteHeader`? Once *any* byte of body is
written, what status has Go already committed to the wire? The `ResponseWriter`
interface has a strict contract — `WriteHeader` (and `Header().Set`) must come
**before** the first `Write`.

<details>
<summary>Solution & why</summary>

`http.ResponseWriter` writes an HTTP response as a stream: status line, then
headers, then body — in that order, on the wire. The interface enforces this:

- `Header()` returns the header map you may mutate — but only **before** the
  headers are sent.
- `WriteHeader(status)` sends the status line and freezes the headers.
- `Write([]byte)` sends body bytes. **If `WriteHeader` has not been called yet,
  `Write` implicitly calls `WriteHeader(http.StatusOK)` first.**

So the moment the bugged handler does `fmt.Fprintf(w, ...)`, Go commits a
`200 OK` status line and locks the header. The later `w.WriteHeader(404)` has
nothing left to set — the status is already on the wire — so it's ignored and a
live server logs `superfluous response.WriteHeader call`.

```go
// BUG: body first -> implicit 200 commits the status, freezing the header.
func getProduct(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    if id == "missing" {
        fmt.Fprintf(w, "product %q not found\n", id) // commits 200 here
        w.WriteHeader(http.StatusNotFound)           // too late — ignored
        return
    }
    w.Write([]byte("product found\n"))
}
```

The fix is to honor the ordering contract: set any headers, call
`WriteHeader(status)`, *then* write the body.

```go
// FIX: Header().Set -> WriteHeader(status) -> Write(body)
func getProduct(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    if id == "missing" {
        w.Header().Set("Content-Type", "text/plain; charset=utf-8")
        w.WriteHeader(http.StatusNotFound) // status committed first
        fmt.Fprintf(w, "product %q not found\n", id)
        return
    }
    w.Write([]byte("product found\n"))
}
```

**Rules of thumb:**

- The order is always `Header().Set(...)` → `WriteHeader(status)` → `Write(body)`.
  Headers and status must be set before the first body byte.
- The first `Write` implicitly sends `200 OK`. If you want any other status,
  call `WriteHeader` *before* writing.
- Use `http.Error(w, msg, status)` for error responses — it sets the
  Content-Type, calls `WriteHeader`, and writes the body in the correct order
  for you (then `return`).
- Test status codes with `httptest.NewRecorder()` — `rec.Code` makes this class
  of bug deterministic and catchable in CI, not something you eyeball in a
  browser.

</details>
