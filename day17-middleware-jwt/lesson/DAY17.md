# Day 17 — Middleware, JWT Authentication, and Request Validation

> Mentor note: Authentication is where security bugs live. Every year, millions of dollars are lost to JWT misconfiguration, missing validations, and auth middleware that's bypassed by a single misrouted path. Today we build it correctly from first principles — then you'll be able to spot the bugs in code reviews, not just write the happy path.

---

## 0. The Mental Model

```
Request
  │
  ▼
[requestLogger]  ← middleware 1: log, pass through
  │
  ▼
[authRequired]   ← middleware 2: check JWT, abort if invalid
  │
  ▼
[handler]        ← your actual business logic
  │
  ▼
[authRequired]   ← post-handler? No — middleware wraps; it RETURNS here
  │
  ▼
[requestLogger]  ← logging can measure duration here (before return)
  │
  ▼
Response sent
```

Middleware is just a function that wraps a handler. The wrapping creates a chain. Understanding this mental model means you can debug any middleware system in any language.

---

## 1. The Handler-Wrapping Pattern (stdlib)

### Theory

The pattern is always the same shape:

```go
func myMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // PRE: runs before the next handler
        // Example: log start time, check auth, set headers

        next.ServeHTTP(w, r) // call the chain

        // POST: runs after the next handler returns
        // Example: log duration, measure response size
    })
}
```

### Chaining

```go
// Manual chaining (right to left, innermost is called first):
handler := requestLogger(rateLimiter(authRequired(productHandler)))

// Reading it: incoming request → requestLogger → rateLimiter → authRequired → productHandler
//             outgoing response: returns through the stack in reverse
```

### Why this works

`http.HandlerFunc` is a function type that satisfies `http.Handler`. So any function with the right signature is a handler. And a function that returns a handler IS middleware. No special types needed — just the one interface from Day 16.

---

## 2. Gin Middleware — The Same Pattern, Gin Flavor

### Theory

In Gin, middleware is a `gin.HandlerFunc`:

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // PRE
        c.Next()   // call next handler in chain
        // POST (runs after c.Next() returns)
    }
}
```

`c.Abort()` stops the chain immediately — no further handlers run. `c.AbortWithStatus(401)` stops the chain and sends a status. Use this for auth failures.

`c.Set("key", value)` / `c.Get("key")` pass data between middleware and handlers via the request context. This is how you pass the authenticated user from auth middleware to your handler.

### Registering middleware

```go
r := gin.New()

// Global: applies to ALL routes
r.Use(gin.Recovery()) // catch panics
r.Use(requestLogger())

// Group: applies to a subset
auth := r.Group("/api")
auth.Use(authRequired())
{
    auth.GET("/profile", getProfile) // protected
}

// Route-level: applies to one endpoint
r.GET("/admin", adminOnly(), getAdminDashboard)
```

### Middleware ordering matters

```go
// WRONG: Rate limiter runs before logger — if rate limiter blocks, 
//        we never log the request. Hard to debug in production.
r.Use(rateLimiter())
r.Use(requestLogger())

// RIGHT: Log everything, then apply business rules
r.Use(requestLogger())
r.Use(rateLimiter())
r.Use(authRequired())
```

---

## 3. JWT Authentication

### What is JWT?

A **JSON Web Token** is a compact, URL-safe representation of claims. Structure:

```
header.payload.signature
```

- **Header:** `{"alg":"HS256","typ":"JWT"}` — base64url encoded
- **Payload:** `{"sub":"user-123","role":"admin","exp":1720000000}` — base64url encoded
- **Signature:** `HMAC-SHA256(header + "." + payload, secretKey)` — proves authenticity

The server creates the token and gives it to the client on login. The client sends it on every subsequent request. The server verifies the signature — if valid, trusts the claims inside.

**Critical:** JWT payload is NOT encrypted. It is only signed. Anyone with the token can base64-decode the payload and read it. Never put sensitive data (passwords, PII) in a JWT payload.

### Why not sessions?

| Sessions | JWT |
|----------|-----|
| State stored on server (DB or memory) | Stateless — server stores nothing |
| Horizontal scaling needs shared session store | Any server can verify with just the secret |
| Easy to invalidate (delete session) | Hard to invalidate before expiry |
| Safe with cookies | Bearer token in Authorization header |

JWT is the right choice for microservices and APIs. Sessions are better for monolithic web apps with server-side rendering.

### The `golang-jwt/jwt` library

```go
import "github.com/golang-jwt/jwt/v5"

// Claims struct — embed jwt.RegisteredClaims for standard fields.
type Claims struct {
    UserID string `json:"user_id"`
    Role   string `json:"role"`
    jwt.RegisteredClaims // exp, iat, iss, sub, etc.
}

// Create a token:
claims := Claims{
    UserID: "user-123",
    Role:   "admin",
    RegisteredClaims: jwt.RegisteredClaims{
        ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
        IssuedAt:  jwt.NewNumericDate(time.Now()),
        Issuer:    "my-service",
    },
}
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
tokenString, err := token.SignedString([]byte(secretKey))

// Verify a token:
token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
    // CRITICAL: check the signing method BEFORE accepting the token.
    // The alg:none attack: an attacker sets alg to "none", 
    // bypassing signature verification entirely.
    if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
    }
    return []byte(secretKey), nil
})
if err != nil {
    // token is expired, invalid, tampered — reject the request
}
claims, ok := token.Claims.(*Claims)
```

### Common JWT mistakes

1. **No expiry:** A JWT without `exp` is valid forever. If leaked, attacker has permanent access.
2. **alg:none attack:** Not checking the signing method allows an attacker to forge tokens without knowing the secret. ALWAYS check `t.Method`.
3. **Weak secret:** A short or guessable secret key can be brute-forced. Use 32+ bytes of cryptographically random data.
4. **Storing in localStorage:** Vulnerable to XSS. Use `HttpOnly` cookies for web apps.
5. **No token refresh logic:** Short-lived access tokens (15min) + long-lived refresh tokens is the production pattern.

---

## 4. Request Validation with `go-playground/validator`

### Theory

Raw JSON binding gives you the data. Validation ensures it's meaningful.

```go
type CreateUserRequest struct {
    Email    string `json:"email"    validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Name     string `json:"name"     validate:"required,min=2,max=100"`
    Age      int    `json:"age"      validate:"omitempty,min=13,max=120"`
    Role     string `json:"role"     validate:"required,oneof=user admin moderator"`
}
```

### Validation tags

| Tag | Meaning |
|-----|---------|
| `required` | Field must be present and non-zero |
| `omitempty` | Skip validation if zero value |
| `min=N,max=N` | Length (strings) or value (numbers) |
| `email` | Valid email format |
| `url` | Valid URL |
| `oneof=a b c` | Must be one of these values |
| `uuid` | Valid UUID format |
| `len=N` | Exact length |
| `gt=0` | Greater than 0 |

### Using the validator

```go
import "github.com/go-playground/validator/v10"

var validate = validator.New()

func validateRequest(req any) error {
    return validate.Struct(req)
}

// In your handler:
var req CreateUserRequest
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(400, gin.H{"error": "invalid JSON"})
    return
}
if err := validate.Struct(req); err != nil {
    // Parse validation errors into readable messages
    var errs validator.ValidationErrors
    if errors.As(err, &errs) {
        msgs := make([]string, len(errs))
        for i, e := range errs {
            msgs[i] = fmt.Sprintf("%s: %s", e.Field(), e.Tag())
        }
        c.JSON(422, gin.H{"errors": msgs})
        return
    }
    c.JSON(400, gin.H{"error": err.Error()})
    return
}
```

### Gin + validator integration

Gin uses `go-playground/validator` internally for `binding` tags:

```go
type LoginRequest struct {
    Email    string `json:"email"    binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}
// c.ShouldBindJSON runs the binding tags automatically
```

`binding` tags are Gin's built-in path to validation. `validate` tags require manual `validate.Struct()` call. Both use the same validator library under the hood.

---

## 5. Putting It All Together: Auth Flow

```
POST /auth/login
  body: {"email":"...", "password":"..."}
  → validate credentials against DB/store
  → issue JWT access token (15min) + refresh token (7 days)
  → return {"access_token":"...","expires_in":900}

GET /api/profile  (protected)
  header: Authorization: Bearer <access_token>
  → authMiddleware: parse token, validate signature + expiry
  → authMiddleware: c.Set("user", claims) — inject claims into context
  → handler: userID := c.MustGet("user").(*Claims).UserID
```

---

## 6. Role-Based Access Control (RBAC) with Middleware

```go
// requireRole returns a middleware that enforces a minimum role.
func requireRole(role string) gin.HandlerFunc {
    return func(c *gin.Context) {
        claims, ok := c.MustGet("claims").(*Claims)
        if !ok || claims.Role != role {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
                "error": "insufficient permissions",
            })
            return
        }
        c.Next()
    }
}

// Usage:
admin := r.Group("/admin")
admin.Use(authRequired(), requireRole("admin"))
{
    admin.DELETE("/users/:id", deleteUser) // admin only
}
```

---

## 7. Performance Implications

- JWT verification is a cryptographic operation (HMAC-SHA256). It's fast — sub-microsecond on modern hardware. Not a bottleneck.
- Middleware runs on EVERY request. Keep middleware lean — no DB calls in auth middleware if you can cache the user lookup.
- Consider caching user data in a `sync.Map` or Redis after JWT verification to avoid a DB hit per request (see Day 19).
- `go-playground/validator` is fast and caches reflection data on first use. Initialize `validator.New()` once at startup, not per request.

---

## 8. Expert Thinking Mode

- **Beginner:** "I check if the token exists and if it's signed."
- **Intermediate:** "I check signature, expiry, and put claims in context."
- **Senior:** "I also check: signing algorithm (not 'none'), issuer claim, and that the user still exists/is not banned (token revocation via DB or Redis blocklist). I use short-lived tokens (15min) with refresh."
- **Staff:** "Token revocation is the hard problem of JWT. My team decided: for high-value actions (password change, logout all devices), I issue a new token family and the old one becomes invalid via a version field in the DB. For normal expiry, I accept the latency window."

---

## 9. Real-World Use

- **GitHub API:** Bearer tokens in Authorization header — same pattern we're building.
- **AWS Cognito / Auth0:** Both issue JWTs. Your middleware is the verification layer regardless of who issues the token.
- **Internal microservices:** Service-to-service auth also uses JWTs (different `iss` claim, service account subject). The middleware is reusable.
- **Our E-Commerce backend:** Day 20 uses exactly this pattern — `authRequired` middleware on all /api routes, `requireRole("admin")` on product management endpoints.

---

## 10. Interview Questions

1. Explain the JWT structure. Is the payload encrypted? What happens if I base64-decode it?
2. What is the "alg:none" attack? How do you prevent it in `golang-jwt`?
3. What is `c.Abort()` vs `c.AbortWithStatus()` vs `return` in Gin middleware?
4. How do you pass data from middleware to a handler in Gin?
5. What's the difference between `binding:"required"` and `validate:"required"` tags?
6. Why use short-lived JWTs + refresh tokens instead of a single long-lived token?
7. Design an RBAC system for an e-commerce API with roles: `customer`, `vendor`, `admin`. Which endpoints need which roles?

---

## Your Tasks for Today

Go to `../exercises/`. You'll implement a login endpoint that issues a JWT and a protected profile endpoint that verifies it. The challenge builds RBAC. Fill in the TODOs.

Don't peek at `../solutions/` until you've tried.
