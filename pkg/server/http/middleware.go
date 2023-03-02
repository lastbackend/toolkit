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
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
	"github.com/lastbackend/toolkit/pkg/server"
	"net/http"
)

type Middlewares struct {
	log    logger.Logger
	global []string
	items  map[string]server.HttpServerMiddleware
}

func (m *Middlewares) SetGlobal(name ...string) {
	m.global = append(m.global, name...)
}

func (m *Middlewares) Add(name string, h server.HttpServerMiddleware) {
	m.items[name] = h
}

func (m *Middlewares) apply(h http.Handler, opts []server.HTTPServerOption) http.Handler {

	for _, g := range m.global {
		if middleware, ok := m.items[g]; ok {
			h = middleware(h)
		} else {
			m.log.Errorf("can not find global middleware: %s", g)
		}
	}

	for _, opt := range opts {

		if opt.Kind() != optionKindMiddleware {
			continue
		}

		o, ok := opt.(optionMiddleware)

		if !ok {
			continue
		}

		if middleware, ok := m.items[o.middleware]; ok {
			h = middleware(h)
		} else {
			m.log.Errorf("can not find global middleware: %s", o.middleware)
		}
	}

	return h
}
func newMiddlewares(log logger.Logger) *Middlewares {
	middlewares := Middlewares{
		log:    log,
		global: make([]string, 0),
		items:  make(map[string]server.HttpServerMiddleware),
	}

	return &middlewares
}
