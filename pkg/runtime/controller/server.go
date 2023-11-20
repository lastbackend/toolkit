package controller

import (
  "context"
  "github.com/lastbackend/toolkit/pkg/runtime"
  "github.com/lastbackend/toolkit/pkg/runtime/logger"
  "github.com/lastbackend/toolkit/pkg/server"
  "github.com/lastbackend/toolkit/pkg/server/grpc"
  "github.com/lastbackend/toolkit/pkg/server/http"
  "go.uber.org/fx"
  "regexp"
)

type serverManager struct {
  runtime.Server

  runtime runtime.Runtime
  log     logger.Logger

  http map[string]server.HTTPServer
  grpc map[string]server.GRPCServer
}

func (c *serverManager) HTTP() server.HTTPServer {
  slug := getSlug(c.runtime.Meta().GetName())
  return c.http[slug]
}

func (c *serverManager) GRPC() server.GRPCServer {
  slug := getSlug(c.runtime.Meta().GetName())
  return c.grpc[slug]
}

func (c *serverManager) HTTPGet(name string) server.HTTPServer {
  slug := getSlug(name)
  return c.http[slug]
}

func (c *serverManager) HTTPNew(name string, options *server.HTTPServerOptions) server.HTTPServer {

  if name == "" {
    name = c.runtime.Meta().GetSlug()
  }

  srv := http.NewServer(name, c.runtime, options)
  slug := getSlug(name)
  c.http[slug] = srv
  return c.http[slug]
}

func (c *serverManager) HTTPList() map[string]server.HTTPServer {
  return c.http
}

func (c *serverManager) GRPCGet(name string) server.GRPCServer {
  slug := getSlug(name)
  return c.grpc[slug]
}

func (c *serverManager) GRPCNew(name string, options *server.GRPCServerOptions) server.GRPCServer {
  if name == "" {
    name = c.runtime.Meta().GetSlug()
  }
  srv := grpc.NewServer(c.runtime, name, options)
  slug := getSlug(name)
  c.grpc[slug] = srv
  return c.grpc[slug]
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

      interceptors := s.GetInterceptors()
      for _, interceptor := range interceptors {
        provides = append(provides, fx.Annotate(
          interceptor,
          fx.As(new(server.GRPCInterceptor)),
          fx.ResultTags(`group:"interceptors"`),
        ),
        )
      }

    }
  }

  for _, s := range c.http {
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
    provides = append(provides, fx.Annotate(
      s.GetInterceptorsConstructor(),
      fx.ParamTags(`group:"interceptors"`)))
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
      c.log.V(5).Errorf("GRPC server start failed: %v", err)
      return err
    }
  }

  for _, s := range c.http {
    if err := s.Start(ctx); err != nil {
      c.log.V(5).Errorf("HTTP server start failed: %v", err)
      return err
    }
  }

  return nil
}

func (c *serverManager) Stop(ctx context.Context) error {

  for _, s := range c.http {
    if err := s.Stop(ctx); err != nil {
      return err
    }
  }

  for _, s := range c.grpc {
    if err := s.Stop(); err != nil {
      return err
    }
  }

  return nil
}

func getSlug(s string) string {
  return regexp.MustCompile(`[^_a-zA-Z0-9 ]+`).ReplaceAllString(s, "_")
}

func newServerController(_ context.Context, runtime runtime.Runtime) runtime.Server {

  pl := new(serverManager)

  pl.runtime = runtime
  pl.log = runtime.Log()

  pl.http = make(map[string]server.HTTPServer, 0)
  pl.grpc = make(map[string]server.GRPCServer, 0)

  return pl
}
