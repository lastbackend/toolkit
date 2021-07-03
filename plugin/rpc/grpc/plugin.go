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

package grpc

import (
	"context"
	"github.com/lastbackend/engine/plugin"
)

const (
	PluginName    = "grpc"
	defaultPrefix = "grpc"
)

// Register - registers the plugin implements rpc client using gRPC as a transport
func Register(f plugin.RegisterFunc) plugin.CreatorFunc {
	return func(o plugin.Option) interface{} {
		p := newRpc(o.Prefix)
		f(p)
		return p.getClient()
	}
}

type Storage interface {
	NewRequest(service, method string, req interface{}, reqOpts ...RequestOption) Request
	Call(ctx context.Context, req Request, rsp interface{}, opts ...CallOption) error
}

type Request interface {
	Service() string
	Method() string
	Endpoint() string
	Body() interface{}
	Stream() bool
}

type RequestOption interface {
}

type CallOption func(*CallOptions)
