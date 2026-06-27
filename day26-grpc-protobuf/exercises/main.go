// Day 26 — YOUR exercises. Fill in the TODOs.
//
// Run with: go run main.go
// This file has a pre-wired server skeleton and JSON codec.
// Your job: implement the missing methods and the client demo.
//
// Don't peek at ../solutions/ until you've genuinely tried each one.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"day26/exercises/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/status"
)

// JSON codec — do not touch, makes hand-authored stubs work over the wire.
func init() { encoding.RegisterCodec(JSONCodec{}) }

type JSONCodec struct{}

func (JSONCodec) Name() string                        { return "proto" }
func (JSONCodec) Marshal(v interface{}) ([]byte, error) { return json.Marshal(v) }
func (JSONCodec) Unmarshal(data []byte, v interface{}) error { return json.Unmarshal(data, v) }

// ── In-memory store (provided — no changes needed) ────────────────────────

type productStore struct {
	mu       sync.RWMutex
	products map[string]*pb.OrderResponse // reusing OrderResponse as product for simplicity
	orders   map[string]*pb.OrderResponse
	counter  int
}

func newStore() *productStore {
	s := &productStore{
		products: make(map[string]*pb.OrderResponse),
		orders:   make(map[string]*pb.OrderResponse),
	}
	// Seed some orders for listing
	s.orders["ORD-0001"] = &pb.OrderResponse{OrderID: "ORD-0001", CustomerID: "alice", Status: "PENDING", TotalUSD: 99.99}
	s.orders["ORD-0002"] = &pb.OrderResponse{OrderID: "ORD-0002", CustomerID: "alice", Status: "SHIPPED", TotalUSD: 49.50}
	s.orders["ORD-0003"] = &pb.OrderResponse{OrderID: "ORD-0003", CustomerID: "bob", Status: "PENDING", TotalUSD: 25.00}
	s.counter = 3
	return s
}

func (s *productStore) nextID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter++
	return fmt.Sprintf("ORD-%04d", s.counter)
}

func (s *productStore) saveOrder(o *pb.OrderResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders[o.OrderID] = o
}

func (s *productStore) getOrder(id string) (*pb.OrderResponse, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.orders[id]
	return o, ok
}

func (s *productStore) ordersForCustomer(customerID string) []*pb.OrderResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*pb.OrderResponse
	for _, o := range s.orders {
		if o.CustomerID == customerID {
			result = append(result, o)
		}
	}
	return result
}

// =====================================================================
// EXERCISE 1 — Implement CreateOrder (Unary RPC)
//
// Requirements:
//   - Validate: CustomerID must not be empty → codes.InvalidArgument
//   - Validate: Items must not be empty → codes.InvalidArgument
//   - Calculate total from items (quantity * price_usd)
//   - Generate an ID with store.nextID()
//   - Save to store and return the OrderResponse with Status="PENDING"
//   - Log: "[server] created <id> for <customer>"
// =====================================================================

// =====================================================================
// EXERCISE 2 — Implement GetOrder (Unary RPC)
//
// Requirements:
//   - Validate OrderID not empty → codes.InvalidArgument
//   - Look up in store → if missing return codes.NotFound
//   - Return the found order
// =====================================================================

// =====================================================================
// EXERCISE 3 — Implement ListOrders (Server Streaming RPC)
//
// Requirements:
//   - Validate CustomerID not empty → codes.InvalidArgument
//   - Call store.ordersForCustomer to get the slice
//   - For each order: check stream.Context().Err(), then stream.Send(order)
//   - Return nil when done (EOF to client)
// =====================================================================

// =====================================================================
// CHALLENGE — Implement CancelOrder (Unary RPC)
//
// Requirements:
//   - Validate OrderID not empty → codes.InvalidArgument
//   - If not found → codes.NotFound
//   - If status is "SHIPPED" or "DELIVERED" → codes.FailedPrecondition
//     message: "cannot cancel order in status <X>"
//   - Otherwise: set status to "CANCELLED", save, return updated order
//   - Log the reason
// =====================================================================

// orderServer — your server implementation lives here.
type orderServer struct {
	pb.UnimplementedOrderServiceServer
	store *productStore
}

func (s *orderServer) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.OrderResponse, error) {
	// TODO: Exercise 1
	return nil, status.Error(codes.Unimplemented, "not implemented yet")
}

func (s *orderServer) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.OrderResponse, error) {
	// TODO: Exercise 2
	return nil, status.Error(codes.Unimplemented, "not implemented yet")
}

func (s *orderServer) ListOrders(req *pb.ListOrdersRequest, stream pb.OrderService_ListOrdersServer) error {
	// TODO: Exercise 3
	return status.Error(codes.Unimplemented, "not implemented yet")
}

func (s *orderServer) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.OrderResponse, error) {
	// TODO: Challenge
	return nil, status.Error(codes.Unimplemented, "not implemented yet")
}

// ── Main — wire server + write your client demo ───────────────────────────

func main() {
	fmt.Println("=== Day 26 Exercises ===")

	store := newStore()

	// Server setup — provided, no changes needed
	srv := grpc.NewServer()
	pb.RegisterOrderServiceServer(srv, &orderServer{store: store})
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	must(err, "listen")
	addr := lis.Addr().String()
	fmt.Printf("[main] server on %s\n", addr)
	go func() {
		if err := srv.Serve(lis); err != nil && !strings.Contains(err.Error(), "use of closed") {
			log.Printf("[server] stopped: %v", err)
		}
	}()
	defer srv.GracefulStop()
	time.Sleep(50 * time.Millisecond)

	// Client setup — provided, no changes needed
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	must(err, "dial")
	defer conn.Close()
	client := pb.NewOrderServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ── Exercise 1: Create an order ───────────────────────────────────────
	fmt.Println("\n── Exercise 1: CreateOrder ──")
	// TODO: Call client.CreateOrder with a valid request (customer "carol", 2 items)
	// Print the returned order ID, status, and total.
	_ = client
	_ = ctx

	// ── Exercise 2: Get the order you just created ────────────────────────
	fmt.Println("\n── Exercise 2: GetOrder ──")
	// TODO: Call client.GetOrder for the ID from exercise 1.
	// Then call it with a bogus ID and print the error code.

	// ── Exercise 3: List orders for "alice" ───────────────────────────────
	fmt.Println("\n── Exercise 3: ListOrders (streaming) ──")
	// TODO: Call client.ListOrders for customer "alice".
	// Loop with stream.Recv() until io.EOF and print each order.
	_ = io.EOF

	// ── Challenge: Cancel an order ────────────────────────────────────────
	fmt.Println("\n── Challenge: CancelOrder ──")
	// TODO: Cancel the order you created (should succeed).
	// Then try to cancel "ORD-0002" (status=SHIPPED) and print the FailedPrecondition error.
}

func must(err error, msg string) {
	if err != nil {
		log.Fatalf("[fatal] %s: %v", msg, err)
	}
}
