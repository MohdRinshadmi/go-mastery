// Day 26 — reference solutions. Try the exercises yourself FIRST.
// Run with: go run main.go

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

	"day26/solutions/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/status"
)

func init() { encoding.RegisterCodec(JSONCodec{}) }

type JSONCodec struct{}

func (JSONCodec) Name() string                          { return "proto" }
func (JSONCodec) Marshal(v interface{}) ([]byte, error) { return json.Marshal(v) }
func (JSONCodec) Unmarshal(data []byte, v interface{}) error { return json.Unmarshal(data, v) }

// ── Store ─────────────────────────────────────────────────────────────────

type productStore struct {
	mu      sync.RWMutex
	orders  map[string]*pb.OrderResponse
	counter int
}

func newStore() *productStore {
	s := &productStore{orders: make(map[string]*pb.OrderResponse)}
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

// ── Server implementation ─────────────────────────────────────────────────

type orderServer struct {
	pb.UnimplementedOrderServiceServer
	store *productStore
}

// Exercise 1 — CreateOrder
func (s *orderServer) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.OrderResponse, error) {
	// Input validation — always validate early, return codes.InvalidArgument
	if req.CustomerID == "" {
		return nil, status.Error(codes.InvalidArgument, "customer_id is required")
	}
	if len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "order must contain at least one item")
	}

	// Calculate total
	var total float64
	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return nil, status.Errorf(codes.InvalidArgument, "item %q: quantity must be > 0", item.ProductID)
		}
		total += float64(item.Quantity) * item.PriceUSD
	}

	order := &pb.OrderResponse{
		OrderID:    s.store.nextID(),
		CustomerID: req.CustomerID,
		Status:     "PENDING",
		TotalUSD:   total,
	}
	s.store.saveOrder(order)
	log.Printf("[server] created %s for %s", order.OrderID, order.CustomerID)
	return order, nil
}

// Exercise 2 — GetOrder
func (s *orderServer) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.OrderResponse, error) {
	if req.OrderID == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}
	order, ok := s.store.getOrder(req.OrderID)
	if !ok {
		// NotFound is a permanent failure — callers should not retry
		return nil, status.Errorf(codes.NotFound, "order %q not found", req.OrderID)
	}
	return order, nil
}

// Exercise 3 — ListOrders (server streaming)
func (s *orderServer) ListOrders(req *pb.ListOrdersRequest, stream pb.OrderService_ListOrdersServer) error {
	if req.CustomerID == "" {
		return status.Error(codes.InvalidArgument, "customer_id is required")
	}

	orders := s.store.ordersForCustomer(req.CustomerID)
	log.Printf("[server] streaming %d orders for %s", len(orders), req.CustomerID)

	for _, o := range orders {
		// Honor client cancellations and deadlines
		if err := stream.Context().Err(); err != nil {
			return status.FromContextError(err).Err()
		}
		if err := stream.Send(o); err != nil {
			return err // client disconnected
		}
	}
	return nil // nil closes stream with EOF on the client side
}

// Challenge — CancelOrder
func (s *orderServer) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.OrderResponse, error) {
	if req.OrderID == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}
	order, ok := s.store.getOrder(req.OrderID)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "order %q not found", req.OrderID)
	}
	// Cannot cancel orders already in transit or delivered
	if order.Status == "SHIPPED" || order.Status == "DELIVERED" {
		return nil, status.Errorf(codes.FailedPrecondition,
			"cannot cancel order in status %s", order.Status)
	}
	order.Status = "CANCELLED"
	s.store.saveOrder(order)
	log.Printf("[server] cancelled %s reason=%q", order.OrderID, req.Reason)
	return order, nil
}

// ── Main ──────────────────────────────────────────────────────────────────

func main() {
	fmt.Println("=== Day 26 Solutions ===")

	store := newStore()
	srv := grpc.NewServer()
	pb.RegisterOrderServiceServer(srv, &orderServer{store: store})
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	must(err, "listen")
	fmt.Printf("[main] server on %s\n", lis.Addr())
	go func() {
		if err := srv.Serve(lis); err != nil && !strings.Contains(err.Error(), "use of closed") {
			log.Printf("[server] stopped: %v", err)
		}
	}()
	defer srv.GracefulStop()
	time.Sleep(50 * time.Millisecond)

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	must(err, "dial")
	defer conn.Close()
	client := pb.NewOrderServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Exercise 1
	fmt.Println("\n── Exercise 1: CreateOrder ──")
	created, err := client.CreateOrder(ctx, &pb.CreateOrderRequest{
		CustomerID: "carol",
		Items: []*pb.OrderItem{
			{ProductID: "book-go", Quantity: 2, PriceUSD: 39.99},
			{ProductID: "notebook", Quantity: 1, PriceUSD: 14.99},
		},
	})
	must(err, "CreateOrder")
	fmt.Printf("  Created: id=%s status=%s total=%.2f\n",
		created.OrderID, created.Status, created.TotalUSD)

	// Exercise 2
	fmt.Println("\n── Exercise 2: GetOrder ──")
	got, err := client.GetOrder(ctx, &pb.GetOrderRequest{OrderID: created.OrderID})
	must(err, "GetOrder")
	fmt.Printf("  Got: id=%s customer=%s status=%s\n", got.OrderID, got.CustomerID, got.Status)

	_, err = client.GetOrder(ctx, &pb.GetOrderRequest{OrderID: "ORD-9999"})
	if st, ok := status.FromError(err); ok {
		fmt.Printf("  Expected NotFound: code=%s\n", st.Code())
	}

	// Exercise 3
	fmt.Println("\n── Exercise 3: ListOrders (streaming) ──")
	stream, err := client.ListOrders(ctx, &pb.ListOrdersRequest{CustomerID: "alice"})
	must(err, "ListOrders")
	for {
		o, err := stream.Recv()
		if err == io.EOF {
			break
		}
		must(err, "Recv")
		fmt.Printf("  [stream] id=%s status=%s total=%.2f\n", o.OrderID, o.Status, o.TotalUSD)
	}

	// Challenge
	fmt.Println("\n── Challenge: CancelOrder ──")
	cancelled, err := client.CancelOrder(ctx, &pb.CancelOrderRequest{
		OrderID: created.OrderID,
		Reason:  "customer requested refund",
	})
	must(err, "CancelOrder")
	fmt.Printf("  Cancelled: id=%s status=%s\n", cancelled.OrderID, cancelled.Status)

	// Try to cancel a SHIPPED order — should fail with FailedPrecondition
	_, err = client.CancelOrder(ctx, &pb.CancelOrderRequest{
		OrderID: "ORD-0002",
		Reason:  "too late",
	})
	if st, ok := status.FromError(err); ok {
		fmt.Printf("  FailedPrecondition: code=%s msg=%q\n", st.Code(), st.Message())
	}

	fmt.Println("\n=== Solutions complete ===")
}

func must(err error, msg string) {
	if err != nil {
		log.Fatalf("[fatal] %s: %v", msg, err)
	}
}
