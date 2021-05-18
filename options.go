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
	"gitlab.com/lastbackend/engine/cmd/flags"
	"gitlab.com/lastbackend/engine/config"
	"gitlab.com/lastbackend/engine/server"
	"gitlab.com/lastbackend/engine/storage"

	"context"
)

type Options struct {
	Name        string
	Description string
	Version     string

	Cmd     cmd.Cmd
	Client  client.Client
	Server  server.Server
	Storage storage.Storage
	Config  config.Config

	BeforeStart []func() error
	BeforeStop  []func() error
	AfterStart  []func() error
	AfterStop   []func() error

	Context context.Context

	Signal bool
}

func newOptions(opts ...Option) Options {
	opt := Options{
		Cmd: cmd.NewCmd(),
		//Client:  client.DefaultClient,
		Server: server.DefaultServer,
		//Storage: storage.DefaultStorage,
		Context: context.Background(),
	}

	for _, o := range opts {
		o(&opt)
	}

	return opt
}

func WithName(name string) Option {
	return func(o *Options) {
		o.Name = name
	}
}

func WithVersion(version string) Option {
	return func(o *Options) {
		o.Version = version
	}
}

func WithDescription(desc string) Option {
	return func(o *Options) {
		o.Description = desc
	}
}

func WithFlags(flags ...flags.Flag) Option {
	return func(o *Options) {
		for _, flag := range flags {
			*o.Cmd.Get().Flags = append(*o.Cmd.Get().Flags, flag)
		}
	}
}
