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

package server

import (
	"crypto/tls"
	"github.com/google/uuid"
	"github.com/lastbackend/engine/codec"
	"github.com/lastbackend/engine/logger"
	"google.golang.org/grpc"

	"context"
	"net"
	"sync"
)

type Server interface {
	Init(...Option) error
	Options() Options
	Register(h Handler) error
	NewHandler(h interface{}, opts ...HandlerOption) Handler
	Handle(h Handler) error
	Start() error
	Stop() error
}

type server struct {
	sync.RWMutex

	srv *grpc.Server

	router *router

	handlers map[string]Handler

	started bool

	exit chan chan error

	wg *sync.WaitGroup

	opts Options
}

type Option func(*Options)

var (
	DefaultAddress          = ":0"
	DefaultName             = "lb.engine.server"
	DefaultVersion          = "latest"
	DefaultId               = uuid.New().String()
	DefaultServer           = newServer()
)

type Request interface {
	Service() string
	Method() string
	Endpoint() string
	ContentType() string
	Header() map[string]string
	Body() interface{}
	Read() ([]byte, error)
	Codec() codec.Reader
	Stream() bool
}

type Response interface {
	Codec() codec.Writer
	WriteHeader(map[string]string)
	Write([]byte) error
}

type Stream interface {
	Context() context.Context
	Request() Request
	Send(interface{}) error
	Recv(interface{}) error
	Error() error
	Close() error
}

type Handler interface {
	register(*server)
	Name() string
	Handler() interface{}
	Options() HandlerOptions
}

func newServer(opts ...Option) Server {
	options := newOptions(opts...)

	//router := newRpcRouter()
	//router.hdlrWrappers = options.HdlrWrappers
	//router.subWrappers = options.SubWrappers

	var gOpts []grpc.ServerOption
	{
		gOpts = []grpc.ServerOption{}
	}

	grpcServer := grpc.NewServer(gOpts...)

	return &server{
		opts: options,
		srv:  grpcServer,
		//router:      router,
		//handlers:    make(map[string]Handler),
		//subscribers: make(map[Subscriber][]broker.Subscriber),
		//exit:        make(chan chan error),
		//wg:          wait(options.Context),
	}
}

func (s *server) Options() Options {
	s.RLock()
	opts := s.opts
	s.RUnlock()
	return opts
}

func (s *server) Init(opts ...Option) error {
	s.Lock()
	defer s.Unlock()

	for _, opt := range opts {
		opt(&s.opts)
	}

	// update router if its the default
	//if s.opts.Router == nil {
	//	r := newRpcRouter()
	//	r.hdlrWrappers = s.opts.HdlrWrappers
	//	r.serviceMap = s.router.serviceMap
	//	r.subWrappers = s.opts.SubWrappers
	//	s.router = r
	//}
	//
	//s.rsvc = nil

	return nil
}

func (s *server) Start() error {
	s.RLock()
	if s.started {
		s.RUnlock()
		return nil
	}
	s.RUnlock()

	config := s.Options()

	var ts net.Listener

	if l := s.getListener(); l != nil {
		ts = l
	} else {
		var err error

		// check the tls config for secure connect
		if tc := config.TLSConfig; tc != nil {
			ts, err = tls.Listen("tcp", config.Address, tc)
		} else {
			ts, err = net.Listen("tcp", config.Address)
		}
		if err != nil {
			return err
		}
	}

	//if g.opts.Context != nil {
	//	if c, ok := g.opts.Context.Value(maxConnKey{}).(int); ok && c > 0 {
	//		ts = netutil.LimitListener(ts, c)
	//	}
	//}

	logger.DefaultLogger.Infof("Server [grpc] Listening on %s", ts.Addr().String())

	s.Lock()
	s.opts.Address = ts.Addr().String()
	s.Unlock()

	// only connect if we're subscribed
	//if len(g.subscribers) > 0 {
	//	// connect to the broker
	//	if err := config.Broker.Connect(); err != nil {
	//		logger.DefaultLogger.Errorf("Broker [%s] connect error: %v", config.Broker.String(), err)
	//		return err
	//	}
	//
	//	logger.DefaultLogger.Infof("Broker [%s] Connected to %s", config.Broker.String(), config.Broker.Address())
	//}

	// announce self to the world
	//if err := g.Register(); err != nil {
	//	logger.DefaultLogger.Errorf("Server register error: %v", err)
	//}

	go func() {
		if err := s.srv.Serve(ts); err != nil {
			logger.DefaultLogger.Errorf("gRPC Server start error: %v", err)
		}
	}()

	//go func() {
	//	t := new(time.Ticker)
	//
	//	// only process if it exists
	//	if g.opts.RegisterInterval > time.Duration(0) {
	//		// new ticker
	//		t = time.NewTicker(g.opts.RegisterInterval)
	//	}
	//
	//	// return error chan
	//	var ch chan error
	//
	//Loop:
	//	for {
	//		select {
	//		// register self on interval
	//		case <-t.C:
	//			if err := g.Register(); err != nil {
	//				logger.DefaultLogger.Error("Server register error: ", err)
	//			}
	//		// wait for exit
	//		case ch = <-g.exit:
	//			break Loop
	//		}
	//	}
	//
	//	// deregister self
	//	if err := g.Deregister(); err != nil {
	//		logger.DefaultLogger.Error("Server deregister error: ", err)
	//	}
	//
	//	// wait for waitgroup
	//	if g.wg != nil {
	//		g.wg.Wait()
	//	}
	//
	//	// stop the grpc server
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
	//	// close transport
	//	ch <- nil
	//
	//	//logger.DefaultLogger.Infof("Broker [%s] Disconnected from %s", config.Broker.String(), config.Broker.Address())
	//
	//	// disconnect broker
	//	if err := config.Broker.Disconnect(); err != nil {
	//		logger.DefaultLogger.Errorf("Broker [%s] disconnect error: %v", config.Broker.String(), err)
	//	}
	//}()

	// mark the server as started
	s.Lock()
	s.started = true
	s.Unlock()

	return nil
}

func (s *server) Stop() error {
	s.RLock()
	if !s.started {
		s.RUnlock()
		return nil
	}
	s.RUnlock()

	ch := make(chan error)
	s.exit <- ch

	var err error
	select {
	case err = <-ch:
		s.Lock()
		s.started = false
		s.Unlock()
	}

	return err
}

func (s *server) Register(h Handler) error {
	s.Lock()
	defer s.Unlock()

	if err := s.router.Handle(h); err != nil {
		return err
	}

	s.handlers[h.Name()] = h
	h.register(s)

	return nil
}

func (s *server) NewHandler(h interface{}, opts ...HandlerOption) Handler {
	return s.router.NewHandler(h, opts...)
}

func (s *server) Handle(h Handler) error {
	s.Lock()
	defer s.Unlock()

	if err := s.router.Handle(h); err != nil {
		return err
	}

	s.handlers[h.Name()] = h

	return nil
}

type netListener struct{}

func (s *server) getListener() net.Listener {
	if s.opts.Context == nil {
		return nil
	}

	if l, ok := s.opts.Context.Value(netListener{}).(net.Listener); ok && l != nil {
		return l
	}

	return nil
}
