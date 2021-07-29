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
	"github.com/lastbackend/engine/context/metadata"
	"github.com/lastbackend/engine/network/resolver"
	"github.com/lastbackend/engine/util/backoff"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/encoding"
	grpc_md "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	encoding.RegisterCodec(protoCodec{})
}

const (
	// The default resolver service for search service hosts
	defaultResolverService = "local"
	// The default number of times a request is tried
	defaultRetries = 5 * time.Second
	// The default request timeout
	defaultRequestTimeout = 15 * time.Second
	// The connection pool size
	defaultPoolSize = 100
	// The connection pool ttl
	defaultPoolTTL = time.Minute
	// DefaultPoolMaxStreams maximum streams on a connection (20)
	defaultPoolMaxStreams = 20
	// DefaultPoolMaxIdle maximum idle conns of a pool (50)
	defaultPoolMaxIdle = 50
	// DefaultMaxRecvMsgSize maximum message that client can receive (16 MB).
	defaultMaxRecvMsgSize = 1024 * 1024 * 16
	// DefaultMaxSendMsgSize maximum message that client can send (16 MB).
	defaultMaxSendMsgSize = 1024 * 1024 * 16
)

type client struct {
	prefix string
	opts   Options
	pool   *pool
}

func newClient(prefix string) *client {
	return &client{
		opts:   defaultOptions(),
		prefix: prefix,
		pool:   newPool(),
	}
}

func (c *client) Init(opts Options) error {
	c.opts = opts
	c.pool.Init(opts.Pool)
	return nil
}

func (c *client) Call(ctx context.Context, service, method string, body, resp interface{}, opts ...CallOption) error {
	if body == nil {
		return status.Error(codes.Internal, "request is nil")
	}
	if resp == nil {
		return status.Error(codes.Internal, "response is nil")
	}

	callOpts := c.opts.CallOptions
	for _, opt := range opts {
		opt(&callOpts)
	}

	ctx, cancel := context.WithTimeout(ctx, callOpts.RequestTimeout)
	defer cancel()

	headers := c.makeHeaders(ctx, service, callOpts)
	req := newRequest(service, method, body, headers)

	routes, err := resolver.DefaultResolver.Lookup(req.service)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	next, err := c.opts.Selector.Select(routes.Addresses())
	if err != nil {
		return err
	}

	invokeFunc := c.invoke

	b := &backoff.Backoff{
		Max: callOpts.Retries,
	}

	for {
		select {
		case <-ctx.Done():
			return status.Error(codes.Canceled, ctx.Err().Error())
		default:
			err := invokeFunc(ctx, next(), req, resp, callOpts)
			if err != nil {
				d := b.Duration()
				if d.Seconds() >= callOpts.Retries.Seconds() {
					return status.Error(codes.Internal, err.Error())
				}
				time.Sleep(d)
				continue
			}
			b.Reset()
			return nil
		}
	}

}

func (c *client) Stream(ctx context.Context, service, method string, body interface{}, opts ...CallOption) (Stream, error) {
	if body == nil {
		return nil, status.Error(codes.Internal, "request is nil")
	}

	callOpts := c.opts.CallOptions
	for _, opt := range opts {
		opt(&callOpts)
	}

	streamFunc := c.stream

	headers := c.makeHeaders(ctx, service, callOpts)
	req := newRequest(service, method, body, headers)

	routes, err := resolver.DefaultResolver.Lookup(req.service)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	next, err := c.opts.Selector.Select(routes.Addresses())
	if err != nil {
		return nil, err
	}

	b := &backoff.Backoff{
		Max: callOpts.Retries,
	}

	for {
		select {
		case <-ctx.Done():
			return nil, status.Error(codes.Canceled, ctx.Err().Error())
		default:
			s, err := streamFunc(ctx, next(), req, callOpts)
			if err != nil {
				d := b.Duration()
				if d.Seconds() >= callOpts.Retries.Seconds() {
					return nil, status.Error(codes.Internal, err.Error())
				}
				time.Sleep(d)
				continue
			}
			b.Reset()
			return s, nil
		}
	}
}

func (c *client) Close() error {
	return nil
}

func (c *client) invoke(ctx context.Context, addr string, req *request, rsp interface{}, opts CallOptions) error {

	md := grpc_md.New(req.headers)
	ctx = grpc_md.NewOutgoingContext(ctx, md)

	var gErr error
	conn, err := c.pool.getConn(addr, c.makeGrpcDialOptions()...)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Failed sending request: %v", err))
	}
	defer conn.pool.release(addr, conn, gErr)

	ch := make(chan error, 1)
	go func() {
		ch <- conn.Invoke(ctx, req.Method(), req.Body(), rsp, c.makeGrpcCallOptions(opts)...)
	}()

	select {
	case err := <-ch:
		gErr = err
	case <-ctx.Done():
		gErr = status.Error(codes.Internal, ctx.Err().Error())
	}

	return gErr
}

func (c *client) stream(ctx context.Context, addr string, req *request, opts CallOptions) (Stream, error) {

	md := grpc_md.New(req.Headers())
	ctx = grpc_md.NewOutgoingContext(ctx, md)

	cc, err := c.pool.getConn(addr, c.makeGrpcDialOptions()...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	desc := &grpc.StreamDesc{
		StreamName:    req.Method(),
		ClientStreams: true,
		ServerStreams: true,
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)

	st, err := cc.NewStream(ctx, desc, req.Method(), c.makeGrpcCallOptions(opts)...)
	if err != nil {
		cancel()
		c.pool.release(addr, cc, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	s := &stream{
		ClientStream: st,
		context:      ctx,
		request:      req,
		conn:         cc,
		close: func(err error) {
			if err != nil {
				cancel()
			}
			c.pool.release(addr, cc, err)
		},
	}

	if err := st.SendMsg(req.body); err != nil {
		return nil, err
	}

	if err := st.CloseSend(); err != nil {
		return nil, err
	}

	return s, nil
}

func (c *client) withPrefix(name string) string {
	return fmt.Sprintf("%s-%s", c.prefix, name)
}

func (c *client) withEnvPrefix(name string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s", c.prefix, name))
}

func (c *client) makeGrpcCallOptions(opts CallOptions) []grpc.CallOption {
	grpcCallOptions := make([]grpc.CallOption, 0)

	if opts.MaxCallRecvMsgSize > 0 {
		grpcCallOptions = append(grpcCallOptions, grpc.MaxCallRecvMsgSize(opts.MaxCallRecvMsgSize))
	}
	if opts.MaxCallSendMsgSize > 0 {
		grpcCallOptions = append(grpcCallOptions, grpc.MaxCallSendMsgSize(opts.MaxCallSendMsgSize))
	}
	if opts.MaxRetryRPCBufferSize > 0 {
		grpcCallOptions = append(grpcCallOptions, grpc.MaxRetryRPCBufferSize(opts.MaxRetryRPCBufferSize))
	}
	if opts.CallContentSubtype != "" {
		grpcCallOptions = append(grpcCallOptions, grpc.CallContentSubtype(opts.CallContentSubtype))
	}

	return grpcCallOptions
}

func (c *client) makeGrpcDialOptions() []grpc.DialOption {
	grpcDialOptions := make([]grpc.DialOption, 0)

	// TODO: implement auths
	grpcDialOptions = append(grpcDialOptions, grpc.WithInsecure())

	if c.opts.MaxRecvMsgSize != nil || c.opts.MaxSendMsgSize != nil {
		var defaultCallOpts = make([]grpc.CallOption, 0)
		if c.opts.MaxRecvMsgSize != nil {
			defaultCallOpts = append(defaultCallOpts, grpc.MaxCallRecvMsgSize(*c.opts.MaxRecvMsgSize))
		}
		if c.opts.MaxSendMsgSize != nil {
			defaultCallOpts = append(defaultCallOpts, grpc.MaxCallSendMsgSize(*c.opts.MaxSendMsgSize))
		}
		grpcDialOptions = append(grpcDialOptions, grpc.WithDefaultCallOptions(defaultCallOpts...))
	}
	if c.opts.WriteBufferSize != nil {
		grpcDialOptions = append(grpcDialOptions, grpc.WithWriteBufferSize(*c.opts.WriteBufferSize))
	}
	if c.opts.ReadBufferSize != nil {
		grpcDialOptions = append(grpcDialOptions, grpc.WithReadBufferSize(*c.opts.ReadBufferSize))
	}
	if c.opts.InitialWindowSize != nil {
		grpcDialOptions = append(grpcDialOptions, grpc.WithInitialWindowSize(*c.opts.InitialWindowSize))
	}
	if c.opts.InitialConnWindowSize != nil {
		grpcDialOptions = append(grpcDialOptions, grpc.WithInitialConnWindowSize(*c.opts.InitialConnWindowSize))
	}
	if c.opts.UserAgent != nil {
		grpcDialOptions = append(grpcDialOptions, grpc.WithUserAgent(*c.opts.UserAgent))
	}
	if c.opts.MaxHeaderListSize != nil {
		grpcDialOptions = append(grpcDialOptions, grpc.WithMaxHeaderListSize(uint32(*c.opts.MaxHeaderListSize)))
	}

	return grpcDialOptions
}

func (c *client) makeHeaders(ctx context.Context, service string, opts CallOptions) map[string]string {
	var headers = make(map[string]string, 0)

	if md, ok := metadata.LoadFromContext(ctx); ok {
		for k, v := range md {
			headers[strings.ToLower(k)] = v
		}
	}
	if opts.Headers != nil {
		for k, v := range opts.Headers {
			headers[strings.ToLower(k)] = v
		}
	}

	if _, ok := headers["content-type"]; !ok {
		headers["content-type"] = c.opts.ContentType
	}

	headers["x-service-name"] = service

	return headers
}
