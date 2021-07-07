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
	"github.com/lastbackend/engine/network/resolver"
	"github.com/lastbackend/engine/network/resolver/local"
	"github.com/lastbackend/engine/plugin/rpc/grpc/selector"
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

func retryNever(ctx context.Context, req *request, retryCount int, err error) (bool, error) {
	return false, nil
}

type CallFunc func(ctx context.Context, addr string, req *request, rsp interface{}, opts CallOptions) error
type CallMiddlewareFunc func(CallFunc) CallFunc
type MiddlewareFunc func(Client) Client
type LookupFunc func(context.Context, *request, CallOptions) ([]string, error)

type Options struct {
	Context context.Context

	Addresses       []string
	ContentType     string
	Proxy           string
	ResolverService string
	MaxRecvMsgSize  int
	MaxSendMsgSize  int

	Selector selector.Selector
	Resolver resolver.Resolver

	Middlewares []MiddlewareFunc

	Pool        PoolOptions
	CallOptions CallOptions
}

type BackoffFunc func(ctx context.Context, req *request, attempts int) (time.Duration, error)
type RetryFunc func(ctx context.Context, req *request, retryCount int, err error) (bool, error)

type CallOptions struct {
	AuthToken      bool
	Backoff        BackoffFunc
	Retry          RetryFunc
	Retries        time.Duration
	DialTimeout    time.Duration
	RequestTimeout time.Duration
	StreamTimeout  time.Duration
	Middlewares    []CallMiddlewareFunc
	Context        context.Context
}

func defaultOptions() Options {
	slc, _ := selector.New(selector.RoundRobin)
	return Options{
		Context:     context.Background(),
		ContentType: "application/protobuf",
		Selector:    slc,
		Resolver:    local.NewResolver(),
		CallOptions: CallOptions{
			Backoff:        exponentialBackoff,
			Retry:          retryNever,
			Retries:        defaultRetries,
			RequestTimeout: defaultRequestTimeout,
		},
		Pool: PoolOptions{
			Size: converter.ToIntPointer(defaultPoolSize),
			Ttl:  converter.ToDurationPointer(defaultPoolTTL),
		},
	}
}
