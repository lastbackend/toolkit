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

package grpc

import (
	"github.com/google/uuid"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"

	"crypto/tls"
	"time"
)

const (
	defaultAddress          = ":0"
	defaultName             = "go.engine.server"
	defaultRegisterInterval = time.Second * 30
	defaultRegisterTTL      = time.Second * 90
)

type Options struct {
	Id      string
	Name    string
	Address string

	MaxConnKey     int
	MaxRecvMsgSize int
	MaxSendMsgSize int

	GrpcOptions []grpc.ServerOption
	TLSConfig   *tls.Config

	GRPCWebAddr    string
	GrpcWebOptions []grpcweb.Option

	RegisterInterval time.Duration
	RegisterTTL      time.Duration
}

func patchOptions(opts, patch Options) Options {
	if len(patch.Id) > 0 {
		opts.Id = patch.Id
	}
	if len(patch.Name) > 0 {
		opts.Name = patch.Name
	}
	if len(patch.Address) > 0 {
		opts.Address = patch.Address
	}
	if patch.MaxConnKey > 0 {
		opts.MaxConnKey = patch.MaxConnKey
	}
	if patch.MaxRecvMsgSize > 0 {
		opts.MaxRecvMsgSize = patch.MaxRecvMsgSize
	}
	if patch.MaxSendMsgSize > 0 {
		opts.MaxSendMsgSize = patch.MaxSendMsgSize
	}
	if patch.GrpcOptions != nil {
		opts.GrpcOptions = patch.GrpcOptions
	}
	if patch.TLSConfig != nil {
		opts.TLSConfig = patch.TLSConfig
	}
	if len(patch.GRPCWebAddr) > 0 {
		opts.GRPCWebAddr = patch.GRPCWebAddr
	}
	if patch.GrpcWebOptions != nil {
		opts.GrpcWebOptions = patch.GrpcWebOptions
	}
	if patch.RegisterInterval > time.Duration(0) {
		opts.RegisterInterval = patch.RegisterInterval
	}
	if patch.RegisterTTL > time.Duration(0) {
		opts.RegisterTTL = patch.RegisterTTL
	}
	return opts
}

func defaultOptions() Options {
	return Options{
		Id:               uuid.New().String(),
		Name:             defaultName,
		Address:          defaultAddress,
		RegisterInterval: defaultRegisterInterval,
		RegisterTTL:      defaultRegisterTTL,
	}
}
