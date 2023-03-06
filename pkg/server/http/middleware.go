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
	"fmt"
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
	"github.com/lastbackend/toolkit/pkg/server"
	"go.uber.org/fx"
	"net/http"
)

const MiddlewareNotFoundError string = "Can not apply middleware router: %s Can not find global server middleware: %s. To " +
	"register middleware, please add Server().HTTP().SetMiddleware(\"%s\", http.Handler) to runtime."

type Middlewares struct {
	log          logger.Logger
	global       []server.KindMiddleware
	constructors []interface{}
	items        map[server.KindMiddleware]server.HttpServerMiddleware
}

func (m *Middlewares) SetGlobal(middlewares ...server.KindMiddleware) {
	for _, item := range middlewares {
		if item != "" {
			m.global = append(m.global, item)
		}
	}
}

func (m *Middlewares) AddConstructor(h interface{}) {
	m.constructors = append(m.constructors, h)
}

func (m *Middlewares) Add(h server.HttpServerMiddleware) {
	m.items[h.Kind()] = h
}

func (m *Middlewares) apply(handler server.HTTPServerHandler) (http.HandlerFunc, error) {

	h := handler.Handler

	var (
		exclude = make(map[server.KindMiddleware]server.HttpServerMiddleware, 0)
		mws     = make(map[server.KindMiddleware]server.HttpServerMiddleware, 0)
	)

	for _, opt := range handler.Options {
		if opt.Kind() != optionKindMiddleware {
			continue
		}

		o, ok := opt.(*optionMiddleware)

		if !ok {
			continue
		}

		for _, g := range m.global {
			if g == o.middleware {
				continue
			}
		}

		middleware, ok := m.items[o.middleware]
		if !ok {
			m.log.Errorf(MiddlewareNotFoundError, handler.Path, o.middleware, o.middleware)
			return h, fmt.Errorf("can not find global server middleware: %s", o.middleware)
		}

		mws[o.middleware] = middleware
	}

	for _, opt := range handler.Options {
		if opt.Kind() != optionKindExcludeGlobalMiddleware {
			continue
		}

		o, ok := opt.(*optionExcludeGlobalMiddleware)

		if !ok {
			continue
		}

		middleware, ok := m.items[o.middleware]
		if !ok {
			m.log.Errorf(MiddlewareNotFoundError, handler.Path, o.middleware, o.middleware)
			continue
		}
		exclude[o.middleware] = middleware
	}

	for _, g := range m.global {

		if _, ok := exclude[g]; ok {
			continue
		}

		middleware, ok := m.items[g]
		if !ok {
			m.log.Errorf(MiddlewareNotFoundError, handler.Path, g, g)
			return h, fmt.Errorf("can not find global server middleware: %s", g)
		}

		mws[g] = middleware
	}

	for n, mw := range mws {
		m.log.V(5).Infof("apply middleware %s to %s", n, handler.Path)
		h = mw.Apply(h)
	}

	return h, nil
}

type MiddlewareRegistry struct {
	fx.In
}

func newMiddlewares(log logger.Logger) *Middlewares {
	middlewares := Middlewares{
		log:          log,
		global:       make([]server.KindMiddleware, 0),
		constructors: make([]interface{}, 0),
		items:        make(map[server.KindMiddleware]server.HttpServerMiddleware),
	}

	return &middlewares
}
