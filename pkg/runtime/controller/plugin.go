package controller

import (
	"context"
	"errors"
	"fmt"
	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
	"github.com/lastbackend/toolkit/pkg/util/types"
	"golang.org/x/sync/errgroup"
	"reflect"
)

const PluginHookMethodPreStart = "PreStart"
const PluginHookMethodOnStart = "OnStart"
const PluginHookMethodOnStartSync = "OnStartSync"
const PluginHookMethodOnStop = "OnStop"
const PluginHookMethodOnStopSync = "OnStopSync"

type pluginManager struct {
	runtime.Plugin

	log          logger.Logger
	constructors []any
	plugins      []toolkit.Plugin
}

func (c *pluginManager) Provide(constructor ...any) {
	c.constructors = append(c.constructors, constructor...)
}

func (c *pluginManager) Constructors() []any {
	return c.constructors
}

func (c *pluginManager) Register(plugins []toolkit.Plugin) {
	c.log.V(5).Info("pluginManager.Register.start")
	c.plugins = append(c.plugins, plugins...)
	c.log.V(5).Infof("pluginManager.Register.plugins %v", c.plugins)
	c.log.V(5).Info("pluginManager.Register.end")
	return
}

func (c *pluginManager) PreStart(ctx context.Context) error {
	c.log.V(5).Info("pluginManager.PreStart.start")
	err := c.hook(ctx, PluginHookMethodPreStart, true)
	if err != nil {
		return err
	}
	c.log.V(5).Info("pluginManager.PreStart.end")
	return nil
}

func (c *pluginManager) OnStart(ctx context.Context) error {
	c.log.V(5).Info("pluginManager.OnStart.start")

	err := c.hook(ctx, PluginHookMethodOnStartSync, true)
	if err != nil {
		return err
	}

	err = c.hook(ctx, PluginHookMethodOnStart, false)
	if err != nil {
		return err
	}

	c.log.V(5).Info("pluginManager.OnStart.end")
	return nil
}

func (c *pluginManager) OnStop(ctx context.Context) error {
	c.log.V(5).Info("pluginManager.OnStop.start")

	err := c.hook(ctx, PluginHookMethodOnStopSync, true)
	if err != nil {
		return err
	}

	err = c.hook(ctx, PluginHookMethodOnStop, false)
	if err != nil {
		return err
	}

	c.log.V(5).Info("pluginManager.OnStop.end")
	return nil
}

func (c *pluginManager) hook(ctx context.Context, kind string, sync bool) error {

	if sync {
		c.log.V(5).Infof("pluginManager.%s.start:sync", kind)
		defer func() {
			c.log.V(5).Infof("pluginManager.%s.end:sync", kind)
		}()
	} else {
		c.log.V(5).Infof("pluginManager.%s.start:async", kind)
		defer func() {
			c.log.V(5).Infof("pluginManager.%s.end:async", kind)
		}()
	}

	// start all non-sync methods
	g, ctx := errgroup.WithContext(ctx)
	for i := 0; i < len(c.plugins); i++ {

		if c.plugins[i] == nil {
			continue
		}

		plugin := c.plugins[i]

		if sync {
			if err := c.call(ctx, plugin, kind); err != nil {
				return err
			}
		} else {
			g.Go(func() error {
				return c.call(ctx, plugin, kind)
			})
		}
	}

	if err := g.Wait(); err != nil {
		c.log.V(5).Errorf("can not start toolkit:", err.Error())
		return err
	}

	return nil

}

func (c *pluginManager) call(ctx context.Context, pkg toolkit.Plugin, kind string) error {

	args := []reflect.Value{reflect.ValueOf(ctx)}
	meth := reflect.ValueOf(pkg).MethodByName(kind)
	name := types.Type(pkg)

	if !reflect.ValueOf(meth).IsZero() {
		c.log.V(5).Infof("pluginManager.%s.call: %s", kind, name)
		res := meth.Call(args)

		if len(res) < 1 {
			return errors.New(fmt.Sprintf("pluginManager.%s.call:%s:err: error is not declared as return result in method", kind, name))
		}

		if len(res) > 1 {
			return errors.New(fmt.Sprintf("pluginManager.%s.call:%s:err: method results are not supported. Only error is supported", kind, name))
		}

		var err error
		if v := res[0].Interface(); v != nil {
			err = v.(error)
		}
		return err
	}

	return nil
}

func newPluginController(_ context.Context, runtime runtime.Runtime) runtime.Plugin {
	pl := new(pluginManager)
	pl.log = runtime.Log()
	pl.constructors = make([]any, 0)
	pl.plugins = make([]toolkit.Plugin, 0)
	return pl
}
