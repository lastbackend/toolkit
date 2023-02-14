// Code generated by protoc-gen-toolkit. DO NOT EDIT.
// source: github.com/lastbackend/toolkit/examples/wss/apis/server.proto

package serverpb

import (
	context "context"

	"github.com/lastbackend/toolkit/examples/helloworld/gen"
	"github.com/lastbackend/toolkit/examples/wss/gen/server"
	grpc "github.com/lastbackend/toolkit/pkg/client/grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// Suppress "imported and not used" errors
var _ context.Context
var _ emptypb.Empty

// Client gRPC API for Router service
func NewRouterRPCClient(service string, c grpc.Client) RouterRPCClient {
	return &routerGrpcRPCClient{service, c}
}

// Client gRPC API for Router service
type RouterRPCClient interface {
	Subscribe(ctx context.Context, req *SubscribeRequest, opts ...grpc.CallOption) (*SubscribeResponse, error)

	HelloWorld(ctx context.Context, req *servicepb.HelloRequest, opts ...grpc.CallOption) (*servicepb.HelloReply, error)
}

type routerGrpcRPCClient struct {
	service string
	cli     grpc.Client
}

func (c *routerGrpcRPCClient) Subscribe(ctx context.Context, req *SubscribeRequest, opts ...grpc.CallOption) (*SubscribeResponse, error) {
	resp := new(SubscribeResponse)
	if err := c.cli.Call(ctx, c.service, Router_SubscribeMethod, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *routerGrpcRPCClient) HelloWorld(ctx context.Context, req *servicepb.HelloRequest, opts ...grpc.CallOption) (*servicepb.HelloReply, error) {
	resp := new(servicepb.HelloReply)
	if err := c.cli.Call(ctx, c.service, Router_HelloWorldMethod, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (routerGrpcRPCClient) mustEmbedUnimplementedRouterClient() {}

// Client methods for Router service
const (
	Router_SubscribeMethod = "/gateway.Router/Subscribe"

	Router_HelloWorldMethod = "/gateway.Router/HelloWorld"
)