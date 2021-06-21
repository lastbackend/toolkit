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
	"context"
	"fmt"
	"github.com/lastbackend/engine/cmd"
	"github.com/lastbackend/engine/logger"
	"github.com/lastbackend/engine/plugin"
	"github.com/lastbackend/engine/plugin/manager"
	"github.com/lastbackend/engine/service/client"
	"github.com/lastbackend/engine/service/server"
	"github.com/lastbackend/engine/transport"
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

	pm         manager.Manager
	transports []transport.Transport

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
	s.transports = make([]transport.Transport, 0)
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

	for _, t := range s.transports {
		s.cli.AddFlags(t.Flags()...)
	}

	if err := s.cli.Execute(); err != nil {
		return err
	}

	return nil
}

func (s *service) Register(i interface{}, props map[string]map[string]ServiceProps) error {

	initField := func(tech string, service ServiceProps, valueField reflect.Value) error {
		if valueField.Kind() == reflect.Ptr && valueField.IsNil() {
			t := reflect.TypeOf(valueField.Interface()).Elem()
			valueField.Set(reflect.New(t))
		}

		switch tech {
		case "Storage":
			fallthrough
		case "Cache":
			if err := s.pm.Register(valueField.Interface(), service.Func.(func(f plugin.RegisterFunc) plugin.CreatorFunc), service.Options.(plugin.Option)); err != nil {
				return err
			}
		case "Broker":
		}

		return nil
	}

	valueIface := reflect.ValueOf(i)

	// Check if the passed interface is a pointer
	if valueIface.Type().Kind() != reflect.Ptr {
		return fmt.Errorf("the argument must be a pointer")
	}
	if valueIface.IsNil() {
		return fmt.Errorf("the argument must not be nil")
	}

	for tech, fields := range props {

		valueField := valueIface.Elem().FieldByName(tech)
		if !valueField.IsValid() {
			return fmt.Errorf("interface `%s` does not have the field `%s`", valueIface.Type(), tech)
		}

		typeField := reflect.TypeOf(valueIface.Elem().FieldByName(tech))

		if typeField.Kind() == reflect.Interface {
			return fmt.Errorf("the argument %s must not be interface", valueIface.Elem().FieldByName(tech).Type())
		}

		techValueField := valueIface.Elem().FieldByName(tech)

		if techValueField.Kind() == reflect.Ptr && techValueField.IsNil() {

			if len(fields) != 1 {
				return fmt.Errorf("interface `%s` does not inplement custom options structure", techValueField.Type())
			}

			vField := techValueField

			// Get first element from map
			keys := reflect.ValueOf(fields).MapKeys()
			serviceProp := fields[keys[0].String()]

			if err := initField(tech, serviceProp, vField); err != nil {
				return fmt.Errorf("can not init %s argument: %v", vField.Type(), err)
			}

			continue
		}

		if techValueField.Kind() == reflect.Struct {

			if techValueField.NumField() == 1 {

				elemValueFiled := techValueField.Field(0)
				if elemValueFiled.Kind() == reflect.Interface {
					if techValueField.Kind() != reflect.Ptr {
						return fmt.Errorf("using unaddressable value %s", techValueField.Type().Field(0).Name)
					}
					elemValueFiled = techValueField
				}
				if elemValueFiled.Kind() == reflect.Struct {
					if techValueField.Kind() != reflect.Ptr {
						return fmt.Errorf("using unaddressable value %s", techValueField.Type().Field(0).Name)
					}
				}

				// Get first element from map
				keys := reflect.ValueOf(fields).MapKeys()
				serviceProp := fields[keys[0].String()]

				if err := initField(tech, serviceProp, elemValueFiled); err != nil {
					return fmt.Errorf("can not init %s argument: %v", elemValueFiled.Type(), err)
				}

				continue
			}
		}

		for srv, serviceProp := range fields {
			srvField := valueIface.Elem().FieldByName(tech).FieldByName(srv)

			if srvField.Kind() != reflect.Interface {
				if !srvField.IsValid() {
					return fmt.Errorf("interface `%s` does not have the field `%s`", valueField.Type(), srv)
				}
				if srvField.Kind() != reflect.Ptr {
					return fmt.Errorf("the `%s` must be a pointer`", srv)
				}
				if srvField.IsNil() {
					t := reflect.TypeOf(srvField.Interface()).Elem()
					srvField.Set(reflect.New(t))
				}
			}

			if err := initField(tech, serviceProp, srvField); err != nil {
				return fmt.Errorf("can not init %s argument: %v", srvField.Type(), err)
			}

		}

	}

	return nil
}

func (s *service) Transport(t transport.Transport) error {
	valueIface := reflect.ValueOf(t)

	// Check if the passed interface is a pointer
	if valueIface.Type().Kind() != reflect.Ptr {
		return fmt.Errorf("the argument must be a pointer")
	}
	if valueIface.IsNil() {
		return fmt.Errorf("the argument must not be nil")
	}

	s.transports = append(s.transports, t)

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

	for _, t := range s.transports {
		if err := t.Start(); err != nil {
			return err
		}
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

	for _, t := range s.transports {
		if err := t.Stop(); err != nil {
			return err
		}
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
