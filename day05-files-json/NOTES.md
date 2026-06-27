# Day 05 Notes — Quick Reference

## Reading files
```go
// Whole file (small, bounded)
data, err := os.ReadFile("config.json")  // []byte

// Streaming (big / unbounded), line by line
f, err := os.Open("big.log")
if err != nil { return err }
defer f.Close()
sc := bufio.NewScanner(f)
for sc.Scan() {
    line := sc.Text()      // process one line
}
if err := sc.Err(); err != nil { return err }   // ALWAYS check
```
Default `Scanner` token limit ~64 KB → `sc.Buffer(buf, max)` for huge lines.

## Writing files
```go
// Whole file (0644 = rw-r--r--)
err := os.WriteFile("out.txt", []byte("hello\n"), 0644)

// Streaming / append
f, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
if err != nil { return err }
defer f.Close()
w := bufio.NewWriter(f)
fmt.Fprintln(w, "a line")
w.Flush()                  // MUST flush a bufio.Writer
```

## io.Reader / io.Writer — the universal interfaces
```go
type Reader interface { Read(p []byte) (n int, err error) }
type Writer interface { Write(p []byte) (n int, err error) }

io.Copy(dst, src)          // stream any Reader -> any Writer
io.ReadAll(r)              // drain a Reader to []byte
strings.NewReader("...")   // a Reader for tests (no disk)
```
**Accept `io.Reader`/`io.Writer` in signatures** — testable & reusable for free.

## JSON encode / decode
```go
type Product struct {
    ID    string   `json:"id"`
    Name  string   `json:"name"`
    Price float64  `json:"price"`
    Tags  []string `json:"tags,omitempty"` // omit when empty
    sku   string                            // unexported -> NEVER serialized
}

b, err := json.Marshal(p)                 // compact
b, err := json.MarshalIndent(p, "", "  ") // pretty

var p Product
err := json.Unmarshal(data, &p)           // note &p (pointer!)
```

## Struct tag rules
| Tag | Effect |
|-----|--------|
| `json:"id"` | rename the wire key |
| `json:"-"` | exclude the field entirely |
| `json:",omitempty"` | omit when zero-valued on encode |
- **Only exported (Capitalized) fields are (de)serialized.**
- Tag format is rigid: `json:"name"` — no space after the colon.

## Streaming JSON (from/to a Reader/Writer)
```go
err := json.NewDecoder(r).Decode(&p)  // r = file, HTTP body, ...
err = json.NewEncoder(w).Encode(p)
```

## Dynamic JSON
```go
var m map[string]interface{}
json.Unmarshal(data, &m)   // numbers -> float64, objects -> maps
```
Prefer a typed struct; reserve this for truly dynamic payloads.

## Present-but-zero vs absent
```go
type T struct{ Count *int `json:"count"` } // nil = absent, &0 = present-zero
```

## Key terms
- **`io.Reader` / `io.Writer`** — one-method byte source / sink interfaces.
- **Buffered writer** — `bufio.Writer`; needs `Flush()`.
- **`bufio.Scanner`** — line/token streaming reader (check `Err()`).
- **Marshal / Unmarshal** — Go ↔ JSON over a `[]byte`.
- **Decoder / Encoder** — Go ↔ JSON streaming over a Reader/Writer.
- **Struct tag** — metadata like `json:"id"` controlling the wire key.
- **`omitempty`** — drop zero-valued fields on encode.
