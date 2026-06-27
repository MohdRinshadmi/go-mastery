# Day 26 — Interview Q&A (gRPC + protobuf)

<details>
<summary><strong>1. Difference between Unary and Server Streaming? A real use case for each.</strong></summary>

Unary is one request → one response, identical semantics to a REST call (`GetOrder`, `CreateOrder`). Server streaming is one request → a *stream* of responses over a single call (`ListOrders` that yields results as they're found, log/event tailing, large-result pagination, live price feeds). Server streaming avoids buffering a huge result set in memory and lets the client start consuming immediately.
</details>

<details>
<summary><strong>2. Why does gRPC use HTTP/2? What does multiplexing give you over HTTP/1.1?</strong></summary>

HTTP/2 gives binary framing, header compression (HPACK), and **multiplexing**: many independent streams over one TCP connection. In HTTP/1.1 a connection handles one request at a time, so a slow request causes head-of-line blocking and clients open many connections. HTTP/2 lets thousands of concurrent RPCs share one connection without blocking each other, which is what makes gRPC efficient for high-frequency internal calls. It also enables first-class streaming.
</details>

<details>
<summary><strong>3. Explain protobuf field numbers. Why can you never change one once deployed?</strong></summary>

Each field has a number (the tag) that is its identity on the wire — the binary encoding stores the number, not the field name. Encoders/decoders match fields by number across versions, which is what makes schema evolution work: an old client just skips an unknown number. If you change a field's number (or reuse a freed one), bytes written under the old number get decoded into the wrong field — silent data corruption. To retire a field, mark it `reserved` so the number is never reused.
</details>

<details>
<summary><strong>4. How do you load-balance gRPC in production? Why is L4 round-robin insufficient?</strong></summary>

gRPC runs over one long-lived HTTP/2 connection that multiplexes all calls. An L4 (TCP) load balancer balances *connections*, so it pins every RPC from a client to whichever pod got the connection — no real balancing. You need either **client-side load balancing** (the gRPC client resolves all backends via DNS/service discovery and round-robins requests itself) or an **L7 proxy** (Envoy, Linkerd, nginx grpc_pass) that understands HTTP/2 streams and balances per-request.
</details>

<details>
<summary><strong>5. `codes.Unavailable` vs `codes.Internal` — why does it matter for retries?</strong></summary>

`Unavailable` means "transient, the request didn't succeed but retrying may work" (server restarting, connection dropped). Clients and service meshes automatically retry it. `Internal` means a server-side bug/invariant violation — retrying won't help and may amplify the problem. Returning `Unavailable` for a permanent failure creates retry storms; returning `Internal` for a transient blip means clients give up on a request that would have succeeded. Choosing the code correctly *is* the retry policy.
</details>

<details>
<summary><strong>6. What are gRPC interceptors? Write a unary server interceptor signature.</strong></summary>

Interceptors are gRPC's middleware — they wrap every RPC for cross-cutting concerns (logging, auth, metrics, panic recovery, tracing). Unary server interceptor:

```go
func(
    ctx context.Context,
    req interface{},
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (resp interface{}, err error)
```

You call `handler(ctx, req)` to proceed, wrapping it with your logic before/after. There are matching client interceptors and streaming variants.
</details>

<details>
<summary><strong>7. You need to add a new required field to a deployed proto message. How, safely?</strong></summary>

proto3 has no `required`, and you shouldn't make a field hard-required across a deployed boundary anyway — that's a breaking change (old producers won't set it). Add it as a normal new field with a **new number**, treat its zero value as "not provided", and enforce the requirement in application validation, not the schema. Roll out producers to set it before consumers depend on it. If you need to distinguish "not set" from "set to zero", make it `optional`. Use `buf breaking` in CI to catch anything that actually breaks the wire contract.
</details>

<details>
<summary><strong>8. Why is `optional` (field presence) needed for partial updates?</strong></summary>

Because a proto3 scalar can't distinguish "unset" from "set to zero", an update handler can't tell whether `price == 0` means "set price to 0" or "leave price alone". `optional` generates a pointer (`*float64`); `nil` means unset, so the handler applies only the fields actually provided. Alternatively use a `FieldMask` to name the paths being updated — the canonical Google AIP `Update` pattern.
</details>

<details>
<summary><strong>9. When would you NOT use gRPC?</strong></summary>

Public/external APIs consumed by browsers, mobile apps, or third parties — they expect REST/JSON, and gRPC-Web adds a proxy + complexity. Simple, low-frequency CRUD where REST's tooling (curl, Postman) and debuggability win. Teams without the tooling maturity (protoc/buf in CI, generated code in VCS, a schema registry). gRPC shines for high-frequency, contract-critical *internal* service-to-service calls.
</details>

<details>
<summary><strong>10. How do deadlines propagate in gRPC, and why set them?</strong></summary>

When a client sets a context deadline, gRPC sends it as call metadata; the server's handler `ctx` carries that deadline, and any downstream gRPC calls it makes inherit a child deadline. So a 500ms budget at the edge cascades through the whole chain, and when it expires every level gets a cancellation. Without deadlines a single hung downstream can block goroutines and connections all the way up. Always set a `context.WithTimeout` on client calls and respect `ctx` in handlers.
</details>
