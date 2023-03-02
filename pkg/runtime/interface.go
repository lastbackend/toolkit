package runtime

import (
	"context"
	"github.com/caarlos0/env/v7"
	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
	"github.com/lastbackend/toolkit/pkg/runtime/meta"
)

type Runtime interface {
	Meta() *meta.Meta
	Log() logger.Logger

	Client() Client
	Package() Package
	Server() Server
	Config() Config
	Plugin() Plugin

	//Tools()

	Service() toolkit.Service

	Start(ctx context.Context, fn ...interface{}) error
}

type Client interface {
	toolkit.Client
}

type Config interface {
	toolkit.Config

	SetMeta(meta *meta.Meta)
	Parse(v interface{}, prefix string, opts ...env.Options) error
	Print(v interface{}, prefix string)

	Configs() []interface{}
}

type Server interface {
	toolkit.Server

	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	Provides() []interface{}
	Constructors() []interface{}
}

type Plugin interface {
	toolkit.Plugin

	Plugins() []interface{}

	Provide(interface{})
	Provides() []interface{}

	PreStart(ctx context.Context) error
	OnStart(ctx context.Context) error
	OnStop(ctx context.Context) error
}

type Package interface {
	toolkit.Package

	Provides() []interface{}
	Packages() []interface{}

	PreStart(ctx context.Context) error
	OnStart(ctx context.Context) error
	OnStop(ctx context.Context) error
}

//type Tools interface {
//	Metrics() metrics.Metrics
//	Probes() probes.Probes
//	Traces() trace.Trace
//}
