package controller

import (
	"context"
	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
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

	app  fx.App
	meta *meta.Meta

	service toolkit.Service

	logger logger.Logger

	client runtime.Client
	config runtime.Config
	server runtime.Server
	plugin runtime.Plugin
	pkg    runtime.Package

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

	opts = append(opts, fx.Invoke(func(ctx context.Context) error {
		return c.Plugin().PreStart(ctx)
	}))

	for _, f := range fn {
		opts = append(opts, fx.Invoke(f))
	}

	opts = append(opts, fx.Invoke(func(ctx context.Context) error {
		return c.Package().PreStart(ctx)
	}))

	// get constructors from servers
	constructors := c.Server().Constructors()
	for _, c := range constructors {
		opts = append(opts, fx.Invoke(c))
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

	fx.New(
		fx.Options(opts...),
		fx.Options(
			fx.Provide(func() logger.Logger { return c.logger }),
		),
		//fx.NopLogger,
	).Run()

	return nil
}

func (c *controller) onStart(ctx context.Context) error {
	c.Plugin().OnStart(ctx)
	c.Package().OnStart(ctx)

	if err := c.Tools().OnStart(ctx); err != nil {
		return err
	}
	return c.Server().Start(ctx)
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

func (c *controller) Tools() runtime.Tools {
	return c.tools
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

	rt.meta = new(meta.Meta)
	rt.meta.SetName(name)
	rt.fillMeta(opts...)

	rt.logger = logger.DefaultLogger

	rt.client = newClientController(ctx, rt.logger)
	rt.config = newConfigController(ctx, rt.logger)
	rt.plugin = newPluginController(ctx, rt.logger)
	rt.pkg = newPackageController(ctx, rt.logger)

	rt.server = newServerController(ctx, rt, rt.logger)

	rt.config.SetMeta(rt.meta)

	logOpts := logger.Options{}
	if err := rt.config.Parse(&logOpts, ""); err != nil {
		return nil, err
	}
	rt.logger.Init(logOpts)
	rt.logger.WithFields(logger.Fields{
		"microservice": name,
	})

	svc := new(service)
	svc.runtime = rt
	rt.service = svc

	if rt.tools, err = newToolsRegistration(rt); err != nil {
		return nil, err
	}

	return rt, nil
}
