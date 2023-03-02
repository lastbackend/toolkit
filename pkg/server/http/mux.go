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
	logger2 "github.com/lastbackend/toolkit/pkg/runtime/logger"
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

func NewServer(runtime runtime.Runtime, _ *server.HTTPServerOptions) server.HTTPServer {

	s := &httpServer{
		runtime:      runtime,
		prefix:       defaultPrefix,
		marshalerMap: GetMarshalerMap(),
		exit:         make(chan chan error),

		corsHandlerFunc:      corsHandlerFunc,
		grpcErrorHandlerFunc: errors.GrpcErrorHandlerFunc,

		middlewares: newMiddlewares(runtime.Log()),
		wsManager:   websockets.NewManager(),
		handlers:    make(map[string]server.HTTPServerHandler, 0),
	}

	if err := runtime.Config().Parse(&s.opts, s.prefix); err != nil {
		return nil
	}

	runtime.Config().Print(&s.opts, s.prefix)

	return s
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

	for _, h := range s.handlers {

		if logger2.V(logger2.DebugLevel, logger2.DefaultLogger) {
			logger2.Debugf("register [http] route: %s", h.Path)
		}

		s.middlewares.apply(r, h.Options)
		r.Handle(h.Path, h.Handler).Methods(h.Method)
	}

	go func() {
		logger2.Infof("server [http] listening on %s", listener.Addr().String())
		if err := http.Serve(listener, r); err != nil {
			if logger2.V(logger2.ErrorLevel, logger2.DefaultLogger) {
				logger2.Errorf("server [http] start error: %v", err)
			}
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

func (s *httpServer) UseMiddleware(middlewares ...string) {
	s.middlewares.SetGlobal(middlewares...)
}

func (s *httpServer) AddMiddleware(name string, middleware server.HttpServerMiddleware) {
	s.middlewares.Add(name, middleware)
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

func (s *httpServer) handle(h http.Handler) http.Handler {
	headers := func(h http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if s.opts.EnableCORS {
				s.corsHandlerFunc(w, r)
			}
			h.ServeHTTP(w, r)
		}
	}
	h = headers(h)
	return h
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
