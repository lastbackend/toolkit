// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.12
// source: github.com/lastbackend/toolkit/examples/service/apis/example.proto

package servicepb

import (
	context "context"

	ptypes "github.com/lastbackend/toolkit/examples/service/gen/ptypes"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ExampleClient is the client API for Example service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ExampleClient interface {
	// Example methods
	HelloWorld(ctx context.Context, in *ptypes.HelloWorldRequest, opts ...grpc.CallOption) (*ptypes.HelloWorldResponse, error)
}

type exampleClient struct {
	cc grpc.ClientConnInterface
}

func NewExampleClient(cc grpc.ClientConnInterface) ExampleClient {
	return &exampleClient{cc}
}

func (c *exampleClient) HelloWorld(ctx context.Context, in *ptypes.HelloWorldRequest, opts ...grpc.CallOption) (*ptypes.HelloWorldResponse, error) {
	out := new(ptypes.HelloWorldResponse)
	err := c.cc.Invoke(ctx, "/lastbackend.example.Example/HelloWorld", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ExampleServer is the server API for Example service.
// All implementations should embed UnimplementedExampleServer
// for forward compatibility
type ExampleServer interface {
	// Example methods
	HelloWorld(context.Context, *ptypes.HelloWorldRequest) (*ptypes.HelloWorldResponse, error)
}

// UnimplementedExampleServer should be embedded to have forward compatible implementations.
type UnimplementedExampleServer struct{}

func (UnimplementedExampleServer) HelloWorld(context.Context, *ptypes.HelloWorldRequest) (*ptypes.HelloWorldResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HelloWorld not implemented")
}

// UnsafeExampleServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ExampleServer will
// result in compilation errors.
type UnsafeExampleServer interface {
	mustEmbedUnimplementedExampleServer()
}

func RegisterExampleServer(s grpc.ServiceRegistrar, srv ExampleServer) {
	s.RegisterService(&Example_ServiceDesc, srv)
}

func _Example_HelloWorld_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ptypes.HelloWorldRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExampleServer).HelloWorld(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lastbackend.example.Example/HelloWorld",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExampleServer).HelloWorld(ctx, req.(*ptypes.HelloWorldRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Example_ServiceDesc is the grpc.ServiceDesc for Example service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Example_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "lastbackend.example.Example",
	HandlerType: (*ExampleServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "HelloWorld",
			Handler:    _Example_HelloWorld_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "github.com/lastbackend/toolkit/examples/service/apis/example.proto",
}
