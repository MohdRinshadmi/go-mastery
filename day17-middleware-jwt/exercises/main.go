// Day 17 — YOUR exercises. Fill in the TODOs.
//
// Run with:   go run main.go
// Don't peek at ../solutions/ until you've genuinely tried each one.
package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// =====================================================================
// EXERCISE 1 (beginner) — Generate and parse a JWT
//
// Define a Claims struct with: UserID string, Role string,
// and embed jwt.RegisteredClaims.
//
// Implement:
//   GenerateToken(userID, role string, secret []byte) (string, error)
//   ParseToken(tokenString string, secret []byte) (*Claims, error)
//
// In main: generate a token for userID="usr-1" role="customer",
// parse it back, and print the claims. Also try parsing a tampered
// token and confirm you get an error.
// =====================================================================

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(userID, role string, secret []byte) (string, error) {
	// TODO: create claims with 1 hour expiry, sign with HS256
	return "", nil
}

func ParseToken(tokenString string, secret []byte) (*Claims, error) {
	// TODO: parse and validate. MUST check signing method!
	return nil, nil
}

// =====================================================================
// EXERCISE 2 (beginner) — Middleware: request ID injection
//
// Write a Gin middleware AddRequestID() that:
//   - Generates a unique ID per request (use fmt.Sprintf("req-%d", time.Now().UnixNano()))
//   - Sets it as response header "X-Request-ID"
//   - Stores it in the Gin context as "request_id"
//   - Then calls c.Next()
//
// Write a handler that reads "request_id" from context and returns it in JSON.
// Register: GET /ping with the middleware and handler.
// Test using httptest.NewRecorder.
// =====================================================================

func AddRequestID() gin.HandlerFunc {
	// TODO
	return nil
}

// =====================================================================
// EXERCISE 3 (intermediate) — Auth middleware + protected route
//
// Using your GenerateToken/ParseToken from Exercise 1:
//
// Write AuthRequired() gin.HandlerFunc that:
//   - Reads "Authorization: Bearer <token>" header
//   - Parses the token (use secret = []byte("exercise-secret"))
//   - On success: c.Set("claims", claims), c.Next()
//   - On failure: c.AbortWithStatusJSON(401, ...) — do NOT call c.Next()
//
// Build a Gin router with:
//   POST /auth/token  — body: {"user_id":"u1","role":"admin"}
//                       returns {"token":"..."}  (no password check — it's an exercise)
//   GET  /me          — protected by AuthRequired, returns claims as JSON
//
// Use httptest to verify the full flow in main (no live server needed).
// =====================================================================

func AuthRequired(secret []byte) gin.HandlerFunc {
	// TODO
	return nil
}

func buildAuthRouter(secret []byte) *gin.Engine {
	// TODO
	return nil
}

// =====================================================================
// CHALLENGE — Validation + RBAC
//
// Extend buildAuthRouter to add:
//
//   POST /products (admin only)
//     body: {"name": "string", "price": float, "stock": int}
//     Validation rules (use go-playground/validator):
//       - name: required, min=2, max=200
//       - price: required, gt=0
//       - stock: required, gte=0
//     Returns 201 with the product + a generated ID.
//     Returns 422 with validation error messages on bad input.
//
// Write a RequireRole(role string) gin.HandlerFunc middleware.
// Chain: AuthRequired → RequireRole("admin") → handler
//
// Test in main:
//   - Valid admin POST → 201
//   - Customer POST → 403
//   - Admin POST with empty name → 422
// =====================================================================

type CreateProductRequest struct {
	Name  string  `json:"name"  validate:"required,min=2,max=200"`
	Price float64 `json:"price" validate:"required,gt=0"`
	Stock int     `json:"stock" validate:"required,gte=0"`
}

func RequireRole(role string) gin.HandlerFunc {
	// TODO
	return nil
}

func main() {
	fmt.Println("== Exercise 1: JWT generate + parse ==")
	// TODO

	fmt.Println("== Exercise 2: Request ID middleware ==")
	// TODO

	fmt.Println("== Exercise 3: Auth flow ==")
	// TODO

	fmt.Println("== Challenge: Validation + RBAC ==")
	// TODO

	// Suppress unused import hints
	_ = http.StatusOK
	_ = gin.Default
}
