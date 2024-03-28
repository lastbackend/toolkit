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

package server

import (
	"context"
	"crypto/tls"
	"github.com/lastbackend/toolkit/pkg/server/http/marshaler"
	"github.com/lastbackend/toolkit/pkg/server/http/websockets"
	"google.golang.org/grpc"
	"net/http"
)

type HTTPServer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	UseMiddleware(...KindMiddleware)
	UseMarshaler(contentType string, marshaler marshaler.Marshaler) error

	GetMiddlewares() []any
	SetMiddleware(middleware any)

	SetService(service interface{})
	GetService() interface{}

	AddHandler(method, path string, h http.HandlerFunc, opts ...HTTPServerOption)

	GetConstructor() interface{}

	Subscribe(event string, h websockets.EventHandler)
	Info() ServerInfo

	ServerWS(w http.ResponseWriter, r *http.Request)
	SetCorsHandlerFunc(hf http.HandlerFunc)
	SetErrorHandlerFunc(hf func(http.ResponseWriter, error))
}

type HTTPServerOptions struct {
	Host string
	Port int

	TLSConfig *tls.Config
}

type ServerInfo struct {
	Kind ServerKind
	Host string
	Port int

	TLSConfig *tls.Config
}

type HTTPServerHandler struct {
	Method  string
	Path    string
	Handler http.HandlerFunc
	Options []HTTPServerOption
}

type HttpOptionKind string

type HTTPServerOption interface {
	Kind() HttpOptionKind
}

type DefaultHttpServerMiddleware struct{}

func (DefaultHttpServerMiddleware) Order() int {
	return 0
}

type HttpServerMiddleware interface {
	Apply(h http.HandlerFunc) http.HandlerFunc
	Kind() KindMiddleware
	Order() int
}

type KindMiddleware string
type KindInterceptor string

type GRPCServer interface {
	Start(ctx context.Context) error
	Stop() error

	SetDescriptor(descriptor grpc.ServiceDesc)

	SetService(constructor interface{})
	SetConstructor(fn interface{})

	GetInterceptors() []interface{}
	SetInterceptor(interceptor any)

	GetService() interface{}
	GetConstructor() interface{}
	GetInterceptorsConstructor() interface{}

	RegisterService(service interface{})
	Info() ServerInfo
}

type GRPCServerOptions struct {
	Host string
	Port int

	TLSConfig *tls.Config
}

type ServerKind string

type GRPCInterceptor interface {
	Kind() KindInterceptor
	Order() int
	Interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error)
}

const (
	ServerKindHTTPServer = "http"
	ServerKindGRPCServer = "grpc"
)
