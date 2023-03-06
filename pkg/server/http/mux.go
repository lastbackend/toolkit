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

package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/server"
	"github.com/lastbackend/toolkit/pkg/server/http/errors"
	"github.com/lastbackend/toolkit/pkg/server/http/marshaler"
	"github.com/lastbackend/toolkit/pkg/server/http/websockets"
	"mime"
	"net"
	"net/http"
	"sync"
)

const (
	defaultPrefix = "http"
)

var (
	acceptHeader      = http.CanonicalHeaderKey("Accept")
	contentTypeHeader = http.CanonicalHeaderKey("Content-Type")
)

type httpServer struct {
	runtime runtime.Runtime

	sync.RWMutex

	opts Config

	prefix    string
	isRunning bool

	handlers     map[string]server.HTTPServerHandler
	marshalerMap map[string]marshaler.Marshaler

	middlewares *Middlewares

	corsHandlerFunc      http.HandlerFunc
	grpcErrorHandlerFunc func(http.ResponseWriter, error)

	wsManager *websockets.Manager

	service interface{}
	exit    chan chan error
}

func NewServer(name string, runtime runtime.Runtime, options *server.HTTPServerOptions) server.HTTPServer {

	s := &httpServer{
		runtime:      runtime,
		prefix:       defaultPrefix,
		marshalerMap: GetMarshalerMap(),
		exit:         make(chan chan error),

		corsHandlerFunc:      corsHandlerFunc,
		grpcErrorHandlerFunc: errors.GrpcErrorHandlerFunc,

		middlewares: newMiddlewares(runtime.Log()),
		wsManager:   websockets.NewManager(runtime.Log()),
		handlers:    make(map[string]server.HTTPServerHandler, 0),
	}

	if name != "" {
		s.prefix = name
	}

	if err := runtime.Config().Parse(&s.opts, s.prefix); err != nil {
		return nil
	}

	if options != nil {
		s.parseOptions(options)
	}

	return s
}

func (s *httpServer) parseOptions(options *server.HTTPServerOptions) {

	if options != nil {
		if options.Host != "" {
			s.opts.Host = options.Host
		}

		if options.Port > 0 {
			s.opts.Port = options.Port
		}

		if options.TLSConfig != nil {
			s.opts.TLSConfig = options.TLSConfig
		}
	}
}

func (s *httpServer) Info() server.ServerInfo {
	return server.ServerInfo{
		Kind:      server.ServerKindHTTPServer,
		Host:      s.opts.Host,
		Port:      s.opts.Port,
		TLSConfig: s.opts.TLSConfig,
	}
}

func (s *httpServer) Start(_ context.Context) error {

	s.RLock()
	if s.isRunning {
		s.RUnlock()
		return nil
	}
	s.RUnlock()

	var (
		listener net.Listener
		err      error
	)

	if s.grpcErrorHandlerFunc != nil {
		errors.GrpcErrorHandlerFunc = s.grpcErrorHandlerFunc
	}

	if transportConfig := s.opts.TLSConfig; transportConfig != nil {
		listener, err = tls.Listen("tcp", fmt.Sprintf("%s:%d", s.opts.Host, s.opts.Port), transportConfig)
	} else {
		listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", s.opts.Host, s.opts.Port))
	}
	if err != nil {
		return err
	}

	r := mux.NewRouter()

	if s.opts.EnableCORS {
		r.Methods(http.MethodOptions).HandlerFunc(s.corsHandlerFunc)
	}

	r.NotFoundHandler = s.methodNotFoundHandler()
	r.MethodNotAllowedHandler = s.methodNotAllowedHandler()

	if s.opts.EnableCORS {
		s.middlewares.Add(&corsMiddleware{handler: s.corsHandlerFunc})
	}

	for _, h := range s.handlers {

		s.runtime.Log().Infof("register [http] route: %s", h.Path)

		handler, err := s.middlewares.apply(h)
		if err != nil {
			return err
		}
		r.Handle(h.Path, handler).Methods(h.Method)

		s.runtime.Log().Infof("bind handler: method: %s, path: %s", h.Method, h.Path)
	}

	go func() {
		s.runtime.Log().Infof("server [http] listening on %s", listener.Addr().String())
		if err := http.Serve(listener, r); err != nil {
			s.runtime.Log().Errorf("server [http] start error: %v", err)
		}
	}()

	s.Lock()
	s.isRunning = true
	s.Unlock()

	go func() {
		ch := <-s.exit
		ch <- listener.Close()
	}()

	return nil
}

func (s *httpServer) Stop() error {
	s.RLock()
	if !s.isRunning {
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
		s.isRunning = false
		s.Unlock()
	}

	return err
}

func (s *httpServer) UseMiddleware(middlewares ...server.KindMiddleware) {
	s.middlewares.SetGlobal(middlewares...)
}

func (s *httpServer) GetMiddlewares() []interface{} {
	return s.middlewares.constructors
}

func (s *httpServer) GetConstructor() interface{} {
	return s.constructor
}

func (s *httpServer) SetMiddleware(middleware any) {
	s.middlewares.AddConstructor(middleware)
}

func (s *httpServer) AddHandler(method string, path string, h http.HandlerFunc, opts ...server.HTTPServerOption) {
	key := fmt.Sprintf("%s:%s", method, path)
	s.handlers[key] = server.HTTPServerHandler{Method: method, Path: path, Handler: h, Options: opts}
}

func (s *httpServer) SetService(fn interface{}) {
	s.service = fn
}

func (s *httpServer) GetService() interface{} {
	return s.service
}

func (s *httpServer) SetCorsHandlerFunc(hf http.HandlerFunc) {
	s.corsHandlerFunc = hf
}

func (s *httpServer) SetErrorHandlerFunc(hf func(http.ResponseWriter, error)) {
	s.grpcErrorHandlerFunc = hf
}

func (s *httpServer) Subscribe(event string, h websockets.EventHandler) {
	s.wsManager.AddEventHandler(event, h)
}

func (s *httpServer) ServerWS(w http.ResponseWriter, r *http.Request) {
	s.wsManager.ServeWS(w, r)
}

func (s *httpServer) constructor(mws ...server.HttpServerMiddleware) {
	for _, mw := range mws {
		s.middlewares.Add(mw)
	}
}

func GetMarshaler(s server.HTTPServer, req *http.Request) (inbound, outbound marshaler.Marshaler) {
	for _, acceptVal := range req.Header[acceptHeader] {
		if m, ok := s.(*httpServer).marshalerMap[acceptVal]; ok {
			outbound = m
			break
		}
	}

	for _, contentTypeVal := range req.Header[contentTypeHeader] {
		contentType, _, err := mime.ParseMediaType(contentTypeVal)
		if err != nil {
			continue
		}
		if m, ok := s.(*httpServer).marshalerMap[contentType]; ok {
			inbound = m
			break
		}
	}

	if inbound == nil {
		inbound = DefaultMarshaler
	}
	if outbound == nil {
		outbound = inbound
	}

	return inbound, outbound
}
