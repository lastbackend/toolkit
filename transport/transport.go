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

package transport

import (
	"github.com/lastbackend/engine/service/server/codec"

	"context"
	"crypto/tls"
	"time"
)

type TransportOption func(*TransportOptions)
type DialOption func(*DialOptions)
type ListenOption func(*ListenOptions)

type TransportOptions struct {
	Secure    bool
	Addrs     []string
	Codec     codec.Marshaler
	TLSConfig *tls.Config
	Timeout   time.Duration
	Context   context.Context
}

type DialOptions struct {
	Stream  bool
	Timeout time.Duration
	Context context.Context
}

type ListenOptions struct {
	Context context.Context
}

type Transport interface {
	Init(...TransportOption) error
	Dial(addr string, opts ...DialOption) (Client, error)
	Listen(addr string, opts ...ListenOption) (Listener, error)
	String() string
}

type Message struct {
	Header map[string]string
	Body   []byte
}

type Socket interface {
	Recv(*Message) error
	Send(*Message) error
	Close() error
	Local() string
	Remote() string
}

type Client interface {
	Socket
}

type Listener interface {
	Addr() string
	Close() error
	Accept(func(Socket)) error
}
