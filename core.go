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
	"github.com/lastbackend/engine/cmd"
	"github.com/lastbackend/engine/logger"
	"github.com/lastbackend/engine/server"
)

type ServiceType uint8

type ServiceProps struct {
	Func    interface{}
	Options interface{}
}

type HandlerProps struct {
	Transport string
	Options   interface{}
}

type Service interface {
	Name() string
	Version() string
	Meta() Meta
	Init() error
	Logger() logger.Logger
	SetContext(ctx context.Context)
	Register(i interface{}, props map[string]map[string]ServiceProps) error
	Transport(t server.Server) error
	Run() error
}

type Meta interface {
	SetVersion(version string)
	SetEnvPrefix(prefix string)
	SetShortDescription(desc string)
	SetLongDescription(desc string)
}

type CLI interface {
	cmd.FlagSet
}

func NewService(name string) Service {
	return newService(name)
}
