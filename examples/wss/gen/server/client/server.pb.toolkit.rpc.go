// Code generated by protoc-gen-toolkit. DO NOT EDIT.
// source: github.com/lastbackend/toolkit/examples/wss/apis/server.proto

package serverpb

import (
	context "context"

	"github.com/lastbackend/toolkit/examples/helloworld/gen"
	client "github.com/lastbackend/toolkit/pkg/client"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// Suppress "imported and not used" errors
var _ context.Context
var _ emptypb.Empty

// Client gRPC API for Router service
func NewRouterRPCClient(service string, c client.GRPCClient) RouterRPCClient {
	return &routerGrpcRPCClient{service, c}
}

// Client gRPC API for Router service
type RouterRPCClient interface {
	SayHello(ctx context.Context, req *servicepb.HelloRequest, opts ...client.GRPCCallOption) (*servicepb.HelloReply, error)

	HelloWorld(ctx context.Context, req *servicepb.HelloRequest, opts ...client.GRPCCallOption) (*servicepb.HelloReply, error)
}

type routerGrpcRPCClient struct {
	service string
	cli     client.GRPCClient
}

func (c *routerGrpcRPCClient) SayHello(ctx context.Context, req *servicepb.HelloRequest, opts ...client.GRPCCallOption) (*servicepb.HelloReply, error) {
	resp := new(servicepb.HelloReply)
	if err := c.cli.Call(ctx, c.service, Router_SayHelloMethod, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *routerGrpcRPCClient) HelloWorld(ctx context.Context, req *servicepb.HelloRequest, opts ...client.GRPCCallOption) (*servicepb.HelloReply, error) {
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

	Router_SayHelloMethod = "/gateway.Router/SayHello"

	Router_HelloWorldMethod = "/gateway.Router/HelloWorld"
)
