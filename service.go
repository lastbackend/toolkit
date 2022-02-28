/*
Copyright [2014] - [2021] The Last.Backend authors.

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

package engine

import (
	"github.com/lastbackend/engine/cmd"
	"github.com/lastbackend/engine/logger"

	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"
)

type service struct {
	*meta
	context     context.Context
	cli         cmd.CLI
	logger      logger.Logger
	clients     []Client
	servers     []Server
	plugins     []Plugin
	controllers []Controller
	signal      bool
}

func newService(name string) Service {
	s := new(service)
	s.meta = new(meta)
	s.meta.SetName(name)
	s.context = context.Background()
	s.logger = logger.DefaultLogger
	s.cli = cmd.New()
	s.clients = make([]Client, 0)
	s.servers = make([]Server, 0)
	s.plugins = make([]Plugin, 0)
	return s
}

func (s *service) Meta() Meta {
	return s.meta
}

func (s *service) CLI() CLI {
	return s.cli
}

func (s *service) PluginRegister(plug Plugin) error {
	valueIface := reflect.ValueOf(plug)

	// Check if the passed interface is a pointer
	if valueIface.Type().Kind() != reflect.Ptr {
		return fmt.Errorf("the argument must be a pointer")
	}
	if valueIface.IsNil() {
		return fmt.Errorf("the argument must not be nil")
	}

	s.plugins = append(s.plugins, plug)
	return nil
}

func (s *service) ClientRegister(cli Client) error {

	valueIface := reflect.ValueOf(cli)

	// Check if the passed interface is a pointer
	if valueIface.Type().Kind() != reflect.Ptr {
		return fmt.Errorf("the argument must be a pointer")
	}
	if valueIface.IsNil() {
		return fmt.Errorf("the argument must not be nil")
	}

	s.clients = append(s.clients, cli)

	return nil
}

func (s *service) ServerRegister(srv Server) error {
	valueIface := reflect.ValueOf(srv)

	// Check if the passed interface is a pointer
	if valueIface.Type().Kind() != reflect.Ptr {
		return fmt.Errorf("the argument must be a pointer")
	}
	if valueIface.IsNil() {
		return fmt.Errorf("the argument must not be nil")
	}

	s.servers = append(s.servers, srv)

	return nil
}

func (s *service) ControllerRegister(ctrl Controller) error {
	valueIface := reflect.ValueOf(ctrl)

	// Check if the passed interface is a pointer
	if valueIface.Type().Kind() != reflect.Ptr {
		return fmt.Errorf("the argument must be a pointer")
	}
	if valueIface.IsNil() {
		return fmt.Errorf("the argument must not be nil")
	}

	s.controllers = append(s.controllers, ctrl)

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
	s.cli.SetName(s.meta.Name)
	s.cli.SetEnvPrefix(s.meta.EnvPrefix)
	s.cli.SetVersion(s.meta.Version)
	s.cli.SetShortDescription(s.meta.ShorDescription)
	s.cli.SetLongDescription(s.meta.LongDescription)

	return s.cli.Run(func() error {
		for _, t := range s.plugins {
			if err := t.Start(s.context); err != nil {
				return err
			}
		}
		for _, t := range s.controllers {
			if err := t.Start(s.context); err != nil {
				return err
			}
		}
		for _, t := range s.servers {
			if err := t.Start(); err != nil {
				return err
			}
		}
		for _, t := range s.clients {
			if err := t.Start(); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *service) Stop() error {
	for _, t := range s.controllers {
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
