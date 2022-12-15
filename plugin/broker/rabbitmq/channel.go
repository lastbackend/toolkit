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
	"errors"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

type amqpChannel struct {
	uuid       string
	connection *amqp.Connection
	channel    *amqp.Channel
}

func newRabbitChannel(conn *amqp.Connection, prefetchCount int, prefetchGlobal bool) (*amqpChannel, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	rabbitCh := &amqpChannel{
		uuid:       id.String(),
		connection: conn,
	}
	if err := rabbitCh.Connect(prefetchCount, prefetchGlobal); err != nil {
		return nil, err
	}
	return rabbitCh, nil
}

func (a *amqpChannel) Connect(prefetchCount int, prefetchGlobal bool) (err error) {
	a.channel, err = a.connection.Channel()
	if err != nil {
		return err
	}
	err = a.channel.Qos(prefetchCount, 0, prefetchGlobal)
	if err != nil {
		return err
	}
	return nil
}

func (a *amqpChannel) Close() error {
	if a.channel == nil {
		return errors.New("channel is nil")
	}
	return a.channel.Close()
}

func (a *amqpChannel) Publish(exchange, key string, message amqp.Publishing) error {
	if a.channel == nil {
		return errors.New("channel is nil")
	}
	return a.channel.Publish(exchange, key, false, false, message)
}

func (a *amqpChannel) DeclareExchange(exchange, kind string) error {
	return a.channel.ExchangeDeclare(
		exchange, // name
		kind,     // kind
		false,    // durable
		false,    // autoDelete
		false,    // internal
		false,    // noWait
		nil,      // args
	)
}

func (a *amqpChannel) DeclareDurableExchange(exchange, kind string) error {
	return a.channel.ExchangeDeclare(
		exchange, // name
		kind,     // kind
		true,     // durable
		false,    // autoDelete
		false,    // internal
		false,    // noWait
		nil,      // args
	)
}

func (a *amqpChannel) DeclareQueue(queue string, args amqp.Table) error {
	_, err := a.channel.QueueDeclare(
		queue, // name
		false, // durable
		true,  // autoDelete
		false, // exclusive
		false, // noWait
		args,  // args
	)
	return err
}

func (a *amqpChannel) DeclareDurableQueue(queue string, args amqp.Table) error {
	_, err := a.channel.QueueDeclare(
		queue, // name
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		args,  // args
	)
	return err
}

func (a *amqpChannel) DeclareReplyQueue(queue string) error {
	_, err := a.channel.QueueDeclare(
		queue, // name
		false, // durable
		true,  // autoDelete
		true,  // exclusive
		false, // noWait
		nil,   // args
	)
	return err
}

func (a *amqpChannel) ConsumeQueue(queue string, autoAck bool) (<-chan amqp.Delivery, error) {
	return a.channel.Consume(
		queue,   // queue
		a.uuid,  // consumer
		autoAck, // autoAck
		false,   // exclusive
		false,   // nolocal
		false,   // nowait
		nil,     // args
	)
}

func (a *amqpChannel) BindQueue(queue, key, exchange string, args amqp.Table) error {
	return a.channel.QueueBind(
		queue,    // name
		key,      // key
		exchange, // exchange
		false,    // noWait
		args,     // args
	)
}
