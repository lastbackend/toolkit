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

package router

import (
	"context"
	"github.com/lastbackend/toolkit/pkg/router/ws"
	"net/http"
)

type Server interface {
	Start(ctx context.Context) error
	Stop() error
	AddMiddleware(h func(http.Handler) http.Handler)
	Handle(method string, path string, h http.HandlerFunc, opts HandleOptions)
	Subscribe(event string, h ws.EventHandler)
	ServerWS(w http.ResponseWriter, r *http.Request)
	SetCorsHandlerFunc(hf http.HandlerFunc)
	SetErrorHandlerFunc(hf func(http.ResponseWriter, error))
}
