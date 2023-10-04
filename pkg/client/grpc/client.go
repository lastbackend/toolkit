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
  "time"

  "github.com/lastbackend/toolkit/pkg/client"
  "github.com/lastbackend/toolkit/pkg/client/grpc/resolver"
  "github.com/lastbackend/toolkit/pkg/client/grpc/resolver/file"
  "github.com/lastbackend/toolkit/pkg/client/grpc/resolver/local"
  "github.com/lastbackend/toolkit/pkg/context/metadata"
  "github.com/lastbackend/toolkit/pkg/runtime"
  "github.com/lastbackend/toolkit/pkg/util/backoff"
  "google.golang.org/grpc"
  "google.golang.org/grpc/codes"
  "google.golang.org/grpc/credentials/insecure"
  "google.golang.org/grpc/encoding"
  grpc_md "google.golang.org/grpc/metadata"
  "google.golang.org/grpc/status"
)

func init() {
  encoding.RegisterCodec(protoCodec{})
}

const (
  // default prefix
  defaultPrefix = "GRPC_CLIENT"
  // default pool name
  defaultPoolName = ""
  // default GRPC port
  defaultPort = 9000
  // The default number of times a request is tried
  defaultRetries = 0 * time.Second
  // The default request timeout
  defaultRequestTimeout = 15 * time.Second
  // The connection pool size
  defaultPoolSize = 100
  // The connection pool ttl
  defaultPoolTTL = time.Minute
  // DefaultMaxRecvMsgSize maximum message that client can receive (16 MB).
  defaultMaxRecvMsgSize = 1024 * 1024 * 16
  // DefaultMaxSendMsgSize maximum message that client can send (16 MB).
  defaultMaxSendMsgSize = 1024 * 1024 * 16
)

type grpcClient struct {
  ctx      context.Context
  runtime  runtime.Runtime
  resolver resolver.Resolver

  opts Options
  pool map[string]*pool
}

func NewClient(ctx context.Context, runtime runtime.Runtime) client.GRPCClient {

  client := &grpcClient{
    ctx:     ctx,
    runtime: runtime,
    opts:    defaultOptions(),
    pool:    make(map[string]*pool, 0),
  }

  client.pool[defaultPoolName] = newPool()
  runtime.Config().Parse(&client.opts, defaultPrefix)

  if client.opts.Resolver == "local" {
    client.resolver = local.NewResolver(runtime)
  }

  if client.opts.Resolver == "file" {
    client.resolver = file.NewResolver(runtime)
  }

  return client
}

func (c *grpcClient) Conn(service string) (grpc.ClientConnInterface, error) {
  var p *pool
  p, ok := c.pool[service]
  if !ok {
    p = newPool()
    c.pool[service] = p

  }

  routes, err := c.getResolver().Lookup(service)
  if err != nil && !strings.HasSuffix(err.Error(), "route not found") {
    return nil, status.Error(codes.Unavailable, err.Error())
  }

  addresses := routes.Addresses()
  if len(addresses) == 0 {
    addresses = []string{fmt.Sprintf(":%d", defaultPort)}
  }

  next, err := c.opts.Selector.Select(addresses)
  if err != nil {
    return nil, err
  }

  return p.getConn(c.ctx, next(), c.makeGrpcDialOptions()...)

}

func (c *grpcClient) GetResolver() resolver.Resolver {
  return c.resolver
}

func (c *grpcClient) SetResolver(resolver resolver.Resolver) {
  c.resolver = resolver
}

func (c *grpcClient) Call(ctx context.Context, service, method string, body, resp interface{}, opts ...client.GRPCCallOption) error {
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
  req := client.NewGRPCRequest(service, method, body, headers)

  routes, err := c.getResolver().Lookup(req.Service())
  if err != nil && !strings.HasSuffix(err.Error(), "route not found") {
    return status.Error(codes.Unavailable, err.Error())
  }

  addresses := routes.Addresses()
  if len(addresses) == 0 {
    addresses = []string{fmt.Sprintf(":%d", defaultPort)}
  }

  next, err := c.opts.Selector.Select(addresses)
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
          return err
        }
        time.Sleep(d)
        continue
      }
      b.Reset()
      return nil
    }
  }

}

func (c *grpcClient) Stream(ctx context.Context, service, method string, body interface{}, opts ...client.GRPCCallOption) (grpc.ClientStream, error) {

  callOpts := c.opts.CallOptions
  for _, opt := range opts {
    opt(&callOpts)
  }

  streamFunc := c.stream

  headers := c.makeHeaders(ctx, service, callOpts)
  req := client.NewGRPCRequest(service, method, body, headers)

  routes, err := c.getResolver().Lookup(req.Service())
  if err != nil && !strings.HasSuffix(err.Error(), "route not found") {
    return nil, status.Error(codes.Unavailable, err.Error())
  }

  addresses := routes.Addresses()
  if len(addresses) == 0 {
    addresses = []string{fmt.Sprintf(":%d", defaultPort)}
  }

  next, err := c.opts.Selector.Select(addresses)
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
          return nil, err
        }
        time.Sleep(d)
        continue
      }
      b.Reset()
      return s, nil
    }
  }
}

func (c *grpcClient) invoke(ctx context.Context, addr string, req *client.GRPCRequest, rsp interface{}, opts client.GRPCCallOptions) error {

  md := grpc_md.New(req.Headers())
  ctx = grpc_md.NewOutgoingContext(ctx, md)

  var gErr error
  conn, err := c.pool[defaultPoolName].getConn(ctx, addr, c.makeGrpcDialOptions()...)
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
    gErr = status.Error(codes.Canceled, ctx.Err().Error())
  }

  return gErr
}

func (c *grpcClient) stream(ctx context.Context, addr string, req *client.GRPCRequest, opts client.GRPCCallOptions) (grpc.ClientStream, error) {

  md := grpc_md.New(req.Headers())
  ctx = grpc_md.NewOutgoingContext(ctx, md)
  ctx, cancel := context.WithCancel(ctx)

  cc, err := c.pool[defaultPoolName].getConn(ctx, addr, c.makeGrpcDialOptions()...)
  if err != nil {
    cancel()
    return nil, status.Error(codes.Internal, err.Error())
  }

  desc := &grpc.StreamDesc{
    StreamName:    req.Method(),
    ClientStreams: true,
    ServerStreams: true,
  }

  st, err := cc.NewStream(ctx, desc, req.Method(), c.makeGrpcCallOptions(opts)...)
  if err != nil {
    cancel()
    c.pool[defaultPoolName].release(addr, cc, err)
    return nil, status.Error(codes.Canceled, err.Error())
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
      c.pool[defaultPoolName].release(addr, cc, err)
    },
  }

  // wait for error response
  ch := make(chan error, 1)

  go func() {
    // send the first message
    ch <- st.SendMsg(req.Body())
  }()

  var grr error

  select {
  case err := <-ch:
    grr = err
  case <-ctx.Done():
    grr = ctx.Err()
  }

  if grr != nil {
    _ = st.CloseSend()
    return nil, grr
  }

  return s, nil
}

func (c *grpcClient) makeGrpcCallOptions(opts client.GRPCCallOptions) []grpc.CallOption {
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

func (c *grpcClient) makeGrpcDialOptions() []grpc.DialOption {
  grpcDialOptions := make([]grpc.DialOption, 0)

  // TODO: implement auths
  grpcDialOptions = append(grpcDialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))

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

func (c *grpcClient) makeHeaders(ctx context.Context, service string, opts client.GRPCCallOptions) map[string]string {
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

func (c *grpcClient) getResolver() resolver.Resolver {
  if c.resolver == nil {
    c.resolver = local.NewResolver(c.runtime)
  }

  return c.resolver
}
