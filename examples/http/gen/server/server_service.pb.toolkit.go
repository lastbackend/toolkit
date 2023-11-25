// Code generated by protoc-gen-toolkit. DO NOT EDIT.
// source: github.com/lastbackend/toolkit/examples/http/apis/server.proto

package serverpb

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	toolkit "github.com/lastbackend/toolkit"
	client "github.com/lastbackend/toolkit/pkg/client"
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
	_ http.Handler
	_ errors.Err
	_ io.Reader
	_ json.Marshaler
	_ tk_ws.Client
	_ tk_http.Handler
	_ client.GRPCClient
)

// Definitions

// Service Http define
type serviceHttp struct {
	runtime runtime.Runtime
}

func NewHttpService(name string, opts ...runtime.Option) (_ toolkit.Service, err error) {
	app := new(serviceHttp)

	app.runtime, err = controller.NewRuntime(context.Background(), name, opts...)
	if err != nil {
		return nil, err
	}

	// loop over plugins and initialize plugin instance

	// loop over plugins and register plugin in toolkit

	// create new Http HTTP server
	app.runtime.Server().HTTPNew(name, nil)

	return app.runtime.Service(), nil
}
