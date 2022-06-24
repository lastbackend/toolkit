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
	"github.com/lastbackend/toolkit/cmd"
	"github.com/lastbackend/toolkit/logger"
	"github.com/lastbackend/toolkit/probe"
	"github.com/lastbackend/toolkit/probe/types"
	"golang.org/x/sync/errgroup"

	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"
)

type service struct {
	context  context.Context
	cli      cmd.CLI
	logger   logger.Logger
	clients  []Client
	servers  []Server
	plugins  []Plugin
	packages []Package
	signal   bool
	probe    types.Probe
}

func newService(name string) Service {
	s := new(service)
	s.context = context.Background()
	s.logger = logger.DefaultLogger
	s.cli = cmd.New(name)
	s.probe = probe.NewProbe()
	s.clients = make([]Client, 0)
	s.servers = make([]Server, 0)
	s.plugins = make([]Plugin, 0)
	s.packages = make([]Package, 0)
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

func (s *service) PackageRegister(ctrl Package) error {
	valueIface := reflect.ValueOf(ctrl)

	// Check if the passed interface is a pointer
	if valueIface.Type().Kind() != reflect.Ptr {
		return fmt.Errorf("the package must be a pointer")
	}
	if valueIface.IsNil() {
		return fmt.Errorf("the package must not be nil")
	}

	s.packages = append(s.packages, ctrl)

	return nil
}

func (s *service) Logger() logger.Logger {
	return s.logger
}

func (s *service) SetContext(ctx context.Context) {
	s.context = ctx
}

func (s *service) Run() error {

	if err := s.Start(); err != nil {
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
	case <-s.context.Done():
	}

	return s.Stop()
}

func (s *service) Start() error {

	s.probe.Init(s.Meta().GetEnvPrefix(), s.CLI())

	err := s.cli.PreRun(func() error {
		for _, t := range s.packages {
			_, ok := reflect.TypeOf(t).MethodByName("PreStart")
			if ok {
				args := []reflect.Value{reflect.ValueOf(s.context)}
				res := reflect.ValueOf(t).MethodByName("PreStart").Call(args)
				if len(res) == 1 {
					if v := res[0].Interface(); v != nil {
						return v.(error)
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = s.cli.Run(func() error {

		for _, t := range s.plugins {
			if err := t.Start(s.context); err != nil {
				return err
			}
		}

		group, _ := errgroup.WithContext(s.context)

		for _, t := range s.packages {
			group.Go(func() error {
				return t.Start(s.context)
			})
		}

		for _, t := range s.servers {
			group.Go(func() error {
				return t.Start()
			})
		}

		if err := group.Wait(); err != nil {
			return err
		}

		for _, t := range s.clients {
			if err := t.Start(); err != nil {
				return err
			}
		}

		return s.probe.Start(s.context)
	})
	if err != nil {
		return err
	}

	err = s.cli.PostRun(func() error {
		for _, t := range s.packages {
			_, ok := reflect.TypeOf(t).MethodByName("PostStart")
			if ok {
				args := []reflect.Value{reflect.ValueOf(s.context)}
				res := reflect.ValueOf(t).MethodByName("PostStart").Call(args)
				if len(res) == 1 {
					if v := res[0].Interface(); v != nil {
						return v.(error)
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return s.cli.Execute()
}

func (s *service) Stop() error {
	if err := s.probe.Stop(); err != nil {
		return err
	}

	for _, t := range s.packages {
		if err := t.Stop(); err != nil {
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
	return nil
}

var shutdownSignals = []os.Signal{
	syscall.SIGTERM,
	syscall.SIGINT,
	syscall.SIGQUIT,
	syscall.SIGKILL,
}
