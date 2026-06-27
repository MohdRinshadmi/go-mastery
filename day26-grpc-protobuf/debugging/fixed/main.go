// Day 26 debugging — FIXED: use field presence (proto3 `optional`) so the
// server can tell "unset" from "set to zero" and only update what was sent.
//
// In real proto3, marking a field `optional` makes protoc generate a POINTER
// (*string, *float64, ...) — a nil pointer means "not set". We model that here
// with pointer fields. The server applies a field only when its pointer is
// non-nil, so a partial update touches exactly the fields the client provided.
//
// (An alternative real-world fix is a FieldMask listing which paths to update;
// the pointer/optional approach is shown here as the simplest correct model.)
//
// Run: go run .
package main

import (
	"errors"
	"fmt"
)

// UpdateProductRequest with `optional`-style presence: nil pointer == unset.
type UpdateProductRequest struct {
	Id    string
	Name  *string  // nil == not set
	Price *float64 // nil == not set
	Stock *int32   // nil == not set
}

type Product struct {
	Id    string
	Name  string
	Price float64
	Stock int32
}

type server struct {
	store map[string]*Product
}

func (s *server) UpdateProduct(req *UpdateProductRequest) (*Product, error) {
	p, ok := s.store[req.Id]
	if !ok {
		return nil, errors.New("not found")
	}
	// FIX: only apply fields that were actually set (non-nil pointer).
	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.Price != nil {
		p.Price = *req.Price
	}
	if req.Stock != nil {
		p.Stock = *req.Stock
	}
	return p, nil
}

// small helpers, like protobuf's proto.String / proto.Float64 wrappers.
func f64(v float64) *float64 { return &v }

func main() {
	s := &server{store: map[string]*Product{
		"sku-1": {Id: "sku-1", Name: "Mechanical Keyboard", Price: 129.99, Stock: 42},
	}}

	fmt.Printf("before: %+v\n", *s.store["sku-1"])

	// Partial update: only Price set; Name and Stock are nil == "leave alone".
	updated, err := s.UpdateProduct(&UpdateProductRequest{Id: "sku-1", Price: f64(99.99)})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("after : %+v\n", *updated)

	if updated.Name == "Mechanical Keyboard" && updated.Stock == 42 && updated.Price == 99.99 {
		fmt.Println("OK: only Price changed; Name and Stock preserved")
	} else {
		fmt.Println("unexpected result")
	}
}
