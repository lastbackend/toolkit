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

package grpc

import (
	"context"
	"fmt"
	"strings"
)

const (
	defaultPrefix = "grpc"
)

type RPCClient interface {
	Start(context.Context) error
	Stop() error
	Client() *GRPCClient
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

type rpcClient struct {
	client *GRPCClient

	prefix string
	opts   Options
}

type ClientOptions struct {
	Name string
}

// NewClient - client the plugin implements rpc client using gRPC as a transport
func NewClient(opts ClientOptions) *GRPCClient {
	c := new(rpcClient)
	if len(opts.Name) == 0 {
		c.prefix = defaultPrefix
	}
	c.prefix = opts.Name
	c.opts = defaultOptions()
	c.client = newClient(opts.Name)

	return c.Client()
}

func (s *rpcClient) Client() *GRPCClient {
	return s.client
}

func (s *rpcClient) Start(_ context.Context) error {
	return s.client.Init(s.opts)
}

func (s *rpcClient) Stop() error {
	return s.client.Close()
}

func (s *rpcClient) withPrefix(name string) string {
	return fmt.Sprintf("%s-%s", s.prefix, name)
}

func (s *rpcClient) withEnvPrefix(name string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s", s.prefix, name))
}
