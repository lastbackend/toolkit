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

package router

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/lastbackend/toolkit/pkg/cmd"
	"github.com/lastbackend/toolkit/pkg/logger"
	"github.com/lastbackend/toolkit/pkg/router/errors"
	"github.com/lastbackend/toolkit/pkg/router/marshaler"
	"github.com/lastbackend/toolkit/pkg/router/ws"
	"mime"
	"net"
	"net/http"
	"strings"
	"sync"
)

const (
	defaultPrefix = "http"
)

var (
	acceptHeader      = http.CanonicalHeaderKey("Accept")
	contentTypeHeader = http.CanonicalHeaderKey("Content-Type")
)

type server struct {
	sync.RWMutex

	opts Options

	prefix    string
	address   string
	isRunning bool

	handlers     map[string]handler
	marshalerMap map[string]marshaler.Marshaler

	middlewares Middleware

	corsHandlerFunc      http.HandlerFunc
	grpcErrorHandlerFunc func(http.ResponseWriter, error)

	wsManager *ws.Manager

	exit chan chan error
}

type handler struct {
	path          string
	method        string
	handler       http.HandlerFunc
	handleOptions HandleOptions
}

func New(cli cmd.FlagSet) Server {
	s := &server{
		prefix:               defaultPrefix,
		marshalerMap:         GetMarshalerMap(),
		opts:                 defaultOptions(),
		exit:                 make(chan chan error),
		corsHandlerFunc:      corsHandlerFunc,
		grpcErrorHandlerFunc: errors.GrpcErrorHandlerFunc,
		wsManager:            ws.NewManager(),
		handlers:             make(map[string]handler, 0),
	}

	s.addFlags(cli)

	return s
}

func (s *server) Start(_ context.Context) error {

	if s.opts.IsDisable {
		return nil
	}

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
		listener, err = tls.Listen("tcp", s.opts.Address, transportConfig)
	} else {
		listener, err = net.Listen("tcp", s.opts.Address)
	}
	if err != nil {
		return err
	}

	s.Lock()
	s.address = listener.Addr().String()
	s.Unlock()

	r := mux.NewRouter()

	if s.opts.EnableCORS {
		r.Methods(http.MethodOptions).HandlerFunc(s.corsHandlerFunc)
	}

	r.NotFoundHandler = s.methodNotFoundHandler()
	r.MethodNotAllowedHandler = s.methodNotAllowedHandler()

	for _, h := range s.handlers {
		if logger.V(logger.DebugLevel, logger.DefaultLogger) {
			logger.Debugf("register [http] route: %s", h.path)
		}
		hndl := h.handleOptions.Middlewares.apply(h.handler)
		hndl = s.middlewares.apply(hndl)
		r.Handle(h.path, s.handle(hndl)).Methods(h.method)
	}

	go func() {
		logger.Infof("server [http] listening on %s", listener.Addr().String())
		if err := http.Serve(listener, r); err != nil {
			if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
				logger.Errorf("server [http] start error: %v", err)
			}
		}
	}()

	s.Lock()
	s.isRunning = true
	s.Unlock()

	ch := <-s.exit
	ch <- listener.Close()

	return nil
}

func (s *server) Stop() error {
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

func (s *server) Handle(method string, path string, h http.HandlerFunc, opts HandleOptions) {
	key := fmt.Sprintf("%s:%s", method, path)
	s.handlers[key] = handler{method: method, path: path, handler: h, handleOptions: opts}
}

func (s *server) SetCorsHandlerFunc(hf http.HandlerFunc) {
	s.corsHandlerFunc = hf
}

func (s *server) SetErrorHandlerFunc(hf func(http.ResponseWriter, error)) {
	s.grpcErrorHandlerFunc = hf
}

func (s *server) AddMiddleware(h func(h http.Handler) http.Handler) {
	s.middlewares.Add(h)
}

func (s *server) Subscribe(event string, h ws.EventHandler) {
	s.wsManager.AddEventHandler(event, h)
}

func (s *server) ServerWS(w http.ResponseWriter, r *http.Request) {
	s.wsManager.ServeWS(w, r)
}

func (s *server) addFlags(cli cmd.FlagSet) {
	cli.AddStringFlag(s.withPrefix("address"), &s.opts.Address).
		Env(s.withEnvPrefix("ADDRESS")).
		Usage("Server address for listening").
		Default(defaultAddress)

	cli.AddBoolFlag(s.withPrefix("enable-cors"), &s.opts.EnableCORS).
		Env(s.withEnvPrefix("ENABLE_CORS")).
		Usage("Server with CORS")

	cli.AddBoolFlag(s.withPrefix("disable"), &s.opts.IsDisable).
		Env(s.withEnvPrefix("DISABLE")).
		Usage("Sets the disable server")
}

func (s *server) withPrefix(name string) string {
	return fmt.Sprintf("%s-%s", s.prefix, name)
}

func (s *server) withEnvPrefix(name string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s", s.prefix, name))
}

func (s *server) handle(h http.Handler) http.Handler {
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

func GetMarshaler(s Server, req *http.Request) (inbound, outbound marshaler.Marshaler) {
	for _, acceptVal := range req.Header[acceptHeader] {
		if m, ok := s.(*server).marshalerMap[acceptVal]; ok {
			outbound = m
			break
		}
	}

	for _, contentTypeVal := range req.Header[contentTypeHeader] {
		contentType, _, err := mime.ParseMediaType(contentTypeVal)
		if err != nil {
			continue
		}
		if m, ok := s.(*server).marshalerMap[contentType]; ok {
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
