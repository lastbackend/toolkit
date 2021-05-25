/*
Copyright [2014] - [2021] The Last.Backend authors.

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

package client

import (
	"github.com/lastbackend/engine/service/codec"

	"context"
)

var (
	DefaultClient Client
)

type CallOption func(*CallOptions)
type PublishOption func(*PublishOptions)
type MessageOption func(*MessageOptions)
type RequestOption func(*RequestOptions)

type Client interface {
	NewMessage(topic string, msg interface{}, opts ...MessageOption) Message
	NewRequest(service, endpoint string, req interface{}, reqOpts ...RequestOption) Request
	Call(ctx context.Context, req Request, rsp interface{}, opts ...CallOption) error
	Stream(ctx context.Context, req Request, opts ...CallOption) (Stream, error)
	Publish(ctx context.Context, msg Message, opts ...PublishOption) error
	String() string
}

type Message interface {
	Topic() string
	Payload() interface{}
	ContentType() string
}

type Request interface {
	Service() string
	Method() string
	Endpoint() string
	ContentType() string
	Body() interface{}
	Codec() codec.Writer
	Stream() bool
}

type Response interface {
	Codec() codec.Reader
	Header() map[string]string
	Read() ([]byte, error)
}

type Stream interface {
	Context() context.Context
	Request() Request
	Response() Response
	Send(interface{}) error
	Recv(interface{}) error
	Error() error
	Close() error
}
