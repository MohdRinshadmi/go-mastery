# Day 17 — Middleware & JWT Pitfalls (Trap → Why → Fix)

Auth bugs are quiet. Nothing panics; the request just succeeds when it should
have been refused. Here are the ones that show up in real code reviews.

---

## 1. Middleware that forgets to call the next handler

**Trap.**

```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !valid(r) {
            http.Error(w, "no", http.StatusUnauthorized)
            return
        }
        // forgot: next.ServeHTTP(w, r)
    })
}
```

**Why.** The chain stops dead. There's no error — the handler simply never runs,
so the client gets an empty `200` (or a hung-feeling blank response). In Gin the
same bug is a missing `c.Next()`.

**Fix.** Every code path that doesn't reject must pass control downstream.

```go
next.ServeHTTP(w, r) // stdlib
// or
c.Next()             // gin
```

---

## 2. `==` on signatures or secrets

**Trap.**

```go
if providedSig == expectedSig { // string compare
    // accept
}
```

**Why.** `==` short-circuits at the first differing byte. The time it takes to
fail leaks how many leading bytes matched — a timing oracle an attacker can use
to reconstruct the signature byte by byte.

**Fix.** Constant-time comparison.

```go
import "crypto/hmac"

if hmac.Equal(providedSig, expectedSig) { // always compares all bytes
    // accept
}
```

---

## 3. Not pinning the signing method (alg:none)

**Trap.**

```go
token, _ := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
    return secret, nil // trusts whatever alg the token declares
})
```

**Why.** An attacker sets the header `alg` to `"none"`, drops the signature, and
your parser "verifies" a token nobody signed. They forge any claims they want.

**Fix.** Pin the expected method before returning the key.

```go
func(t *jwt.Token) (any, error) {
    if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
    }
    return secret, nil
}
```

---

## 4. No expiry check

**Trap.** Verify the signature, trust the claims, done — `exp` never looked at.

**Why.** A valid signature proves *integrity*, not *freshness*. In stateless JWT,
`exp` is your only built-in revocation. Skip it and any leaked token grants
permanent access. A token with no `exp` at all is valid forever.

**Fix.** Always validate expiry (libraries do this if `exp` is set — so always
set it, and reject a missing `exp`).

```go
RegisteredClaims: jwt.RegisteredClaims{
    ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
}
```

---

## 5. Putting secrets in the JWT payload

**Trap.**

```go
claims := Claims{Password: user.Password, SSN: user.SSN} // in the token
```

**Why.** A JWT is **signed, not encrypted**. The payload is base64url — anyone
holding the token decodes it instantly. Paste it into jwt.io and read it. You've
shipped PII to the client and into every log that records the header.

**Fix.** Put only non-sensitive identifiers in claims (`sub`, `role`). Look up
anything sensitive server-side from the `sub`.

```go
claims := Claims{UserID: user.ID, Role: user.Role} // safe to expose
```

---

## 6. `c.Abort()` without `return`

**Trap.**

```go
func authMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if !valid(c) {
            c.AbortWithStatus(http.StatusUnauthorized)
            // no return — keeps going
        }
        c.Next() // runs even on the failure path
    }
}
```

**Why.** `c.Abort()` flags the chain to stop *after the current handler returns*
— it does not stop the current function. Without `return`, the rest of your
middleware body still executes (here, calling `c.Next()` anyway).

**Fix.** Abort *and* return.

```go
if !valid(c) {
    c.AbortWithStatus(http.StatusUnauthorized)
    return
}
c.Next()
```

---

## 7. Middleware ordering: auth before logging

**Trap.**

```go
r.Use(authRequired())  // rejects here...
r.Use(requestLogger()) // ...so this never runs for rejected requests
```

**Why.** When auth rejects a request, the logger never sees it. Your access logs
silently omit every blocked/attacking request — exactly the ones you most want
to see when investigating an incident.

**Fix.** Log first, then enforce business rules.

```go
r.Use(requestLogger())
r.Use(authRequired())
```
