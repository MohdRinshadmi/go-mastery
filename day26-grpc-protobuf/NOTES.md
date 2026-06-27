# Day 26 — Quick Reference (gRPC + protobuf)

## The four RPC types
| Type | Shape | Use case |
|---|---|---|
| Unary | 1 req → 1 resp | 99% of calls; REST-like |
| Server streaming | 1 req → N resp | feeds, log tailing, large results |
| Client streaming | N req → 1 resp | uploads, batch ingest |
| Bidirectional | N ↔ N | chat, games, collaboration |

## protobuf essentials
- `.proto` is the **contract**; `protoc` generates Go structs + stubs.
- Every field has a **number** = wire identity. Never change/reuse it; `reserved` to retire.
- proto3 scalars have **no presence**: zero value == unset. Use `optional` (→ pointer), wrappers, or `FieldMask` when you must tell them apart.
- Safe evolution: add fields, add RPCs, add enum values. Unsafe: change number/type, remove/rename without `reserved`.
- Binary, ~5x smaller and faster to parse than JSON; not human-readable (need tooling/grpcurl).

## Server / client skeleton
```go
// server
lis, _ := net.Listen("tcp", ":50051")
srv := grpc.NewServer(grpc.UnaryInterceptor(logging))
pb.RegisterOrderServiceServer(srv, &orderServer{})
srv.Serve(lis)

// client
conn, _ := grpc.NewClient("localhost:50051",
    grpc.WithTransportCredentials(insecure.NewCredentials()))
defer conn.Close()
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
resp, err := pb.NewOrderServiceClient(conn).CreateOrder(ctx, req)
```

## Errors = status codes (not HTTP)
- Return `status.Errorf(codes.NotFound, ...)`, never raw `errors.New`.
- Client: `st, ok := status.FromError(err); st.Code()`.
- `Unavailable` = retry; `Internal`/`InvalidArgument`/`FailedPrecondition` = don't.

## Production checklist
- TLS/mTLS always; deadlines on every call (they propagate).
- L7 or client-side load balancing (L4 pins to one pod).
- Interceptors for auth/logging/metrics/tracing/recovery.
- `buf lint` + `buf breaking` in CI; protos in a shared `api/` repo.

## Key terms
**Stub** — generated client/server code. **Interceptor** — gRPC middleware.
**Field number** — wire identity of a field. **Field presence / `optional`** — ability to tell "unset" from "zero". **FieldMask** — list of paths to update. **Multiplexing** — many HTTP/2 streams on one TCP connection. **Reserved** — marking a retired field number/name unusable. **Deadline propagation** — child calls inherit the parent's timeout.
