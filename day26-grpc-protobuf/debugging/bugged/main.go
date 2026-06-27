// Day 26 debugging — proto3 "zero value vs unset" field-presence bug.
//
// We simulate a gRPC UpdateProduct RPC. The proto3 request message has plain
// scalar fields (string Name, double Price, int32 Stock). The client wants to
// do a PARTIAL update: change only the price, leaving name and stock alone.
//
// In proto3, an unset scalar and a scalar set to its zero value are
// INDISTINGUISHABLE on the wire. The server here naively copies every field
// from the request onto the stored product. So when the client sends a request
// with only Price set, the un-set Name ("") and Stock (0) silently CLOBBER the
// stored values.
//
// No external deps — the "generated protobuf message" is a plain struct.
// Run: go run .
package main

import (
	"errors"
	"fmt"
)

// UpdateProductRequest mimics a proto3-generated message: all plain scalars.
// There is NO way to tell "field not set" from "field set to zero".
type UpdateProductRequest struct {
	Id    string
	Name  string  // unset == ""
	Price float64 // unset == 0
	Stock int32   // unset == 0
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
	// BUG: blindly copy every field. A partial update that only sets Price
	// will overwrite Name with "" and Stock with 0, because proto3 cannot
	// distinguish "the client didn't set this" from "the client set zero".
	p.Name = req.Name
	p.Price = req.Price
	p.Stock = req.Stock
	return p, nil
}

func main() {
	s := &server{store: map[string]*Product{
		"sku-1": {Id: "sku-1", Name: "Mechanical Keyboard", Price: 129.99, Stock: 42},
	}}

	fmt.Printf("before: %+v\n", *s.store["sku-1"])

	// Client wants to change ONLY the price to 99.99. It leaves Name and Stock
	// unset — which on the wire means "" and 0.
	updated, err := s.UpdateProduct(&UpdateProductRequest{Id: "sku-1", Price: 99.99})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("after : %+v\n", *updated)

	if updated.Name == "" || updated.Stock == 0 {
		fmt.Println("BUG: partial update clobbered Name and/or Stock with zero values!")
	} else {
		fmt.Println("OK: only Price changed")
	}
}
