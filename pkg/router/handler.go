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
	"encoding/json"
	"net/http"
)

type HandleOption func(*HandleOptions)

type HandleOptions struct {
	Middlewares Middleware
}

func corsHandlerFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Methods", "OPTIONS,GET,POST,PUT,DELETE")
	w.Header().Add("Access-Control-Expose-Headers", "Content-Disposition")
	w.Header().Add("Access-Control-Allow-Headers", "Authorization,Content-Type,Origin,X-Tools-Name,X-Requested-With,Content-Name,Accept,Accept-Range,Range")
	w.Header().Add("Content-Type", "application/json")
}

func (s *server) methodNotAllowedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.opts.EnableCORS {
			corsHandlerFunc(w, r)
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write(makeResponse(http.StatusMethodNotAllowed, "Method Not Allowed"))
	})
}

func (s *server) methodNotFoundHandler() http.Handler {
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
