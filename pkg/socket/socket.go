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
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Socket struct {
	sync.Mutex

	ctx context.Context

	socket  *websocket.Conn
	session string

	write chan []byte

	event chan *Message

	ping chan []byte
	pong chan []byte

	close chan bool

	done chan *Socket
}

type Message struct {
	Data   []byte
	Socket *Socket
}

type MSG struct {
	MT int
	MD []byte
}

const (
	writeWait = 10 * time.Second
	pongWait  = 10 * time.Second
)

func (s *Socket) listen() {

	// Create listener to pipe message to hub
	pipe := make(chan MSG)
	go func() {
		for {
			select {
			case p := <-pipe:

				if err := s.socket.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
					fmt.Println(err)
				}

				s.Lock()
				err := s.socket.WriteMessage(p.MT, p.MD)
				s.Unlock()

				if err != nil {

					if err := s.disconnect(); err != nil {
						//fmt.Println(err)
					}
				}

			}
		}
	}()

	go func() {

		for {

			select {

			case m := <-s.write:
				pipe <- MSG{websocket.TextMessage, m}

			case p := <-s.ping:
				pipe <- MSG{websocket.PingMessage, p}

			case p := <-s.pong:
				pipe <- MSG{websocket.PongMessage, p}

			case <-s.close:
				if err := s.socket.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
					fmt.Println(err)
				}

				if err := s.socket.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
					//fmt.Println(err)
				}

				if err := s.disconnect(); err != nil {
					//fmt.Println(err)
				}
			}
		}
	}()

	s.socket.SetPongHandler(func(string) error {
		return s.socket.SetReadDeadline(time.Now().Add(pongWait))
	})

	go func() {
		for {

			m, b, err := s.socket.ReadMessage()

			if m == websocket.PingMessage {
				s.pong <- []byte{}
				continue
			}

			if ce, ok := err.(*websocket.CloseError); ok {
				switch ce.Code {
				case websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived:
					s.done <- s
					return
				}
			}

			if err != nil {
				s.done <- s
				//fmt.Println(err)
				break
			}

			if s.event != nil {
				s.event <- &Message{Data: b, Socket: s}
			}

		}
	}()

}

func (s *Socket) pump(p MSG) error {

	limit := 5

	for i := 0; i <= limit; i++ {

		if err := s.socket.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
			return err
		}

		s.Lock()
		err := s.socket.WriteMessage(p.MT, p.MD)
		s.Unlock()

		if err != nil {

			if i < limit {
				ticker := time.NewTicker(time.Second)
				select {
				case <-ticker.C:
				}

				continue
			}

			return err
		}

		return nil
	}

	return nil
}

func (s *Socket) disconnect() error {

	if s.socket != nil {
		if err := s.socket.Close(); err != nil {
			s.done <- s
			return err
		}
	}

	s.done <- s

	return nil
}

func (s *Socket) Write(msg []byte) {
	s.write <- msg
}

func (s *Socket) Context() context.Context {
	return s.ctx
}

func (s *Socket) Close() {
	s.close <- true
}

// Ping connection to be online
func (s *Socket) Ping() {
	s.ping <- []byte{}
}

// New - return new Socket connection
func New(ctx context.Context, socket *websocket.Conn, leave chan *Socket, event chan *Message) *Socket {
	s := new(Socket)
	s.ctx = ctx

	s.socket = socket

	s.write = make(chan []byte)
	s.ping = make(chan []byte)
	s.pong = make(chan []byte)

	s.close = make(chan bool)

	s.event = event
	s.done = leave

	s.listen()

	return s
}
