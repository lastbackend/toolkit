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
	"github.com/lastbackend/toolkit/pkg/server/http/websockets"
	"google.golang.org/grpc"
	"net/http"
)

type HTTPServer interface {
	Start(ctx context.Context) error
	Stop() error

	UseMiddleware(...KindMiddleware)

	SetMiddleware(name KindMiddleware, middleware HttpServerMiddleware)
	AddHandler(method, path string, h http.HandlerFunc, opts ...HTTPServerOption)

	SetService(fn interface{})
	GetService() interface{}

	Subscribe(event string, h websockets.EventHandler)

	ServerWS(w http.ResponseWriter, r *http.Request)
	SetCorsHandlerFunc(hf http.HandlerFunc)
	SetErrorHandlerFunc(hf func(http.ResponseWriter, error))
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

type HttpServerMiddleware func(h http.Handler) http.Handler
type KindMiddleware string

type GRPCServer interface {
	Start(ctx context.Context) error
	Stop() error

	SetDescriptor(descriptor grpc.ServiceDesc)

	SetService(constructor interface{})
	SetConstructor(fn interface{})

	GetService() interface{}
	GetConstructor() interface{}

	RegisterService(service interface{})
}

type GRPCServerDecorator interface {
	Start(ctx context.Context) error
	Stop() error
	SetService(constructor interface{})
}
