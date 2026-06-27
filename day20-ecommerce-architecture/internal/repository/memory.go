// Package repository: in-memory implementations of the repo interfaces the
// service layer defines. Swap these for Postgres (Day 19) without touching
// the service or transport layers — that's the point of the abstraction.
package repository

import (
	"context"
	"sync"

	"ecommerce/internal/domain"
)

type UserRepo struct {
	mu     sync.RWMutex
	byID   map[string]domain.User
	emails map[string]string // email -> id
}

func NewUserRepo() *UserRepo {
	return &UserRepo{byID: map[string]domain.User{}, emails: map[string]string{}}
}
func (r *UserRepo) Create(_ context.Context, u domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.emails[u.Email]; ok {
		return domain.ErrConflict
	}
	r.byID[u.ID] = u
	r.emails[u.Email] = u.ID
	return nil
}
func (r *UserRepo) GetByID(_ context.Context, id string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.byID[id]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return u, nil
}
func (r *UserRepo) GetByEmail(_ context.Context, email string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.emails[email]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return r.byID[id], nil
}
func (r *UserRepo) List(_ context.Context) ([]domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.User, 0, len(r.byID))
	for _, u := range r.byID {
		out = append(out, u)
	}
	return out, nil
}

type ProductRepo struct {
	mu   sync.RWMutex
	byID map[string]domain.Product
}

func NewProductRepo() *ProductRepo {
	return &ProductRepo{byID: map[string]domain.Product{}}
}
func (r *ProductRepo) Create(_ context.Context, p domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[p.ID] = p
	return nil
}
func (r *ProductRepo) Update(_ context.Context, p domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byID[p.ID]; !ok {
		return domain.ErrNotFound
	}
	r.byID[p.ID] = p
	return nil
}
func (r *ProductRepo) GetByID(_ context.Context, id string) (domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.byID[id]
	if !ok {
		return domain.Product{}, domain.ErrNotFound
	}
	return p, nil
}
func (r *ProductRepo) List(_ context.Context) ([]domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Product, 0, len(r.byID))
	for _, p := range r.byID {
		out = append(out, p)
	}
	return out, nil
}

type OrderRepo struct {
	mu   sync.RWMutex
	byID map[string]domain.Order
}

func NewOrderRepo() *OrderRepo { return &OrderRepo{byID: map[string]domain.Order{}} }
func (r *OrderRepo) Create(_ context.Context, o domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[o.ID] = o
	return nil
}
func (r *OrderRepo) GetByID(_ context.Context, id string) (domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	o, ok := r.byID[id]
	if !ok {
		return domain.Order{}, domain.ErrNotFound
	}
	return o, nil
}
