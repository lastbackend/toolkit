/*
Copyright [2014] - [2022] The Last.Backend authors.

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

package rabbitmq

import (
	"github.com/streadway/amqp"

	"crypto/tls"
	"strings"
	"sync"
	"time"
)

type brokerOptions struct {
	Endpoint  string
	TLSCert   string
	TLSKey    string
	TLSVerify bool
}

type broker struct {
	sync.Mutex

	conn *amqp.Connection

	endpoint  string
	tlsConfig *tls.Config

	connected bool
	close     chan bool

	waitConnection chan struct{}
}

func newBroker(opts brokerOptions) (*broker, error) {
	var tlsConfig *tls.Config

	if opts.TLSVerify {
		cer, err := tls.LoadX509KeyPair(opts.TLSCert, opts.TLSKey)
		if err != nil {
			return nil, err
		}
		tlsConfig = &tls.Config{Certificates: []tls.Certificate{cer}}
	}

	return &broker{
		endpoint:  opts.Endpoint,
		tlsConfig: tlsConfig,
	}, nil
}

func (b *broker) Connect() error {

	config := amqp.Config{}

	if config.TLSClientConfig != nil || strings.HasPrefix(b.endpoint, "amqps://") {
		if config.TLSClientConfig == nil {
			config.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}

		b.endpoint = strings.Replace(b.endpoint, "amqp://", "amqps://", 1)
	}

	if err := b.tryConnect(config); err != nil {
		return err
	}

	b.Lock()
	b.connected = true
	b.Unlock()

	go b.reconnect(config)

	return nil
}

func (b *broker) Disconnect() error {
	return b.conn.Close()
}

func (b *broker) tryConnect(config amqp.Config) error {

	conn, err := amqp.DialConfig(b.endpoint, config)
	if err != nil {
		return err
	}

	b.conn = conn

	return nil
}

func (b *broker) reconnect(config amqp.Config) {
	var connect bool

	for {
		if connect {
			if err := b.tryConnect(config); err != nil {
				time.Sleep(1 * time.Second)
				continue
			}

			b.Lock()
			b.connected = true
			b.Unlock()

			close(b.waitConnection)
		}

		connect = true
		notifyClose := make(chan *amqp.Error)
		b.conn.NotifyClose(notifyClose)

		select {
		case <-notifyClose:
			b.Lock()
			b.connected = false
			b.waitConnection = make(chan struct{})
			b.Unlock()
		case <-b.close:
			return
		}
	}
}
