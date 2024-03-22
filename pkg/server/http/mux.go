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
	"fmt"
	"mime"
	"net/http"
	"regexp"
	"sync"

	"github.com/gorilla/mux"
	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/server"
	"github.com/lastbackend/toolkit/pkg/server/http/errors"
	"github.com/lastbackend/toolkit/pkg/server/http/marshaler"
	"github.com/lastbackend/toolkit/pkg/server/http/websockets"
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

	// fn for init user-defined service
	// fn for server registration
	service interface{}

	middlewares *Middlewares

	corsHandlerFunc http.HandlerFunc

	wsManager *websockets.Manager

	server *http.Server
	exit   chan chan error

	r *mux.Router
}

func NewServer(name string, runtime runtime.Runtime, options *server.HTTPServerOptions) server.HTTPServer {

	s := &httpServer{
		runtime:      runtime,
		prefix:       defaultPrefix,
		marshalerMap: GetMarshalerMap(),
		exit:         make(chan chan error),

		corsHandlerFunc: corsHandlerFunc,

		middlewares: newMiddlewares(runtime.Log()),
		wsManager:   websockets.NewManager(runtime.Log()),
		handlers:    make(map[string]server.HTTPServerHandler, 0),

		r: mux.NewRouter(),
	}

	name = regexp.MustCompile(`[^_a-zA-Z0-9 ]+`).ReplaceAllString(name, "_")

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

	if s.opts.EnableCORS {
		s.r.Methods(http.MethodOptions).HandlerFunc(s.corsHandlerFunc)
		s.middlewares.global = append(s.middlewares.global, corsMiddlewareKind)
		s.middlewares.Add(&corsMiddleware{handler: s.corsHandlerFunc})
	}

	s.r.NotFoundHandler = s.methodNotFoundHandler()
	s.r.MethodNotAllowedHandler = s.methodNotAllowedHandler()

	s.server = &http.Server{
		Addr:      fmt.Sprintf("%s:%d", s.opts.Host, s.opts.Port),
		Handler:   s.r,
		TLSConfig: s.opts.TLSConfig,
	}

	for _, h := range s.handlers {
		if err := s.registerHandler(h); err != nil {
			return err
		}
	}

	s.Lock()
	s.isRunning = true
	s.Unlock()

	go func() {
		s.runtime.Log().V(5).Infof("server [http] [%s] started", s.server.Addr)
		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			s.runtime.Log().Errorf("server [http] [%s] start error: %v", s.server.Addr, err)
		}
		s.runtime.Log().V(5).Infof("server [http] [%s] stopped", s.server.Addr)
		s.Lock()
		s.isRunning = false
		s.Unlock()
	}()

	return nil
}

func (s *httpServer) registerHandler(h server.HTTPServerHandler) error {
	s.runtime.Log().V(5).Infof("register [http] route: %s", h.Path)

	handler, err := s.middlewares.apply(h)
	if err != nil {
		return err
	}
	s.r.Handle(h.Path, handler).Methods(h.Method)

	s.runtime.Log().V(5).Infof("bind handler: method: %s, path: %s", h.Method, h.Path)

	return nil
}

func (s *httpServer) Stop(ctx context.Context) error {
	s.runtime.Log().V(5).Infof("server [http] [%s] stop call start", s.server.Addr)

	if err := s.server.Shutdown(ctx); err != nil {
		s.runtime.Log().Errorf("server [http] [%s] stop call error: %v", s.server.Addr, err)
		return err
	}

	s.runtime.Log().V(5).Infof("server [http] [%s] stop call end", s.server.Addr)
	return nil
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
	if !s.isRunning {
		s.handlers[key] = server.HTTPServerHandler{Method: method, Path: path, Handler: h, Options: opts}
	} else {
		_ = s.registerHandler(server.HTTPServerHandler{Method: method, Path: path, Handler: h, Options: opts})
	}
}

func (s *httpServer) SetCorsHandlerFunc(hf http.HandlerFunc) {
	s.corsHandlerFunc = hf
}

func (s *httpServer) SetErrorHandlerFunc(hf func(http.ResponseWriter, error)) {
	errors.GrpcErrorHandlerFunc = hf
}

func (s *httpServer) Subscribe(event string, h websockets.EventHandler) {
	s.wsManager.AddEventHandler(event, h)
}

func (s *httpServer) ServerWS(w http.ResponseWriter, r *http.Request) {
	s.wsManager.ServeWS(w, r)
}

// SetService - set user-defined handlers
func (s *httpServer) SetService(service interface{}) {
	s.service = service
	return
}

// GetService - set user-defined handlers
func (s *httpServer) GetService() interface{} {
	return s.service
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
