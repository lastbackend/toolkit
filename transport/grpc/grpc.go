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
	"crypto/tls"
	"fmt"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/lastbackend/engine/cmd"
	"golang.org/x/net/netutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
	"net/http"
	"strings"
	"sync"
)

const (
	ServiceName = "grpc"
)

type grpcServer struct {
	sync.RWMutex

	prefix string

	opts Options

	srv        *grpc.Server
	started    bool
	registered bool
	handlers   map[string]Handler

	exit chan chan error
}

func NewTransport(prefix string) *grpcServer {
	return newServer(prefix)
}

func (g *grpcServer) Name() string {
	return ServiceName
}

func (g *grpcServer) NewHandler(h interface{}, opts ...HandlerOption) *Handler {
	return newHandler(h, opts...)
}

func (g *grpcServer) Register(sd *grpc.ServiceDesc, ss interface{}) error {
	g.srv.RegisterService(sd, ss)
	return nil
}

func (g *grpcServer) Start() error {
	g.configure()

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

	if g.opts.MaxConnSize > 0 {
		listener = netutil.LimitListener(listener, g.opts.MaxConnSize)
	}

	fmt.Println(fmt.Sprintf("server [grpc] Listening on %s", listener.Addr().String()))

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
		g.started = false
		g.Unlock()
	}

	return err
}

func (g *grpcServer) Flags() []cmd.Flag {
	return []cmd.Flag{
		&cmd.StringFlag{
			Name:        g.withPrefix("address"),
			EnvVars:     []string{g.withEnvPrefix("ADDRESS")},
			Usage:       "Server address for listening",
			Required:    false,
			Value:       defaultAddress,
			Destination: &g.opts.Address,
		},
		&cmd.StringFlag{
			Name:        g.withPrefix("name"),
			EnvVars:     []string{g.withEnvPrefix("NAME")},
			Usage:       "Server name",
			Required:    false,
			Value:       defaultName,
			Destination: &g.opts.Name,
		},
		&cmd.IntFlag{
			Name:        g.withPrefix("max-recv-msg-size"),
			EnvVars:     []string{g.withEnvPrefix("MAC-RECV-MSG-SIZE")},
			Usage:       "Sets the max message size in bytes the server can receive (default 16 MB)",
			Required:    false,
			Destination: &g.opts.MaxRecvMsgSize,
		},
		&cmd.IntFlag{
			Name:        g.withPrefix("max-send-msg-size"),
			EnvVars:     []string{g.withEnvPrefix("MAC-RECV-MSG-SIZE")},
			Usage:       "Sets the max message size in bytes the server can send (default 16 MB)",
			Required:    false,
			Destination: &g.opts.MaxSendMsgSize,
		},
		&cmd.IntFlag{
			Name:        g.withPrefix("max-conn-size"),
			EnvVars:     []string{g.withEnvPrefix("MAX-CONN-SIZE")},
			Usage:       "Sets the max simultaneous connections for server (default unlimited)",
			Required:    false,
			Destination: &g.opts.MaxConnSize,
		},
	}
}

func (g *grpcServer) Commands() []cmd.Command {
	return make([]cmd.Command, 0)
}

func newServer(prefix string) *grpcServer {
	srv := &grpcServer{
		prefix:   prefix,
		opts:     defaultOptions(),
		exit:     make(chan chan error),
	}

	srv.configure()

	return srv
}

// TODO: need implement defaultHandler method
func (g *grpcServer) defaultHandler(srv interface{}, stream grpc.ServerStream) error {
	return nil
}

func (g *grpcServer) configure() {
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
