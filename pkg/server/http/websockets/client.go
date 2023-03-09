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
	"encoding/json"
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
	"time"

	"github.com/gorilla/websocket"
)

// The message types are defined in RFC 6455, section 11.8.
const (
	TextMessage   = websocket.TextMessage
	BinaryMessage = websocket.BinaryMessage
	CloseMessage  = websocket.CloseMessage
	PingMessage   = websocket.PingMessage
	PongMessage   = websocket.PongMessage
)

type ClientList map[*Client]bool

// Client is a websocket client, basically a frontend visitor
type Client struct {
	ctx        context.Context
	log        logger.Logger
	connection *websocket.Conn
	manager    *Manager
}

var (
	// pongWait is how long we will await a pong response from client
	pongWait = 10 * time.Second
	// pingInterval has to be less than pongWait, We cant multiply by 0.9 to get 90% of time
	// Because that can make decimals, so instead *9 / 10 to get 90%
	// The reason why it has to be less than PingRequency is because otherwise it will send
	// a new Ping before getting response
	pingInterval = (pongWait * 9) / 10
)

// NewClient is used to initialize a new Client with all required values initialized
func NewClient(ctx context.Context, log logger.Logger, conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		ctx:        ctx,
		log:        log,
		connection: conn,
		manager:    manager,
	}
}

// WriteMessage is a helper method for getting a writer using NextWriter.
func (c *Client) WriteMessage(messageType int, data []byte) error {
	if err := c.connection.WriteMessage(messageType, data); err != nil {
		c.log.Errorf("failed write message: %v", err)
		return err
	}
	return nil
}

// WriteJSON writes the JSON encoding of v to the connection.
func (c *Client) WriteJSON(i interface{}) error {
	if err := c.connection.WriteJSON(i); err != nil {
		c.log.Errorf("failed write JSON message: %v", err)
		return err
	}
	return nil
}

// readMessages will start the client to read messages and handle them
// appropriately.
// This is supposed to be run as a goroutine
func (c *Client) readMessages() {
	defer func() {
		c.manager.removeClient(c)
	}()

	c.connection.SetReadLimit(512)

	if err := c.connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		c.log.Errorf("configure Wait time for Pong response failed %v", err)
		return
	}

	c.connection.SetPongHandler(c.pongHandler)

	for {
		_, payload, err := c.connection.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.log.Errorf("failed reading message: %v", err)
			}
			break
		}

		var request Event
		if err := json.Unmarshal(payload, &request); err != nil {
			c.log.Errorf("failed marshalling message: %v", err)
			c.connection.WriteMessage(TextMessage, []byte(`{"error":"invalid JSON format"}`))
			continue
		}

		if err := c.manager.routeEvent(request, c); err != nil {
			c.log.Errorf("failed handling message: %v", err)
		}
	}
}

// pongHandler is used to handle PongMessages for the Client
func (c *Client) pongHandler(_ string) error {
	return c.connection.SetReadDeadline(time.Now().Add(pongWait))
}

// writeMessages is a process that listens for new messages to output to the Client
func (c *Client) writeMessages() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.manager.removeClient(c)
	}()

	for {
		select {
		case <-ticker.C:
			if err := c.connection.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				c.log.Errorf("failed write ping message: %v", err)
				return
			}
		}

	}
}
