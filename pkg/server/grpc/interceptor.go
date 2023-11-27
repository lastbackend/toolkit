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

package grpc

import (
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
	"github.com/lastbackend/toolkit/pkg/server"
)

type Interceptors struct {
	log          logger.Logger
	constructors []interface{}
	items        map[server.KindInterceptor]server.GRPCInterceptor
}

func (i *Interceptors) AddConstructor(h interface{}) {
	i.constructors = append(i.constructors, h)
}

func (i *Interceptors) Add(h server.GRPCInterceptor) {
	i.items[h.Kind()] = h
}

func newInterceptors(log logger.Logger) *Interceptors {
	interceptors := Interceptors{
		log:          log,
		constructors: make([]interface{}, 0),
		items:        make(map[server.KindInterceptor]server.GRPCInterceptor),
	}

	return &interceptors
}
