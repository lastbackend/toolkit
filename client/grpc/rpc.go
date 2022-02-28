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
	"fmt"
	"github.com/lastbackend/engine"
	"strings"
)

const (
	defaultPrefix = "grpc"
)

type RPCClient interface {
	engine.Client

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
func NewClient(app engine.Service, opts *ClientOptions) RPCClient {
	c := new(rpcClient)
	if len(opts.Name) == 0 {
		c.prefix = defaultPrefix
	}
	c.prefix = opts.Name
	c.opts = defaultOptions()
	c.client = newClient(opts.Name)

	c.addFlags(app)

	return c
}

func (s *rpcClient) Client() *GRPCClient {
	return s.client
}

func (s *rpcClient) Start() error {
	return s.client.Init(s.opts)
}

func (s *rpcClient) Stop() error {
	return s.client.Close()
}

func (s *rpcClient) addFlags(app engine.Service) {

	app.CLI().AddIntFlag(s.withPrefix("pool-size"), s.opts.Pool.Size).
		Env(s.withEnvPrefix("POOL_SIZE")).
		Usage("Set pool size").
		Default(defaultPoolSize)

	app.CLI().AddDurationFlag(s.withPrefix("pool-ttl"), s.opts.Pool.TTL).
		Env(s.withEnvPrefix("POOL_TTL")).
		Usage("Set pool ttl").
		Default(defaultPoolTTL)

	app.CLI().AddIntFlag(s.withPrefix("max-recv-msg-size"), s.opts.MaxRecvMsgSize).
		Env(s.withEnvPrefix("MAX_RECV_MSG_SIZE")).
		Usage("Sets the maximum message size in bytes the client can receive (default 16 MB)").
		Default(defaultMaxRecvMsgSize)

	app.CLI().AddIntFlag(s.withPrefix("max-send-msg-size"), s.opts.MaxSendMsgSize).
		Env(s.withEnvPrefix("MAX_SEND_MSG_SIZE")).
		Usage("Sets the maximum message size in bytes the client can send (default 16 MB)").
		Default(defaultMaxSendMsgSize)

	app.CLI().AddInt32Flag(s.withPrefix("max-header-list-size"), s.opts.MaxHeaderListSize).
		Env(s.withEnvPrefix("MAX_HEADER_LIST_SIZE")).
		Usage("Sets the specifies the maximum (uncompressed) size of header list that the client is prepared to accept")

	app.CLI().AddIntFlag(s.withPrefix("write-buffer-size"), s.opts.WriteBufferSize).
		Env(s.withEnvPrefix("WRITE_BUFFER_SIZE")).
		Usage("Sets the how much data can be batched before doing a write on the wire.\nThe corresponding memory allocation for this buffer will be twice the size to keep syscalls low.\nZero will disable the write buffer (default 32 KB)")

	app.CLI().AddIntFlag(s.withPrefix("read-buffer-size"), s.opts.ReadBufferSize).
		Env(s.withEnvPrefix("READ_BUFFER_SIZE")).
		Usage("Sets the size of the reading buffer, this determines how\nmuch data can be read at most for each read syscall. Zero will disable read buffer (default 32 KB)")

	app.CLI().AddInt32Flag(s.withPrefix("initial-window-size"), s.opts.InitialWindowSize).
		Env(s.withEnvPrefix("INITIAL_WINDOW_SIZE")).
		Usage("Sets the value for initial window size on a stream.\nThe lower bound for window size is 64K and any value smaller than that will be ignored.")

	app.CLI().AddInt32Flag(s.withPrefix("initial-conn-window-size"), s.opts.InitialConnWindowSize).
		Env(s.withEnvPrefix("INITIAL_CONN_WINDOW_SIZE")).
		Usage("Sets the value for initial window size on a connection.\nThe lower bound for window size is 64K and any value smaller than that will be ignored.")

	app.CLI().AddStringFlag(s.withPrefix("user-agent"), s.opts.UserAgent).
		Env(s.withEnvPrefix("USER_AGENT")).
		Usage("Sets the specifies a user agent string for all the RPCs")

}

func (s *rpcClient) withPrefix(name string) string {
	return fmt.Sprintf("%s-%s", s.prefix, name)
}

func (s *rpcClient) withEnvPrefix(name string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s", s.prefix, name))
}
