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

package http

import (
	"crypto/tls"
	"github.com/lastbackend/toolkit/pkg/server"
	"net/http"
)

const (
	defaultAddress                                          = ":8080"
	optionKindMiddleware              server.HttpOptionKind = "middleware"
	optionKindExcludeGlobalMiddleware server.HttpOptionKind = "excludeGlobalMiddleware"
)

const (
	MethodGet     = "GET"
	MethodHead    = "HEAD"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodPatch   = "PATCH" // RFC 5789
	MethodDelete  = "DELETE"
	MethodConnect = "CONNECT"
	MethodOptions = "OPTIONS"
	MethodTrace   = "TRACE"
)

type optionMiddleware struct {
	kind       server.HttpOptionKind
	middleware server.KindMiddleware
}

func (optionMiddleware) Kind() server.HttpOptionKind {
	return optionKindMiddleware
}

func WithMiddleware(middleware server.KindMiddleware) server.HTTPServerOption {
	return &optionMiddleware{kind: optionKindMiddleware, middleware: middleware}
}

type optionExcludeGlobalMiddleware struct {
	kind       server.HttpOptionKind
	middleware server.KindMiddleware
}

func (optionExcludeGlobalMiddleware) Kind() server.HttpOptionKind {
	return optionKindExcludeGlobalMiddleware
}

func WithExcludeGlobalMiddleware(middleware server.KindMiddleware) server.HTTPServerOption {
	return &optionMiddleware{kind: optionKindMiddleware, middleware: middleware}
}

type Config struct {
	Id string

	Host string `env:"SERVER_LISTEN" envDefault:"0.0.0.0" comment:"Set HTTP server listen host"`
	Port int    `env:"SERVER_PORT" envDefault:"8080" comment:"Set HTTP server listen port"`

	Prefix string

	EnableCORS bool
	IsDisable  bool

	TLSConfig *tls.Config
}

type Wrapper func(h http.Handler) http.Handler
