package controller

import (
	"context"
	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/pkg/client/grpc"
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
)

type clientManager struct {
	log  logger.Logger
	grpc *grpc.GRPCClient
}

func (c *clientManager) GRPC() *grpc.GRPCClient {
	return c.grpc
}

func newClientController(_ context.Context, log logger.Logger) toolkit.Client {
	cl := new(clientManager)
	cl.log = log
	return cl
}
