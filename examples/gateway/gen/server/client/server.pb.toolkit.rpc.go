// Code generated by protoc-gen-toolkit. DO NOT EDIT.
// source: github.com/lastbackend/toolkit/examples/gateway/apis/server.proto

package serverpb

import (
	context "context"

	"github.com/lastbackend/toolkit/examples/helloworld/gen"
	grpc "github.com/lastbackend/toolkit/pkg/client/grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// Suppress "imported and not used" errors
var _ context.Context
var _ emptypb.Empty

// Client gRPC API for ProxyGateway service
func NewProxyGatewayRPCClient(service string, c grpc.Client) ProxyGatewayRPCClient {
	return &proxygatewayGrpcRPCClient{service, c}
}

// Client gRPC API for ProxyGateway service
type ProxyGatewayRPCClient interface {
	HelloWorld(ctx context.Context, req *servicepb.HelloRequest, opts ...grpc.CallOption) (*servicepb.HelloReply, error)
}

type proxygatewayGrpcRPCClient struct {
	service string
	cli     grpc.Client
}

func (c *proxygatewayGrpcRPCClient) HelloWorld(ctx context.Context, req *servicepb.HelloRequest, opts ...grpc.CallOption) (*servicepb.HelloReply, error) {
	resp := new(servicepb.HelloReply)
	if err := c.cli.Call(ctx, c.service, ProxyGateway_HelloWorldMethod, req, resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

func (proxygatewayGrpcRPCClient) mustEmbedUnimplementedProxyGatewayClient() {}

// Client methods for ProxyGateway service
const (
	ProxyGateway_HelloWorldMethod = "/gateway.ProxyGateway/HelloWorld"
)