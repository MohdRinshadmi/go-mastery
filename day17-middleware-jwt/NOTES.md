# Day 17 — Middleware & JWT Cheatsheet

Fast reference. For the why, read the lesson; for the traps, read PITFALLS.md.

---

## The middleware wrapping shape (stdlib)

```go
func mw(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // PRE  — runs before next
        next.ServeHTTP(w, r)   // call the chain (skip this = chain stops)
        // POST — runs after next returns (unwinding the stack)
    })
}
```

Middleware is just `func(http.Handler) http.Handler`. `http.HandlerFunc` adapts a
plain function into an `http.Handler`.

## Chaining order

```go
h := requestLogger(rateLimiter(authRequired(productHandler)))
// request  flows: logger → rateLimiter → auth → handler   (outer → inner)
// response flows: handler → auth → rateLimiter → logger    (inner → outer)
```

Outermost starts first and finishes last — it's a call stack. Log first, enforce
auth after, so rejected requests still get logged.

## Gin flavor

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // PRE
        c.Next()   // call next handler
        // POST
    }
}
```

| Call | Effect |
|------|--------|
| `c.Next()` | run the rest of the chain |
| `c.Abort()` | stop later handlers (no response written, doesn't stop current func) |
| `c.AbortWithStatus(401)` | abort + write status |
| `c.AbortWithStatusJSON(403, h)` | abort + status + JSON body |
| `c.Set(k, v)` / `c.MustGet(k)` | pass data middleware → handler |

Always `Abort...()` **then** `return` on the failure path.

## JWT: 3 parts + HMAC

```
header.payload.signature      (each segment is base64url)
```

- header  `{"alg":"HS256","typ":"JWT"}`
- payload `{"sub":"u-123","role":"admin","exp":1720000000}`  ← signed, NOT encrypted
- sig     `HMAC-SHA256(header + "." + payload, secret)`

Anyone can decode the payload. Never put secrets/PII in it.

## stdlib verify checklist (in order)

1. **alg** — pin the method (`*jwt.SigningMethodHMAC`); reject `alg:none`.
2. **signature** — recompute the HMAC, compare with `hmac.Equal` (constant-time, never `==`).
3. **exp** — reject if expired *or* if `exp` is missing (forever-valid token).
4. **iss / aud** — check the issuer/audience is who you expect.
5. then trust the claims (`sub`, `role`, ...).

```go
mac := hmac.New(sha256.New, secret)
mac.Write([]byte(payloadB64))
if !hmac.Equal(got, mac.Sum(nil)) { /* reject */ }
if time.Now().Unix() >= claims.Exp  { /* reject */ }
```

## validator: `binding` vs `validate`

| | `binding:"..."` | `validate:"..."` |
|--|----------------|------------------|
| Library | go-playground/validator (under the hood) | go-playground/validator |
| Who runs it | Gin, automatically in `c.ShouldBindJSON(&req)` | you, via `validate.Struct(req)` |
| Typical use | Gin request structs | non-Gin / out-of-band validation |
| Init | built into Gin | `var validate = validator.New()` once at startup |

Common tags: `required`, `omitempty`, `min=N`, `max=N`, `email`, `url`,
`oneof=a b c`, `uuid`, `len=N`, `gt=0`.

```go
type LoginRequest struct {
    Email    string `json:"email"    binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}
```

## RBAC in one snippet

```go
func requireRole(role string) gin.HandlerFunc {
    return func(c *gin.Context) {
        claims := c.MustGet("claims").(*Claims)
        if claims.Role != role {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
            return
        }
        c.Next()
    }
}
admin := r.Group("/admin")
admin.Use(authRequired(), requireRole("admin"))
```

---

## Key terms

- **Middleware** — a function that wraps a handler (`func(http.Handler) http.Handler`) to run logic before/after it; chaining them forms a request pipeline.
- **JWT** — JSON Web Token: compact `header.payload.signature`, signed claims passed as a bearer token.
- **Claims** — the key/values in the JWT payload (`sub`, `role`, `exp`, ...). Signed, not encrypted.
- **HMAC-SHA256 (HS256)** — keyed hash used to sign/verify a JWT with a shared secret.
- **alg:none** — an attack where the token declares no signature algorithm; prevented by pinning the expected signing method.
- **RBAC** — Role-Based Access Control: authorize requests by the user's role, enforced via middleware (`requireRole`).
- **hmac.Equal / constant-time compare** — compares all bytes regardless of content, defeating timing attacks; use instead of `==` on secrets/signatures.
- **RegisteredClaims / exp** — standard JWT claims (`exp`, `iat`, `iss`, `sub`, ...); `exp` is the Unix expiry timestamp and the only built-in revocation in stateless JWT.
