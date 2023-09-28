package client

import (
  "context"
  "github.com/lastbackend/toolkit/pkg/client/grpc/resolver"
  "google.golang.org/grpc"
  "time"
)

type GRPCStream grpc.ClientStream
type GRPCClient interface {
  GetResolver() resolver.Resolver
  SetResolver(resolver resolver.Resolver)
  Conn(pool string) (grpc.ClientConnInterface, error)
  Call(ctx context.Context, service, method string, req, rsp interface{}, opts ...GRPCCallOption) error
  Stream(ctx context.Context, service, method string, body interface{}, opts ...GRPCCallOption) (grpc.ClientStream, error)
}

type GRPCCallOption func(*GRPCCallOptions)

type GRPCCallOptions struct {
  Backoff               GRPCBackoffFunc
  Retries               time.Duration
  RequestTimeout        time.Duration
  Context               context.Context
  Headers               map[string]string
  MaxCallSendMsgSize    int
  MaxCallRecvMsgSize    int
  MaxRetryRPCBufferSize int
  CallContentSubtype    string
}

func GRPCOptionHeaders(h map[string]string) GRPCCallOption {
  return func(o *GRPCCallOptions) {
    o.Headers = h
  }
}

func GRPCOptionMaxCallSendMsgSize(bytes int) GRPCCallOption {
  return func(o *GRPCCallOptions) {
    o.MaxCallSendMsgSize = bytes
  }
}

func GRPCOptionMaxCallRecvMsgSize(bytes int) GRPCCallOption {
  return func(o *GRPCCallOptions) {
    o.MaxCallRecvMsgSize = bytes
  }
}

func GRPCOptionRequestTimeout(timeout time.Duration) GRPCCallOption {
  return func(o *GRPCCallOptions) {
    o.RequestTimeout = timeout
  }
}

type GRPCBackoffFunc func(ctx context.Context, req *GRPCRequest, attempts int) (time.Duration, error)
type GRPCRetryFunc func(ctx context.Context, req *GRPCRequest, retryCount int, err error) (bool, error)

type GRPCRequest struct {
  service string
  method  string
  headers map[string]string
  body    interface{}
}

func NewGRPCRequest(method, service string, body interface{}, headers map[string]string) *GRPCRequest {
  r := new(GRPCRequest)
  r.service = method
  r.method = service
  r.body = body
  if headers == nil {
    headers = make(map[string]string, 0)
  }
  r.headers = headers
  return r
}

func (r *GRPCRequest) Service() string {
  return r.service
}

func (r *GRPCRequest) Method() string {
  return r.method
}

func (r *GRPCRequest) Body() interface{} {
  return r.body
}

func (r *GRPCRequest) Headers() map[string]string {
  return r.headers
}

type HTTPClient interface {
  Get() error
  Post() error
}
