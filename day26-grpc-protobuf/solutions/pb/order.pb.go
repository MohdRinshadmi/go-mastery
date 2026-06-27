// Code equivalent to what protoc --go_out would generate for order.proto.
// Hand-authored so the module builds without protoc installed.
// In production: run protoc and commit the generated files.
//
// Key differences from raw generated code:
//   - No protoreflect wiring (saves ~200 lines of boilerplate)
//   - No proto.Message interface implementation
//   - Pure Go structs — usable directly with google.golang.org/grpc
//
// The grpc framework only needs the structs for encoding; we use
// encoding/json as a stand-in serialization for the in-process demo.
// Real generated code uses protobuf binary encoding automatically.

package pb

// ── Message structs ─────────────────────────────────────────────────────

// OrderItem represents a line-item in an order.
type OrderItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int32   `json:"quantity"`
	PriceUSD  float64 `json:"price_usd"`
}

// CreateOrderRequest — payload for the CreateOrder RPC.
type CreateOrderRequest struct {
	CustomerID string       `json:"customer_id"`
	Items      []*OrderItem `json:"items"`
}

// GetOrderRequest — payload for the GetOrder RPC.
type GetOrderRequest struct {
	OrderID string `json:"order_id"`
}

// ListOrdersRequest — payload for the ListOrders server-streaming RPC.
type ListOrdersRequest struct {
	CustomerID string `json:"customer_id"`
}

// CancelOrderRequest — payload for the CancelOrder RPC.
type CancelOrderRequest struct {
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}

// OrderResponse — returned by all order RPCs.
type OrderResponse struct {
	OrderID    string  `json:"order_id"`
	CustomerID string  `json:"customer_id"`
	Status     string  `json:"status"`
	TotalUSD   float64 `json:"total_usd"`
}
