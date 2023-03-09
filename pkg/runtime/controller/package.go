package controller

import (
	"context"
	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
	"reflect"
)

const PackageHookMethodPreStart = "PreStart"
const PackageHookMethodOnStart = "OnStart"
const PackageHookMethodOnStartSync = "OnStartSync"
const PackageHookMethodOnStop = "OnStop"
const PackageHookMethodOnStopSync = "OnStopSync"

type packageController struct {
	runtime.Package

	log logger.Logger

	constructors []any
	packages     []toolkit.Package
}

func (c *packageController) Provide(constructor ...any) {
	c.constructors = append(c.constructors, constructor...)
	return
}

func (c *packageController) Constructors() []any {
	return c.constructors
}

func (c *packageController) Register(packages []toolkit.Package) {
	c.log.V(5).Info("packageManager.Register.start")
	c.packages = append(c.packages, packages...)
	c.log.V(5).Infof("packageManager.Register.packages: %v", c.packages)
	c.log.V(5).Info("packageManager.Register.end")
}

func (c *packageController) PreStart(ctx context.Context) error {
	c.log.V(5).Info("packageManager.PreStart.start")
	c.hook(ctx, PackageHookMethodPreStart, true)
	c.log.V(5).Info("packageManager.PreStart.end")
	return nil
}

func (c *packageController) OnStart(ctx context.Context) error {
	c.log.V(5).Info("packageManager.OnStart.start")

	c.hook(ctx, PackageHookMethodOnStart, true)
	c.hook(ctx, PackageHookMethodOnStartSync, true)

	c.log.V(5).Info("packageManager.OnStart.end")
	return nil
}

func (c *packageController) OnStop(ctx context.Context) error {
	c.log.V(5).Info("packageManager.OnStop.start")

	c.hook(ctx, PackageHookMethodOnStop, true)
	c.hook(ctx, PackageHookMethodOnStopSync, true)

	c.log.V(5).Info("packageManager.OnStop.end")
	return nil
}

func (c *packageController) hook(ctx context.Context, kind string, sync bool) error {

	for i := 0; i < len(c.packages); i++ {

		if c.packages[i] == nil {
			continue
		}

		if sync {
			c.call(ctx, c.packages[i], kind)
		} else {
			go c.call(ctx, c.packages[i], kind)
		}

	}

	return nil
}

func (c *packageController) call(ctx context.Context, pkg toolkit.Package, kind string) {
	args := []reflect.Value{reflect.ValueOf(ctx)}
	meth := reflect.ValueOf(pkg).MethodByName(kind)
	if !reflect.ValueOf(meth).IsZero() {
		c.log.V(5).Infof("packageManager.%s.call", kind)
		meth.Call(args)
	}
}

func newPackageController(_ context.Context, runtime runtime.Runtime) runtime.Package {
	pl := new(packageController)
	pl.log = runtime.Log()
	pl.constructors = make([]any, 0)
	pl.packages = make([]toolkit.Package, 0)
	return pl
}
