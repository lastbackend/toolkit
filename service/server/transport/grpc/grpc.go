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
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/lastbackend/engine/service/server"
	"golang.org/x/net/netutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sync"
)

const (
	ServiceName = "grpc"
)

var (
	// DefaultMaxRecvMsgSize maximum message that client can receive (16 MB).
	DefaultMaxRecvMsgSize = 1024 * 1024 * 16
	// DefaultMaxSendMsgSize maximum message that client can send (16 MB).
	DefaultMaxSendMsgSize = 1024 * 1024 * 16
	// DefaultMaxMsgSize define maximum message size that server can send
	// or receive.  Default value is 4MB.
	DefaultMaxMsgSize = 1024 * 1024 * 4
)

type grpcServer struct {
	sync.RWMutex

	opts Options

	srv        *grpc.Server
	started    bool
	registered bool
	handlers   map[string]server.Handler
	//rpc        *rServer
	//rsvc       *registry.Service

	exit chan chan error
}

func NewServer(opts Options) server.Server {
	return newServer(opts)
}

func (g *grpcServer) Name() string {
	return ServiceName
}

func (g *grpcServer) Init(opts Options) error {
	g.opts = patchOptions(g.opts, opts)
	g.configure()
	return nil
}

func (g *grpcServer) NewHandler(h interface{}, opts ...server.HandlerOption) server.Handler {
	return nil
}

func (g *grpcServer) Handle(h server.Handler) error {
	return nil
}

func (g *grpcServer) Start() error {

	g.RLock()
	if g.started {
		g.RUnlock()
		return nil
	}
	g.RUnlock()

	var (
		listener net.Listener
		err      error
	)

	if transportConfig := g.opts.TLSConfig; transportConfig != nil {
		listener, err = tls.Listen("tcp", g.opts.Address, transportConfig)
	} else {
		listener, err = net.Listen("tcp", g.opts.Address)
	}
	if err != nil {
		return err
	}

	if g.opts.MaxConnKey > 0 {
		listener = netutil.LimitListener(listener, g.opts.MaxConnKey)
	}

	fmt.Println(fmt.Sprintf("server [grpc] Listening on %s", listener.Addr().String()))

	g.Lock()
	g.opts.Address = listener.Addr().String()
	g.Unlock()

	//if err := g.register(); err != nil {
	//	fmt.Println(fmt.Sprintf("server register error: %v", err))
	//}

	if len(g.opts.GRPCWebAddr) > 0 {
		//gRPCWebAddr := ":8082"
		gRPCWebAddr := g.opts.GRPCWebAddr

		if len(g.opts.GRPCWebAddr) > 0 {
			grpcWebOptions := make([]grpcweb.Option, 0)
			if g.opts.GrpcWebOptions != nil {
				grpcWebOptions = g.opts.GrpcWebOptions
			}
			wrappedGrpc := grpcweb.WrapServer(g.srv, grpcWebOptions...)
			webGRPCServer := &http.Server{
				Addr:      gRPCWebAddr,
				TLSConfig: g.opts.TLSConfig,
				Handler:   http.Handler(wrappedGrpc),
			}

			go webGRPCServer.ListenAndServe()

			fmt.Println(fmt.Sprintf("server [gRPC-Web] Listening on %s", gRPCWebAddr))
		}

	}

	go func() {
		if err := g.srv.Serve(listener); err != nil {
			fmt.Println(fmt.Sprintf("gRPC server start error: %v", err))
		}
	}()

	//go func() {
	//	t := new(time.Ticker)
	//
	//	// only process if exists
	//	if g.opts.RegisterInterval > time.Duration(0) {
	//		t = time.NewTicker(g.opts.RegisterInterval)
	//	}
	//
	//	var chanCLoseTransport chan error
	//
	//Loop:
	//	for {
	//		select {
	//		case <-t.C:
	//			if err := g.register(); err != nil {
	//				fmt.Println(fmt.Sprintf("server register error: %v", err))
	//			}
	//		case chanCLoseTransport = <-g.exit:
	//			break Loop
	//		}
	//	}
	//
	//	if err := g.deregister(); err != nil {
	//		fmt.Println(fmt.Sprintf("server deregister error: %v", err))
	//	}
	//
	//	exit := make(chan bool)
	//
	//	go func() {
	//		g.srv.GracefulStop()
	//		close(exit)
	//	}()
	//
	//	select {
	//	case <-exit:
	//	case <-time.After(time.Second):
	//		g.srv.Stop()
	//	}
	//
	//	chanCLoseTransport <- nil
	//}()

	g.Lock()
	g.started = true
	g.Unlock()

	return nil
}

func (g *grpcServer) Stop() error {
	g.RLock()
	if !g.started {
		g.RUnlock()
		return nil
	}
	g.RUnlock()

	ch := make(chan error)
	g.exit <- ch

	var err error
	select {
	case err = <-ch:
		g.Lock()
		//g.rsvc = nil
		g.started = false
		g.Unlock()
	}

	return err
}

func newServer(opts Options) server.Server {
	srv := &grpcServer{
		opts: patchOptions(defaultOptions(), opts),
		//rpc: &rServer{
		//	serviceMap: make(map[string]*service),
		//},
		handlers: make(map[string]server.Handler),
		exit:     make(chan chan error),
	}

	srv.configure()

	return srv
}

func (g *grpcServer) handler(srv interface{}, stream grpc.ServerStream) error {
	return nil
}

func (g *grpcServer) configure() {
	g.Lock()
	defer g.Unlock()

	if g.srv != nil {
		return
	}

	maxRecvMsgSize := DefaultMaxRecvMsgSize
	if g.opts.MaxRecvMsgSize != 0 {
		maxRecvMsgSize = g.opts.MaxRecvMsgSize
	}

	maxSendMsgSize := DefaultMaxSendMsgSize
	if g.opts.MaxSendMsgSize != 0 {
		maxSendMsgSize = g.opts.MaxSendMsgSize
	}

	gopts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(maxRecvMsgSize),
		grpc.MaxSendMsgSize(maxSendMsgSize),
		grpc.UnknownServiceHandler(g.handler),
	}

	if g.opts.TLSConfig != nil {
		gopts = append(gopts, grpc.Creds(credentials.NewTLS(g.opts.TLSConfig)))
	}

	if g.opts.GrpcOptions != nil && len(g.opts.GrpcOptions) > 0 {
		gopts = append(gopts, g.opts.GrpcOptions...)
	}

	//g.rsvc = nil
	g.srv = grpc.NewServer(gopts...)
}
