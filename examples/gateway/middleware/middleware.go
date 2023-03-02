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

package middleware

import (
	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/examples/gateway/config"
)

type Middleware struct {
	//config *config.Config
	//svc    toolkit.Service
}

func New(cfg *config.Config, svc toolkit.Service) Middleware {
	//mw := Middleware{
	//	config: cfg,
	//	svc:    svc,
	//}

	//svc.Router().AddMiddleware(mw.RequestID)

	// Example entities
	//pb.HelloWorldMiddlewareAdd(mw.ExampleMiddleware)

	//return mw
	return Middleware{}
}
