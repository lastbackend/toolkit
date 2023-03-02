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

package websockets

import (
	"context"
	logger2 "github.com/lastbackend/toolkit/pkg/runtime/logger"
	"net/http"
	"reflect"
	"sync"
	"unsafe"

	"github.com/gorilla/websocket"
	"github.com/lastbackend/toolkit/pkg/server/http/errors"
	"github.com/lastbackend/toolkit/pkg/util/converter"
)

var (
	// upgrader is used to upgrade incoming HTTP requests
	// into a persistent websocket connection
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

var (
	ErrEventNotSupported = errors.New("this event type is not supported")
)

const (
	RequestHeaders = "x-http-req-headers"
)

// Manager is used to hold references to all Clients Registered, and Broadcasting etc
type Manager struct {
	sync.RWMutex

	clients  ClientList
	handlers map[string]EventHandler
}

// NewManager is used to initialize all the values inside the manager
func NewManager() *Manager {
	m := &Manager{
		clients:  make(ClientList),
		handlers: make(map[string]EventHandler),
	}
	return m
}

// AddEventHandler configures and adds all handlers
func (m *Manager) AddEventHandler(event string, handler EventHandler) {
	m.Lock()
	defer m.Unlock()
	m.handlers[event] = handler
}

// routeEvent is used to make sure the correct event goes into the correct handler
func (m *Manager) routeEvent(event Event, c *Client) error {
	if handler, ok := m.handlers[event.Type]; ok {
		if err := handler(c.ctx, event, c); err != nil {
			return err
		}
		return nil
	} else {
		return ErrEventNotSupported
	}
}

// ServeWS is an HTTP Handler that the has the Manager that allows connections
func (m *Manager) ServeWS(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	// Begin by upgrading the HTTP request
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if logger2.V(logger2.ErrorLevel, logger2.DefaultLogger) {
			logger2.Errorf("upgrading the HTTP request failed %v", err)
		}
		return
	}

	// Copy request context data
	ctx := context.Background()
	contextValues := reflect.ValueOf(r.Context()).Elem()
	contextKeys := reflect.TypeOf(r.Context()).Elem()

	if contextKeys.Kind() == reflect.Struct {
		for i := 0; i < contextValues.NumField(); i++ {
			reflectValue := contextValues.Field(i)
			reflectValue = reflect.NewAt(reflectValue.Type(), unsafe.Pointer(reflectValue.UnsafeAddr())).Elem()

			reflectField := contextKeys.Field(i)
			if reflectField.Name == "key" {
				ctx = context.WithValue(ctx, reflectValue.Interface(), r.Context().Value(reflectValue.Interface()))
			}
		}
	}

	headers, err := converter.PrepareHeaderFromRequest(r)
	if err != nil {
		errors.HTTP.InternalServerError(w)
		return
	}

	ctx = context.WithValue(ctx, RequestHeaders, headers)

	client := NewClient(ctx, conn, m)

	m.addClient(client)

	go client.readMessages()
	go client.writeMessages()
}

// addClient will add clients to our clients list
func (m *Manager) addClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	m.clients[client] = true
}

// removeClient will remove the client and clean up
func (m *Manager) removeClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.clients[client]; ok {
		if err := client.connection.Close(); err != nil {
			if logger2.V(logger2.ErrorLevel, logger2.DefaultLogger) {
				logger2.Errorf("close client failed %v", err)
			}
		}
		delete(m.clients, client)
	}
}
