/*
Copyright [2014] - [2023] The Last.Backend authors.

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
	"net/http"

	"github.com/lastbackend/toolkit/pkg/client/grpc"
	"github.com/lastbackend/toolkit/pkg/cmd"
	"github.com/lastbackend/toolkit/pkg/logger"
	probe "github.com/lastbackend/toolkit/pkg/probe/types"
	"github.com/lastbackend/toolkit/pkg/router"
)

type ServiceType uint8

type HandlerProps struct {
	Transport string
	Options   interface{}
}

type Service interface {
	Meta() Meta
	CLI() CLI
	Probe() probe.Probe
	Logger() logger.Logger
	Router() router.Server
	Client() grpc.Client
	PluginRegister(plug Plugin) error
	ServerRegister(srv Server) error
	ClientRegister(cli Client) error
	PreRunHookFunc(fn HookFunc) error
	PostRunHookFunc(fn HookFunc) error
	PreStopHookFunc(fn HookFunc) error
	PostStopHookFunc(fn HookFunc) error
	Run(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type Meta cmd.MetaInfo

type CLI interface {
	cmd.FlagSet
	cmd.CommandSet
}

type Server interface {
	Start(ctx context.Context) error
	Stop() error
}

type Client interface {
	Start(ctx context.Context) error
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

type Probe interface {
	AddLivenessFunc(name string, fn probe.HandleFunc)
	AddReadinessFunc(name string, fn probe.HandleFunc)
	LiveEndpoint(http.ResponseWriter, *http.Request)
	ReadyEndpoint(http.ResponseWriter, *http.Request)
}

func NewService(name string) Service {
	return newService(name)
}
