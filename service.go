/*
Copyright [2014] - [2022] The Last.Backend authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package toolkit

import (
	"github.com/lastbackend/toolkit/pkg/cmd"
	"github.com/lastbackend/toolkit/pkg/logger"
	"github.com/lastbackend/toolkit/pkg/probe"
	"github.com/lastbackend/toolkit/pkg/probe/types"
	"golang.org/x/sync/errgroup"

	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"
)

type HookFunc func(ctx context.Context) error

type service struct {
	cli           cmd.CLI
	logger        logger.Logger
	clients       []Client
	servers       []Server
	plugins       []Plugin
	preRunFuncs   []HookFunc
	postRunFuncs  []HookFunc
	preStopFuncs  []HookFunc
	postStopFuncs []HookFunc
	signal        bool
	probe         types.Probe
}

func newService(name string) Service {
	s := new(service)
	s.logger = logger.DefaultLogger
	s.cli = cmd.New(name)
	s.probe = probe.NewProbe()
	s.clients = make([]Client, 0)
	s.servers = make([]Server, 0)
	s.plugins = make([]Plugin, 0)
	s.preRunFuncs = make([]HookFunc, 0)
	s.postRunFuncs = make([]HookFunc, 0)
	s.preStopFuncs = make([]HookFunc, 0)
	s.postStopFuncs = make([]HookFunc, 0)
	return s
}

func (s *service) Meta() Meta {
	return s.cli.GetMeta()
}

func (s *service) CLI() CLI {
	return s.cli
}

func (s *service) Probe() types.Probe {
	return s.probe
}

func (s *service) PluginRegister(plug Plugin) error {
	valueIface := reflect.ValueOf(plug)

	// Check if the passed interface is a pointer
	if valueIface.Type().Kind() != reflect.Ptr {
		return fmt.Errorf("the plugin must be a pointer")
	}
	if valueIface.IsNil() {
		return fmt.Errorf("the plugin must not be nil")
	}

	s.plugins = append(s.plugins, plug)
	return nil
}

func (s *service) ClientRegister(cli Client) error {

	valueIface := reflect.ValueOf(cli)

	// Check if the passed interface is a pointer
	if valueIface.Type().Kind() != reflect.Ptr {
		return fmt.Errorf("the client must be a pointer")
	}
	if valueIface.IsNil() {
		return fmt.Errorf("the client must not be nil")
	}

	s.clients = append(s.clients, cli)

	return nil
}

func (s *service) ServerRegister(srv Server) error {
	valueIface := reflect.ValueOf(srv)

	// Check if the passed interface is a pointer
	if valueIface.Type().Kind() != reflect.Ptr {
		return fmt.Errorf("the server must be a pointer")
	}
	if valueIface.IsNil() {
		return fmt.Errorf("the server must not be nil")
	}

	s.servers = append(s.servers, srv)

	return nil
}

func (s *service) Logger() logger.Logger {
	return s.logger
}

func (s *service) PreRunHookFunc(fn HookFunc) error {
	if fn == nil {
		return fmt.Errorf("the hook function must not be nil")
	}
	s.preRunFuncs = append(s.preRunFuncs, fn)
	return nil
}

func (s *service) PostRunHookFunc(fn HookFunc) error {
	if fn == nil {
		return fmt.Errorf("the hook function must not be nil")
	}
	s.postRunFuncs = append(s.postRunFuncs, fn)
	return nil
}

func (s *service) PreStopHookFunc(fn HookFunc) error {
	if fn == nil {
		return fmt.Errorf("the hook function must not be nil")
	}
	s.preStopFuncs = append(s.preStopFuncs, fn)
	return nil
}

func (s *service) PostStopHookFunc(fn HookFunc) error {
	if fn == nil {
		return fmt.Errorf("the hook function must not be nil")
	}
	s.postStopFuncs = append(s.postStopFuncs, fn)
	return nil
}

func (s *service) Run(ctx context.Context) error {

	if err := s.Start(ctx); err != nil {
		return err
	}

	ch := make(chan os.Signal, 1)
	if s.signal {
		signal.Notify(ch, shutdownSignals...)
	}

	select {
	// wait on kill signal
	case <-ch:
	// wait on context cancel
	case <-ctx.Done():
	}

	return s.Stop(ctx)
}

func (s *service) Start(ctx context.Context) error {

	s.probe.Init(s.Meta().GetEnvPrefix(), s.CLI())

	err := s.cli.PreRun(func() error {
		group, _ := errgroup.WithContext(ctx)

		for _, fn := range s.preRunFuncs {
			var _fn = fn
			group.Go(func() error {
				return _fn(ctx)
			})
		}
		return group.Wait()
	})
	if err != nil {
		return err
	}

	err = s.cli.Run(func() error {

		for _, t := range s.plugins {
			if err := t.Start(ctx); err != nil {
				return err
			}
		}

		group, _ := errgroup.WithContext(ctx)

		for _, t := range s.servers {
			var _t = t
			group.Go(func() error {
				return _t.Start(ctx)
			})
		}

		if err := group.Wait(); err != nil {
			return err
		}

		for _, t := range s.clients {
			if err := t.Start(ctx); err != nil {
				return err
			}
		}

		return s.probe.Start(ctx)
	})
	if err != nil {
		return err
	}

	err = s.cli.PostRun(func() error {
		group, _ := errgroup.WithContext(ctx)

		for _, fn := range s.postRunFuncs {
			var _fn = fn
			group.Go(func() error {
				return _fn(ctx)
			})
		}

		return group.Wait()
	})
	if err != nil {
		return err
	}

	return s.cli.Execute()
}

func (s *service) Stop(ctx context.Context) error {
	if err := s.probe.Stop(); err != nil {
		return err
	}

	for _, t := range s.preStopFuncs {
		if err := t(ctx); err != nil {
			return err
		}
	}
	for _, t := range s.servers {
		if err := t.Stop(); err != nil {
			return err
		}
	}
	for _, t := range s.plugins {
		if err := t.Stop(); err != nil {
			return err
		}
	}
	for _, t := range s.postStopFuncs {
		if err := t(ctx); err != nil {
			return err
		}
	}
	return nil
}

var shutdownSignals = []os.Signal{
	syscall.SIGTERM,
	syscall.SIGINT,
	syscall.SIGQUIT,
	syscall.SIGKILL,
}
