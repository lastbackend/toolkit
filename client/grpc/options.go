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
	"github.com/lastbackend/engine/client/grpc/selector"
	"github.com/lastbackend/engine/util/converter"

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

	ContentType string

	// Grpc DialOptions
	WriteBufferSize       *int
	ReadBufferSize        *int
	InitialWindowSize     *int32
	InitialConnWindowSize *int32
	MaxHeaderListSize     *int32
	MaxRecvMsgSize        *int
	MaxSendMsgSize        *int
	UserAgent             *string

	Selector selector.Selector

	Pool        PoolOptions
	CallOptions CallOptions
}

type BackoffFunc func(ctx context.Context, req *request, attempts int) (time.Duration, error)
type RetryFunc func(ctx context.Context, req *request, retryCount int, err error) (bool, error)

type CallOptions struct {
	Backoff BackoffFunc
	Retries time.Duration
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
			Ttl:  converter.ToDurationPointer(defaultPoolTTL),
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
