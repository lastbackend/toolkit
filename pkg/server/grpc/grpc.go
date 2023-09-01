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
	"github.com/lastbackend/toolkit/pkg/server"
	"net"
	"net/http"
	"regexp"
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
	handler interface{}

	interceptors *Interceptors

	grpc    *grpc.Server
	options *server.GRPCServerOptions

	isRunning bool

	wait *sync.WaitGroup
}

// NewServer - init and return new grpc server instance
func NewServer(runtime runtime.Runtime, name string, options *server.GRPCServerOptions) server.GRPCServer { //nolint

	if name == "" {
		name = serviceName
	}

	name = regexp.MustCompile(`[^_a-zA-Z0-9 ]+`).ReplaceAllString(name, "_")

	srv := &grpcServer{
		runtime:      runtime,
		prefix:       name,
		opts:         defaultOptions(),
		wait:         &sync.WaitGroup{},
		options:      options,
		interceptors: newInterceptors(runtime.Log()),
	}

	if err := runtime.Config().Parse(&srv.opts, srv.prefix); err != nil {
		return nil
	}

	return srv
}

func (g *grpcServer) Info() server.ServerInfo {
	return server.ServerInfo{
		Kind:      server.ServerKindGRPCServer,
		Host:      g.opts.Host,
		Port:      g.opts.Port,
		TLSConfig: g.opts.TLSConfig,
	}
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
	g.handler = service
	return
}

// SetConstructor - set fx handlers definition
func (g *grpcServer) SetConstructor(fn interface{}) {
	g.provide = fn
	return
}

func (s *grpcServer) GetInterceptors() []interface{} {
	return s.interceptors.constructors
}

// SetInterceptor - set fx handlers definition
func (g *grpcServer) SetInterceptor(interceptor any) {
	g.interceptors.AddConstructor(interceptor)
}

// GetConstructor - set fx handlers definition
func (g *grpcServer) GetConstructor() interface{} {
	return g.provide
}

func (g *grpcServer) GetInterceptorsConstructor() interface{} {
	return g.constructor
}

func (g *grpcServer) constructor(interceptors ...server.GRPCInterceptor) {
	for _, interceptor := range interceptors {
		g.interceptors.Add(interceptor)
	}
}

// parseOptions - get options from config and convert them to grpc server options
func (g *grpcServer) parseOptions(options *server.GRPCServerOptions) []grpc.ServerOption {

	if options != nil {
		if options.Host != "" {
			g.opts.Host = options.Host
		}

		if options.Port > 0 {
			g.opts.Port = options.Port
		}

		if options.TLSConfig != nil {
			g.opts.TLSConfig = options.TLSConfig
		}
	}

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

	var interceptors = make([]grpc.UnaryServerInterceptor, 0)
	for _, i := range g.interceptors.items {
		interceptors = append(interceptors, i.Interceptor)
	}

	gopts = append(gopts, grpc.ChainUnaryInterceptor(interceptors...))

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

	g.runtime.Log().V(5).Infof("server [grpc] Listening on %s", listener.Addr().String())

	g.Lock()
	g.address = listener.Addr().String()
	g.Unlock()

	g.grpc = grpc.NewServer(g.parseOptions(g.options)...)
	g.grpc.RegisterService(&g.descriptor, g.handler)

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
			g.wait.Add(1)
			g.runtime.Log().V(5).Infof("server [gRPC-Web] [%s:%d] started", g.opts.GRPCWebHost, g.opts.GRPCWebPort)
			if err = webGRPCServer.ListenAndServe(); err != nil {
				g.runtime.Log().Errorf("server [grpc] [%s:%d]  start error: %v", g.opts.GRPCWebHost, g.opts.GRPCWebPort, err)
			}
			g.runtime.Log().V(5).Infof("server [gRPC-Web] [%s:%d] stopped", g.opts.GRPCWebHost, g.opts.GRPCWebPort)
			g.wait.Done()
		}()

	}

	go func() {
		g.wait.Add(1)
		g.runtime.Log().V(5).Infof("server [grpc] [%s:%d] started", g.opts.Host, g.opts.Port)
		if err := g.grpc.Serve(listener); err != nil {
			g.runtime.Log().Errorf("server [grpc] start error: %v", err)
		}
		g.wait.Done()
		g.runtime.Log().V(5).Infof("server [grpc] [%s:%d] stopped", g.opts.Host, g.opts.Port)
	}()

	g.Lock()
	g.isRunning = true
	g.Unlock()

	return nil
}

func (g *grpcServer) Stop() error {

	g.runtime.Log().V(5).Infof("server [grpc] [%s:%d] stop call start", g.opts.Host, g.opts.Port)
	g.grpc.GracefulStop()
	g.wait.Wait()
	g.runtime.Log().V(5).Infof("server [grpc] [%s:%d] stop call end", g.opts.Host, g.opts.Port)

	return nil
}

// TODO: need implement defaultHandler method
func (g *grpcServer) defaultHandler(_ interface{}, _ grpc.ServerStream) error {
	return nil
}
