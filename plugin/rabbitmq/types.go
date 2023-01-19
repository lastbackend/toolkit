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

import "context"

type SubscriberHandler func(ctx context.Context, event string, payload []byte)

type Event interface {
	Message() *Message
	Name() string
	Ack() error
	Error() error
}

type Message struct {
	Header map[string]string
	Body   []byte
}

type PublishOptions struct {
	Headers map[string]interface{}
}

type SubscribeOptions struct {
	DurableQueue   bool
	RequeueOnError bool
	Headers        map[string]interface{}
}

type Subscriber interface {
	Unsubscribe() error
}
