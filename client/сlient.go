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
	"context"
	"github.com/lastbackend/engine/codec"
)

type Client interface {
	Init(...Option) error
	Options() Options
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

// Option used by the Client
type Option func(*Options)

// CallOption used by Call or Stream
type CallOption func(*CallOptions)

// PublishOption used by Publish
type PublishOption func(*PublishOptions)

// MessageOption used by NewMessage
type MessageOption func(*MessageOptions)

// RequestOption used by NewRequest
type RequestOption func(*RequestOptions)

// Request is the interface for a synchronous request used by Call or Stream
type Request interface {
	// The service to call
	Service() string
	// The action to take
	Method() string
	// The endpoint to invoke
	Endpoint() string
	// The content type
	ContentType() string
	// The unencoded request body
	Body() interface{}
	// Write to the encoded request writer. This is nil before a call is made
	Codec() codec.Writer
	// indicates whether the request will be a streaming one rather than unary
	Stream() bool
}

// Response is the response received from a service
type Response interface {
	// Read the response
	Codec() codec.Reader
	// read the header
	Header() map[string]string
	// Read the undecoded response
	Read() ([]byte, error)
}

// Stream is the inteface for a bidirectional synchronous stream
type Stream interface {
	// Context for the stream
	Context() context.Context
	// The request made
	Request() Request
	// The response read
	Response() Response
	// Send will encode and send a request
	Send(interface{}) error
	// Recv will decode and read a response
	Recv(interface{}) error
	// Error returns the stream error
	Error() error
	// Close closes the stream
	Close() error
}