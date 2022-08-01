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
	"sync"
	"time"

	"github.com/lastbackend/toolkit/logger"
	"github.com/streadway/amqp"
)

type consumer struct {
	done         bool
	mtx          sync.Mutex
	exchange     string
	queue        string
	key          string
	durableQueue bool
	autoAck      bool
	broker       *broker
	ch           *amqpChannel
	fn           func(msg amqp.Delivery)
	headers      map[string]interface{}
	queueArgs    map[string]interface{}
}

func (c *consumer) Unsubscribe() error {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.done = true

	if c.ch != nil {
		return c.ch.Close()
	}

	return nil
}

func (c *consumer) resubscribe() {

	minDelay := 100 * time.Millisecond
	maxDelay := 30 * time.Second
	expFactor := time.Duration(2)
	delay := minDelay

	for {
		if c.done {
			return
		}

		select {
		case <-c.broker.conn.close:
			return
		case <-c.broker.conn.waitConnection:
		case <-time.After(time.Second):
			continue
		}

		c.broker.mtx.Lock()
		if !c.broker.conn.connected {
			c.broker.mtx.Unlock()
			continue
		}

		ch, sub, err := c.broker.conn.Consume(
			c.exchange,
			c.queue,
			c.key,
			c.headers,
			c.queueArgs,
			c.autoAck,
			c.durableQueue,
		)

		c.broker.mtx.Unlock()

		if err != nil {
			if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
				logger.Error(err)
			}

			if delay > maxDelay {
				delay = maxDelay
			}

			time.Sleep(delay)
			delay *= expFactor

			continue
		}

		delay = minDelay

		c.mtx.Lock()
		c.ch = ch
		c.mtx.Unlock()

		for d := range sub {
			c.broker.wg.Add(1)
			c.fn(d)
			c.broker.wg.Done()
		}
	}
}
