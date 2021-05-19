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
	"gitlab.com/lastbackend/engine/cmd"
	"gitlab.com/lastbackend/engine/config"
	"gitlab.com/lastbackend/engine/server"
	"gitlab.com/lastbackend/engine/storage"

	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type service struct {
	*meta

	context context.Context
	once    sync.Once

	cmd     cmd.Cmd
	client  client.Client
	server  server.Server
	storage storage.Storage
	config  config.Config

	signal bool
}

func newService(name string) Service {
	s := new(service)
	s.meta = new(meta)
	s.meta.Name = name
	s.cmd = cmd.New()
	s.context = context.Background()
	s.server = server.DefaultServer
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
	return s.cmd
}

func (s *service) Init() error {
	s.cmd.SetName(s.meta.Name)
	s.cmd.SetEnvPrefix(s.meta.EnvPrefix)
	s.cmd.SetVersion(s.meta.Version)
	s.cmd.SetShortDescription(s.meta.ShorDescription)
	s.cmd.SetLongDescription(s.meta.LongDescription)

	if err := s.cmd.Execute(); err != nil {
		return err
	}
	
	return nil
}

func (s *service) Client() client.Client {
	return s.client
}

func (s *service) Server() server.Server {
	return s.server
}

func (s *service) Storage() storage.Storage {
	return s.storage
}

func (s *service) Config() config.Config {
	return s.config
}

func (s *service) SetContext(ctx context.Context) {
	s.context = ctx
}

func (s *service) Run() error {

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

	return s.server.Stop()
}

var shutdownSignals = []os.Signal{
	syscall.SIGTERM,
	syscall.SIGINT,
	syscall.SIGQUIT,
	syscall.SIGKILL,
}
