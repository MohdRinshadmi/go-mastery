// Code equivalent to what protoc --go-grpc_out would generate for order.proto.
// Hand-authored — semantically identical to protoc output.
//
// This file defines:
//   - OrderServiceClient  (what callers use)
//   - OrderServiceServer  (what implementations must satisfy)
//   - registration helper RegisterOrderServiceServer
//   - stream types for server streaming

package pb

import (
	"context"

	"google.golang.org/grpc"
)

// ── Service description ──────────────────────────────────────────────────

// OrderService_ServiceDesc is the grpc.ServiceDesc for OrderService.
// Registered with grpc.Server so the framework routes incoming RPCs.
var OrderService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "order.OrderService",
	HandlerType: (*OrderServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateOrder",
			Handler:    _OrderService_CreateOrder_Handler,
		},
		{
			MethodName: "GetOrder",
			Handler:    _OrderService_GetOrder_Handler,
		},
		{
			MethodName: "CancelOrder",
			Handler:    _OrderService_CancelOrder_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ListOrders",
			Handler:       _OrderService_ListOrders_Handler,
			ServerStreams: true,
		},
	},
}

// ── Server interface ─────────────────────────────────────────────────────

// OrderServiceServer must be implemented by the service you write.
// Embed UnimplementedOrderServiceServer to get forward-compatible defaults.
type OrderServiceServer interface {
	CreateOrder(context.Context, *CreateOrderRequest) (*OrderResponse, error)
	GetOrder(context.Context, *GetOrderRequest) (*OrderResponse, error)
	CancelOrder(context.Context, *CancelOrderRequest) (*OrderResponse, error)
	ListOrders(*ListOrdersRequest, OrderService_ListOrdersServer) error
	mustEmbedUnimplementedOrderServiceServer()
}

// UnimplementedOrderServiceServer — embed this in your implementation.
// Future proto additions won't break your code.
type UnimplementedOrderServiceServer struct{}

func (UnimplementedOrderServiceServer) CreateOrder(context.Context, *CreateOrderRequest) (*OrderResponse, error) {
	return nil, nil
}
func (UnimplementedOrderServiceServer) GetOrder(context.Context, *GetOrderRequest) (*OrderResponse, error) {
	return nil, nil
}
func (UnimplementedOrderServiceServer) CancelOrder(context.Context, *CancelOrderRequest) (*OrderResponse, error) {
	return nil, nil
}
func (UnimplementedOrderServiceServer) ListOrders(*ListOrdersRequest, OrderService_ListOrdersServer) error {
	return nil
}
func (UnimplementedOrderServiceServer) mustEmbedUnimplementedOrderServiceServer() {}

// UnsafeOrderServiceServer — opt-in to skip compatibility checks (not recommended).
type UnsafeOrderServiceServer interface {
	mustEmbedUnimplementedOrderServiceServer()
}

// ── Registration helper ──────────────────────────────────────────────────

// RegisterOrderServiceServer wires your implementation into the gRPC server.
func RegisterOrderServiceServer(s grpc.ServiceRegistrar, srv OrderServiceServer) {
	s.RegisterService(&OrderService_ServiceDesc, srv)
}

// ── RPC handlers (called by the gRPC framework) ──────────────────────────

func _OrderService_CreateOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateOrderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrderServiceServer).CreateOrder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/order.OrderService/CreateOrder",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrderServiceServer).CreateOrder(ctx, req.(*CreateOrderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OrderService_GetOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetOrderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrderServiceServer).GetOrder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/order.OrderService/GetOrder",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrderServiceServer).GetOrder(ctx, req.(*GetOrderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OrderService_CancelOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CancelOrderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrderServiceServer).CancelOrder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/order.OrderService/CancelOrder",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrderServiceServer).CancelOrder(ctx, req.(*CancelOrderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OrderService_ListOrders_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ListOrdersRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(OrderServiceServer).ListOrders(m, &orderServiceListOrdersServer{stream})
}

// ── Stream types ──────────────────────────────────────────────────────────

// OrderService_ListOrdersServer — implemented by the gRPC framework.
// Your server implementation calls Send() for each response in the stream.
type OrderService_ListOrdersServer interface {
	Send(*OrderResponse) error
	grpc.ServerStream
}

type orderServiceListOrdersServer struct {
	grpc.ServerStream
}

func (x *orderServiceListOrdersServer) Send(m *OrderResponse) error {
	return x.ServerStream.SendMsg(m)
}

// ── Client interface ──────────────────────────────────────────────────────

// OrderServiceClient is what your caller code uses.
type OrderServiceClient interface {
	CreateOrder(ctx context.Context, in *CreateOrderRequest, opts ...grpc.CallOption) (*OrderResponse, error)
	GetOrder(ctx context.Context, in *GetOrderRequest, opts ...grpc.CallOption) (*OrderResponse, error)
	CancelOrder(ctx context.Context, in *CancelOrderRequest, opts ...grpc.CallOption) (*OrderResponse, error)
	ListOrders(ctx context.Context, in *ListOrdersRequest, opts ...grpc.CallOption) (OrderService_ListOrdersClient, error)
}

type orderServiceClient struct {
	cc grpc.ClientConnInterface
}

// NewOrderServiceClient creates a client stub from a gRPC connection.
func NewOrderServiceClient(cc grpc.ClientConnInterface) OrderServiceClient {
	return &orderServiceClient{cc}
}

func (c *orderServiceClient) CreateOrder(ctx context.Context, in *CreateOrderRequest, opts ...grpc.CallOption) (*OrderResponse, error) {
	out := new(OrderResponse)
	err := c.cc.Invoke(ctx, "/order.OrderService/CreateOrder", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *orderServiceClient) GetOrder(ctx context.Context, in *GetOrderRequest, opts ...grpc.CallOption) (*OrderResponse, error) {
	out := new(OrderResponse)
	err := c.cc.Invoke(ctx, "/order.OrderService/GetOrder", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *orderServiceClient) CancelOrder(ctx context.Context, in *CancelOrderRequest, opts ...grpc.CallOption) (*OrderResponse, error) {
	out := new(OrderResponse)
	err := c.cc.Invoke(ctx, "/order.OrderService/CancelOrder", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *orderServiceClient) ListOrders(ctx context.Context, in *ListOrdersRequest, opts ...grpc.CallOption) (OrderService_ListOrdersClient, error) {
	stream, err := c.cc.NewStream(ctx, &OrderService_ServiceDesc.Streams[0], "/order.OrderService/ListOrders", opts...)
	if err != nil {
		return nil, err
	}
	x := &orderServiceListOrdersClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// OrderService_ListOrdersClient — returned by the client's ListOrders call.
type OrderService_ListOrdersClient interface {
	Recv() (*OrderResponse, error)
	grpc.ClientStream
}

type orderServiceListOrdersClient struct {
	grpc.ClientStream
}

func (x *orderServiceListOrdersClient) Recv() (*OrderResponse, error) {
	m := new(OrderResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}
