package controller

import (
	"context"
	"github.com/common-nighthawk/go-figure"
	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
	zp "github.com/lastbackend/toolkit/pkg/runtime/logger/zap"
	"github.com/lastbackend/toolkit/pkg/runtime/meta"
	"go.uber.org/fx"
	"os"
	"os/signal"
	"syscall"
)

var shutdownSignals = []os.Signal{
	syscall.SIGTERM,
	syscall.SIGINT,
	syscall.SIGQUIT,
	syscall.SIGKILL,
}

type controller struct {
	fx.Out
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

	onStartHook []func(ctx context.Context) error
	onStopHook  []func(ctx context.Context) error

	onStartSyncHook []func(ctx context.Context) error
	onStopSyncHook  []func(ctx context.Context) error

	tools runtime.Tools
	done  chan error
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

func (c *controller) Start(ctx context.Context) error {
	if c.help() {
		return nil
	}

	return c.start(ctx)
}

func (c *controller) start(ctx context.Context) error {

	banner := figure.NewFigure(c.meta.GetName(), "", true)
	banner.Print()

	c.Log().V(5).Info("runtime.controller.start")

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
			)),
		fx.Provide(
			fx.Annotate(
				func() logger.Logger {
					return c.logger
				},
			)),
		fx.Provide(
			fx.Annotate(
				func() toolkit.Client {
					return c.service.Client()
				},
			)),
		fx.Provide(
			fx.Annotate(
				func() toolkit.Server {
					return c.service.Server()
				},
			)),
		fx.Provide(func() context.Context { return ctx }))

	cfgs := c.Config().Configs()

	c.Log().V(5).Info("runtime.controller: configs supply")
	for _, c := range cfgs {
		opts = append(opts, fx.Supply(c))
	}

	c.Log().V(5).Info("runtime.controller: user custom provide")
	for _, p := range c.providers {
		opts = append(opts, fx.Provide(p))
	}

	// Provide plugins
	c.Log().V(5).Info("runtime.controller: plugins provide start")

	plugins := c.Plugin().Constructors()
	for _, p := range plugins {
		opts = append(opts, fx.Provide(
			fx.Annotate(
				build(p, new(toolkit.Plugin)),
				fx.ResultTags(``, `group:"plugins"`))))
	}

	c.Log().V(5).Info("runtime.controller: plugins provide end")

	// Provide packages
	c.Log().V(5).Info("runtime.controller: packages provide start")
	packages := c.Package().Constructors()
	for _, p := range packages {
		opts = append(opts, fx.Provide(
			fx.Annotate(build(p, new(toolkit.Package)),
				fx.ResultTags(``, `group:"packages"`))))
	}
	c.Log().V(5).Info("runtime.controller: packages provide end")

	// Provide servers
	c.Log().V(5).Info("runtime.controller: servers provide start")
	servers := c.Server().Provides()
	for _, s := range servers {
		opts = append(opts, fx.Provide(s))
	}
	c.Log().V(5).Info("runtime.controller: servers provide end")

	// Start invocations
	// Invoke plugin PreStart
	c.Log().V(5).Info("runtime.controller: plugins invoke registration")
	opts = append(opts, fx.Invoke(fx.Annotate(
		c.Plugin().Register,
		fx.ParamTags(`group:"plugins"`))))

	c.Log().V(5).Info("runtime.controller: plugins invoke PreStart")
	opts = append(opts, fx.Invoke(c.Plugin().PreStart))

	// Invoke packages PreStart
	c.Log().V(5).Info("runtime.controller: package invoke registration")
	opts = append(opts, fx.Invoke(fx.Annotate(
		c.Package().Register,
		fx.ParamTags(`group:"packages"`))))

	c.Log().V(5).Info("runtime.controller: package invoke PreStart")
	opts = append(opts, fx.Invoke(c.Package().PreStart))

	c.Log().V(5).Info("runtime.controller: user custom invoke")
	for _, p := range c.invokes {
		opts = append(opts, fx.Invoke(p))
	}

	// get constructors from servers
	c.Log().V(5).Info("runtime.controller: servers constructors invoke")
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

	c.app = fx.New(
		fx.Options(opts...),
		fx.WithLogger(c.logger.Fx),
	)

	go func() {
		c.app.Run()
	}()

	defer c.app.Stop(context.Background())
	c.Log().V(5).Info("runtime.controller.started")

	sign := make(chan os.Signal)
	signal.Notify(sign, shutdownSignals...)
	select {
	case <-sign:
	case err := <-c.done:
		if err != nil {
			c.Log().Errorf("runtime.controller: stop with err: %v", err)
			return err
		}
	}

	c.Log().V(5).Info("runtime.controller.stopped")
	return nil
}

func (c *controller) onStart(ctx context.Context) error {

	c.Log().V(5).Info("runtime.controller.onStart: start")

	c.Log().V(5).Info("runtime.controller.onStart: tools OnStart")
	if err := c.Tools().OnStart(ctx); err != nil {
		return err
	}

	c.Log().V(5).Info("runtime.controller.onStart: server start")
	c.Server().Start(ctx)

	c.Log().V(5).Info("runtime.controller.onStart: resolver OnStart call")
	c.client.GRPC().GetResolver().OnStart(ctx)

	c.Log().V(5).Info("runtime.controller.onStart: plugin OnStart call")
	c.Plugin().OnStart(ctx)

	c.Log().V(5).Info("runtime.controller.onStart: package OnStart call")
	c.Package().OnStart(ctx)

	for _, fn := range c.onStartHook {
		go func() {
			if err := fn(ctx); err != nil {
				c.Log().Error(err)
			}
		}()
	}

	for _, fn := range c.onStartSyncHook {
		if err := fn(ctx); err != nil {
			return err
		}
	}
	c.Log().V(5).Info("runtime.controller: started")
	return nil
}

func (c *controller) onStop(ctx context.Context) error {

	c.Plugin().OnStop(ctx)
	c.Package().OnStop(ctx)

	c.Server().Stop(ctx)

	for _, fn := range c.onStopHook {
		go func() {
			if err := fn(ctx); err != nil {
				c.Log().Error(err)
			}
		}()
	}

	for _, fn := range c.onStopSyncHook {
		if err := fn(ctx); err != nil {
			return err
		}
	}

	c.Log().V(0).Info("runtime.controller: stopped")
	return nil
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

func (c *controller) Stop(_ context.Context, err error) {
	c.done <- err
	return
}

func (c *controller) RegisterOnStartHook(fn ...func(ctx context.Context) error) {
	c.onStartHook = append(c.onStartHook, fn...)
}

func (c *controller) RegisterOnStopHook(fn ...func(ctx context.Context) error) {
	c.onStopHook = append(c.onStopHook, fn...)
}

func (c *controller) RegisterOnStartSyncHook(fn ...func(ctx context.Context) error) {
	c.onStartSyncHook = append(c.onStartSyncHook, fn...)
}

func (c *controller) RegisterOnStopSyncHook(fn ...func(ctx context.Context) error) {
	c.onStopSyncHook = append(c.onStopSyncHook, fn...)
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

	rt.done = make(chan error)

	rt.providers = make([]interface{}, 0)
	rt.invokes = make([]interface{}, 0)

	rt.onStartHook = make([]func(context.Context) error, 0)
	rt.onStopHook = make([]func(context.Context) error, 0)

	rt.onStartSyncHook = make([]func(context.Context) error, 0)
	rt.onStopSyncHook = make([]func(context.Context) error, 0)

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
