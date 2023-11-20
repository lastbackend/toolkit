package controller

import (
  "context"
  "fmt"
  "github.com/lastbackend/toolkit"
  "github.com/lastbackend/toolkit/pkg/runtime"
  "github.com/lastbackend/toolkit/pkg/runtime/logger"
  "github.com/lastbackend/toolkit/pkg/util/types"
  "github.com/pkg/errors"
  "golang.org/x/sync/errgroup"
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

func (c *packageController) Register(packages []toolkit.PackageItem) {
  c.log.V(5).Info("packageManager.Register.start")
  c.packages = make([]toolkit.Package, len(packages))
  for _, pkg := range packages {
    c.packages[pkg.Index] = pkg.Source
  }

  c.log.V(5).Infof("packageManager.Register.packages: %v", c.packages)
  c.log.V(5).Info("packageManager.Register.end")
}

func (c *packageController) PreStart(ctx context.Context) error {
  c.log.V(5).Info("packageManager.PreStart.start")
  err := c.hook(ctx, PackageHookMethodPreStart, true)
  if err != nil {
    return err
  }
  c.log.V(5).Info("packageManager.PreStart.end")
  return nil
}

func (c *packageController) OnStart(ctx context.Context) error {
  c.log.V(5).Info("packageManager.OnStart.start")

  err := c.hook(ctx, PackageHookMethodOnStartSync, true)
  if err != nil {
    return err
  }

  err = c.hook(ctx, PackageHookMethodOnStart, false)
  if err != nil {
    return err
  }

  c.log.V(5).Info("packageManager.OnStart.end")
  return nil
}

func (c *packageController) OnStop(ctx context.Context) error {
  c.log.V(5).Info("packageManager.OnStop.start")

  err := c.hook(ctx, PackageHookMethodOnStopSync, true)
  if err != nil {
    return err
  }

  err = c.hook(ctx, PackageHookMethodOnStop, false)
  if err != nil {
    return err
  }

  c.log.V(5).Info("packageManager.OnStop.end")
  return nil
}

func (c *packageController) hook(ctx context.Context, kind string, sync bool) error {

  ctx, cancel := context.WithCancelCause(ctx)

  if sync {
    c.log.V(5).Infof("packageManager.%s.start:sync", kind)
    defer func() {
      c.log.V(5).Infof("packageManager.%s.end:sync", kind)
    }()
  } else {
    c.log.V(5).Infof("packageManager.%s.start:async", kind)
    defer func() {
      c.log.V(5).Infof("packageManager.%s.end:async", kind)
    }()
  }

  // start all non-sync methods
  g := errgroup.Group{}

  for i := 0; i < len(c.packages); i++ {

    if c.packages[i] == nil {
      continue
    }

    pkg := c.packages[i]

    if sync {
      if err := c.call(ctx, pkg, kind); err != nil {
        return err
      }
    } else {
      g.Go(func() error {
        return c.call(ctx, pkg, kind)
      })
    }
  }

  if err := g.Wait(); err != nil {
    c.log.V(5).Errorf("can not start toolkit:", err.Error())
    cancel(err)
    return err
  }

  return nil
}

func (c *packageController) call(ctx context.Context, pkg toolkit.Package, kind string) error {

  args := []reflect.Value{reflect.ValueOf(ctx)}
  meth := reflect.ValueOf(pkg).MethodByName(kind)
  name := types.Type(pkg)

  if !reflect.ValueOf(meth).IsZero() {
    c.log.V(5).Infof("packageManager.%s.call: %s", kind, name)

    res := meth.Call(args)
    if len(res) < 1 {
      return nil
    }

    if len(res) > 1 {
      return errors.New(fmt.Sprintf("packageManager.%s.call:%s:err: method results are not supported. Only error is supported", kind, name))
    }

    if v := res[0].Interface(); v != nil {
      if err, ok := v.(error); ok && err != nil {
        return err
      }
    }
  }

  return nil
}

func newPackageController(_ context.Context, runtime runtime.Runtime) runtime.Package {
  pl := new(packageController)
  pl.log = runtime.Log()
  pl.constructors = make([]any, 0)
  pl.packages = make([]toolkit.Package, 0)
  return pl
}
