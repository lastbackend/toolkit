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

package http

import (
	"github.com/lastbackend/engine/cmd"
	"github.com/lastbackend/engine/logger"

	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
)

const (
	ServiceName = "http"
)

type httpServer struct {
	sync.RWMutex

	prefix string

	opts Options

	srv *http.ServeMux

	isRunnning bool
	registered bool

	exit chan chan error
}

func NewServer(prefix string) Server {
	return newServer(prefix)
}

func (h *httpServer) Name() string {
	return ServiceName
}

func (h *httpServer) Register(path string, handler http.Handler) error {
	for _, wrapper := range h.opts.Wrappers {
		handler = wrapper(handler)
	}

	if h.opts.EnableCORS {
		handler = AddCorsHandler(handler)
	}

	h.srv.Handle(path, handler)
	return nil
}

func (h *httpServer) Start() error {
	h.init()

	h.RLock()
	if h.isRunnning {
		h.RUnlock()
		return nil
	}
	h.RUnlock()

	var (
		listener net.Listener
		err      error
	)

	if transportConfig := h.opts.TLSConfig; transportConfig != nil {
		listener, err = tls.Listen("tcp", h.opts.Address, transportConfig)
	} else {
		listener, err = net.Listen("tcp", h.opts.Address)
	}
	if err != nil {
		return err
	}

	if logger.V(logger.InfoLevel, logger.DefaultLogger) {
		logger.Infof("server [http] Listening on %s", listener.Addr().String())
	}

	h.Lock()
	h.opts.Address = listener.Addr().String()
	h.Unlock()

	go func() {
		if err := http.Serve(listener, h.srv); err != nil {
			if logger.V(logger.InfoLevel, logger.DefaultLogger) {
				logger.Errorf("server [http] start error: %v", err)
			}
		}
	}()

	go func() {
		ch := <-h.exit
		ch <- listener.Close()
	}()

	h.Lock()
	h.isRunnning = true
	h.Unlock()

	return nil
}

func (h *httpServer) Stop() error {
	h.RLock()
	if !h.isRunnning {
		h.RUnlock()
		return nil
	}
	h.RUnlock()

	ch := make(chan error)
	h.exit <- ch

	var err error
	select {
	case err = <-ch:
		h.Lock()
		h.isRunnning = false
		h.Unlock()
	}

	return err
}

func (h *httpServer) Flags() []cmd.Flag {
	return []cmd.Flag{
		&cmd.StringFlag{
			Name:        h.withPrefix("address"),
			EnvVar:      h.withEnvPrefix("ADDRESS"),
			Usage:       "Server address for listening",
			Required:    false,
			Value:       defaultAddress,
			Destination: &h.opts.Address,
		},
		&cmd.StringFlag{
			Name:        h.withPrefix("name"),
			EnvVar:      h.withEnvPrefix("NAME"),
			Usage:       "Server name",
			Required:    false,
			Value:       defaultName,
			Destination: &h.opts.Name,
		},
	}
}

func newServer(prefix string) *httpServer {
	srv := &httpServer{
		prefix: prefix,
		opts:   defaultOptions(),
		exit:   make(chan chan error),
	}

	srv.init()

	return srv
}

func (h *httpServer) init() {
	h.Lock()
	defer h.Unlock()

	h.srv = http.NewServeMux()
}

func (h *httpServer) withPrefix(name string) string {
	return fmt.Sprintf("%s-%s", h.prefix, name)
}

func (h *httpServer) withEnvPrefix(name string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s", h.prefix, name))
}
