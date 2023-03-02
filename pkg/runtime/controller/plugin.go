package controller

import (
	"context"
	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
	"reflect"
)

const PluginHookMethodPreStart = "PreStart"
const PluginHookMethodOnStart = "OnStart"
const PluginHookMethodOnStop = "OnStop"

type pluginManager struct {
	runtime.Plugin

	log     logger.Logger
	provide []interface{}
	plugins []interface{}
}

func (c *pluginManager) Provide(plugin interface{}) {
	c.provide = append(c.provide, plugin)
}

func (c *pluginManager) Provides() []interface{} {
	return c.provide
}

func (c *pluginManager) Register(plugin interface{}) {
	c.plugins = append(c.plugins, plugin)
	return
}

func (c *pluginManager) Plugins() []interface{} {
	return c.plugins
}

func (c *pluginManager) PreStart(ctx context.Context) error {
	c.log.Debug(" -- plugin manager : pre start -- ")
	args := []reflect.Value{reflect.ValueOf(ctx)}
	for _, p := range c.plugins {
		meth := reflect.ValueOf(p).MethodByName(PluginHookMethodPreStart)
		if !reflect.ValueOf(meth).IsZero() {
			c.log.Debug(" -- plugin manager : call preStart -- ")
			meth.Call(args)
		}
	}
	return nil
}

func (c *pluginManager) OnStart(ctx context.Context) error {
	c.log.Debug(" -- plugin manager : on start hook : start -- ")
	args := []reflect.Value{reflect.ValueOf(ctx)}
	for _, p := range c.plugins {
		meth := reflect.ValueOf(p).Elem().MethodByName(PluginHookMethodOnStart)
		if !reflect.ValueOf(meth).IsZero() {
			c.log.Debug(" -- plugin manager : call onStart -- ")
			meth.Call(args)
		}
	}
	c.log.Debug(" -- plugin manager : on start hook : stop -- ")
	return nil
}

func (c *pluginManager) OnStop(ctx context.Context) error {
	args := []reflect.Value{reflect.ValueOf(ctx)}
	for _, p := range c.plugins {
		c.log.Debug(reflect.ValueOf(p).Elem().Type())
		meth := reflect.ValueOf(p).Elem().MethodByName(PluginHookMethodOnStop)
		if !reflect.ValueOf(meth).IsZero() {
			c.log.Debug(" -- plugin manager : call onStop -- ")
			meth.Call(args)
		}
	}
	return nil
}

func newPluginController(_ context.Context, log logger.Logger) runtime.Plugin {
	pl := new(pluginManager)
	pl.log = log
	pl.provide = make([]interface{}, 0)
	pl.plugins = make([]interface{}, 0)
	return pl
}
