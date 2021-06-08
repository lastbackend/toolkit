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
	"github.com/lastbackend/engine/plugin"
	"github.com/lastbackend/engine/plugin/manager"
	"github.com/lastbackend/engine/service/client"
	"github.com/lastbackend/engine/service/server"

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

	client client.Client
	server server.Server
	logger logger.Logger

	pm manager.Manager

	signal bool
}

func newService(name string) Service {
	s := new(service)
	s.meta = new(meta)
	s.meta.Name = name
	s.cli = cmd.New()
	s.context = context.Background()
	s.server = server.New()
	s.logger = logger.DefaultLogger
	s.pm = manager.NewManager()
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

func (s *service) Init() error {
	s.cli.SetName(s.meta.Name)
	s.cli.SetEnvPrefix(s.meta.EnvPrefix)
	s.cli.SetVersion(s.meta.Version)
	s.cli.SetShortDescription(s.meta.ShorDescription)
	s.cli.SetLongDescription(s.meta.LongDescription)

	s.cli.AddFlags(s.server.Flags()...)
	s.cli.AddCommands(s.server.Commands()...)

	s.cli.AddFlags(s.pm.Flags()...)
	s.cli.AddCommands(s.pm.Commands()...)

	if err := s.cli.Execute(); err != nil {
		return err
	}

	return nil
}

func (s *service) RegisterPlugin(p plugin.Plugin) {
	s.pm.RegisterPlugin(p)
}

func (s *service) Register(i interface{}, props map[string]map[string]ServiceProps) error {

	valueIface := reflect.ValueOf(i)

	// Check if the passed interface is a pointer
	if valueIface.Type().Kind() != reflect.Ptr {
		return fmt.Errorf("the argument must be a pointer")
	}
	if valueIface.IsNil() {
		return fmt.Errorf("the argument must not be nil")
	}

	for tech, fields := range props {

		storageField := valueIface.Elem().FieldByName(tech)
		if !storageField.IsValid() {
			return fmt.Errorf("interface `%s` does not have the field `%s`", valueIface.Type(), tech)
		}

		for srv := range fields {
			srvField := valueIface.Elem().FieldByName(tech).FieldByName(srv)
			if !srvField.IsValid() {
				return fmt.Errorf("interface `%s` does not have the field `%s`", storageField.Type(), srv)
			}
			if srvField.Kind() != reflect.Ptr {
				return fmt.Errorf("the `%s` must be a pointer`", srv)
			}
			if srvField.IsNil() {
				t := reflect.TypeOf(srvField.Interface()).Elem()
				srvField.Set(reflect.New(t))
			}

			switch tech {
			case "Storage":
				fallthrough
			case "Cache":
				if err := s.pm.Register(srvField.Interface(), props[tech][srv].Func.(func(f plugin.RegisterFunc) plugin.CreatorFunc), props[tech][srv].Options.(plugin.Option)); err != nil {
					return err
				}
			case "Broker":
			case "Service":

			}

		}
	}

	return nil
}

func (s *service) Client() client.Client {
	return s.client
}

func (s *service) Server() server.Server {
	return s.server
}

func (s *service) Logger() logger.Logger {
	return s.logger
}

func (s *service) SetContext(ctx context.Context) {
	s.context = ctx
}

func (s *service) Run() error {

	if err := s.pm.Start(); err != nil {
		return err
	}

	if err := s.server.Start(); err != nil {
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

	s.pm.Stop()

	return s.server.Stop()
}

var shutdownSignals = []os.Signal{
	syscall.SIGTERM,
	syscall.SIGINT,
	syscall.SIGQUIT,
	syscall.SIGKILL,
}
