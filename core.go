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
	"gitlab.com/lastbackend/engine/plugin"
	"gitlab.com/lastbackend/engine/server"

	"context"
)

type Storage interface {
	plugin.Plugin
	Client()
}

type Service interface {
	Name() string
	Version() string
	Meta() Meta
	CLI() CLI
	Init() error
	Client() client.Client
	Server() server.Server
	SetContext(ctx context.Context)
	Register(i interface{}) error
	Run() error
}

type Meta interface {
	SetVersion(string)
	SetEnvPrefix(string)
	SetShortDescription(string)
	SetLongDescription(string)
}

type CLI interface {
	cmd.FlagSet
}

func NewService(name string) Service {
	return newService(name)
}
