# Day 17 — Middleware & JWT Interview Questions

The lesson's seven, plus the follow-ups a senior interviewer asks when your
first answer is correct. Answers are written the way you'd say them out loud.

---

### 1. Explain the JWT structure. Is the payload encrypted? What happens if I base64-decode it?

<details>
<summary>Answer</summary>

A JWT is three base64url segments joined by dots: `header.payload.signature`.

- **Header** — `{"alg":"HS256","typ":"JWT"}`: the signing algorithm and type.
- **Payload** — the claims: `{"sub":"user-123","role":"admin","exp":...}`.
- **Signature** — `HMAC-SHA256(header + "." + payload, secret)` for HS256.

The payload is **not encrypted, only signed**. Base64url is an *encoding*, not
encryption — anyone can decode it. If you base64-decode the middle segment you
read the claims in plaintext. So never put passwords, PII, or anything secret in
a JWT. The signature doesn't hide the data; it only proves the data wasn't
altered and was issued by someone holding the secret.

</details>

---

### 2. What is the "alg:none" attack? How do you prevent it in `golang-jwt`?

<details>
<summary>Answer</summary>

The JWT spec allows `alg:"none"` — a token with no signature. The attack: take a
valid token, change the header `alg` to `none`, edit the payload to whatever you
like (`"role":"admin"`), drop the signature. If the server verifies "whatever
algorithm the token says," it accepts a token nobody signed.

Prevention: **pin the expected signing method** in the key function before
returning the secret.

```go
jwt.ParseWithClaims(s, &Claims{}, func(t *jwt.Token) (any, error) {
    if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
    }
    return secret, nil
})
```

Modern `golang-jwt` also lets you pass `jwt.WithValidMethods([]string{"HS256"})`
as a parser option for the same effect.

</details>

---

### 3. What is `c.Abort()` vs `c.AbortWithStatus()` vs `return` in Gin middleware?

<details>
<summary>Answer</summary>

- `c.Abort()` flags the chain so **no later handlers run**, but it does *not*
  stop the current function and writes no response by itself.
- `c.AbortWithStatus(code)` does the same flagging *and* writes a status code.
  `c.AbortWithStatusJSON(code, body)` also writes a JSON body.
- `return` only exits the **current function**. On its own it does not stop the
  chain — Gin continues to the next handler.

The correct auth-failure pattern is to **abort and return**: abort stops the
downstream handlers, return stops the rest of *this* function. Forget the return
and the code after the abort still runs.

</details>

---

### 4. How do you pass data from middleware to a handler in Gin?

<details>
<summary>Answer</summary>

Through the request context with `c.Set(key, value)` in the middleware and
`c.Get(key)` (or `c.MustGet(key)`) in the handler.

```go
// in auth middleware, after verifying the token:
c.Set("claims", claims)

// in the handler:
claims := c.MustGet("claims").(*Claims)
userID := claims.UserID
```

`MustGet` panics if the key is absent, so only use it where the middleware that
sets it is guaranteed to have run. In stdlib the equivalent is
`context.WithValue` on `r.Context()` and `r = r.WithContext(...)`.

</details>

---

### 5. What's the difference between `binding:"required"` and `validate:"required"` tags?

<details>
<summary>Answer</summary>

Both use the same `go-playground/validator` library under the hood — the
difference is *who triggers it*.

- `binding` tags are Gin's built-in path: `c.ShouldBindJSON(&req)` parses the
  body **and** runs the binding rules automatically in one call.
- `validate` tags are not run by Gin. You must call `validate.Struct(req)`
  yourself with your own `validator.New()` instance.

Use `binding` for request structs handled by Gin; use `validate` when validating
structs outside the bind step (e.g. config, internal data, non-Gin code).

</details>

---

### 6. Why use short-lived JWTs + refresh tokens instead of one long-lived token?

<details>
<summary>Answer</summary>

Stateless JWTs can't be revoked before they expire — there's no server session
to delete. So expiry time is a trade-off:

- **Long-lived token:** convenient, but if it leaks the attacker has access for
  its whole lifetime, and you can't easily cut it off.
- **Short-lived access token (e.g. 15 min) + long-lived refresh token:** the
  access token limits the blast radius of a leak to minutes. The refresh token
  is used only against a dedicated endpoint, can be stored more carefully, and —
  crucially — *is* revocable because you track it server-side (DB/Redis). Revoke
  the refresh token and the user can't mint new access tokens.

You get statelessness on the hot path (every API call verifies a JWT with no DB
hit) while keeping a revocation lever on the refresh path.

</details>

---

### 7. Design an RBAC system for an e-commerce API with roles `customer`, `vendor`, `admin`. Which endpoints need which roles?

<details>
<summary>Answer</summary>

Put the role in the JWT claims, set it at login, and enforce it with a
`requireRole` middleware layered after `authRequired`.

| Endpoint | customer | vendor | admin |
|----------|:---:|:---:|:---:|
| `GET /products` (browse) | ✅ | ✅ | ✅ |
| `POST /orders` (buy) | ✅ | ✅ | ✅ |
| `GET /me/orders` | ✅ | ✅ | ✅ |
| `POST /products` (list a product) | ❌ | ✅ | ✅ |
| `PUT /products/:id` (own product) | ❌ | ✅ (own) | ✅ |
| `GET /vendor/sales` | ❌ | ✅ | ✅ |
| `DELETE /users/:id` | ❌ | ❌ | ✅ |
| `GET /admin/reports` | ❌ | ❌ | ✅ |

```go
admin := r.Group("/admin")
admin.Use(authRequired(), requireRole("admin"))

vendor := r.Group("/vendor")
vendor.Use(authRequired(), requireRole("vendor"))
```

Key nuance: vendor edits must also check **ownership** ("vendor can edit *their*
products"), which is resource-level authorization the role check alone can't
express — do that in the handler against `claims.UserID`. Prefer permission
checks over hard role-name comparisons if the role set is likely to grow.

</details>

---

### 8. Why does `hmac.Equal` matter — why not just `==` on the signature?

<details>
<summary>Answer</summary>

`==` (and `bytes.Equal`) short-circuits: it stops at the first byte that
differs. That makes comparison time depend on how many leading bytes matched —
a **timing side channel**. An attacker who can measure response times can
reconstruct a valid signature/MAC one byte at a time by finding the value that's
slightly slower to reject.

`hmac.Equal` (built on `subtle.ConstantTimeCompare`) always examines every byte,
so the time it takes is independent of the contents. Use it for any comparison
of secrets, MACs, or tokens.

</details>

---

### 9. How does middleware chaining map to a call stack? Where does the "post-handler" code run?

<details>
<summary>Answer</summary>

`requestLogger(auth(handler))` builds a nested set of function calls. A request
descends the stack: logger's PRE → auth's PRE → handler. The response unwinds
back up: handler returns → auth's POST → logger's POST.

The POST code (everything *after* `next.ServeHTTP(w, r)` or `c.Next()`) runs as
the call returns — it's the unwinding half of the stack. That's why a logger can
measure request duration by recording `start` before `next` and computing the
elapsed time after: the "after" line executes once the entire inner chain has
returned. It's exactly a function call stack — outermost middleware is the first
to start and the last to finish.

</details>

---

### 10. JWTs are stateless — so how do you revoke one before it expires?

<details>
<summary>Answer</summary>

You can't truly revoke a stateless token, so you reintroduce a little state on
purpose. Common strategies:

- **Short expiry + refresh tokens:** accept a small window where a token is still
  valid; revoke at the refresh layer (most common).
- **Denylist / blocklist:** store revoked token IDs (`jti`) in Redis until they'd
  naturally expire; auth middleware checks the list. Adds a lookup per request.
- **Token version / `tokenVersion` claim:** keep a counter on the user row;
  embed it in the token; bump it on "log out everywhere" / password change so all
  older tokens fail a cheap comparison.
- **Short rotating signing keys:** rotate the secret to invalidate whole
  generations of tokens (blunt instrument).

The Staff-level point: every option trades away some of JWT's statelessness. Pick
the smallest amount of state that meets your revocation requirement.

</details>

---

### 11. Where should the client store the token — HttpOnly cookie or localStorage?

<details>
<summary>Answer</summary>

- **localStorage:** readable by any JavaScript on the page, so it's exposed to
  **XSS** — one injected script exfiltrates the token. Convenient for SPAs and
  cross-origin APIs, but you own the XSS risk entirely.
- **HttpOnly cookie:** not readable by JavaScript, so XSS can't steal it. The
  trade-off is **CSRF** — cookies are sent automatically, so you need `SameSite`
  (Lax/Strict) and/or CSRF tokens. Also `Secure` so it's HTTPS-only.

The generally safer default for browser apps is an `HttpOnly`, `Secure`,
`SameSite` cookie with CSRF protection — you defend XSS at the storage layer and
handle CSRF explicitly. Pure-API/mobile clients that aren't browsers don't have
the XSS surface and typically carry the token in the `Authorization` header.

</details>
