package controller

import (
	"context"
	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
	"reflect"
)

const PackageHookMethodPreStart = "PreStart"
const PackageHookMethodOnStart = "OnStart"
const PackageHookMethodOnStop = "OnStop"

type oackageController struct {
	runtime.Package
	log logger.Logger

	provide  []interface{}
	packages []interface{}
}

func (c *oackageController) Provide(pkg interface{}) {
	c.provide = append(c.provide, pkg)
	return
}

func (c *oackageController) Provides() []interface{} {
	return c.provide
}

func (c *oackageController) Register(pkg interface{}) {
	c.packages = append(c.packages, pkg)
}

func (c *oackageController) Packages() []interface{} {
	return c.packages
}

func (c *oackageController) PreStart(ctx context.Context) error {
	c.log.Debug(" -- package manager : pre start hook : start  -- ")

	args := []reflect.Value{reflect.ValueOf(ctx)}
	for _, p := range c.packages {

		c.log.Debug(reflect.ValueOf(p).Type())

		meth := reflect.ValueOf(p).MethodByName(PackageHookMethodPreStart)
		if !reflect.ValueOf(meth).IsZero() {
			c.log.Debug(" -- package manager : call preStart -- ")
			meth.Call(args)
		}
	}

	c.log.Debug(" -- package manager : pre start hook : stop  -- ")

	return nil
}

func (c *oackageController) OnStart(ctx context.Context) error {

	c.log.Debug(" -- package manager : on start hook: start -- ")

	args := []reflect.Value{reflect.ValueOf(ctx)}
	for _, p := range c.packages {
		meth := reflect.ValueOf(p).MethodByName(PackageHookMethodOnStart)
		if !reflect.ValueOf(meth).IsZero() {
			c.log.Debug(" -- package manager : call onStart -- ")
			meth.Call(args)
		}
	}

	c.log.Debug(" -- package manager : on start hook: stop -- ")

	return nil
}

func (c *oackageController) OnStop(ctx context.Context) error {

	c.log.Debug(" -- package manager : on stop hook : start -- ")

	args := []reflect.Value{reflect.ValueOf(ctx)}
	for _, p := range c.packages {
		item := reflect.ValueOf(p)
		meth := item.MethodByName(PackageHookMethodOnStop)
		if !reflect.ValueOf(meth).IsZero() {
			c.log.Debug(" -- package manager : call onStop -- ")
			meth.Call(args)
		}
	}

	c.log.Debug(" -- package manager : on stop hook : stop -- ")

	return nil
}

func newPackageController(_ context.Context, runtime runtime.Runtime) runtime.Package {
	pl := new(oackageController)
	pl.log = runtime.Log()
	pl.provide = make([]interface{}, 0)
	pl.packages = make([]interface{}, 0)
	return pl
}
