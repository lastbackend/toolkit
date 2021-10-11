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
	"github.com/lastbackend/engine/client"

	"context"
)

const (
	PluginName    = "grpc"
	defaultPrefix = "grpc"
)

// Register - client the plugin implements rpc client using gRPC as a transport
func Register(f client.RegisterFunc) client.CreatorFunc {
	return func(o client.Option) interface{} {
		p := newRpc(o.Prefix)
		f(p)
		return p.getClient()
	}
}

type Client interface {
	Call(ctx context.Context, service, method string, req, rsp interface{}, opts ...CallOption) error
	Stream(ctx context.Context, service, method string, body interface{}, opts ...CallOption) (Stream, error)
}

type CallOption func(*CallOptions)

type Stream interface {
	Context() context.Context
	Headers() map[string]string
	SendMsg(m interface{}) error
	RecvMsg(m interface{}) error
	CloseSend() error
}
