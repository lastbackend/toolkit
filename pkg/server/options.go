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

package server

import (
	"github.com/google/uuid"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"

	"crypto/tls"
	"time"
)

const (
	defaultAddress          = ":50005"
	defaultName             = "go.toolkit.server"
	defaultRegisterInterval = time.Second * 30
	defaultRegisterTTL      = time.Second * 90
	// DefaultMaxRecvMsgSize maximum message that client can receive (16 MB).
	defaultMaxRecvMsgSize = 1024 * 1024 * 16
	// DefaultMaxSendMsgSize maximum message that client can send (16 MB).
	defaultMaxSendMsgSize = 1024 * 1024 * 16
)

type Options struct {
	ID      string
	Name    string
	Address string

	MaxConnSize    int
	MaxRecvMsgSize int
	MaxSendMsgSize int

	IsDisable bool

	GrpcOptions []grpc.ServerOption
	TLSConfig   *tls.Config

	GRPCWebAddr    string
	GrpcWebOptions []grpcweb.Option

	RegisterInterval time.Duration
	RegisterTTL      time.Duration
}

func defaultOptions() Options {
	return Options{
		ID:               uuid.New().String(),
		Name:             defaultName,
		Address:          defaultAddress,
		RegisterInterval: defaultRegisterInterval,
		RegisterTTL:      defaultRegisterTTL,
	}
}
