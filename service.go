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
	"gitlab.com/lastbackend/engine/client"
	"gitlab.com/lastbackend/engine/config"
	"gitlab.com/lastbackend/engine/server"
	"gitlab.com/lastbackend/engine/storage"

	"os"
	"os/signal"
	"sync"
	"syscall"
)

type service struct {
	opts Options
	once sync.Once
}

func newService(opts ...Option) (Service, error) {
	s := new(service)
	s.opts = newOptions(opts...)

	if err := s.opts.Cmd.Execute(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *service) Name() string {
	return s.opts.Name
}

func (s *service) Version() string {
	return s.opts.Version
}

func (s *service) Options() Options {
	return s.opts
}

func (s *service) Client() client.Client {
	return s.opts.Client
}

func (s *service) Server() server.Server {
	return s.opts.Server
}

func (s *service) Storage() storage.Storage {
	return s.opts.Storage
}

func (s *service) Config() config.Config {
	return s.opts.Config
}

func (s *service) Start() error {
	for _, fn := range s.opts.BeforeStart {
		if err := fn(); err != nil {
			return err
		}
	}

	if err := s.opts.Server.Start(); err != nil {
		return err
	}

	for _, fn := range s.opts.AfterStart {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

func (s *service) Stop() error {
	var gerr error

	for _, fn := range s.opts.BeforeStop {
		if err := fn(); err != nil {
			gerr = err
		}
	}

	if err := s.opts.Server.Stop(); err != nil {
		return err
	}

	for _, fn := range s.opts.AfterStop {
		if err := fn(); err != nil {
			gerr = err
		}
	}

	return gerr
}

func (s *service) Run() error {
	if err := s.Start(); err != nil {
		return err
	}

	ch := make(chan os.Signal, 1)

	if s.opts.Signal {
		signal.Notify(ch, shutdown()...)
	}

	select {
	// wait on kill signal
	case <-ch:
	// wait on context cancel
	case <-s.opts.Context.Done():
	}

	return s.Stop()
}

func shutdown() []os.Signal {
	return []os.Signal{
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGKILL,
	}
}
