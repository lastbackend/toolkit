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
	"net/http"
)

const MiddlewareNotFoundError string = "Can not apply middleware router: %s Can not find global server middleware: %s. To " +
	"register middleware, please add Server().HTTP().SetMiddleware(\"%s\", http.Handler) to runtime."

type Middlewares struct {
	log    logger.Logger
	global []server.KindMiddleware

	items map[server.KindMiddleware]server.HttpServerMiddleware
}

func (m *Middlewares) SetGlobal(middlewares ...server.KindMiddleware) {
	for _, item := range middlewares {
		if item != "" {
			m.global = append(m.global, item)
		}
	}
}

func (m *Middlewares) Add(name server.KindMiddleware, h server.HttpServerMiddleware) {
	m.items[name] = h
}

func (m *Middlewares) apply(h http.Handler, handler server.HTTPServerHandler) error {

	for _, opt := range handler.Options {

		if opt.Kind() != optionKindMiddleware {
			continue
		}

		o, ok := opt.(optionMiddleware)

		if !ok {
			continue
		}

		for _, g := range m.global {
			if g == o.middleware {
				continue
			}
		}

		if middleware, ok := m.items[o.middleware]; ok {
			h = middleware(h)
		} else {
			m.log.Errorf(MiddlewareNotFoundError, handler.Path, o.middleware, o.middleware)
			return fmt.Errorf("can not find global server middleware: %s", o.middleware)
		}
	}

	for _, g := range m.global {
		if middleware, ok := m.items[g]; ok {
			h = middleware(h)
		} else {
			m.log.Errorf(MiddlewareNotFoundError, handler.Path, g, g)
			return fmt.Errorf("can not find global server middleware: %s", g)
		}
	}

	return nil
}
func newMiddlewares(log logger.Logger) *Middlewares {
	middlewares := Middlewares{
		log:    log,
		global: make([]server.KindMiddleware, 0),
		items:  make(map[server.KindMiddleware]server.HttpServerMiddleware),
	}

	return &middlewares
}
