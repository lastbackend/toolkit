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
	"context"

	"github.com/lastbackend/toolkit/cmd"
	"github.com/lastbackend/toolkit/logger"
)

type ServiceType uint8

type HandlerProps struct {
	Transport string
	Options   interface{}
}

type Service interface {
	Meta() Meta
	CLI() CLI
	Logger() logger.Logger
	SetContext(ctx context.Context)
	PluginRegister(plug Plugin) error
	ServerRegister(srv Server) error
	ClientRegister(cli Client) error
	PackageRegister(ctrl Package) error
	Run() error
	Start() error
	Stop() error
}

type Meta cmd.MetaInfo

type CLI interface {
	cmd.FlagSet
	cmd.CommandSet
}

type Server interface {
	Start() error
	Stop() error
}

type Client interface {
	Start() error
	Stop() error
}

type Plugin interface {
	Start(ctx context.Context) error
	Stop() error
}

type Package interface {
	PreStart(ctx context.Context) error
	Start(ctx context.Context) error
	PostStart(ctx context.Context) error
	Stop() error
}

func NewService(name string) Service {
	return newService(name)
}
