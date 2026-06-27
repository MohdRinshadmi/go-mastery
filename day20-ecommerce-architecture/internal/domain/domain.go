// Package domain holds the core types and errors. It imports NOTHING from
// http, sql, or other layers — it's the innermost circle.
package domain

import "errors"

// Domain errors — the HTTP layer maps these to status codes in ONE place.
var (
	ErrNotFound     = errors.New("not found")
	ErrValidation   = errors.New("validation failed")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("conflict")
)

type Role string

const (
	RoleCustomer Role = "customer"
	RoleAdmin    Role = "admin"
)

type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	Role         Role   `json:"role"`
	PasswordHash string `json:"-"` // never serialized
}

type Product struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

type OrderItem struct {
	ProductID string  `json:"product_id"`
	Qty       int     `json:"qty"`
	UnitPrice float64 `json:"unit_price"`
}

type Order struct {
	ID     string      `json:"id"`
	UserID string      `json:"user_id"`
	Items  []OrderItem `json:"items"`
	Total  float64     `json:"total"`
}
