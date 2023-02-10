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

package socket

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jpillora/backoff"
)

const (
	defaultMinReconnect      = 5 * time.Second
	defaultMaxReconnect      = 30 * time.Second
	defaultHandshakeTimeout  = 2 * time.Second
	defaultReconnectInterval = 1.5
)

// ErrNotConnected - returned when the application read/writes a message and the connection is closed
var ErrNotConnected = errors.New("websocket: not connected")

type ClientOpts struct {
	Reconnect        bool
	ReconnectAttempt int
}

func NewClient(o ...ClientOpts) *client {
	c := new(client)

	if len(o) == 1 {
		c.IsReconnect = o[0].Reconnect

		if o[0].ReconnectAttempt > 0 {
			c.ReconnectAttempt = o[0].ReconnectAttempt
		}

	}

	return c
}

// The client type represents a reconnecting WebSocket connection.
type client struct {
	lock sync.Mutex

	// Specifies the initial reconnecting interval, default to 2 seconds
	ReconnectMinInterval time.Duration
	// Specifies the maximum reconnecting interval, default to 30 seconds
	ReconnectMaxInterval time.Duration
	// Specifies the rate of increase of the reconnection interval, default to 1.5
	ReconnectRateInterval float64
	// HandshakeTimeout specifies the duration for the handshake to complete, default to 2 seconds
	HandshakeTimeout time.Duration
	// NonVerbose suppress connecting/reconnecting messages.
	NonVerbose bool

	url string

	*websocket.Conn

	reqHeader http.Header
	httpResp  *http.Response
	dialer    *websocket.Dialer

	dialErr error

	IsReconnect bool

	ReconnectAttempt int
}

// Will try to reconnect.
func (c *client) closeAndReconnect() {
	c.Close()

	if !c.IsReconnect {
		return
	}

	b := &backoff.Backoff{
		Min:    c.ReconnectMinInterval,
		Max:    c.ReconnectMaxInterval,
		Factor: c.ReconnectRateInterval,
		Jitter: true,
	}

	rand.Seed(time.Now().UTC().UnixNano())

	attempt := 0

	for {
		nextInterval := b.Duration()

		err := c.connect()
		if err == nil {
			if !c.NonVerbose {
				fmt.Println(fmt.Sprintf("dial: connection was successfully established with %s", c.url))
			}
			break
		}

		if c.ReconnectAttempt > 0 {
			attempt++
			if attempt > c.ReconnectAttempt {
				break
			}
			fmt.Println(fmt.Sprintf("dial: reconnect attepmt %d", attempt))
		} else {
			if !c.NonVerbose {
				fmt.Println("dial: will try again in", nextInterval, "seconds.")
			}
		}

		time.Sleep(nextInterval)
	}

}

// Close network connection without sending or waiting for a close frame.
func (c *client) Close() {
	c.lock.Lock()
	if c.Conn != nil {
		if err := c.Conn.Close(); err != nil {
			fmt.Println("Close err: ", err)
		}
	}
	c.lock.Unlock()
}

// ReadMessage is a helper method for getting a reader using NextReader
func (c *client) ReadMessage() (messageType int, message []byte, err error) {
	err = ErrNotConnected
	messageType, message, err = c.Conn.ReadMessage()
	if err != nil {
		c.closeAndReconnect()
	}
	return
}

// WriteMessage is a helper method for getting a writer using NextWriter.
func (c *client) WriteMessage(messageType int, data []byte) error {
	err := ErrNotConnected
	err = c.Conn.WriteMessage(messageType, data)
	if err != nil {
		c.closeAndReconnect()
	}
	return err
}

// WriteJSON writes the JSON encoding of v to the connection.
func (c *client) WriteJSON(v interface{}) error {
	err := ErrNotConnected
	err = c.Conn.WriteJSON(v)
	if err != nil {
		c.closeAndReconnect()
	}
	return err
}

// GetHTTPResponse returned the http response from the handshake
func (c *client) GetHTTPResponse() *http.Response {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.httpResp
}

// GetDialError -  Returns the last dialer error.
// nil on successful connection.
func (c *client) GetDialError() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.dialErr
}

// Dial creates a new client connection.
// The URI of the host and request.
// Use requestHeader to specify headers and more.
func (c *client) Dial(uri string, reqHeader http.Header) error {
	uri, err := c.parseURL(uri)
	if err != nil {
		fmt.Println(fmt.Sprintf("dial: %v", err))
	}

	c.url = uri

	if c.ReconnectMinInterval == 0 {
		c.ReconnectMinInterval = defaultMinReconnect
	}

	if c.ReconnectMaxInterval == 0 {
		c.ReconnectMaxInterval = defaultMaxReconnect
	}

	if c.HandshakeTimeout == 0 {
		c.HandshakeTimeout = defaultHandshakeTimeout
	}

	if c.ReconnectRateInterval == 0 {
		c.ReconnectRateInterval = defaultReconnectInterval
	}

	c.dialer = websocket.DefaultDialer
	c.dialer.HandshakeTimeout = c.HandshakeTimeout

	if err := c.connect(); err != nil {
		return err
	}

	// wait on first attempt
	time.Sleep(c.HandshakeTimeout)

	return nil
}

func (c *client) connect() error {

	conn, resp, err := c.dialer.Dial(c.url, c.reqHeader)
	if err != nil {
		return err
	}

	c.Conn = conn
	c.dialErr = err
	c.httpResp = resp

	return nil
}

func (c *client) parseURL(uri string) (string, error) {
	if uri == "" {
		return "", errors.New("dial: url cannot be empty")
	}

	u, err := url.Parse(uri)
	if err != nil {
		return "", errors.New("url: " + err.Error())
	}

	if u.Scheme != "ws" && u.Scheme != "wss" {
		return "", errors.New("url: websocket uris must start with ws or wss scheme")
	}

	if u.User != nil {
		return "", errors.New("url: user name and password are not allowed in websocket URIs")
	}

	return uri, nil
}
