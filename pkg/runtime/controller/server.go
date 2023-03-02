package controller

import (
	"context"
	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
	"github.com/lastbackend/toolkit/pkg/server"
	"github.com/lastbackend/toolkit/pkg/server/grpc"
	"github.com/lastbackend/toolkit/pkg/server/http"
)

const DefaultGRPCServer = "default"

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
	srv := http.NewServer(c.runtime, options)
	c.http[name] = srv
	return c.http[name]
}

func (c *serverManager) GRPCGet(name string) server.GRPCServer {
	return c.grpc[name]
}

func (c *serverManager) GRPCNew(name string, options *server.GRPCServerOptions) server.GRPCServer {
	srv := grpc.NewServer(c.runtime, options)
	c.grpc[name] = srv
	return c.grpc[name]
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

	}

	return provides
}

func (c *serverManager) Constructors() []interface{} {

	provides := make([]interface{}, 0)

	for _, s := range c.grpc {
		provides = append(provides, s.GetConstructor())
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

func (c *serverManager) Stop(ctx context.Context) error {

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

func newServerController(_ context.Context, runtime runtime.Runtime, log logger.Logger) runtime.Server {

	pl := new(serverManager)

	pl.runtime = runtime
	pl.log = log

	pl.http = make(map[string]server.HTTPServer, 0)
	pl.grpc = make(map[string]server.GRPCServer, 0)
	return pl
}
