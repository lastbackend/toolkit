package controller

import (
	"context"
	"github.com/lastbackend/toolkit/pkg/client"
	"github.com/lastbackend/toolkit/pkg/client/grpc"
	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
)

type clientManager struct {
	log  logger.Logger
	grpc client.GRPCClient
	http client.HTTPClient
}

func (c *clientManager) GRPC() client.GRPCClient {
	return c.grpc
}

func (c *clientManager) HTTP() client.HTTPClient {
	return c.http
}

func newClientController(ctx context.Context, runtime runtime.Runtime) runtime.Client {
	cl := new(clientManager)

	cl.log = runtime.Log()
	cl.grpc = grpc.NewClient(ctx, runtime)

	return cl
}
