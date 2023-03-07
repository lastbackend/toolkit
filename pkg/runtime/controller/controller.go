package controller

import (
	"context"
	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
	zp "github.com/lastbackend/toolkit/pkg/runtime/logger/zap"
	"github.com/lastbackend/toolkit/pkg/runtime/meta"
	"go.uber.org/fx"
	"os"
	"syscall"
)

var shutdownSignals = []os.Signal{
	syscall.SIGTERM,
	syscall.SIGINT,
	syscall.SIGQUIT,
	syscall.SIGKILL,
}

type controller struct {
	runtime.Runtime

	app  *fx.App
	meta *meta.Meta

	service toolkit.Service

	logger logger.Logger

	client runtime.Client
	config runtime.Config
	server runtime.Server
	plugin runtime.Plugin
	pkg    runtime.Package

	providers []interface{}
	invokes   []interface{}

	tools runtime.Tools
}

func (c *controller) Service() toolkit.Service {
	return c.service
}

func (c *controller) Meta() *meta.Meta {
	return c.meta
}

func (c *controller) Log() logger.Logger {
	return c.logger
}

func (c *controller) Start(ctx context.Context, fn ...interface{}) error {

	if c.help() {
		return nil
	}

	return c.start(ctx, fn...)
}

func (c *controller) start(ctx context.Context, fn ...interface{}) error {
	c.Log().Info("controller start")

	opts := make([]fx.Option, 0)
	opts = append(opts, fx.Provide(
		fx.Annotate(
			func() runtime.Runtime {
				return c
			},
		)),
		fx.Provide(
			fx.Annotate(
				func() toolkit.Service {
					return c.Service()
				},
			),
		),
		fx.Provide(func() context.Context { return ctx }))

	cfgs := c.Config().Configs()
	for _, c := range cfgs {
		opts = append(opts, fx.Supply(c))
	}

	for _, p := range c.providers {
		opts = append(opts, fx.Provide(p))
	}

	plugins := c.Plugin().Provides()
	for _, p := range plugins {
		opts = append(opts, fx.Provide(p))
	}

	packages := c.Package().Provides()
	for _, p := range packages {
		opts = append(opts, fx.Provide(p))
	}

	// get provides from servers
	servers := c.Server().Provides()
	for _, s := range servers {
		opts = append(opts, fx.Provide(s))
	}

	opts = append(opts, fx.Provide(func() logger.Logger { return c.logger }))

	opts = append(opts, fx.Invoke(func(ctx context.Context) error {
		return c.Plugin().PreStart(ctx)
	}))

	opts = append(opts, fx.Invoke(func(ctx context.Context) error {
		return c.Package().PreStart(ctx)
	}))

	for _, p := range c.invokes {
		opts = append(opts, fx.Invoke(p))
	}

	// get constructors from servers
	constructors := c.Server().Constructors()
	for _, c := range constructors {
		opts = append(opts, fx.Invoke(c))
	}

	opts = append(opts, fx.Invoke(func(ctx context.Context) error {
		if err := c.Tools().OnStart(ctx); err != nil {
			return err
		}
		c.Server().Start(ctx)
		c.Plugin().OnStart(ctx)
		return nil
	}))

	for _, f := range fn {
		opts = append(opts, fx.Invoke(f))
	}

	opts = append(opts, fx.Invoke(func(lc fx.Lifecycle) error {

		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return c.onStart(ctx)
			},
			OnStop: func(ctx context.Context) error {
				return c.onStop(ctx)
			},
		})
		return nil
	}))

	c.app = fx.New(
		fx.Options(opts...),
		fx.WithLogger(c.logger.Fx),
	)
  c.app.Run()

	return nil
}

func (c *controller) onStart(ctx context.Context) error {
	c.client.GRPC().GetResolver().OnStart(ctx)
	c.Package().OnStart(ctx)
	return nil
}

func (c *controller) onStop(ctx context.Context) error {
	c.Plugin().OnStop(ctx)
	c.Package().OnStop(ctx)
	return c.Server().Stop(ctx)
}

func (c *controller) Config() runtime.Config {
	return c.config
}

func (c *controller) Plugin() runtime.Plugin {
	return c.plugin
}

func (c *controller) Package() runtime.Package {
	return c.pkg
}

func (c *controller) Client() runtime.Client {
	return c.client
}

func (c *controller) Server() runtime.Server {
	return c.server
}

func (c *controller) Provide(constructor interface{}) {
	c.providers = append(c.providers, constructor)
}

func (c *controller) Invoke(constructor interface{}) {
	c.invokes = append(c.invokes, constructor)
}

func (c *controller) Tools() runtime.Tools {
	return c.tools
}

func (c *controller) Stop(ctx context.Context) error {
	return c.app.Stop(ctx)
}

func (c *controller) fillMeta(opts ...runtime.Option) {
	for _, opt := range opts {
		switch opt.Name() {
		case runtime.MetaOptionEnvPrefix:
			c.meta.SetEnvPrefix(opt.Value())
		case runtime.MetaOptionVersion:
			c.meta.SetVersion(opt.Value())
		case runtime.MetaOptionDescription:
			c.meta.SetDescription(opt.Value())
		}
	}
}

func NewRuntime(ctx context.Context, name string, opts ...runtime.Option) (runtime.Runtime, error) {

	var (
		rt  = new(controller)
		err error
	)

	rt.providers = make([]interface{}, 0)
	rt.invokes = make([]interface{}, 0)

	rt.meta = new(meta.Meta)
	rt.meta.SetName(name)
	rt.fillMeta(opts...)

	rt.config = newConfigController(ctx, rt)
	rt.config.SetMeta(rt.meta)

	rt.logger = zp.NewLogger(rt, logger.Fields{
		"microservice": name,
	})

	rt.client = newClientController(ctx, rt)
	rt.plugin = newPluginController(ctx, rt)
	rt.pkg = newPackageController(ctx, rt)
	rt.server = newServerController(ctx, rt)

	svc := new(service)
	svc.runtime = rt
	rt.service = svc

	if rt.tools, err = newToolsRegistration(rt); err != nil {
		return nil, err
	}

	return rt, nil
}
