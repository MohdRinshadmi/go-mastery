// Package http: the transport layer. Translates HTTP <-> service calls.
// Thin: decode, call service, map domain errors to status codes, encode.
// NO business logic lives here.
package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"ecommerce/internal/domain"
	"ecommerce/internal/service"
)

type ctxKey string

const userKey ctxKey = "user"

type Server struct {
	auth     *service.AuthService
	products *service.ProductService
	orders   *service.OrderService
}

func NewServer(a *service.AuthService, p *service.ProductService, o *service.OrderService) *Server {
	return &Server{auth: a, products: p, orders: o}
}

// Routes builds the mux (Go 1.22 method+pattern routing).
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", s.handleRegister)
	mux.HandleFunc("POST /login", s.handleLogin)
	mux.HandleFunc("GET /products", s.handleListProducts)
	mux.Handle("POST /products", s.authed(http.HandlerFunc(s.handleCreateProduct)))
	mux.Handle("POST /orders", s.authed(http.HandlerFunc(s.handlePlaceOrder)))
	mux.Handle("GET /orders/{id}", s.authed(http.HandlerFunc(s.handleGetOrder)))
	return mux
}

// ---- error mapping: domain error -> status code, in ONE place -----------
func statusFor(err error) int {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrValidation):
		return http.StatusUnprocessableEntity
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func writeErr(w http.ResponseWriter, err error) {
	writeJSON(w, statusFor(err), map[string]any{
		"error": map[string]string{"message": err.Error()},
	})
}

// ---- auth middleware: validate token, put user in context ---------------
func (s *Server) authed(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token == "" {
			writeErr(w, domain.ErrUnauthorized)
			return
		}
		u, err := s.auth.Authenticate(r.Context(), token)
		if err != nil {
			writeErr(w, domain.ErrUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userKey, u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func userFrom(ctx context.Context) domain.User {
	u, _ := ctx.Value(userKey).(domain.User)
	return u
}

// ---- handlers -----------------------------------------------------------
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct{ Email, Password, Name string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	u, err := s.auth.Register(r.Context(), req.Email, req.Password, req.Name, domain.RoleCustomer)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, u)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct{ Email, Password string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	token, err := s.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (s *Server) handleListProducts(w http.ResponseWriter, r *http.Request) {
	ps, err := s.products.List(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ps)
}

func (s *Server) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string
		Price float64
		Stock int
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	p, err := s.products.Create(r.Context(), userFrom(r.Context()), req.Name, req.Price, req.Stock)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (s *Server) handlePlaceOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Items []service.LineItem
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	o, err := s.orders.Place(r.Context(), userFrom(r.Context()), req.Items)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, o)
}

func (s *Server) handleGetOrder(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	o, err := s.orders.GetByID(r.Context(), userFrom(r.Context()), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, o)
}
