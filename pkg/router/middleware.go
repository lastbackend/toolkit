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
	"net/http"
)

type Middleware struct {
	items []func(h http.Handler) http.Handler
}

func (m *Middleware) Add(h func(h http.Handler) http.Handler) {
	m.items = append(m.items, h)
}

func (m *Middleware) apply(h http.Handler) http.Handler {
	wrapped := h
	for i := len(m.items) - 1; i >= 0; i-- {
		wrapped = m.items[i](wrapped)
	}
	return wrapped
}
