# Day 26 — Pitfalls (gRPC + protobuf)

Format: **Trap → Why → Fix**.

### 1. Proto3 zero value ≠ unset
**Trap:** A plain scalar set to `""`/`0`/`false` is indistinguishable from one that was never set, so partial-update handlers clobber fields the client didn't send.
**Why:** proto3 removed per-scalar field presence; the wire carries no "is set" bit for default-valued scalars.
**Fix:** Use `optional` (generates pointers; `nil` == unset), wrapper types, or a `FieldMask` to express presence. Apply only present fields.

### 2. Changing a field number
**Trap:** Renumbering or reusing a field number to "clean up" the proto silently corrupts data for any peer on the old schema.
**Why:** Field numbers are the wire identity. Old bytes tagged `2` get decoded into whatever your new field `2` is.
**Fix:** Never change a number. To remove a field, `reserved 2;` and `reserved "old_name";` so nobody reuses it.

### 3. Returning raw `errors.New` from a handler
**Trap:** Clients can't branch on the failure; everything looks like `codes.Unknown`.
**Why:** gRPC carries a status *code*; a plain Go error has none, so it's mapped to `Unknown`.
**Fix:** Return `status.Errorf(codes.NotFound, ...)` etc. On the client, `status.FromError(err)` and switch on `st.Code()`.

### 4. No deadline on the client call
**Trap:** A slow or hung downstream blocks the caller forever, exhausting goroutines/connections.
**Why:** `context.Background()` never cancels; gRPC will wait indefinitely.
**Fix:** `ctx, cancel := context.WithTimeout(parent, d); defer cancel()`. Deadlines propagate down the call chain.

### 5. L4 load balancing a gRPC service
**Trap:** All RPCs pile onto one backend pod even with many replicas.
**Why:** HTTP/2 multiplexes every call over one long-lived TCP connection; an L4 LB sees one connection and pins it.
**Fix:** Use client-side LB (`grpc.WithDefaultServiceConfig` + a resolver) or an L7/HTTP-2-aware proxy (Envoy, Linkerd).

### 6. Confusing `codes.Unavailable` with `codes.Internal`
**Trap:** Returning `Unavailable` for a permanent error triggers infinite client/mesh retries (a retry storm).
**Why:** `Unavailable` is the standard "transient, safe to retry" signal; meshes auto-retry it.
**Fix:** `Unavailable` only for genuinely transient conditions; use `Internal`/`InvalidArgument`/`FailedPrecondition` for permanent ones.

### 7. Forgetting to close a server stream / leaking the handler goroutine
**Trap:** A server-streaming handler that loops forever (or ignores `ctx.Done()`) leaks a goroutine per client that disconnects.
**Why:** Returning from the handler ends the stream; if you never return and never watch the context, the goroutine lives on.
**Fix:** Return from the handler to end the stream; select on `stream.Context().Done()` to stop early when the client goes away.
