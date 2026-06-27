# Day 20 — Capstone Extension Tasks

Implement these on top of the runnable app in this folder (edit the
`internal/...` packages). Keep handlers thin; put logic in services.

## Task 1 — Order ownership read (RBAC at the service layer)
`GET /orders/{id}` already exists. Verify (and write a test for) the rule:
a `customer` may only read **their own** order; an `admin` may read any.
Returning someone else's order must yield `403` (`domain.ErrForbidden`).
→ Already wired in `OrderService.GetByID`. Your job: add a table-driven
test in `internal/service/` proving customer-vs-admin behavior.

## Task 2 — Stock enforcement
Products have a `Stock` field. `OrderService.Place` must reject a line item
whose `Qty` exceeds stock with `domain.ErrValidation` (→ 422), and decrement
stock on success. Add a test: order within stock succeeds and reduces stock;
order over stock fails and leaves stock unchanged (watch partial-order bugs!).

## Task 3 (challenge) — admin-only `GET /users`
Add an endpoint listing all users, restricted to `admin`. Steps:
1. `UserService.List(ctx, actor)` in the service layer — return
   `domain.ErrForbidden` unless `actor.Role == RoleAdmin`.
2. Wire `GET /users` behind the `authed` middleware in the transport layer.
3. Make sure passwords are never serialized (the `PasswordHash` field is
   already `json:"-"` — verify in the response).

## Definition of done
- `go build ./...` clean, `go vet ./...` clean.
- `go test ./...` green (your new service tests pass).
- Manual smoke test with curl (see lesson) behaves correctly.
- Handlers stay thin; all rules live in services; errors map to correct codes.

Bring it for a PR-style review.
