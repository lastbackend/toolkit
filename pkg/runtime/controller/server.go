package controller

import (
	"context"
	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
	"github.com/lastbackend/toolkit/pkg/server"
	"github.com/lastbackend/toolkit/pkg/server/grpc"
	"github.com/lastbackend/toolkit/pkg/server/http"
	"go.uber.org/fx"
)

type serverManager struct {
	runtime.Server

	runtime runtime.Runtime
	log     logger.Logger

	http map[string]server.HTTPServer
	grpc map[string]server.GRPCServer
}

func (c *serverManager) HTTP() server.HTTPServer {
	return c.http[c.runtime.Meta().GetName()]
}

func (c *serverManager) GRPC() server.GRPCServer {
	return c.grpc[c.runtime.Meta().GetName()]
}

func (c *serverManager) HTTPGet(name string) server.HTTPServer {
	return c.http[name]
}

func (c *serverManager) HTTPNew(name string, options *server.HTTPServerOptions) server.HTTPServer {

	if name == "" {
		name = c.runtime.Meta().GetName()
	}

	srv := http.NewServer(name, c.runtime, options)
	c.http[name] = srv
	return c.http[name]
}

func (c *serverManager) HTTPList() map[string]server.HTTPServer {
	return c.http
}

func (c *serverManager) GRPCGet(name string) server.GRPCServer {
	return c.grpc[name]
}

func (c *serverManager) GRPCNew(name string, options *server.GRPCServerOptions) server.GRPCServer {
	srv := grpc.NewServer(c.runtime, name, options)
	c.grpc[name] = srv
	return c.grpc[name]
}

func (c *serverManager) GRPCList() map[string]server.GRPCServer {
	return c.grpc
}

func (c *serverManager) Provides() []interface{} {

	provides := make([]interface{}, 0)

	for _, s := range c.grpc {
		service := s.GetService()
		if service != nil {
			provides = append(provides, service)
		}
	}

	for _, s := range c.http {
		service := s.GetService()
		if service != nil {
			provides = append(provides, service)
		}

		middlewares := s.GetMiddlewares()
		for _, middleware := range middlewares {
			provides = append(provides, fx.Annotate(
				middleware,
				fx.As(new(server.HttpServerMiddleware)),
				fx.ResultTags(`group:"middlewares"`),
			),
			)
		}

	}

	return provides
}

func (c *serverManager) Constructors() []interface{} {

	provides := make([]interface{}, 0)

	for _, s := range c.grpc {
		provides = append(provides, s.GetConstructor())
	}

	for _, s := range c.http {

		provides = append(provides, fx.Annotate(
			s.GetConstructor(),
			fx.ParamTags(`group:"middlewares"`)))
	}

	return provides
}

func (c *serverManager) Start(ctx context.Context) error {

	for _, s := range c.grpc {
		if err := s.Start(ctx); err != nil {
			return err
		}
	}

	for _, s := range c.http {
		if err := s.Start(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (c *serverManager) Stop(_ context.Context) error {

	for _, s := range c.grpc {
		if err := s.Stop(); err != nil {
			return err
		}
	}

	for _, s := range c.http {
		if err := s.Stop(); err != nil {
			return err
		}
	}

	return nil
}

func newServerController(_ context.Context, runtime runtime.Runtime) runtime.Server {

	pl := new(serverManager)

	pl.runtime = runtime
	pl.log = runtime.Log()

	pl.http = make(map[string]server.HTTPServer, 0)
	pl.grpc = make(map[string]server.GRPCServer, 0)

	return pl
}
