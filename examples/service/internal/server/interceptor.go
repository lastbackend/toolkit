package server

import (
	"context"
	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/server"
	"google.golang.org/grpc"
	"time"
)

type ExampleGRPCServerInterceptor struct {
	server.GRPCInterceptor
	name server.KindInterceptor
	r    runtime.Runtime
}

func NewExampleGRPCServerInterceptor(r runtime.Runtime) server.GRPCInterceptor {
	i := ExampleGRPCServerInterceptor{r: r}
	return &i
}

func (i *ExampleGRPCServerInterceptor) Kind() server.KindInterceptor {
	return i.name
}

func (i *ExampleGRPCServerInterceptor) Interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	i.r.Log().Info("Log: 1")
	time.Sleep(time.Second)

	i.r.Log().Info("Logging")

	res, err := handler(ctx, req)

	i.r.Log().Info("Log: 2")
	time.Sleep(time.Second)

	return res, err
}
