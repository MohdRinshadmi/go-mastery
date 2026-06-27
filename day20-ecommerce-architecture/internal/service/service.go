// Package service holds business logic. It depends on repo INTERFACES
// (defined here, the consumer) — never on concrete repos, http, or sql.
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"ecommerce/internal/domain"
)

// ---- Repo interfaces OWNED by the service (dependency inversion) ---------
type UserRepository interface {
	Create(ctx context.Context, u domain.User) error
	GetByID(ctx context.Context, id string) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	List(ctx context.Context) ([]domain.User, error)
}
type ProductRepository interface {
	Create(ctx context.Context, p domain.Product) error
	Update(ctx context.Context, p domain.Product) error
	GetByID(ctx context.Context, id string) (domain.Product, error)
	List(ctx context.Context) ([]domain.Product, error)
}
type OrderRepository interface {
	Create(ctx context.Context, o domain.Order) error
	GetByID(ctx context.Context, id string) (domain.Order, error)
}

// ---- tiny id + password + token helpers (real apps: UUID, bcrypt, JWT) ---
var idCounter atomic.Int64

func newID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, idCounter.Add(1))
}
func hashPassword(pw string) string {
	sum := sha256.Sum256([]byte("salt:" + pw)) // DEMO only — use bcrypt in prod
	return hex.EncodeToString(sum[:])
}

// ---- AuthService: register/login + opaque token store --------------------
type AuthService struct {
	users  UserRepository
	mu     sync.Mutex
	tokens map[string]string // token -> userID
}

func NewAuthService(users UserRepository) *AuthService {
	return &AuthService{users: users, tokens: map[string]string{}}
}

func (s *AuthService) Register(ctx context.Context, email, password, name string, role domain.Role) (domain.User, error) {
	if !strings.Contains(email, "@") || len(password) < 2 {
		return domain.User{}, fmt.Errorf("%w: email/password invalid", domain.ErrValidation)
	}
	u := domain.User{
		ID: newID("usr"), Email: email, Name: name, Role: role,
		PasswordHash: hashPassword(password),
	}
	if err := s.users.Create(ctx, u); err != nil {
		return domain.User{}, err // ErrConflict bubbles up
	}
	return u, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil || u.PasswordHash != hashPassword(password) {
		return "", domain.ErrUnauthorized // don't leak which part failed
	}
	token := newID("tok")
	s.mu.Lock()
	s.tokens[token] = u.ID
	s.mu.Unlock()
	return token, nil
}

// Authenticate resolves a token to a user (used by middleware).
func (s *AuthService) Authenticate(ctx context.Context, token string) (domain.User, error) {
	s.mu.Lock()
	uid, ok := s.tokens[token]
	s.mu.Unlock()
	if !ok {
		return domain.User{}, domain.ErrUnauthorized
	}
	return s.users.GetByID(ctx, uid)
}

// ---- ProductService: RBAC enforced here (business rule) ------------------
type ProductService struct{ products ProductRepository }

func NewProductService(p ProductRepository) *ProductService { return &ProductService{products: p} }

func (s *ProductService) Create(ctx context.Context, actor domain.User, name string, price float64, stock int) (domain.Product, error) {
	if actor.Role != domain.RoleAdmin {
		return domain.Product{}, fmt.Errorf("%w: admin role required", domain.ErrForbidden)
	}
	if name == "" || price <= 0 {
		return domain.Product{}, fmt.Errorf("%w: name/price invalid", domain.ErrValidation)
	}
	p := domain.Product{ID: newID("prd"), Name: name, Price: price, Stock: stock}
	if err := s.products.Create(ctx, p); err != nil {
		return domain.Product{}, err
	}
	return p, nil
}
func (s *ProductService) List(ctx context.Context) ([]domain.Product, error) {
	return s.products.List(ctx)
}

// ---- OrderService: orchestration + stock rule ----------------------------
type OrderService struct {
	orders   OrderRepository
	products ProductRepository
}

func NewOrderService(o OrderRepository, p ProductRepository) *OrderService {
	return &OrderService{orders: o, products: p}
}

type LineItem struct {
	ProductID string
	Qty       int
}

func (s *OrderService) Place(ctx context.Context, actor domain.User, items []LineItem) (domain.Order, error) {
	if len(items) == 0 {
		return domain.Order{}, fmt.Errorf("%w: order has no items", domain.ErrValidation)
	}
	order := domain.Order{ID: newID("ord"), UserID: actor.ID}
	for _, li := range items {
		p, err := s.products.GetByID(ctx, li.ProductID)
		if err != nil {
			return domain.Order{}, err // ErrNotFound -> 404
		}
		if li.Qty <= 0 || li.Qty > p.Stock {
			return domain.Order{}, fmt.Errorf("%w: qty %d exceeds stock for %s", domain.ErrValidation, li.Qty, p.Name)
		}
		p.Stock -= li.Qty
		_ = s.products.Update(ctx, p)
		order.Items = append(order.Items, domain.OrderItem{ProductID: p.ID, Qty: li.Qty, UnitPrice: p.Price})
		order.Total += p.Price * float64(li.Qty)
	}
	if err := s.orders.Create(ctx, order); err != nil {
		return domain.Order{}, err
	}
	return order, nil
}

// GetByID enforces ownership: a customer may only read their own order.
func (s *OrderService) GetByID(ctx context.Context, actor domain.User, id string) (domain.Order, error) {
	o, err := s.orders.GetByID(ctx, id)
	if err != nil {
		return domain.Order{}, err
	}
	if actor.Role != domain.RoleAdmin && o.UserID != actor.ID {
		return domain.Order{}, fmt.Errorf("%w: not your order", domain.ErrForbidden)
	}
	return o, nil
}
