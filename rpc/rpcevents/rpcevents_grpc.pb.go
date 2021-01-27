// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package rpcevents

import (
	context "context"

	exec "github.com/hyperledger/burrow/execution/exec"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// ExecutionEventsClient is the client API for ExecutionEvents service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ExecutionEventsClient interface {
	// Get StreamEvents (including transactions) for a range of block heights
	Stream(ctx context.Context, in *BlocksRequest, opts ...grpc.CallOption) (ExecutionEvents_StreamClient, error)
	// Get a particular TxExecution by hash
	Tx(ctx context.Context, in *TxRequest, opts ...grpc.CallOption) (*exec.TxExecution, error)
	// GetEvents provides events streaming one block at a time - that is all events emitted in a particular block
	// are guaranteed to be delivered in each GetEventsResponse
	Events(ctx context.Context, in *BlocksRequest, opts ...grpc.CallOption) (ExecutionEvents_EventsClient, error)
}

type executionEventsClient struct {
	cc grpc.ClientConnInterface
}

func NewExecutionEventsClient(cc grpc.ClientConnInterface) ExecutionEventsClient {
	return &executionEventsClient{cc}
}

func (c *executionEventsClient) Stream(ctx context.Context, in *BlocksRequest, opts ...grpc.CallOption) (ExecutionEvents_StreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &_ExecutionEvents_serviceDesc.Streams[0], "/rpcevents.ExecutionEvents/Stream", opts...)
	if err != nil {
		return nil, err
	}
	x := &executionEventsStreamClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type ExecutionEvents_StreamClient interface {
	Recv() (*exec.StreamEvent, error)
	grpc.ClientStream
}

type executionEventsStreamClient struct {
	grpc.ClientStream
}

func (x *executionEventsStreamClient) Recv() (*exec.StreamEvent, error) {
	m := new(exec.StreamEvent)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *executionEventsClient) Tx(ctx context.Context, in *TxRequest, opts ...grpc.CallOption) (*exec.TxExecution, error) {
	out := new(exec.TxExecution)
	err := c.cc.Invoke(ctx, "/rpcevents.ExecutionEvents/Tx", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *executionEventsClient) Events(ctx context.Context, in *BlocksRequest, opts ...grpc.CallOption) (ExecutionEvents_EventsClient, error) {
	stream, err := c.cc.NewStream(ctx, &_ExecutionEvents_serviceDesc.Streams[1], "/rpcevents.ExecutionEvents/Events", opts...)
	if err != nil {
		return nil, err
	}
	x := &executionEventsEventsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type ExecutionEvents_EventsClient interface {
	Recv() (*EventsResponse, error)
	grpc.ClientStream
}

type executionEventsEventsClient struct {
	grpc.ClientStream
}

func (x *executionEventsEventsClient) Recv() (*EventsResponse, error) {
	m := new(EventsResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ExecutionEventsServer is the server API for ExecutionEvents service.
// All implementations must embed UnimplementedExecutionEventsServer
// for forward compatibility
type ExecutionEventsServer interface {
	// Get StreamEvents (including transactions) for a range of block heights
	Stream(*BlocksRequest, ExecutionEvents_StreamServer) error
	// Get a particular TxExecution by hash
	Tx(context.Context, *TxRequest) (*exec.TxExecution, error)
	// GetEvents provides events streaming one block at a time - that is all events emitted in a particular block
	// are guaranteed to be delivered in each GetEventsResponse
	Events(*BlocksRequest, ExecutionEvents_EventsServer) error
	mustEmbedUnimplementedExecutionEventsServer()
}

// UnimplementedExecutionEventsServer must be embedded to have forward compatible implementations.
type UnimplementedExecutionEventsServer struct {
}

func (UnimplementedExecutionEventsServer) Stream(*BlocksRequest, ExecutionEvents_StreamServer) error {
	return status.Errorf(codes.Unimplemented, "method Stream not implemented")
}
func (UnimplementedExecutionEventsServer) Tx(context.Context, *TxRequest) (*exec.TxExecution, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Tx not implemented")
}
func (UnimplementedExecutionEventsServer) Events(*BlocksRequest, ExecutionEvents_EventsServer) error {
	return status.Errorf(codes.Unimplemented, "method Events not implemented")
}
func (UnimplementedExecutionEventsServer) mustEmbedUnimplementedExecutionEventsServer() {}

// UnsafeExecutionEventsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ExecutionEventsServer will
// result in compilation errors.
type UnsafeExecutionEventsServer interface {
	mustEmbedUnimplementedExecutionEventsServer()
}

func RegisterExecutionEventsServer(s grpc.ServiceRegistrar, srv ExecutionEventsServer) {
	s.RegisterService(&_ExecutionEvents_serviceDesc, srv)
}

func _ExecutionEvents_Stream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(BlocksRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ExecutionEventsServer).Stream(m, &executionEventsStreamServer{stream})
}

type ExecutionEvents_StreamServer interface {
	Send(*exec.StreamEvent) error
	grpc.ServerStream
}

type executionEventsStreamServer struct {
	grpc.ServerStream
}

func (x *executionEventsStreamServer) Send(m *exec.StreamEvent) error {
	return x.ServerStream.SendMsg(m)
}

func _ExecutionEvents_Tx_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TxRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExecutionEventsServer).Tx(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpcevents.ExecutionEvents/Tx",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExecutionEventsServer).Tx(ctx, req.(*TxRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ExecutionEvents_Events_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(BlocksRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ExecutionEventsServer).Events(m, &executionEventsEventsServer{stream})
}

type ExecutionEvents_EventsServer interface {
	Send(*EventsResponse) error
	grpc.ServerStream
}

type executionEventsEventsServer struct {
	grpc.ServerStream
}

func (x *executionEventsEventsServer) Send(m *EventsResponse) error {
	return x.ServerStream.SendMsg(m)
}

var _ExecutionEvents_serviceDesc = grpc.ServiceDesc{
	ServiceName: "rpcevents.ExecutionEvents",
	HandlerType: (*ExecutionEventsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Tx",
			Handler:    _ExecutionEvents_Tx_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Stream",
			Handler:       _ExecutionEvents_Stream_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Events",
			Handler:       _ExecutionEvents_Events_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "rpcevents.proto",
}