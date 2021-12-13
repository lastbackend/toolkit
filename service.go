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
	"github.com/lastbackend/engine/client"
	"github.com/lastbackend/engine/cmd"
	"github.com/lastbackend/engine/logger"
	"github.com/lastbackend/engine/server"

	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
)

type service struct {
	*meta

	context context.Context
	once    sync.Once

	cli cmd.CLI

	logger logger.Logger

	servers []server.Server
	clients []client.Client
	plugins []Plugin

	signal bool
}

func newService(name string) Service {
	s := new(service)
	s.meta = new(meta)
	s.meta.Name = name

	s.context = context.Background()
	s.logger = logger.DefaultLogger
	s.cli = cmd.New()
	s.servers = make([]server.Server, 0)
	s.clients = make([]client.Client, 0)
	s.plugins = make([]Plugin, 0)
	return s
}

func (s *service) Name() string {
	return s.meta.Name
}

func (s *service) Version() string {
	return s.meta.Version
}

func (s *service) Meta() Meta {
	return s.meta
}

func (s *service) CLI() CLI {
	return s.cli
}

func (s *service) PluginRegister(plug Plugin) error {
	s.plugins = append(s.plugins, plug)
	return nil
}

func (s *service) Client(i interface{}, f func(f client.RegisterFunc) client.CreatorFunc, o client.Option) error {

	valueIface := reflect.ValueOf(i)
	if valueIface.Kind() != reflect.Ptr || valueIface.IsNil() {
		return fmt.Errorf("using unaddressable value %s", valueIface.Type().Field(0).Name)
	}

	typeField := reflect.TypeOf(i)
	if typeField.Kind() == reflect.Interface {
		return fmt.Errorf("the argument %s must not be interface", valueIface.Elem().Field(0).Type())
	}

	t := reflect.TypeOf(valueIface.Interface()).Elem()
	valueIface.Elem().Field(0).Set(reflect.New(t))

	valIface := valueIface.Elem()
	funcType := reflect.TypeOf(client.RegisterFunc(nil))
	funcRegister := reflect.MakeFunc(funcType, func(args []reflect.Value) []reflect.Value {
		s.clients = append(s.clients, args[0].Interface().(client.Client))
		return nil
	})
	funcCreator := reflect.ValueOf(f).Call([]reflect.Value{funcRegister})
	valueOptions := reflect.ValueOf(o)
	valueInstance := funcCreator[0].Call([]reflect.Value{valueOptions})

	valIface.Field(0).Set(valueInstance[0].Elem().Convert(valIface.Field(0).Type()))

	return nil
}

func (s *service) Server(t server.Server) error {
	valueIface := reflect.ValueOf(t)

	// Check if the passed interface is a pointer
	if valueIface.Type().Kind() != reflect.Ptr {
		return fmt.Errorf("the argument must be a pointer")
	}
	if valueIface.IsNil() {
		return fmt.Errorf("the argument must not be nil")
	}

	s.servers = append(s.servers, t)

	return nil
}

func (s *service) Logger() logger.Logger {
	return s.logger
}

func (s *service) SetContext(ctx context.Context) {
	s.context = ctx
}

func (s *service) Run() error {

	s.cli.SetName(s.meta.Name)
	s.cli.SetEnvPrefix(s.meta.EnvPrefix)
	s.cli.SetVersion(s.meta.Version)
	s.cli.SetShortDescription(s.meta.ShorDescription)
	s.cli.SetLongDescription(s.meta.LongDescription)

	for _, t := range s.clients {
		flags := t.Flags()
		for _, f := range flags {
			s.cli.AddFlag(f)
		}
	}

	for _, t := range s.servers {
		flags := t.Flags()
		for _, f := range flags {
			s.cli.AddFlag(f)
		}
	}

	return s.cli.Run(func() error {

		for _, t := range s.plugins {
			if err := t.Start(s.context); err != nil {
				return err
			}
		}

		for _, t := range s.clients {
			if err := t.Start(); err != nil {
				return err
			}
		}

		for _, t := range s.servers {
			if err := t.Start(); err != nil {
				return err
			}
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
	})
}

var shutdownSignals = []os.Signal{
	syscall.SIGTERM,
	syscall.SIGINT,
	syscall.SIGQUIT,
	syscall.SIGKILL,
}
