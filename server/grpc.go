/*
Copyright [2014] - [2022] The Last.Backend authors.

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

package server

import (
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/logger"
	"golang.org/x/net/netutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
)

const (
	serviceName = "grpc"
)

type ServerOptions struct { // nolint
	Name string
}

type grpcServer struct {
	sync.RWMutex

	prefix string

	opts Options

	srv       *grpc.Server
	isRunning bool

	exit chan chan error
}

func NewServer(app toolkit.Service, opts *ServerOptions) *grpcServer { //nolint
	name := opts.Name
	if name == "" {
		name = serviceName
	}
	s := newServer(name)
	s.addFlags(app)
	return s
}

func (g *grpcServer) Register(sd *grpc.ServiceDesc, ss interface{}) error {
	g.srv.RegisterService(sd, ss)
	return nil
}

func (g *grpcServer) Start() error {

	if g.opts.IsDisable {
		return nil
	}

	g.init()

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

	if transportConfig := g.opts.TLSConfig; transportConfig != nil {
		listener, err = tls.Listen("tcp", g.opts.Address, transportConfig)
	} else {
		listener, err = net.Listen("tcp", g.opts.Address)
	}
	if err != nil {
		return err
	}

	if g.opts.MaxConnSize > 0 {
		listener = netutil.LimitListener(listener, g.opts.MaxConnSize)
	}

	if logger.V(logger.InfoLevel, logger.DefaultLogger) {
		logger.Infof("server [grpc] Listening on %s", listener.Addr().String())
	}

	g.Lock()
	g.opts.Address = listener.Addr().String()
	g.Unlock()

	if len(g.opts.GRPCWebAddr) > 0 {
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

			go func() {
				err = webGRPCServer.ListenAndServe()
				if err != nil {
					logger.Errorf("server [grpc] start error: %v", err)
				}
			}()

			if logger.V(logger.InfoLevel, logger.DefaultLogger) {
				logger.Infof("server [gRPC-Web] Listening on %s", gRPCWebAddr)
			}
		}

	}

	go func() {
		if err := g.srv.Serve(listener); err != nil {
			if logger.V(logger.InfoLevel, logger.DefaultLogger) {
				logger.Errorf("server [grpc] start error: %v", err)
			}
		}
	}()

	go func() {
		ch := <-g.exit
		ch <- listener.Close()
		g.srv.Stop()
	}()

	g.Lock()
	g.isRunning = true
	g.Unlock()

	return nil
}

func (g *grpcServer) Stop() error {
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

func (g *grpcServer) addFlags(app toolkit.Service) {

	app.CLI().AddStringFlag(g.withPrefix("address"), &g.opts.Address).
		Env(g.withEnvPrefix("ADDRESS")).
		Usage("Server address for listening").
		Default(defaultAddress)

	app.CLI().AddStringFlag(g.withPrefix("name"), &g.opts.Name).
		Env(g.withEnvPrefix("NAME")).
		Usage("Server name").
		Default(defaultName)

	app.CLI().AddIntFlag(g.withPrefix("max-recv-msg-size"), &g.opts.MaxRecvMsgSize).
		Env(g.withEnvPrefix("MAX_RECV_MSG_SIZE")).
		Usage("Sets the max message size in bytes the server can receive (default 16 MB)").
		Default(defaultMaxRecvMsgSize)

	app.CLI().AddIntFlag(g.withPrefix("max-send-msg-size"), &g.opts.MaxSendMsgSize).
		Env(g.withEnvPrefix("MAX_RECV_MSG_SIZE")).
		Usage("Sets the max message size in bytes the server can send (default 16 MB)").
		Default(defaultMaxSendMsgSize)

	app.CLI().AddIntFlag(g.withPrefix("max-conn-size"), &g.opts.MaxConnSize).
		Env(g.withEnvPrefix("MAX_CONN_SIZE")).
		Usage("Sets the max simultaneous connections for server (default unlimited)")

	app.CLI().AddBoolFlag(g.withPrefix("disable"), &g.opts.IsDisable).
		Env(g.withEnvPrefix("DISABLE")).
		Usage("Sets the disable server")
}

func newServer(prefix string) *grpcServer {
	srv := &grpcServer{
		prefix: prefix,
		opts:   defaultOptions(),
		exit:   make(chan chan error),
	}

	srv.init()

	return srv
}

// TODO: need implement defaultHandler method
func (g *grpcServer) defaultHandler(srv interface{}, stream grpc.ServerStream) error {
	return nil
}

func (g *grpcServer) init() {
	g.Lock()
	defer g.Unlock()

	if g.srv != nil {
		return
	}

	maxRecvMsgSize := defaultMaxRecvMsgSize
	if g.opts.MaxRecvMsgSize != 0 {
		maxRecvMsgSize = g.opts.MaxRecvMsgSize
	}

	maxSendMsgSize := defaultMaxSendMsgSize
	if g.opts.MaxSendMsgSize != 0 {
		maxSendMsgSize = g.opts.MaxSendMsgSize
	}

	gopts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(maxRecvMsgSize),
		grpc.MaxSendMsgSize(maxSendMsgSize),
		grpc.UnknownServiceHandler(g.defaultHandler),
	}

	if g.opts.TLSConfig != nil {
		gopts = append(gopts, grpc.Creds(credentials.NewTLS(g.opts.TLSConfig)))
	}

	if g.opts.GrpcOptions != nil && len(g.opts.GrpcOptions) > 0 {
		gopts = append(gopts, g.opts.GrpcOptions...)
	}

	g.srv = grpc.NewServer(gopts...)
}

func (g *grpcServer) withPrefix(name string) string {
	return fmt.Sprintf("%s-%s", g.prefix, name)
}

func (g *grpcServer) withEnvPrefix(name string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s", g.prefix, name))
}
