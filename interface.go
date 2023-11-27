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
	"github.com/lastbackend/toolkit/pkg/client"
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
	"github.com/lastbackend/toolkit/pkg/runtime/meta"
	"github.com/lastbackend/toolkit/pkg/server"
)

type Service interface {
	Meta() *meta.Meta

	Log() logger.Logger
	Client() Client
	Server() Server

	RegisterConfig(config ...any) error
	RegisterPlugin(constructor ...any)
	RegisterPackage(constructor ...any)

	Start(ctx context.Context) error
	Stop(ctx context.Context, err error)

	RegisterOnStartHook(...func(ctx context.Context) error)
	RegisterOnStartSyncHook(...func(ctx context.Context) error)

	RegisterOnStopHook(...func(ctx context.Context) error)
	RegisterOnStopSyncHook(...func(ctx context.Context) error)
}

type Client interface {
	GRPC() client.GRPCClient
	HTTP() client.HTTPClient
}

type Server interface {
	HTTP() server.HTTPServer
	GRPC() server.GRPCServer

	HTTPGet(name string) server.HTTPServer
	HTTPNew(name string, options *server.HTTPServerOptions) server.HTTPServer

	GRPCGet(name string) server.GRPCServer
	GRPCNew(name string, options *server.GRPCServerOptions) server.GRPCServer
}

type Config any

type Package any

type PackageItem struct {
	Index  int
	Source any
}

type Plugin any
