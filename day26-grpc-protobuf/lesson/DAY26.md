# Day 26 — gRPC + Protocol Buffers

> Mentor note: gRPC is what separates "I build REST APIs" from "I build high-performance distributed systems." Every major cloud company (Google, Netflix, Uber, Square) runs gRPC for internal service-to-service communication. Today you will understand *why*, not just *how*.

---

## 0. What Problem Does gRPC Solve?

You built REST APIs. REST is great for external-facing APIs. But inside your cluster, when Service A calls Service B 10,000 times per second, REST has real costs:

- **JSON parsing is CPU-expensive** — allocating strings for every field, every call.
- **HTTP/1.1 head-of-line blocking** — one request blocks the next on the same connection.
- **No formal contract** — consumers guess the shape of your API from docs that drift from reality.
- **No streaming** — want server-sent updates? You bolt on SSE or WebSockets separately.

gRPC fixes all four simultaneously.

---

## 1. Protocol Buffers (protobuf)

### Theory
Protocol Buffers (proto3) is a language-neutral, binary serialization format. You define your messages and services in a `.proto` file — this is your **contract**. The `protoc` compiler generates client and server stubs in Go (or 10+ other languages) from that single file.

### Why it exists
Google needed a way to version-safe, language-neutral RPC that was faster than XML/JSON. They open-sourced protobuf in 2008. It is now the dominant internal RPC serialization format.

### Binary vs JSON

| Concern | JSON | protobuf |
|---|---|---|
| Size | ~100 bytes for a simple message | ~20 bytes (field IDs not names) |
| Parse speed | Slow (string scanning) | Fast (binary, schema-assisted) |
| Human readable | Yes | No (need tooling) |
| Schema evolution | Optional (fragile) | Built-in (field numbers) |
| Type safety | No | Yes (generated code) |

**Senior take:** protobuf's field numbers are the schema evolution mechanism. Adding a new field with a new number is backward compatible — old clients ignore unknown fields. Removing a field must mark it `reserved` so no one reuses the number. This is discipline REST/JSON cannot enforce.

### A .proto file

```proto
syntax = "proto3";

package order;
option go_package = "day26/examples/pb;pb";

// Every field has a NUMBER (1, 2, 3). These are the wire identifiers.
// NEVER change a field number — that is a breaking change.
message CreateOrderRequest {
    string customer_id = 1;
    repeated OrderItem items = 2;
}

message OrderItem {
    string product_id = 1;
    int32  quantity   = 2;
    double price_usd  = 3;
}

message OrderResponse {
    string order_id = 1;
    string status   = 2;
    double total    = 3;
}

message GetOrderRequest {
    string order_id = 1;
}

// Streaming example request
message ListOrdersRequest {
    string customer_id = 1;
}

service OrderService {
    // Unary — one request, one response
    rpc CreateOrder(CreateOrderRequest) returns (OrderResponse);
    rpc GetOrder(GetOrderRequest)       returns (OrderResponse);

    // Server streaming — one request, stream of responses
    rpc ListOrders(ListOrdersRequest) returns (stream OrderResponse);
}
```

### The protoc workflow (conceptual)
```bash
# Install (if you have it):
brew install protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate Go code from .proto:
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       order.proto

# This produces two files:
#   order.pb.go       — message structs + marshal/unmarshal
#   order_grpc.pb.go  — client stub + server interface
```

**For this lesson, we hand-author these stubs** so the module compiles without protoc installed. The hand-authored versions are semantically identical to what protoc would generate — you can see the pattern. In real projects you commit the generated files or run protoc in CI.

---

## 2. gRPC Architecture

### The four RPC types

```
1. Unary:            client → ONE request  → server → ONE response
2. Server streaming: client → ONE request  → server → STREAM of responses
3. Client streaming: client → STREAM       → server → ONE response
4. Bidirectional:    client → STREAM       ↔ server → STREAM
```

### When to use which

| Type | Use case |
|---|---|
| Unary | 99% of RPCs — same as REST semantics |
| Server streaming | Real-time feeds, pagination over large datasets, log tailing |
| Client streaming | File upload, sensor data ingestion, batch inserts |
| Bidirectional | Chat, multiplayer games, real-time collaboration |

### When NOT to use gRPC
- **Public external APIs** — your mobile app users, external partners, and browser JavaScript clients expect REST/JSON. gRPC-Web exists but adds complexity.
- **Simple CRUD microservices** — if Service A calls Service B once a minute, REST is fine. gRPC shines at high-frequency internal calls.
- **Teams without tooling maturity** — protoc in CI, generated code in version control, consistent proto schema registry — all require investment.

---

## 3. gRPC vs REST — the Real Tradeoffs

| | REST/JSON | gRPC/protobuf |
|---|---|---|
| Latency | Higher (JSON parse, HTTP/1.1) | Lower (binary, HTTP/2 multiplexing) |
| Throughput | Good | 5-10x better in benchmarks |
| Developer experience | Excellent (curl, Postman) | Good (grpcurl, Postman gRPC) |
| Browser native | Yes | No (gRPC-Web proxy needed) |
| Schema enforcement | Optional (OpenAPI) | Mandatory (proto file) |
| Code generation | Optional | Required (but saves time) |
| Streaming | Bolt-on | First-class |
| Debugging | Easy (human-readable) | Harder (need proto to decode) |
| Load balancing | Simple (HTTP/1.1 L7) | Needs HTTP/2-aware LB |

**Staff engineer take:** In a new microservices system, use gRPC for all inter-service communication and REST for the public gateway. The proto files become your service contracts — they live in a shared `api/` repo and go through code review. Any breaking change to a proto is a versioning event.

**Architect take:** "gRPC everywhere" is not the goal. The goal is *clear contracts*. Proto gives you that. So does OpenAPI. The binary efficiency of gRPC matters at scale (Uber saved ~40% CPU just by switching from Thrift to protobuf). Before that scale, choose the format your team can debug fastest.

---

## 4. gRPC in Go — Key Packages

```go
import (
    "google.golang.org/grpc"                 // core framework
    "google.golang.org/grpc/codes"           // status codes (OK, NotFound, etc.)
    "google.golang.org/grpc/status"          // wrapping errors with gRPC status
    "google.golang.org/grpc/credentials/insecure" // skip TLS for local dev
    "google.golang.org/grpc/metadata"        // request headers (auth, tracing)
)
```

### Server setup pattern
```go
lis, err := net.Listen("tcp", ":50051")
srv := grpc.NewServer(
    grpc.UnaryInterceptor(loggingInterceptor),   // middleware
)
pb.RegisterOrderServiceServer(srv, &orderServer{})
reflection.Register(srv)  // enables grpcurl without proto file
srv.Serve(lis)
```

### Client setup pattern
```go
conn, err := grpc.NewClient(
    "localhost:50051",
    grpc.WithTransportCredentials(insecure.NewCredentials()),
)
client := pb.NewOrderServiceClient(conn)
defer conn.Close()

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

resp, err := client.CreateOrder(ctx, &pb.CreateOrderRequest{...})
```

---

## 5. Error Handling in gRPC

gRPC has its own status codes (not HTTP status codes). **Always use `status.Error`, never raw errors.**

```go
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

func (s *server) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.OrderResponse, error) {
    order, found := s.store[req.OrderId]
    if !found {
        // Return gRPC-native error — client sees codes.NotFound
        return nil, status.Errorf(codes.NotFound, "order %q not found", req.OrderId)
    }
    return order, nil
}

// On the client side:
resp, err := client.GetOrder(ctx, req)
if err != nil {
    st, ok := status.FromError(err)
    if ok && st.Code() == codes.NotFound {
        // handle not found specifically
    }
}
```

### gRPC status codes you'll use most

| Code | Meaning | HTTP equiv |
|---|---|---|
| `OK` | Success | 200 |
| `InvalidArgument` | Bad input | 400 |
| `NotFound` | Resource missing | 404 |
| `AlreadyExists` | Conflict | 409 |
| `PermissionDenied` | Auth failed | 403 |
| `Unavailable` | Temporary failure (retry) | 503 |
| `DeadlineExceeded` | Timeout | 504 |
| `Internal` | Server bug | 500 |

**Senior take:** `Unavailable` is the key retry signal. Clients (and service meshes like Istio) automatically retry on `Unavailable`. Never return `Unavailable` for a permanent error — you'll create infinite retry storms.

---

## 6. Interceptors (gRPC Middleware)

gRPC's equivalent of HTTP middleware. Applied at the server and client level.

```go
// Server-side unary interceptor: add logging + panic recovery
func loggingInterceptor(
    ctx context.Context,
    req interface{},
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (interface{}, error) {
    start := time.Now()
    resp, err := handler(ctx, req)
    log.Printf("RPC: %s | duration: %v | err: %v", info.FullMethod, time.Since(start), err)
    return resp, err
}
```

In production, use `go.uber.org/zap` and the `grpc-ecosystem/go-grpc-middleware` package which provides:
- Authentication interceptors
- Request validation
- Rate limiting
- Prometheus metrics
- Panic recovery
- Distributed tracing (OpenTelemetry)

---

## 7. Production Concerns

### TLS / mTLS
Never run gRPC without TLS in production. For internal services, use **mTLS** (mutual TLS) where both client and server verify each other's certificates. Service meshes like Istio/Linkerd can handle this transparently.

### Load balancing
HTTP/2 multiplexes all RPCs on one TCP connection — your L4 load balancer sees one connection and sends everything to one pod. Use **client-side load balancing** (with a service discovery backend like Consul/DNS) or an **L7 proxy** (Envoy, nginx with grpc_pass) that understands HTTP/2 streams.

### Deadlines — always set them
```go
// BAD: no timeout — a slow downstream hangs this service forever
ctx := context.Background()
resp, _ := client.CreateOrder(ctx, req)

// GOOD: cascading deadline — child cannot outlive parent
ctx, cancel := context.WithTimeout(parent, 500*time.Millisecond)
defer cancel()
```
gRPC propagates deadlines across the call chain. If the parent times out, all children get cancellation signals.

### Proto schema registry
At scale (50+ services), store your `.proto` files in a dedicated `api/` repository. Use tools like **Buf** (buf.build) for linting, breaking-change detection, and schema registry. `buf lint` and `buf breaking` run in CI and prevent accidental contract breaks.

### Schema evolution rules
1. You can: add new fields, add new RPCs, add new enum values (with caution).
2. You cannot: change field numbers, change field types, remove or rename fields without `reserved`.
3. Use `reserved 5;` and `reserved "old_field_name";` when removing fields.

---

## 8. Project Architecture — OrderService

```
┌─────────────────────────────────────────────────────┐
│                  Client (cmd/client)                 │
│  - Demonstrates Unary + Server Streaming calls       │
│  - Shows deadline propagation                        │
│  - Shows error code handling                         │
└──────────────────────┬──────────────────────────────┘
                       │  gRPC / HTTP2 / protobuf wire
                       ▼
┌─────────────────────────────────────────────────────┐
│               Server (cmd/server)                    │
│  - Logging interceptor                               │
│  - In-memory order store (sync.RWMutex)              │
│  - Unary: CreateOrder, GetOrder                      │
│  - Server streaming: ListOrders                      │
└─────────────────────────────────────────────────────┘
```

**Scalability note:** The in-memory store is single-node. At scale you replace it with Redis (Day 28) or a distributed DB. The gRPC layer itself scales horizontally — add more server pods behind an L7 LB.

**Tradeoffs:** Proto adds a codegen step (managed complexity). In exchange you get compile-time contract checking — a mismatch between client and server is a build error, not a runtime surprise.

---

## Common mistakes

1. Treating a proto3 scalar's zero value as "unset" — partial-update handlers then clobber fields the client never sent. Use `optional`, wrapper types, or a `FieldMask`.
2. Changing or reusing a field number to "tidy up" the proto — silent wire corruption for any peer on the old schema. Numbers are forever; `reserved` retired ones.
3. Returning raw `errors.New` from handlers — clients see `codes.Unknown` and can't branch. Always `status.Errorf(codes.X, ...)`.
4. No deadline on client calls (`context.Background()`) — a hung downstream blocks the caller forever. Set `context.WithTimeout`; deadlines propagate down the chain.
5. Load-balancing gRPC at L4 — HTTP/2 multiplexes everything onto one connection, pinning all RPCs to one pod. Use client-side LB or an L7/HTTP-2-aware proxy.
6. Returning `codes.Unavailable` for a permanent error — meshes auto-retry it, creating a retry storm. Reserve `Unavailable` for genuinely transient failures.

---

## Expert Thinking Mode

- **Beginner:** "gRPC is like REST but binary and faster."
- **Senior:** "gRPC is a contract-first RPC framework. The proto file IS the API. I version it like I version code."
- **Staff:** "I choose gRPC for inter-service calls because the performance headroom matters at our call volumes, and because protobuf's schema evolution discipline prevents the API drift that plagues JSON services. I choose REST at the edge because that's what browsers and partners speak."
- **Architect:** "gRPC + a proto schema registry is a distributed type system. When 80 services all import `order.proto` v2, I have a single source of truth for the OrderResponse shape. That's worth more than the binary efficiency — it eliminates entire classes of runtime failures."

---

## Real-world Use

- **Google:** All internal services (Search, Ads, Gmail) use protobuf/gRPC. gRPC was born at Google as "Stubby."
- **Netflix:** Uses gRPC for streaming platform internals; serves millions of streams via gRPC server streaming.
- **Uber:** Switched from Thrift to protobuf; claimed ~40% CPU reduction on serialization-heavy services.
- **Square/Cash App:** Proto as source of truth for all payment service contracts; any API change requires a proto PR review.
- **Buf (buf.build):** The modern protobuf toolchain replacing raw protoc in CI.

---

## Interview Questions

1. What is the difference between gRPC Unary and Server Streaming? Give a real use case for each.
2. Why does gRPC use HTTP/2? What does multiplexing give you that HTTP/1.1 doesn't?
3. Explain protobuf field numbers. Why can you never change a field number once deployed?
4. How do you load-balance gRPC in production? Why is a simple round-robin L4 LB insufficient?
5. What's the difference between `codes.Unavailable` and `codes.Internal` and why does it matter for retry behavior?
6. What are gRPC interceptors? Write the signature of a unary server interceptor.
7. You want to add a new required field to an existing proto message that's already in production. How do you do it safely?

---

## Your Tasks for Today

Go to `../exercises/`. You will:
1. Implement the missing handler methods on the OrderService server.
2. Write a client that calls all three RPCs and prints results.
3. Challenge: add a `CancelOrder` unary RPC with proper `NotFound` / `AlreadyShipped` error handling.

See `../examples/` for a working reference implementation first — run it to understand the shape, then close it and implement exercises yourself.

---

## Day 26 companion files

Self-study companions for this day (in `../`):

- [`debugging/`](../debugging/) — the partial-update-clobbers-fields bug (proto3 zero-value vs unset) with `bugged/` and `fixed/`.
- [`PITFALLS.md`](../PITFALLS.md) — gRPC/protobuf gotchas as Trap → Why → Fix.
- [`INTERVIEW.md`](../INTERVIEW.md) — interview questions with model answers.
- [`NOTES.md`](../NOTES.md) — quick reference + key terms.
- [`RESOURCES.md`](../RESOURCES.md) — curated links (grpc.io, protobuf.dev, Buf).
