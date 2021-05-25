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

package server

import (
	"context"
)

var (
	DefaultServer Server
)

type Server interface {
	Start() error
	Stop() error
	Handle(Handler) error
	NewHandler(interface{}, ...HandlerOption) Handler
}

type Request interface {
	Service() string
	Method() string
	Endpoint() string
	ContentType() string
	Header() map[string]string
	Body() interface{}
	Read() ([]byte, error)
	Stream() bool
}

type Response interface {
	WriteHeader(map[string]string)
	Write([]byte) error
}

type Stream interface {
	Context() context.Context
	Request() Request
	Send(interface{}) error
	Recv(interface{}) error
	Error() error
	Close() error
}

type Handler interface {
	Name() string
	Handler() interface{}
}
