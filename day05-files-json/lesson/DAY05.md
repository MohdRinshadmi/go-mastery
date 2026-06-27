# Day 05 — File Handling, JSON, and the Phase 1 Capstone

> Mentor note: Today you stop writing toy programs and start writing tools. Reading files and (de)serializing JSON is 80% of real backend glue. The hidden lesson is **`io.Reader` / `io.Writer`** — two tiny interfaces that the entire Go ecosystem is built on. Once they click, files, network sockets, HTTP bodies, gzip streams, and buffers all become the same thing. This is the most important interface preview before Phase 2.

---

## 1. Reading files

### The simple way
```go
data, err := os.ReadFile("config.json")  // []byte, reads whole file into memory
if err != nil {
    return fmt.Errorf("reading config: %w", err)
}
```
Great for small files. **Do not** use it on a 4 GB log — you'll load all of it into RAM.

### The streaming way (big files, line by line)
```go
f, err := os.Open("big.log")
if err != nil {
    return err
}
defer f.Close()

scanner := bufio.NewScanner(f)
for scanner.Scan() {
    line := scanner.Text()
    // process one line; constant memory
}
if err := scanner.Err(); err != nil {   // ALWAYS check this after the loop
    return err
}
```

**Senior take:** Choose based on size and bounds. `os.ReadFile` for small, known-bounded config. `bufio.Scanner` for streams and anything user-sized. A junior loads a 10 GB file with `ReadFile` and OOMs the pod. Default to streaming when size is unbounded.

> Gotcha: `bufio.Scanner` has a default max line size (~64 KB). For huge single lines, use `scanner.Buffer(...)` or `bufio.Reader.ReadString('\n')`.

## 2. Writing files

```go
// Whole file at once (perm 0644 = rw-r--r--):
err := os.WriteFile("out.txt", []byte("hello\n"), 0644)

// Streaming / appending:
f, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
if err != nil { return err }
defer f.Close()
w := bufio.NewWriter(f)
fmt.Fprintln(w, "a log line")
w.Flush()   // buffered! you MUST flush or data stays in the buffer
```

**Common mistake:** forgetting `w.Flush()` on a `bufio.Writer` → your output silently never hits disk. And on writes, a failed `Close()` can mean lost data, so capture it (named return + defer closure) when correctness matters.

## 3. `io.Reader` / `io.Writer` — the two interfaces that rule everything

```go
type Reader interface { Read(p []byte) (n int, err error) }
type Writer interface { Write(p []byte) (n int, err error) }
```

A file, a network connection, an HTTP request body, `os.Stdin`, a `bytes.Buffer`, a gzip stream — **all** implement these. So generic plumbing works on any of them:

```go
io.Copy(dst, src)          // copy any Reader into any Writer, streaming
io.ReadAll(r)              // drain a Reader to []byte
json.NewDecoder(r)         // decode JSON straight from any Reader
```

This is *composition over concretion*: write your function to take `io.Reader`, and it works with files in prod and `strings.NewReader(...)` in tests — no disk needed. We lean on this hard in Phase 2.

**Senior take:** Accept `io.Reader`/`io.Writer` in your function signatures instead of `*os.File` or `string`. It makes code testable and reusable for free. This single habit separates idiomatic Go from "ported Python."

---

## 4. JSON — Marshal / Unmarshal

### Encode (Go → JSON)
```go
type Product struct {
    ID    string  `json:"id"`
    Name  string  `json:"name"`
    Price float64 `json:"price"`
    Tags  []string `json:"tags,omitempty"`  // omitted when empty
    sku   string                            // unexported -> NEVER serialized
}

b, err := json.Marshal(p)            // compact
b, err := json.MarshalIndent(p, "", "  ") // pretty
```

### Decode (JSON → Go)
```go
var p Product
err := json.Unmarshal(data, &p)   // note &p — must pass a pointer
```

### Struct tags — the rules that bite newcomers
- `json:"id"` renames the field.
- `json:"-"` excludes it entirely.
- `json:",omitempty"` omits zero-valued fields on encode.
- **Only exported (Capitalized) fields are (de)serialized.** A lowercase field is invisible to JSON — the #1 "why is my field empty?" bug.
- JSON keys match case-insensitively on decode, but tags make it explicit. Always tag.

### Streaming JSON (decode straight from a Reader)
```go
var p Product
err := json.NewDecoder(r).Decode(&p)   // r is any io.Reader (file, HTTP body...)
// encode straight to a Writer:
err = json.NewEncoder(w).Encode(p)
```
This is what you use in HTTP handlers (Phase 4) — never read the whole body into memory first.

### Unknown / dynamic JSON
```go
var any map[string]interface{}
json.Unmarshal(data, &any)   // numbers become float64, objects become maps
```
Use a typed struct whenever you can — `map[string]interface{}` pushes type errors to runtime. Reserve it for genuinely dynamic payloads.

### Common mistakes
1. Lowercase struct fields → silently not serialized.
2. Forgetting the pointer: `json.Unmarshal(data, p)` (should be `&p`).
3. `omitempty` won't omit an empty struct or a `0` you actually meant — for "present but zero vs absent," use a pointer field (`*int`) or `json.RawMessage`.
4. Decoding numbers expecting `int` from `interface{}` — they're `float64`.

### Performance
- `json.Decoder` streams (lower peak memory) vs `Unmarshal` (whole buffer).
- The standard `encoding/json` is reflection-based and fine for 99% of services. At extreme scale, codegen libs (easyjson, jsoniter) cut CPU — but don't reach for them until a profiler tells you to (Phase 2).

---

## Expert Thinking Mode — "read a file / parse JSON"

- **Beginner:** "`os.ReadFile` then `json.Unmarshal`. Done."
- **Senior:** "Is the size bounded? Stream if not. Take `io.Reader` so it's testable. Tag every field; handle the decode error with context."
- **Staff:** "This parsing sits on a request hot path — decode from the body Reader, cap body size, validate after decode. Schema evolution: will old clients break when I add a field?"
- **Architect:** "Serialization format is a contract across services and time. JSON for external/debuggable, protobuf for internal/perf (Phase 6). Versioning and backward-compat are first-class design concerns."

---

## Real-world use

- **Config:** nearly every Go service loads JSON/YAML config at startup and unmarshals into a typed struct.
- **HTTP APIs:** request/response bodies are JSON decoded/encoded straight from `io.Reader`/`io.Writer` (Phase 4).
- **`io.Reader` everywhere:** AWS SDK, gzip, crypto, HTTP — all compose through these interfaces. `io.Copy(os.Stdout, resp.Body)` streams a download with two words.
- **Log processing:** `bufio.Scanner` over multi-GB logs at constant memory is a daily ops tool.

---

## Interview Questions

1. When do you use `os.ReadFile` vs `bufio.Scanner`? What goes wrong if you pick wrong?
2. Why design a function to take `io.Reader` instead of a filename or `*os.File`?
3. Why is my struct field missing from the JSON output? (Name three causes.)
4. What's the difference between `json.Unmarshal` and `json.NewDecoder(r).Decode`?
5. How do you distinguish "field absent" from "field present but zero" in JSON?
6. What Go type do JSON numbers become when decoded into `interface{}`?
7. Why must you check `scanner.Err()` after a scan loop?

---

## Phase 1 Capstone (in `../exercises/`)

A real CLI tool: **`statsgen`** — reads a JSON file of sales records, computes per-category aggregates, and writes a JSON report. It exercises *everything from Phase 1*: slices, maps, structs, pointers, error wrapping, files, and JSON, plus `io.Reader` design. A sample `data.json` is provided. Finish the TODOs, run it against the sample, and bring it for a full PR review. When you pass this, Phase 1 is done and we move to interfaces.

---

## Day 05 companion files

- [Debugging challenge](../debugging/README.md) — unexported struct fields silently vanish from JSON.
- [Pitfalls](../PITFALLS.md) — Trap → Why it bites → Fix.
- [Interview questions](../INTERVIEW.md) — with model answers.
- [Notes / cheatsheet](../NOTES.md) — quick reference.
- [Resources](../RESOURCES.md) — curated links.
