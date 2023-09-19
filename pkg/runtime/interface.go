package runtime

import (
  "context"
  "github.com/caarlos0/env/v7"
  "github.com/lastbackend/toolkit"
  "github.com/lastbackend/toolkit/pkg/client"
  "github.com/lastbackend/toolkit/pkg/runtime/logger"
  "github.com/lastbackend/toolkit/pkg/runtime/meta"
  "github.com/lastbackend/toolkit/pkg/server"
  "github.com/lastbackend/toolkit/pkg/tools/metrics"
  "github.com/lastbackend/toolkit/pkg/tools/probes"
  "github.com/lastbackend/toolkit/pkg/tools/traces"
)

type Runtime interface {
  Meta() *meta.Meta
  Log() logger.Logger

  Client() Client
  Package() Package
  Server() Server
  Config() Config
  Plugin() Plugin

  Tools() Tools

  Service() toolkit.Service

  Provide(constructor interface{})
  Invoke(constructor interface{})

  Start(ctx context.Context) error
  Stop(ctx context.Context, err error)

  RegisterOnStartHook(fn ...func(ctx context.Context) error)
  RegisterOnStartSyncHook(fn ...func(ctx context.Context) error)

  RegisterOnStopHook(fn ...func(ctx context.Context) error)
  RegisterOnStopSyncHook(fn ...func(ctx context.Context) error)
}

type Client interface {
  GRPC() client.GRPCClient
  HTTP() client.HTTPClient
}

type HTTPClient interface {
  Get() error
  Post() error
}

type Server interface {
  toolkit.Server

  Start(ctx context.Context) error
  Stop(ctx context.Context) error

  HTTPList() map[string]server.HTTPServer
  GRPCList() map[string]server.GRPCServer

  Provides() []interface{}
  Constructors() []interface{}
}

type Config interface {
  Provide(...any) error

  SetMeta(meta *meta.Meta)
  Parse(v interface{}, prefix string, opts ...env.Options) error
  Print(v interface{}, prefix string)
  PrintTable(all, nocomments bool) string
  PrintYaml(all, nocomments bool) string

  Configs() []any
}

type Plugin interface {
  Provide(constructor ...interface{})

  Constructors() []interface{}
  Register(plugins []toolkit.Plugin)

  PreStart(ctx context.Context) error
  OnStart(ctx context.Context) error
  OnStop(ctx context.Context) error
}

type Package interface {
  Provide(constructor ...interface{})

  Constructors() []interface{}
  Register(packages []toolkit.PackageItem)

  PreStart(ctx context.Context) error
  OnStart(ctx context.Context) error
  OnStop(ctx context.Context) error
}

type Tools interface {
  OnStart(ctx context.Context) error

  Metrics() metrics.Metrics
  Probes() probes.Probes
  Traces() traces.Traces
}
