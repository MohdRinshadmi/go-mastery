# Day 05 Interview Questions — Files, io, JSON

---

**1. When do you use `os.ReadFile` vs `bufio.Scanner`? What goes wrong if you pick wrong?**

<details><summary>Answer</summary>

`os.ReadFile` reads the entire file into one `[]byte` — great for small, bounded
config. `bufio.Scanner` streams line by line at constant memory — for large or
user-sized input. Pick `ReadFile` on an unbounded file and you load gigabytes into
RAM and OOM. Pick `Scanner` and forget its ~64 KB token limit and a huge single
line errors out.
</details>

---

**2. Why design a function to take `io.Reader` instead of a filename or `*os.File`?**

<details><summary>Answer</summary>

Because `io.Reader` is the common interface implemented by files, network
connections, HTTP bodies, `bytes.Buffer`, `strings.Reader`, gzip streams, and
more. A function that accepts `io.Reader` works with all of them, and is trivially
testable with `strings.NewReader(...)` — no disk needed. It's composition over
concretion, the habit that separates idiomatic Go from "ported Python."
</details>

---

**3. Why is my struct field missing from the JSON output? Name three causes.**

<details><summary>Answer</summary>

(1) The field is **unexported** (lowercase) — `encoding/json` can't see it.
(2) The tag says `json:"-"`, which excludes it entirely.
(3) `omitempty` is set and the field holds its zero value, so it's dropped on
encode. (A malformed tag like `json: "name"` with a space is a fourth cause.)
</details>

---

**4. What's the difference between `json.Unmarshal` and `json.NewDecoder(r).Decode`?**

<details><summary>Answer</summary>

`json.Unmarshal(data, &v)` takes a `[]byte` already in memory. `json.NewDecoder(r)`
reads from any `io.Reader` and streams, so you don't buffer the whole input first
— lower peak memory, and it can read one value at a time from a stream. In HTTP
handlers you decode straight from the request body Reader rather than reading it
all into memory.
</details>

---

**5. How do you distinguish "field absent" from "field present but zero" in JSON?**

<details><summary>Answer</summary>

Use a **pointer field** (`*int`, `*string`): after decode, `nil` means the key was
absent, while a non-nil pointer to `0`/`""` means it was present and zero. Plain
value fields can't distinguish the two (both look like the zero value).
`json.RawMessage` or a custom `UnmarshalJSON` are alternatives.
</details>

---

**6. What Go type do JSON numbers become when decoded into `interface{}`?**

<details><summary>Answer</summary>

`float64`. When the target is `interface{}` (e.g. `map[string]interface{}`), every
JSON number is decoded as a `float64`, so asserting directly to `int` panics — you
must assert to `float64` and convert, or decode into a typed struct.
</details>

---

**7. Why must you check `scanner.Err()` after a scan loop?**

<details><summary>Answer</summary>

`scanner.Scan()` returns `false` on **both** normal end-of-input and read errors.
Without checking `scanner.Err()` afterward, you can't tell a clean EOF from a
truncated or failed read, so you'd silently process partial data as if it were
complete.
</details>

---

**8. What are `io.Reader` and `io.Writer`, and why are they so central?**

<details><summary>Answer</summary>

They're one-method interfaces: `Read(p []byte) (int, error)` and
`Write(p []byte) (int, error)`. Almost every byte source/sink in Go implements one
or both, so generic plumbing — `io.Copy`, `io.ReadAll`, `json.NewDecoder` — works
across files, sockets, buffers, and HTTP bodies uniformly. They're the
composition backbone of the whole standard library.
</details>

---

**9. What does `defer w.Flush()` not protect you from on a buffered writer?**

<details><summary>Answer</summary>

`Flush` pushes buffered bytes to the underlying writer, but if `Flush` (or the
later `Close`) returns an error, a deferred `defer w.Flush()` discards it — so you
can think the write succeeded when bytes never reached disk. Capture and check the
error (named return + deferred closure), especially when correctness matters.
</details>

---

**10. When would you reach for `map[string]interface{}` over a typed struct?**

<details><summary>Answer</summary>

Only for **genuinely dynamic** payloads whose shape you don't know at compile time
(arbitrary third-party JSON, passthrough). A typed struct is preferred whenever you
know the schema: it gives compile-time field names, correct types (e.g. `int` not
`float64`), and pushes errors to decode time instead of scattered runtime
assertions.
</details>

---

**11. How do you stream-copy an HTTP download to stdout in one line, and why does it work?**

<details><summary>Answer</summary>

`io.Copy(os.Stdout, resp.Body)`. `resp.Body` is an `io.ReadCloser` and `os.Stdout`
is an `io.Writer`, so `io.Copy` streams bytes between them in fixed-size chunks
without buffering the whole body. It works precisely because both ends satisfy the
`io.Reader`/`io.Writer` interfaces.
</details>
