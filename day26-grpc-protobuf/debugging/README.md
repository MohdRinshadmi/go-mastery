# Day 26 debugging — the partial update that wiped half the record

**Phase 6 · gRPC + protobuf · field presence / schema evolution**

> Stdlib only. The "protobuf message" is a plain Go struct so this builds and runs
> offline — the bug is the proto3 semantic, not the transport.

## Symptom

An `UpdateProduct` gRPC handler is meant to support **partial updates**: a client
that sends only a new `Price` should change the price and leave `Name` and `Stock`
untouched. Instead, after such a call the product's name comes back empty (`""`)
and its stock is `0`. The fields the client *didn't* send got clobbered.

```bash
cd bugged
go run .
```

Expected: only `Price` changes.
Actual: `Name` and `Stock` are silently reset to their zero values.

## Hint

This is the single most famous proto3 footgun. In proto3, a plain scalar field set
to its **zero value** (`""`, `0`, `false`) and a field that was **never set** look
*identical* on the wire — there is no presence bit. So `req.Name == ""` could mean
"the client cleared the name" or "the client didn't touch the name", and the server
can't tell. What does the handler do with every field of the request, set or not?

## How to reproduce

`go run .` in `bugged/`. It seeds one product, sends an update with only `Price`
set, and prints the before/after — you'll see `Name` and `Stock` wiped.

---

<details>
<summary><strong>Solution &amp; why</strong></summary>

### Root cause

The request is a proto3 message of plain scalars, and the handler copies **every**
field onto the stored record:

```go
p.Name = req.Name   // req.Name is "" because the client didn't set it
p.Price = req.Price // the one field the client actually set
p.Stock = req.Stock // req.Stock is 0 because the client didn't set it
```

proto3 dropped the proto2 concept of explicit field presence for scalars. An unset
`string` deserializes to `""`; an unset `int32` to `0`. The server has no way to
know the client meant "leave this alone" — so a partial update becomes a
full-overwrite-with-blanks. This is a correctness bug that ships data loss.

### The fix

Give the fields **presence**. In proto3 that means marking them `optional`, which
makes `protoc` generate **pointer** fields (`*string`, `*float64`, `*int32`); `nil`
now unambiguously means "not set". The handler applies a field only when present:

```go
if req.Name != nil  { p.Name  = *req.Name  }
if req.Price != nil { p.Price = *req.Price }
if req.Stock != nil { p.Stock = *req.Stock }
```

Now a request with only `Price` set leaves `Name` and `Stock` exactly as they were.

Real-world options, all expressing the same idea — *presence*:

> 1. `optional` scalar fields (proto3.15+) → pointer fields; `nil` == unset. Simplest.
> 2. Wrapper types (`google.protobuf.StringValue`, etc.) — the pre-`optional` way.
> 3. A **`FieldMask`** on the request listing exactly which paths to update — the
>    canonical Google AIP pattern for `Update` RPCs, and the most explicit.

Rules of thumb:

> - In proto3, "zero value" never means "unset" for a plain scalar. If you need to
>   distinguish them (partial updates, "was this provided?"), you need `optional`,
>   a wrapper, or a field mask.
> - Adding `optional` to an existing field is **wire-compatible** — the field number
>   is unchanged, so old and new clients interoperate. This is safe schema evolution.

`fixed/` preserves `Name` and `Stock`; the bugged version destroys them.

</details>
