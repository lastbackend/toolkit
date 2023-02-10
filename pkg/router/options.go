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
	"crypto/tls"
	"net/http"

	"github.com/google/uuid"
)

const (
	defaultAddress = ":8080"
)

type Option func(*Options)

type Options struct {
	Id      string
	Address string
	Prefix  string

	EnableCORS bool

	Wrappers  []Wrapper
	TLSConfig *tls.Config
}

type Wrapper func(h http.Handler) http.Handler

func defaultOptions() Options {
	return Options{
		Id:      uuid.New().String(),
		Address: defaultAddress,
	}
}
