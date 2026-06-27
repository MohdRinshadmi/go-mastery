// Day 26 — gRPC + Protocol Buffers walkthrough
//
// This is a self-contained in-process demo that:
//   - Starts a real gRPC server (on a random port)
//   - Makes real gRPC client calls (unary + server streaming)
//   - Demonstrates interceptors, deadlines, and error codes
//   - Compiles and runs WITHOUT protoc installed
//
// Run with: go run main.go
//
// For a real project with generated code, see order.proto and the
// hand-authored stubs in pb/ — they are structurally identical to
// what protoc --go-grpc_out would produce.

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

	"day26/examples/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/status"
)

// ── JSON codec — replaces protobuf binary encoding ───────────────────────
//
// In production, generated code integrates with the protobuf binary codec
// automatically. Here we register a JSON codec so our hand-authored structs
// marshal correctly over the wire.

func init() {
	encoding.RegisterCodec(JSONCodec{})
}

// JSONCodec implements grpc/encoding.Codec using encoding/json.
type JSONCodec struct{}

func (JSONCodec) Name() string { return "proto" } // "proto" overrides the default

func (JSONCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (JSONCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// ── Order store ───────────────────────────────────────────────────────────

// orderStore is a thread-safe in-memory order database.
// In production: swap for Redis (Day 28) or PostgreSQL.
type orderStore struct {
	mu     sync.RWMutex
	orders map[string]*pb.OrderResponse
}

func newOrderStore() *orderStore {
	return &orderStore{orders: make(map[string]*pb.OrderResponse)}
}

func (s *orderStore) save(o *pb.OrderResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders[o.OrderID] = o
}

func (s *orderStore) get(id string) (*pb.OrderResponse, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.orders[id]
	return o, ok
}

func (s *orderStore) byCustomer(customerID string) []*pb.OrderResponse {
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

// ── orderCounter is used to generate unique IDs ───────────────────────────

var (
	orderMu      sync.Mutex
	orderCounter int
)

func nextOrderID() string {
	orderMu.Lock()
	defer orderMu.Unlock()
	orderCounter++
	return fmt.Sprintf("ORD-%04d", orderCounter)
}

// ── Server implementation ─────────────────────────────────────────────────

// orderServer implements pb.OrderServiceServer.
// It embeds UnimplementedOrderServiceServer for forward compatibility —
// if we add RPCs to the proto later, this struct won't fail to compile.
type orderServer struct {
	pb.UnimplementedOrderServiceServer
	store *orderStore
}

// CreateOrder — Unary RPC.
// Production concern: validate input, persist to DB, publish an OrderCreated
// event to Kafka (Day 27), then return the receipt.
func (s *orderServer) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.OrderResponse, error) {
	if req.CustomerID == "" {
		return nil, status.Error(codes.InvalidArgument, "customer_id is required")
	}
	if len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "order must have at least one item")
	}

	var total float64
	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return nil, status.Errorf(codes.InvalidArgument, "item %q has invalid quantity", item.ProductID)
		}
		total += float64(item.Quantity) * item.PriceUSD
	}

	order := &pb.OrderResponse{
		OrderID:    nextOrderID(),
		CustomerID: req.CustomerID,
		Status:     "PENDING",
		TotalUSD:   total,
	}
	s.store.save(order)

	log.Printf("[server] CreateOrder: %s for customer %s, total=%.2f",
		order.OrderID, order.CustomerID, order.TotalUSD)
	return order, nil
}

// GetOrder — Unary RPC.
func (s *orderServer) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.OrderResponse, error) {
	if req.OrderID == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}
	order, ok := s.store.get(req.OrderID)
	if !ok {
		// codes.NotFound signals the client this is a permanent failure (don't retry)
		return nil, status.Errorf(codes.NotFound, "order %q not found", req.OrderID)
	}
	return order, nil
}

// CancelOrder — Unary RPC.
// Demonstrates multi-case error handling: NotFound + FailedPrecondition.
func (s *orderServer) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.OrderResponse, error) {
	if req.OrderID == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}
	order, ok := s.store.get(req.OrderID)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "order %q not found", req.OrderID)
	}
	if order.Status == "SHIPPED" || order.Status == "DELIVERED" {
		// FailedPrecondition: valid request, but current state doesn't allow it
		return nil, status.Errorf(codes.FailedPrecondition,
			"cannot cancel order %q in status %s", req.OrderID, order.Status)
	}

	order.Status = "CANCELLED"
	s.store.save(order)
	log.Printf("[server] CancelOrder: %s reason=%q", order.OrderID, req.Reason)
	return order, nil
}

// ListOrders — Server streaming RPC.
// The server sends multiple OrderResponse messages on a single RPC call.
// The client reads them in a for loop until io.EOF.
//
// Production use: paginated order history, real-time order feeds,
// streaming large result sets without loading everything into memory.
func (s *orderServer) ListOrders(req *pb.ListOrdersRequest, stream pb.OrderService_ListOrdersServer) error {
	if req.CustomerID == "" {
		return status.Error(codes.InvalidArgument, "customer_id is required")
	}

	orders := s.store.byCustomer(req.CustomerID)
	log.Printf("[server] ListOrders: streaming %d orders for customer %s",
		len(orders), req.CustomerID)

	for _, o := range orders {
		// Check if client cancelled (deadline exceeded, client gone, etc.)
		if err := stream.Context().Err(); err != nil {
			return status.FromContextError(err).Err()
		}
		if err := stream.Send(o); err != nil {
			return err // client disconnected mid-stream
		}
		// In production, add a small sleep between sends to avoid
		// overwhelming a slow consumer. Or implement flow control.
		time.Sleep(10 * time.Millisecond) // artificial delay for demo
	}
	return nil // returning nil closes the stream cleanly (EOF on client side)
}

// ── Interceptors (middleware) ─────────────────────────────────────────────

// loggingInterceptor is a server-side unary interceptor.
// In production, use grpc-ecosystem/go-grpc-middleware for composable chains.
func loggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)

	// Extract gRPC status code for structured logging
	code := codes.OK
	if err != nil {
		if s, ok := status.FromError(err); ok {
			code = s.Code()
		}
	}
	log.Printf("[interceptor] %s | code=%s | duration=%v",
		info.FullMethod, code, time.Since(start))
	return resp, err
}

// ── Main — server + client demo ───────────────────────────────────────────

func main() {
	fmt.Println("=== Day 26: gRPC + Protocol Buffers ===")
	fmt.Println()

	// ── Start server ──────────────────────────────────────────────────────
	store := newOrderStore()
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
	)
	pb.RegisterOrderServiceServer(srv, &orderServer{store: store})

	// Listen on a random available port
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	must(err, "listen")
	addr := lis.Addr().String()
	fmt.Printf("[main] gRPC server listening on %s\n", addr)

	// Run server in background goroutine
	go func() {
		if err := srv.Serve(lis); err != nil && !strings.Contains(err.Error(), "use of closed") {
			log.Printf("[server] stopped: %v", err)
		}
	}()
	defer srv.GracefulStop()

	// Small pause so Serve() is ready
	time.Sleep(50 * time.Millisecond)

	// ── Create client ─────────────────────────────────────────────────────
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	must(err, "dial")
	defer conn.Close()

	client := pb.NewOrderServiceClient(conn)

	// ── Demo 1: Unary — CreateOrder ───────────────────────────────────────
	fmt.Println("\n── Demo 1: Unary CreateOrder ──")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp1, err := client.CreateOrder(ctx, &pb.CreateOrderRequest{
		CustomerID: "customer-alice",
		Items: []*pb.OrderItem{
			{ProductID: "laptop-pro", Quantity: 1, PriceUSD: 1299.99},
			{ProductID: "mouse-ergo", Quantity: 2, PriceUSD: 49.99},
		},
	})
	must(err, "CreateOrder")
	fmt.Printf("[client] Created: id=%s status=%s total=%.2f\n",
		resp1.OrderID, resp1.Status, resp1.TotalUSD)

	// Create a second order for alice
	resp2, err := client.CreateOrder(ctx, &pb.CreateOrderRequest{
		CustomerID: "customer-alice",
		Items: []*pb.OrderItem{
			{ProductID: "keyboard", Quantity: 1, PriceUSD: 149.99},
		},
	})
	must(err, "CreateOrder2")
	fmt.Printf("[client] Created: id=%s status=%s total=%.2f\n",
		resp2.OrderID, resp2.Status, resp2.TotalUSD)

	// ── Demo 2: Unary — GetOrder ──────────────────────────────────────────
	fmt.Println("\n── Demo 2: Unary GetOrder ──")
	got, err := client.GetOrder(ctx, &pb.GetOrderRequest{OrderID: resp1.OrderID})
	must(err, "GetOrder")
	fmt.Printf("[client] Got: id=%s customer=%s status=%s\n",
		got.OrderID, got.CustomerID, got.Status)

	// GetOrder — not found case
	_, err = client.GetOrder(ctx, &pb.GetOrderRequest{OrderID: "ORD-9999"})
	if st, ok := status.FromError(err); ok {
		fmt.Printf("[client] Expected error: code=%s message=%q\n", st.Code(), st.Message())
	}

	// ── Demo 3: Unary — CancelOrder ───────────────────────────────────────
	fmt.Println("\n── Demo 3: Unary CancelOrder ──")
	cancelled, err := client.CancelOrder(ctx, &pb.CancelOrderRequest{
		OrderID: resp1.OrderID,
		Reason:  "customer changed mind",
	})
	must(err, "CancelOrder")
	fmt.Printf("[client] Cancelled: id=%s status=%s\n", cancelled.OrderID, cancelled.Status)

	// Try to cancel again — should get FailedPrecondition (status is now CANCELLED, not allowed)
	// Let's create a "shipped" order to demo that specific case
	shippedOrder, _ := client.CreateOrder(ctx, &pb.CreateOrderRequest{
		CustomerID: "customer-bob",
		Items:      []*pb.OrderItem{{ProductID: "widget", Quantity: 1, PriceUSD: 9.99}},
	})
	// Manually set it to shipped in the store to demo the error path
	store.orders[shippedOrder.OrderID].Status = "SHIPPED"

	_, err = client.CancelOrder(ctx, &pb.CancelOrderRequest{
		OrderID: shippedOrder.OrderID,
		Reason:  "too late",
	})
	if st, ok := status.FromError(err); ok {
		fmt.Printf("[client] Expected FailedPrecondition: code=%s message=%q\n",
			st.Code(), st.Message())
	}

	// ── Demo 4: Server Streaming — ListOrders ─────────────────────────────
	fmt.Println("\n── Demo 4: Server Streaming ListOrders ──")
	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()

	stream, err := client.ListOrders(ctx2, &pb.ListOrdersRequest{CustomerID: "customer-alice"})
	must(err, "ListOrders")

	fmt.Printf("[client] Receiving orders for customer-alice:\n")
	count := 0
	for {
		order, err := stream.Recv()
		if err == io.EOF {
			break // server closed the stream
		}
		if err != nil {
			log.Printf("[client] stream error: %v", err)
			break
		}
		count++
		fmt.Printf("  [%d] id=%s status=%s total=%.2f\n",
			count, order.OrderID, order.Status, order.TotalUSD)
	}
	fmt.Printf("[client] Received %d orders from stream\n", count)

	// ── Demo 5: Input validation error ────────────────────────────────────
	fmt.Println("\n── Demo 5: Input Validation ──")
	_, err = client.CreateOrder(ctx, &pb.CreateOrderRequest{
		CustomerID: "", // intentionally empty
		Items:      []*pb.OrderItem{{ProductID: "x", Quantity: 1, PriceUSD: 1}},
	})
	if st, ok := status.FromError(err); ok {
		fmt.Printf("[client] InvalidArgument: code=%s message=%q\n", st.Code(), st.Message())
	}

	fmt.Println("\n=== All demos complete ===")
	fmt.Println()
	fmt.Println("Key takeaways:")
	fmt.Println("  1. gRPC uses HTTP/2 + binary encoding (protobuf) — faster than REST/JSON")
	fmt.Println("  2. The .proto file is your contract — client + server share the same types")
	fmt.Println("  3. Use status.Error(codes.X, msg) — not raw errors")
	fmt.Println("  4. Always set deadlines — context.WithTimeout propagates across calls")
	fmt.Println("  5. Streaming is first-class — server streaming reads like a range loop")
}

func must(err error, msg string) {
	if err != nil {
		log.Fatalf("[main] %s: %v", msg, err)
	}
}
