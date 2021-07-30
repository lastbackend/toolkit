/*
Copyright [2014] - [2021] The Last.Backend authors.

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
	"net/http"
)

func AddCorsHandler(h http.Handler) http.Handler {
	return corsHandler{h}
}

type corsHandler struct {
	handler http.Handler
}

func (c corsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	SetHeaders(w, r)

	if r.Method == "OPTIONS" {
		return
	}

	c.handler.ServeHTTP(w, r)
}

func SetHeaders(w http.ResponseWriter, r *http.Request) {
	wrapper := func(w http.ResponseWriter, k, v string) {
		if v := w.Header().Get(k); len(v) > 0 {
			return
		}
		w.Header().Set(k, v)
	}

	if origin := r.Header.Get("Origin"); len(origin) > 0 {
		wrapper(w, "Access-Control-Allow-Origin", origin)
	} else {
		wrapper(w, "Access-Control-Allow-Origin", "*")
	}

	wrapper(w, "Access-Control-Allow-Credentials", "true")
	wrapper(w, "Access-Control-Allow-Methods", "POST, PATCH, GET, OPTIONS, PUT, DELETE")
	wrapper(w, "Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}
