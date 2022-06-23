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

import "github.com/streadway/amqp"

type publisher struct {
	delivery amqp.Delivery
	message  *Message
	topic    string
	err      error
}

func (p *publisher) Ack() error {
	return p.delivery.Ack(false)
}

func (p *publisher) Error() error {
	return p.err
}

func (p *publisher) Name() string {
	return p.topic
}

func (p *publisher) Message() *Message {
	return p.message
}
