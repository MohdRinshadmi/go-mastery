// Day 16 — YOUR exercises. Fill in the TODOs.
//
// Run with:   go run main.go
// I (your mentor) will review this like a production PR. Write clean code.
//
// Don't peek at ../solutions/ until you've genuinely tried each one.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// =====================================================================
// EXERCISE 1 (beginner) — Understand the Handler interface
//
// Create a type called HealthHandler that implements http.Handler.
// Its ServeHTTP should write a JSON body: {"status":"ok","service":"catalog"}
// with status 200 and Content-Type: application/json.
// Register it at GET / using a plain http.ServeMux (NOT gin).
// =====================================================================

type HealthHandler struct{}

func (h HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: write JSON {"status":"ok","service":"catalog"} with correct headers
}

// =====================================================================
// EXERCISE 2 (beginner) — Go 1.22 method+pattern routing
//
// Using http.NewServeMux() and Go 1.22 "METHOD /path/{param}" syntax:
//   GET /api/categories      → return a JSON array of 3 category names
//   GET /api/categories/{id} → return JSON {"id":"<id>","name":"Electronics"}
//                              (return the same name for any ID — it's a stub)
//
// Write a runCategoryServer() function that builds the mux and returns it
// as an http.Handler (do NOT call ListenAndServe — just return the handler
// so it can be tested with httptest.NewRecorder in production code).
// =====================================================================

func runCategoryServer() http.Handler {
	// TODO: create mux, register two routes, return it
	return nil
}

// =====================================================================
// EXERCISE 3 (intermediate) — Gin route groups + JSON binding
//
// Using Gin, build a router with these routes under /api/v1:
//
//   POST /api/v1/orders
//     - Bind JSON body: {"product_id": "string", "quantity": int}
//     - Validate: product_id must not be empty, quantity must be > 0
//       (manual validation is fine — Day 17 introduces validator tags)
//     - Return 201 with {"order_id":"ord-001","status":"pending"}
//
//   GET /api/v1/orders/:id
//     - Return 200 with {"order_id":"<id>","status":"pending"}
//
// Write a buildOrderRouter() function that returns a *gin.Engine.
// =====================================================================

type CreateOrderRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

func buildOrderRouter() *gin.Engine {
	// TODO: build the router with the routes above
	return nil
}

// =====================================================================
// CHALLENGE (advanced) — Layered handler: logging + method enforcement
//
// Write a middleware function:
//   func onlyJSON(next http.Handler) http.Handler
//
// It should:
//   1. If the request method is POST/PUT/PATCH AND Content-Type is NOT
//      "application/json", respond 415 Unsupported Media Type and STOP
//      (do not call next).
//   2. Otherwise, call next.ServeHTTP(w, r).
//
// Then write a second middleware:
//   func requestLogger(next http.Handler) http.Handler
//
// It should print: [METHOD] /path
//
// Chain them: requestLogger(onlyJSON(yourPostHandler))
// where yourPostHandler is a handler that just returns 200 {"ok":true}
//
// Register the chain at POST /api/echo on a stdlib mux and start the server.
// =====================================================================

func onlyJSON(next http.Handler) http.Handler {
	// TODO: return an http.HandlerFunc that enforces JSON Content-Type
	return nil
}

func requestLogger(next http.Handler) http.Handler {
	// TODO: return an http.HandlerFunc that logs then calls next
	return nil
}

func main() {
	// --- Exercise 1 test ---
	fmt.Println("== Exercise 1: HealthHandler ==")
	// Test with httptest (no actual server needed):
	// TODO: use httptest.NewRecorder to call HealthHandler{}.ServeHTTP
	//       and print the response status + body.
	//       Hint: import "net/http/httptest" and create httptest.NewRecorder()
	_ = HealthHandler{}

	// --- Exercise 2 test ---
	fmt.Println("== Exercise 2: Category routes ==")
	// TODO: call runCategoryServer() and use httptest to verify the two routes.

	// --- Exercise 3 test ---
	fmt.Println("== Exercise 3: Order router ==")
	// TODO: call buildOrderRouter() and use gin's test helpers or httptest
	//       to send a POST /api/v1/orders and print the response.

	// --- Challenge: start the echo server ---
	fmt.Println("== Challenge: Echo server on :8081 ==")
	// TODO: build the mux with the middleware chain and start the server.
	//       Test with: curl -X POST http://localhost:8081/api/echo
	//                            -H "Content-Type: application/json" -d '{}'
	//                  curl -X POST http://localhost:8081/api/echo
	//                            -H "Content-Type: text/plain" -d 'hello'

	// Silence unused import warnings until you implement above:
	_ = json.Marshal
	_ = gin.Default
	_ = http.NewServeMux
}
