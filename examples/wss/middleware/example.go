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
	"context"
	"fmt"
	"net/http"

	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/server"
)

// RegisterExampleMiddleware - show request info middleware
func RegisterExampleMiddleware(runtime runtime.Runtime) server.HttpServerMiddleware {
	return &ExampleMiddleware{name: "example", runtime: runtime}
}

type ExampleMiddleware struct {
	name    server.KindMiddleware
	runtime runtime.Runtime
}

func (e *ExampleMiddleware) Apply(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Call: ExampleMiddleware")

		// Set example data to request context
		ctx := context.WithValue(r.Context(), "test-data", "example context data")

		h.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (e *ExampleMiddleware) Order() int {
	return 0
}

func (e *ExampleMiddleware) Kind() server.KindMiddleware {
	return e.name
}
