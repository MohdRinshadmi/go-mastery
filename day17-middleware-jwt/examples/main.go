// Day 17 examples — Middleware, JWT Auth, Request Validation
//
// Run with: go run main.go
// Then in another terminal:
//
//   # 1. Login (get a token):
//   curl -s -X POST http://localhost:8080/auth/login \
//        -H "Content-Type: application/json" \
//        -d '{"email":"alice@example.com","password":"secret123"}' | jq .
//
//   # 2. Access protected profile (replace TOKEN with the access_token from step 1):
//   curl -s http://localhost:8080/api/profile \
//        -H "Authorization: Bearer TOKEN" | jq .
//
//   # 3. Try without token (should get 401):
//   curl -s http://localhost:8080/api/profile | jq .
//
//   # 4. Admin-only route with non-admin user (should get 403):
//   curl -s http://localhost:8080/api/admin/dashboard \
//        -H "Authorization: Bearer TOKEN" | jq .
//
//   # 5. Login as admin and try again:
//   curl -s -X POST http://localhost:8080/auth/login \
//        -H "Content-Type: application/json" \
//        -d '{"email":"admin@example.com","password":"admin123"}' | jq .
//
// Stop with Ctrl+C.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
)

// ---- Configuration ---------------------------------------------------------

// jwtSecret would be loaded from env in production (Day 18).
// Must be at least 32 bytes of random data in production.
// NEVER hardcode this in real code — shown here for teaching only.
const jwtSecret = "day17-demo-secret-at-least-32bytes!"

// tokenDuration is intentionally long for demo purposes.
// Production: access tokens are 15 minutes, refresh tokens 7 days.
const tokenDuration = 24 * time.Hour

// ---- Domain types ----------------------------------------------------------

// User represents a user in our system.
type User struct {
	ID       string
	Email    string
	Password string // In production: bcrypt hash, never plaintext
	Role     string // "customer", "vendor", "admin"
}

// Claims embeds jwt.RegisteredClaims — always embed this, never use MapClaims
// for production code (MapClaims loses type safety).
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// ---- Request/Response types ------------------------------------------------

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"` // seconds
}

type ProfileResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

// ---- In-memory user store --------------------------------------------------

// In production this would be a repository backed by Postgres (Day 19).
var userStore = map[string]User{
	"alice@example.com": {
		ID: "usr-1", Email: "alice@example.com",
		Password: "secret123", Role: "customer",
	},
	"admin@example.com": {
		ID: "usr-2", Email: "admin@example.com",
		Password: "admin123", Role: "admin",
	},
}

// ---- JWT helpers -----------------------------------------------------------

// generateToken creates a signed JWT for the given user.
// Keep this function PURE: no side effects, no DB calls, easy to test.
func generateToken(user User) (string, error) {
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ecommerce-service",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// parseToken validates a JWT string and returns the claims.
// This is the function your auth middleware will call.
func parseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{},
		func(t *jwt.Token) (any, error) {
			// SECURITY CRITICAL: verify the signing method.
			// If we skip this, an attacker can set alg:none and forge tokens.
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(jwtSecret), nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// ---- Middleware -------------------------------------------------------------

// contextKey is a typed key for context values.
// Using a custom type prevents key collisions with other packages.
type contextKey string

const claimsKey contextKey = "claims"

// AuthRequired is Gin middleware that validates the Bearer token.
// On success: sets "claims" in the Gin context for downstream handlers.
// On failure: aborts with 401 and stops the chain.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the Bearer token from the Authorization header.
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header required",
			})
			return
		}

		// "Bearer <token>" → split and take the second part.
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header must be 'Bearer <token>'",
			})
			return
		}

		claims, err := parseToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token: " + err.Error(),
			})
			return
		}

		// Inject claims into context for downstream handlers.
		// Use c.Set / c.MustGet — Gin's context storage.
		c.Set(string(claimsKey), claims)
		c.Next() // proceed to the next handler in the chain
	}
}

// RequireRole returns middleware that enforces a specific role.
// Must be placed AFTER AuthRequired in the chain — relies on claims existing.
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get(string(claimsKey))
		if !exists {
			// This is a programming error: RequireRole used without AuthRequired.
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "auth claims not found — check middleware order",
			})
			return
		}

		userClaims := claims.(*Claims)
		if userClaims.Role != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("role '%s' required, you have '%s'",
					role, userClaims.Role),
			})
			return
		}

		c.Next()
	}
}

// RequestLogger logs method, path, status, and duration.
// Notice the pattern: capture start time PRE, log duration POST c.Next().
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next() // execute the rest of the chain

		duration := time.Since(start)
		status := c.Writer.Status()
		log.Printf("[%d] %s %s  %v", status, c.Request.Method, path, duration)
	}
}

// ---- Validator setup -------------------------------------------------------

// validate is initialized once at package level — never per-request.
var validate = validator.New()

// validateStruct returns readable error messages from validator errors.
func validateStruct(s any) ([]string, error) {
	if err := validate.Struct(s); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			msgs := make([]string, len(ve))
			for i, e := range ve {
				msgs[i] = fmt.Sprintf("field '%s' failed '%s'", e.Field(), e.Tag())
			}
			return msgs, err
		}
		return nil, err
	}
	return nil, nil
}

// ---- Handlers --------------------------------------------------------------

func loginHandler(c *gin.Context) {
	var req LoginRequest
	// ShouldBindJSON runs the `binding` tag validators automatically.
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Look up user — in production: repository.FindByEmail(ctx, req.Email)
	user, exists := userStore[req.Email]
	if !exists || user.Password != req.Password {
		// Return the SAME error for wrong email OR wrong password.
		// Never reveal which one — prevents user enumeration attacks.
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	tokenString, err := generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresIn:   int(tokenDuration.Seconds()),
	})
}

func profileHandler(c *gin.Context) {
	// Claims were set by AuthRequired middleware.
	claims := c.MustGet(string(claimsKey)).(*Claims)

	c.JSON(http.StatusOK, ProfileResponse{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
	})
}

func adminDashboardHandler(c *gin.Context) {
	claims := c.MustGet(string(claimsKey)).(*Claims)
	c.JSON(http.StatusOK, gin.H{
		"message": "welcome to the admin dashboard",
		"user_id": claims.UserID,
		"role":    claims.Role,
	})
}

// validateHandler demonstrates the validator library independently of Gin binding.
type CreateProductRequest struct {
	Name        string  `json:"name"         validate:"required,min=2,max=200"`
	Price       float64 `json:"price"        validate:"required,gt=0"`
	Category    string  `json:"category"     validate:"required,oneof=electronics books clothing"`
	Description string  `json:"description"  validate:"omitempty,max=1000"`
}

func createProductHandler(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	// Manual validate.Struct call — more control over error messages.
	msgs, err := validateStruct(req)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": msgs})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":       "prod-001",
		"name":     req.Name,
		"price":    req.Price,
		"category": req.Category,
	})
}

// ---- Router setup ----------------------------------------------------------

func newRouter() *gin.Engine {
	r := gin.New() // gin.New() = no default middleware; we add ours explicitly

	// Global middleware — applies to all routes.
	r.Use(gin.Recovery()) // catch panics, return 500 instead of crashing
	r.Use(RequestLogger())

	// Public routes — no auth required.
	r.POST("/auth/login", loginHandler)

	// Protected routes — auth required for everything in this group.
	api := r.Group("/api")
	api.Use(AuthRequired())
	{
		api.GET("/profile", profileHandler)

		// Admin group: auth + role check.
		admin := api.Group("/admin")
		admin.Use(RequireRole("admin"))
		{
			admin.GET("/dashboard", adminDashboardHandler)
			admin.POST("/products", createProductHandler)
		}
	}

	return r
}

// ---- main ------------------------------------------------------------------

func main() {
	gin.SetMode(gin.ReleaseMode) // suppress debug output for clean demo

	r := newRouter()

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		fmt.Println("Day 17 server on :8080")
		fmt.Println("POST /auth/login  (public)")
		fmt.Println("GET  /api/profile (requires JWT)")
		fmt.Println("GET  /api/admin/dashboard (requires admin role)")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	fmt.Println("Done.")
}
