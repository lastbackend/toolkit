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
	"github.com/lastbackend/toolkit/pkg/client/grpc/selector"
	"github.com/lastbackend/toolkit/pkg/util/converter"

	"context"
	"math"
	"time"
)

func exponentialBackoff(ctx context.Context, req *request, attempts int) (time.Duration, error) {
	if attempts > 10 {
		return 2 * time.Minute, nil
	}
	return time.Duration(math.Pow(float64(attempts), math.E)) * time.Millisecond * 100, nil
}

type CallFunc func(ctx context.Context, addr string, req *request, rsp interface{}, opts CallOptions) error
type CallMiddlewareFunc func(CallFunc) CallFunc
type MiddlewareFunc func(Client) Client
type LookupFunc func(context.Context, *request, CallOptions) ([]string, error)

type Options struct {
	Context context.Context

	ContentType string `env:"CONTENT_TYPE"  envDefault:"application/protobuf" comment:"Set GRPC client request content-type"`

	// Grpc DialOptions
	WriteBufferSize       *int    `env:"WRITE_BUFFER_SIZE" comment:"Sets the how much data can be batched before doing a write on the wire. The corresponding memory allocation for this buffer will be twice the size to keep syscalls low. Zero will disable the write buffer (default 32 KB)"`
	ReadBufferSize        *int    `env:"READ_BUFFER_SIZE" comment:"Sets the size of the reading buffer, this determines how\nmuch data can be read at most for each read syscall. Zero will disable read buffer (default 32 KB)")"`
	InitialWindowSize     *int32  `env:"INITIAL_WINDOW_SIZE" comment:"Sets the value for initial window size on a stream. The lower bound for window size is 64K and any value smaller than that will be ignored."`
	InitialConnWindowSize *int32  `env:"INITIAL_CONN_WINDOW_SIZE" comment:"Sets the value for initial window size on a connection. The lower bound for window size is 64K and any value smaller than that will be ignored."`
	MaxHeaderListSize     *int32  `env:"MAX_HEADER_LIST_SIZE" comment:"Sets the specifies the maximum (uncompressed) size of header list that the client is prepared to accept"`
	MaxRecvMsgSize        *int    `env:"MAX_RECV_MSG_SIZE" comment:"Sets the maximum message size in bytes the client can receive (default 16 MB)"`
	MaxSendMsgSize        *int    `env:"MAX_SEND_MSG_SIZE" comment:"Sets the maximum message size in bytes the client can send (default 16 MB)"`
	UserAgent             *string `env:"USER_AGENT"  envDefault:"application/protobuf" comment:"Sets the specifies a user agent string for all the RPCs"`
	Selector              selector.Selector

	Pool        PoolOptions
	CallOptions CallOptions
}

type BackoffFunc func(ctx context.Context, req *request, attempts int) (time.Duration, error)
type RetryFunc func(ctx context.Context, req *request, retryCount int, err error) (bool, error)

type CallOptions struct {
	Backoff               BackoffFunc
	Retries               time.Duration
	RequestTimeout        time.Duration
	Context               context.Context
	Headers               map[string]string
	MaxCallSendMsgSize    int
	MaxCallRecvMsgSize    int
	MaxRetryRPCBufferSize int
	CallContentSubtype    string
}

func defaultOptions() Options {
	slc, _ := selector.New(selector.RoundRobin)
	return Options{
		Context:     context.Background(),
		ContentType: "application/protobuf",
		Selector:    slc,
		CallOptions: CallOptions{
			Backoff:        exponentialBackoff,
			Retries:        defaultRetries,
			RequestTimeout: defaultRequestTimeout,
		},
		Pool: PoolOptions{
			Size: converter.ToIntPointer(defaultPoolSize),
			TTL:  converter.ToDurationPointer(defaultPoolTTL),
		},
	}
}

func Headers(h map[string]string) CallOption {
	return func(o *CallOptions) {
		o.Headers = h
	}
}

func MaxCallSendMsgSize(bytes int) CallOption {
	return func(o *CallOptions) {
		o.MaxCallSendMsgSize = bytes
	}
}

func MaxCallRecvMsgSize(bytes int) CallOption {
	return func(o *CallOptions) {
		o.MaxCallRecvMsgSize = bytes
	}
}

func RequestTimeout(timeout time.Duration) CallOption {
	return func(o *CallOptions) {
		o.RequestTimeout = timeout
	}
}
