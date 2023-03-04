// Code generated by protoc-gen-toolkit. DO NOT EDIT.
// source: github.com/lastbackend/toolkit/examples/gateway/apis/server.proto

package serverpb

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	toolkit "github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/examples/helloworld/gen"
	grpc "github.com/lastbackend/toolkit/pkg/client/grpc"
	runtime "github.com/lastbackend/toolkit/pkg/runtime"
	controller "github.com/lastbackend/toolkit/pkg/runtime/controller"
	tk_http "github.com/lastbackend/toolkit/pkg/server/http"
	errors "github.com/lastbackend/toolkit/pkg/server/http/errors"
	tk_ws "github.com/lastbackend/toolkit/pkg/server/http/websockets"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the toolkit package it is being compiled against and
// suppress "imported and not used" errors
var (
	_ context.Context
	_ emptypb.Empty
	_ grpc.Client
	_ http.Handler
	_ errors.Err
	_ io.Reader
	_ json.Marshaler
	_ tk_ws.Client
	_ tk_http.Handler
)

// Definitions

// Service ProxyGateway define
type serviceProxyGateway struct {
	runtime runtime.Runtime
}

func NewProxyGatewayService(name string, opts ...runtime.Option) (_ toolkit.Service, err error) {
	app := new(serviceProxyGateway)

	app.runtime, err = controller.NewRuntime(context.Background(), name, opts...)
	if err != nil {
		return nil, err
	}

	// create new ProxyGateway HTTP server
	app.runtime.Server().HTTPNew(name)
	app.runtime.Server().HTTP().UseMiddleware("request_id")
	app.runtime.Server().HTTP().AddHandler(http.MethodPost, "/hello", app.handlerHTTPProxyGatewayHelloWorld,
		tk_http.WithMiddleware("example"))

	return app.runtime.Service(), nil
}

// Define HTTP handlers for Router HTTP server

func (s *serviceProxyGateway) handlerHTTPProxyGatewayHelloWorld(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var protoRequest servicepb.HelloRequest
	var protoResponse servicepb.HelloReply

	im, om := tk_http.GetMarshaler(s.runtime.Server().HTTP(), r)

	reader, err := tk_http.NewReader(r.Body)
	if err != nil {
		errors.HTTP.InternalServerError(w)
		return
	}

	if err := im.NewDecoder(reader).Decode(&protoRequest); err != nil && err != io.EOF {
		errors.HTTP.InternalServerError(w)
		return
	}

	headers, err := tk_http.PrepareHeaderFromRequest(r)
	if err != nil {
		errors.HTTP.InternalServerError(w)
		return
	}

	callOpts := make([]grpc.CallOption, 0)
	callOpts = append(callOpts, grpc.Headers(headers))

	if err := s.runtime.Client().GRPC().Call(ctx, "helloworld", "/helloworld.Greeter/SayHello", &protoRequest, &protoResponse, callOpts...); err != nil {
		errors.GrpcErrorHandlerFunc(w, err)
		return
	}

	buf, err := om.Marshal(protoResponse)
	if err != nil {
		errors.HTTP.InternalServerError(w)
		return
	}

	w.Header().Set("Content-Type", om.ContentType())

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(buf); err != nil {
		return
	}
}
