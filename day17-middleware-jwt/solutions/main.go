// Day 17 — reference solutions. Try the exercises yourself FIRST.
// Run with: go run main.go
package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
)

// ---- Exercise 1: JWT generate + parse --------------------------------------

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(userID, role string, secret []byte) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ParseToken(tokenString string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{},
		func(t *jwt.Token) (any, error) {
			// MUST check signing method — prevents alg:none attack.
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return secret, nil
		},
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}

// ---- Exercise 2: Request ID middleware -------------------------------------

func AddRequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := fmt.Sprintf("req-%d", time.Now().UnixNano())
		c.Header("X-Request-ID", reqID)
		c.Set("request_id", reqID)
		c.Next()
	}
}

// ---- Exercise 3: Auth flow -------------------------------------------------

func AuthRequired(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "must be 'Bearer <token>'"})
			return
		}
		claims, err := ParseToken(parts[1], secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}

// ---- Challenge: Validation + RBAC ------------------------------------------

type CreateProductRequest struct {
	Name  string  `json:"name"  validate:"required,min=2,max=200"`
	Price float64 `json:"price" validate:"required,gt=0"`
	Stock int     `json:"stock" validate:"gte=0"`
}

var validate = validator.New()

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, exists := c.Get("claims")
		if !exists {
			c.AbortWithStatusJSON(http.StatusInternalServerError,
				gin.H{"error": "claims missing — check middleware order"})
			return
		}
		claims := raw.(*Claims)
		if claims.Role != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("role '%s' required", role),
			})
			return
		}
		c.Next()
	}
}

func buildAuthRouter(secret []byte) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Public: issue token (no password check — teaching exercise only)
	r.POST("/auth/token", func(c *gin.Context) {
		var body struct {
			UserID string `json:"user_id" binding:"required"`
			Role   string `json:"role"    binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		tok, err := GenerateToken(body.UserID, body.Role, secret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": tok})
	})

	// Protected group
	api := r.Group("/")
	api.Use(AuthRequired(secret))
	{
		// GET /me — returns claims
		api.GET("/me", func(c *gin.Context) {
			claims := c.MustGet("claims").(*Claims)
			c.JSON(http.StatusOK, gin.H{
				"user_id": claims.UserID,
				"role":    claims.Role,
			})
		})

		// Admin-only: POST /products
		adminRoutes := api.Group("/products")
		adminRoutes.Use(RequireRole("admin"))
		{
			adminRoutes.POST("", func(c *gin.Context) {
				var req CreateProductRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
					return
				}
				if err := validate.Struct(req); err != nil {
					var ve validator.ValidationErrors
					if errors.As(err, &ve) {
						msgs := make([]string, len(ve))
						for i, e := range ve {
							msgs[i] = fmt.Sprintf("'%s' failed '%s'", e.Field(), e.Tag())
						}
						c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": msgs})
						return
					}
				}
				c.JSON(http.StatusCreated, gin.H{
					"id":    "prod-001",
					"name":  req.Name,
					"price": req.Price,
					"stock": req.Stock,
				})
			})
		}
	}

	return r
}

// ---- main: run all exercises without a live server -------------------------

func main() {
	secret := []byte("exercise-secret-32-bytes-minimum!")

	// --- Exercise 1: JWT ---
	fmt.Println("== Exercise 1: JWT generate + parse ==")
	tok, err := GenerateToken("usr-1", "customer", secret)
	if err != nil {
		fmt.Println("  ERROR:", err)
	} else {
		fmt.Printf("  token (truncated): %s...%s\n", tok[:20], tok[len(tok)-10:])
	}

	claims, err := ParseToken(tok, secret)
	if err != nil {
		fmt.Println("  parse error:", err)
	} else {
		fmt.Printf("  parsed → user_id=%s role=%s\n", claims.UserID, claims.Role)
	}

	_, err = ParseToken("tampered.token.value", secret)
	fmt.Printf("  tampered token error: %v\n", err)

	// --- Exercise 2: Request ID ---
	fmt.Println("== Exercise 2: Request ID middleware ==")
	gin.SetMode(gin.TestMode)
	r2 := gin.New()
	r2.Use(AddRequestID())
	r2.GET("/ping", func(c *gin.Context) {
		reqID, _ := c.Get("request_id")
		c.JSON(http.StatusOK, gin.H{"request_id": reqID})
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	r2.ServeHTTP(rec, req)
	fmt.Printf("  status: %d  X-Request-ID: %s\n",
		rec.Code, rec.Header().Get("X-Request-ID"))

	// --- Exercise 3: Auth flow ---
	fmt.Println("== Exercise 3: Auth flow ==")
	router := buildAuthRouter(secret)

	// Get a token
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/auth/token",
		strings.NewReader(`{"user_id":"u1","role":"customer"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	fmt.Printf("  POST /auth/token → %d: %s", rec.Code, rec.Body.String())

	// Extract token (crude substring extraction for the demo)
	body := rec.Body.String()
	tokenStart := strings.Index(body, `"token":"`) + 9
	tokenEnd := strings.Index(body[tokenStart:], `"`) + tokenStart
	userToken := body[tokenStart:tokenEnd]

	// GET /me with valid token
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	router.ServeHTTP(rec, req)
	fmt.Printf("  GET /me → %d: %s", rec.Code, rec.Body.String())

	// GET /me without token
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/me", nil)
	router.ServeHTTP(rec, req)
	fmt.Printf("  GET /me (no token) → %d: %s", rec.Code, rec.Body.String())

	// --- Challenge: RBAC + Validation ---
	fmt.Println("== Challenge: Validation + RBAC ==")

	// Get admin token
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/auth/token",
		strings.NewReader(`{"user_id":"admin1","role":"admin"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	body = rec.Body.String()
	tokenStart = strings.Index(body, `"token":"`) + 9
	tokenEnd = strings.Index(body[tokenStart:], `"`) + tokenStart
	adminToken := body[tokenStart:tokenEnd]

	// Valid admin POST /products
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/products",
		strings.NewReader(`{"name":"Widget Pro","price":29.99,"stock":100}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	router.ServeHTTP(rec, req)
	fmt.Printf("  POST /products (admin, valid) → %d: %s", rec.Code, rec.Body.String())

	// Customer tries to POST /products — should be 403
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/products",
		strings.NewReader(`{"name":"Widget","price":9.99,"stock":5}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+userToken)
	router.ServeHTTP(rec, req)
	fmt.Printf("  POST /products (customer) → %d: %s", rec.Code, rec.Body.String())

	// Admin POST with invalid data — should be 422
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/products",
		strings.NewReader(`{"name":"","price":-5,"stock":0}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	router.ServeHTTP(rec, req)
	fmt.Printf("  POST /products (admin, invalid) → %d: %s", rec.Code, rec.Body.String())
}
