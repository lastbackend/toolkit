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

package rabbitmq

import (
	"crypto/tls"
	"github.com/lastbackend/toolkit/pkg/runtime"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

var (
	DefaultPrefetchCount = 0
	DefaultExchange      = Exchange{
		Name: "toolkit",
	}

	DefaultRabbitURL = "amqp://guest:guest@127.0.0.1:5672"

	defaultHeartbeat = 10 * time.Second
	defaultLocale    = "en_US"

	defaultAmqpConfig = amqp.Config{
		Heartbeat: defaultHeartbeat,
		Locale:    defaultLocale,
	}

	dialConfig = amqp.DialConfig
)

type amqpConn struct {
	sync.Mutex

	runtime         runtime.Runtime
	conn            *amqp.Connection
	channel         *amqpChannel
	exchangeChannel *amqpChannel
	exchange        Exchange
	url             string
	prefetchCount   int
	prefetchGlobal  bool

	err       error
	connected bool
	close     chan bool

	waitConnection chan struct{}
}

type Exchange struct {
	Name    string
	Durable bool
}

func newConnection(runtime runtime.Runtime, ex Exchange, urls []string, prefetchCount int, prefetchGlobal bool) *amqpConn {

	url := DefaultRabbitURL

	if len(urls) > 0 && regexp.MustCompile("^amqp(s)?://.*").MatchString(urls[0]) {
		url = urls[0]
	}

	ret := &amqpConn{
		runtime:        runtime,
		exchange:       ex,
		url:            url,
		prefetchCount:  prefetchCount,
		prefetchGlobal: prefetchGlobal,
		close:          make(chan bool),
		waitConnection: make(chan struct{}),
	}

	close(ret.waitConnection)

	return ret
}

func (a *amqpConn) connect(secure bool, config *amqp.Config) error {
	if err := a.tryConnect(secure, config); err != nil {
		return err
	}

	a.connected = true
	a.err = nil

	go a.reconnect(secure, config)

	return nil
}

func (a *amqpConn) reconnect(secure bool, config *amqp.Config) {

	for {
		if !a.connected {
			if err := a.tryConnect(secure, config); err != nil {
				time.Sleep(1 * time.Second)
				continue
			}

			a.Lock()
			a.connected = true
			a.err = nil
			a.Unlock()

			close(a.waitConnection)
		}

		notifyClose := make(chan *amqp.Error)
		a.conn.NotifyClose(notifyClose)
		chanNotifyClose := make(chan *amqp.Error)
		channel := a.exchangeChannel.channel
		channel.NotifyClose(chanNotifyClose)

		for notifyClose != nil || chanNotifyClose != nil {
			select {
			case err := <-chanNotifyClose:
				a.runtime.Log().Error(err)

				a.Lock()
				a.connected = false
				a.err = err
				a.waitConnection = make(chan struct{})
				a.Unlock()

				chanNotifyClose = nil
			case err := <-notifyClose:
				a.runtime.Log().Error(err)

				a.Lock()
				a.connected = false
				a.err = err
				a.waitConnection = make(chan struct{})
				a.Unlock()

				notifyClose = nil
			case <-a.close:
				return
			}
		}
	}
}

func (a *amqpConn) Connect(secure bool, config *amqp.Config) error {
	a.Lock()
	if a.connected {
		a.Unlock()
		return nil
	}
	a.Unlock()

	select {
	case <-a.close:
		a.close = make(chan bool)
	default:
	}

	return a.connect(secure, config)
}

func (a *amqpConn) Connected() error {
	return a.err
}

func (a *amqpConn) Close() error {
	if a.conn == nil {
		return nil
	}

	select {
	case <-a.close:
		return nil
	default:
		close(a.close)
		a.connected = false
		a.err = errors.New("connection closed")
	}

	return a.conn.Close()
}

func (a *amqpConn) Channel() (*amqp.Channel, error) {
	if !a.connected {
		return nil, errors.New("connection closed")
	}

	return a.conn.Channel()
}

func (a *amqpConn) tryConnect(secure bool, config *amqp.Config) error {
	var err error

	if config == nil {
		config = &defaultAmqpConfig
	}

	url := a.url

	if secure || config.TLSClientConfig != nil || strings.HasPrefix(a.url, "amqps://") {
		if config.TLSClientConfig == nil {
			config.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}

		url = strings.Replace(a.url, "amqp://", "amqps://", 1)
	}

	a.conn, err = dialConfig(url, *config)
	if err != nil {
		return err
	}

	if a.channel, err = newRabbitChannel(a.conn, a.prefetchCount, a.prefetchGlobal); err != nil {
		return err
	}

	if a.exchange.Durable {
		err = a.channel.DeclareDurableExchange(a.exchange.Name, "fanout")
	} else {
		err = a.channel.DeclareExchange(a.exchange.Name, "fanout")
	}
	if err != nil {
		return err
	}

	a.exchangeChannel, err = newRabbitChannel(a.conn, a.prefetchCount, a.prefetchGlobal)

	return err
}

func (a *amqpConn) Consume(exchange, queue, key string, headers amqp.Table, qArgs amqp.Table, autoAck, durableQueue bool) (*amqpChannel, <-chan amqp.Delivery, error) {
	ch, err := newRabbitChannel(a.conn, a.prefetchCount, a.prefetchGlobal)
	if err != nil {
		return nil, nil, err
	}

	if durableQueue {
		err = ch.DeclareDurableQueue(queue, qArgs)
	} else {
		err = ch.DeclareQueue(queue, qArgs)
	}
	if err != nil {
		return nil, nil, err
	}

	consumers, err := ch.ConsumeQueue(queue, autoAck)
	if err != nil {
		return nil, nil, err
	}

	err = ch.BindQueue(queue, key, exchange, headers)
	if err != nil {
		return nil, nil, err
	}

	return ch, consumers, nil
}

func (a *amqpConn) Publish(exchange, key string, msg amqp.Publishing) error {
	return a.exchangeChannel.Publish(exchange, key, msg)
}
