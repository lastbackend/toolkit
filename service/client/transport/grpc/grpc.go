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

// Package grpc provides a grpc transport
package grpc

import (
	"github.com/lastbackend/engine/service/client/transport"

	"crypto/tls"
	"net"
)

type grpcTransport struct {
	opts transport.Options
}

type grpcTransportListener struct {
	listener net.Listener
	secure   bool
	tls      *tls.Config
}

func (t *grpcTransportListener) Addr() string {
	return ""
}

func (t *grpcTransportListener) Close() error {
	return nil
}

func (t *grpcTransportListener) Accept(fn func(transport.Socket)) error {
	return nil
}

func (t *grpcTransport) Dial(addr string, opts ...transport.DialOption) (transport.Client, error) {
	return nil, nil
}

func (t *grpcTransport) Listen(addr string, opts ...transport.ListenOption) (transport.Listener, error) {
	return nil, nil
}

func (t *grpcTransport) Init(opts ...transport.Option) error {
	return nil
}

func (t *grpcTransport) Options() transport.Options {
	return transport.Options{}
}

func (t *grpcTransport) String() string {
	return "grpc"
}
