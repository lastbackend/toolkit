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
	"context"
	"crypto/tls"
	"encoding/json"
	"sync"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type ack struct{}
type reject struct{}

type brokerOptions struct {
	Endpoint        string
	TLSVerify       bool
	TLSCert         string
	TLSKey          string
	PrefetchCount   int
	PrefetchGlobal  bool
	DefaultExchange *Exchange
}

type broker struct {
	mtx sync.Mutex

	conn           *amqpConn
	opts           Config
	endpoints      []string
	prefetchCount  int
	prefetchGlobal bool
	exchange       Exchange

	wg sync.WaitGroup
}

type message struct {
	Event   string `json:"event"`
	Payload string `json:"payload"`
}

func newBroker(opts Config) *broker {

	exchange := DefaultExchange
	if opts.DefaultExchange != nil {
		exchange = *opts.DefaultExchange
	}

	return &broker{
		endpoints: []string{opts.DSN},
		opts:      opts,
		exchange:  exchange,
	}
}

func (r *broker) Ack(ctx context.Context) error {
	fn, ok := ctx.Value(ack{}).(func(bool) error)
	if !ok {
		return errors.New("no acknowledged")
	}
	return fn(false)
}

func (r *broker) Reject(ctx context.Context) error {
	fn, ok := ctx.Value(reject{}).(func(bool) error)
	if !ok {
		return errors.New("no rejected")
	}
	return fn(false)
}

func (r *broker) RejectAndRequeue(ctx context.Context) error {
	fn, ok := ctx.Value(reject{}).(func(bool) error)
	if !ok {
		return errors.New("no rejected")
	}
	return fn(true)
}

func (r *broker) Publish(exchange, event string, payload []byte, opts *PublishOptions) error {

	e := message{
		Event:   event,
		Payload: string(payload),
	}

	body, err := json.Marshal(e)
	if err != nil {
		return err
	}

	if opts == nil {
		opts = new(PublishOptions)
	}

	m := amqp.Publishing{
		Body:    body,
		Headers: amqp.Table{},
	}

	if opts.Headers != nil {
		for k, v := range opts.Headers {
			m.Headers[k] = v
		}
	}

	if r.conn == nil {
		return errors.New("connection is nil")
	}

	return r.conn.Publish(exchange, "*", m)
}

func (r *broker) Subscribe(exchange, queue string, handler SubscriberHandler, opts *SubscribeOptions) (Subscriber, error) {

	if r.conn == nil {
		return nil, errors.New("not connected")
	}

	if opts == nil {
		opts = new(SubscribeOptions)
	}

	fn := func(msg amqp.Delivery) {
		ctx := context.Background()

		e := message{}
		json.Unmarshal(msg.Body, &e)

		headers := make(map[string]string)
		for k, v := range msg.Headers {
			headers[k], _ = v.(string)
		}

		ctx = context.WithValue(ctx, "headers", headers)
		handler(ctx, e.Event, []byte(e.Payload))
	}

	sb := &consumer{
		exchange:     exchange,
		queue:        queue,
		key:          "*",
		autoAck:      true,
		broker:       r,
		fn:           fn,
		headers:      opts.Headers,
		durableQueue: opts.DurableQueue,
	}

	go sb.resubscribe()

	return sb, nil
}

func (r *broker) Connected() error {
	return r.conn.Connected()
}

func (r *broker) Channel() (*amqp.Channel, error) {
	return r.conn.Channel()
}

func (r *broker) Connect() error {
	if r.conn == nil {
		r.conn = newConnection(r.exchange, r.endpoints, r.opts.PrefetchCount, r.opts.PrefetchGlobal)
	}

	conf := defaultAmqpConfig

	if r.opts.TLSVerify {
		cer, err := tls.X509KeyPair([]byte(r.opts.TLSCert), []byte(r.opts.TLSKey))
		if err != nil {
			return err
		}
		conf.TLSClientConfig = &tls.Config{Certificates: []tls.Certificate{cer}}
	}

	return r.conn.Connect(r.opts.TLSVerify, &conf)
}

func (r *broker) Disconnect() error {
	if r.conn == nil {
		return errors.New("not connected")
	}

	err := r.conn.Close()

	r.wg.Wait()

	return err
}
