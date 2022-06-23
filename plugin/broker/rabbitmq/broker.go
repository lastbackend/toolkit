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
	"context"
	"crypto/tls"
	"encoding/json"
	"sync"

	"github.com/lastbackend/toolkit/logger"
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
	opts           brokerOptions
	endpoints      []string
	prefetchCount  int
	prefetchGlobal bool
	exchange       Exchange

	wg sync.WaitGroup
}

func newBroker(opts brokerOptions) *broker {

	exchange := DefaultExchange
	if opts.DefaultExchange != nil {
		exchange = *opts.DefaultExchange
	}

	return &broker{
		endpoints: []string{opts.Endpoint},
		opts:      opts,
		exchange:  exchange,
	}
}

func (r *broker) Ack(ctx context.Context) error {
	fn, ok := ctx.Value(ack{}).(func() error)
	if !ok {
		return errors.New("no acknowledged")
	}
	return fn()
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

func (r *broker) Publish(exchange, topic string, data interface{}, opts *PublishOptions) error {

	body, err := json.Marshal(data)
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

	return r.conn.Publish(exchange, topic, m)
}

func (r *broker) Subscribe(queue, topic string, handler Handler, opts *SubscribeOptions) (Subscriber, error) {
	var ackSuccess bool

	if r.conn == nil {
		return nil, errors.New("not connected")
	}

	if opts == nil {
		opts = new(SubscribeOptions)
	}

	opt := SubscribeOptions{
		AutoAck: opts.AutoAck,
	}

	fn := func(msg amqp.Delivery) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, ack{}, msg.Ack)
		ctx = context.WithValue(ctx, reject{}, msg.Reject)

		header := make(map[string]string)
		for k, v := range msg.Headers {
			header[k], _ = v.(string)
		}

		m := &Message{
			Header: header,
			Body:   msg.Body,
		}

		p := &publisher{
			delivery: msg,
			message:  m,
			topic:    msg.RoutingKey,
			err:      handler(ctx, msg.Body),
		}

		if p.err == nil && ackSuccess && !opt.AutoAck {
			if err := msg.Ack(false); err != nil {
				if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
					logger.Error(err)
				}
			}
		} else if p.err != nil && !opt.AutoAck {
			if err := msg.Nack(false, opts.RequeueOnError); err != nil {
				if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
					logger.Error(err)
				}
			}
		}
	}

	sb := &consumer{
		queue:        queue,
		topic:        topic,
		opts:         opt,
		broker:       r,
		fn:           fn,
		headers:      opts.Headers,
		durableQueue: opts.DurableQueue,
	}

	go sb.resubscribe()

	return sb, nil
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
		cer, err := tls.LoadX509KeyPair(r.opts.TLSCert, r.opts.TLSKey)
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
