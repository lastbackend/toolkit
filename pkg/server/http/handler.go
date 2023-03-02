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
	"encoding/json"
	"github.com/lastbackend/toolkit/pkg/server"
	"net/http"
)

type HandleOption func(*HandleOptions)

type HandleOptions struct {
	Middlewares server.HttpServerMiddleware
}

type Handler func(w http.ResponseWriter, r *http.Request) error

func corsHandlerFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Methods", "OPTIONS,GET,POST,PUT,DELETE")
	w.Header().Add("Access-Control-Expose-Headers", "Content-Disposition")
	w.Header().Add("Access-Control-Allow-Headers", "Authorization,Content-Type,Origin,X-Tools-Name,X-Requested-With,Content-Name,Accept,Accept-Range,Range")
	w.Header().Add("Content-Type", "application/json")
}

func (s *httpServer) methodNotAllowedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.opts.EnableCORS {
			corsHandlerFunc(w, r)
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write(makeResponse(http.StatusMethodNotAllowed, "Method Not Allowed"))
	})
}

func (s *httpServer) methodNotFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.opts.EnableCORS {
			corsHandlerFunc(w, r)
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write(makeResponse(http.StatusNotFound, "Not Found"))
	})
}

type response struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func makeResponse(code int, message string) []byte {
	r, _ := json.Marshal(response{
		Code:    code,
		Status:  http.StatusText(code),
		Message: message,
	})
	return r
}
