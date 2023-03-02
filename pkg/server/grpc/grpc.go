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
	"crypto/tls"
	"fmt"
	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
	"github.com/lastbackend/toolkit/pkg/server"
	"net"
	"net/http"
	"sync"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"golang.org/x/net/netutil"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	serviceName = "grpc"
)

type grpcServer struct {
	runtime runtime.Runtime

	sync.RWMutex

	prefix  string
	address string

	opts Config

	descriptor grpc.ServiceDesc

	// fn for init user-defined service
	// fn for server registration
	provide interface{}
	service interface{}

	grpc      *grpc.Server
	isRunning bool

	exit chan chan error
}

// NewServer - init and return new grpc server instance
func NewServer(runtime runtime.Runtime, opts *server.GRPCServerOptions) server.GRPCServer { //nolint

	var name = serviceName
	if opts != nil {
		name = opts.Name
	}

	srv := &grpcServer{
		runtime: runtime,
		prefix:  name,
		opts:    defaultOptions(),
		exit:    make(chan chan error),
	}

	servopts := srv.parseOptions()
	srv.grpc = grpc.NewServer(servopts...)

	if err := runtime.Config().Parse(&srv.opts, srv.prefix); err != nil {
		return nil
	}

	return srv
}

// SetDescriptor - set generated grpc server descriptor
func (g *grpcServer) SetDescriptor(descriptor grpc.ServiceDesc) {
	g.descriptor = descriptor
}

// SetService - set user-defined handlers
func (g *grpcServer) SetService(service interface{}) {
	g.service = service
	return
}

// GetService - set user-defined handlers
func (g *grpcServer) GetService() interface{} {
	return g.service
}

// RegisterService - set user-defined handlers
func (g *grpcServer) RegisterService(service interface{}) {
	g.grpc.RegisterService(&g.descriptor, service)
	return
}

// SetConstructor - set fx handlers definition
func (g *grpcServer) SetConstructor(fn interface{}) {
	g.provide = fn
	return
}

// GetConstructor - set fx handlers definition
func (g *grpcServer) GetConstructor() interface{} {
	return g.provide
}

// parseOptions - get options from config and convert them to grpc server options
func (g *grpcServer) parseOptions() []grpc.ServerOption {

	gopts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(g.opts.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(g.opts.MaxSendMsgSize),
		grpc.UnknownServiceHandler(g.defaultHandler),
	}

	if g.opts.TLSConfig != nil {
		gopts = append(gopts, grpc.Creds(credentials.NewTLS(g.opts.TLSConfig)))
	}

	if g.opts.GrpcOptions != nil && len(g.opts.GrpcOptions) > 0 {
		gopts = append(gopts, g.opts.GrpcOptions...)
	}

	return gopts
}

func (g *grpcServer) Start(_ context.Context) error {

	g.RLock()
	if g.isRunning {
		g.RUnlock()
		return nil
	}
	g.RUnlock()

	var (
		listener net.Listener
		err      error
	)

	address := fmt.Sprintf("%s:%d", g.opts.Host, g.opts.Port)
	if transportConfig := g.opts.TLSConfig; transportConfig != nil {
		listener, err = tls.Listen("tcp", address, transportConfig)
	} else {
		listener, err = net.Listen("tcp", address)
	}
	if err != nil {
		return err
	}

	if g.opts.MaxConnSize > 0 {
		listener = netutil.LimitListener(listener, g.opts.MaxConnSize)
	}

	g.runtime.Log().Infof("server [grpc] Listening on %s", listener.Addr().String())

	g.Lock()
	g.address = listener.Addr().String()
	g.Unlock()

	if g.opts.GRPCWebPort > 0 {

		grpcWebOptions := make([]grpcweb.Option, 0)
		if g.opts.GrpcWebOptions != nil {
			grpcWebOptions = g.opts.GrpcWebOptions
		}

		wrappedGrpc := grpcweb.WrapServer(g.grpc, grpcWebOptions...)
		webGRPCServer := &http.Server{
			Addr:      fmt.Sprintf("%s:%d", g.opts.GRPCWebHost, g.opts.GRPCWebPort),
			TLSConfig: g.opts.TLSConfig,
			Handler:   http.Handler(wrappedGrpc),
		}

		go func() {
			err = webGRPCServer.ListenAndServe()
			if err != nil {
				logger.Errorf("server [grpc] start error: %v", err)
			}
		}()

		g.runtime.Log().Infof("server [gRPC-Web] Listening on %s:%d", g.opts.GRPCWebHost, g.opts.GRPCWebPort)

	}

	go func() {
		if err := g.grpc.Serve(listener); err != nil {
			g.runtime.Log().Errorf("server [grpc] start error: %v", err)
		}
	}()

	go func() {
		ch := <-g.exit
		ch <- listener.Close()
	}()

	g.Lock()
	g.isRunning = true
	g.Unlock()

	return nil
}

func (g *grpcServer) Stop() error {

	g.grpc.Stop()

	g.RLock()
	if !g.isRunning {
		g.RUnlock()
		return nil
	}
	g.RUnlock()

	ch := make(chan error)
	g.exit <- ch

	var err error
	if err = <-ch; true {
		g.Lock()
		g.isRunning = false
		g.Unlock()
	}

	return err
}

// TODO: need implement defaultHandler method
func (g *grpcServer) defaultHandler(_ interface{}, _ grpc.ServerStream) error {
	return nil
}
