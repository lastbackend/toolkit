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

package grpc

import (
	"github.com/google/uuid"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"

	"crypto/tls"
	"time"
)

const (
	defaultPort             = 9000
	defaultName             = "go.toolkit.server"
	defaultRegisterInterval = time.Second * 30
	defaultRegisterTTL      = time.Second * 90
)

type Config struct {
	ID   string
	Name string `env:"GRPC_SERVER_NAME" envDefault:"" comment:"Set GRPC server name"`

	Host string `env:"GRPC_SERVER_LISTEN" envDefault:"0.0.0.0" comment:"Set GRPC server listen host"`
	Port int    `env:"GRPC_SERVER_PORT" envDefault:"9000" comment:"Set GRPC server listen port"`

	MaxConnSize    int `env:"GRPC_SERVER_MAX_CONNECTION_SIZE"   comment:"Sets the max simultaneous connections for server (default unlimited)"`
	MaxRecvMsgSize int `env:"GRPC_SERVER_MAX_RECEIVE_MESSAGE_SIZE" envDefault:"16777216" comment:"Sets the max message size in bytes the server can receive (default 16 MB)"`
	MaxSendMsgSize int `env:"GRPC_SERVER_MAX_SEND_MESSAGE_SIZE" envDefault:"16777216" comment:"Sets the max message size in bytes the server can send (default 16 MB)"`

	IsDisable bool `env:"GRPC_SERVER_DISABLED" envDefault:"false" comment:"GRPC server disable (default: false)"`

	GrpcOptions []grpc.ServerOption `env:"GRPC_SERVER_OPTIONS" envSeparator:"," comment:"Set GRPC server additional options (key=value,key2=value2)"`
	TLSConfig   *tls.Config

	GRPCWebHost string `env:"GRPC_WEB_SERVER_LISTEN" envDefault:"0.0.0.0" comment:"Set GRPC WEB server listen host"`
	GRPCWebPort int    `env:"GRPC_WEB_SERVER_PORT" comment:"Set GRPC WEB server listen host"`

	GrpcWebOptions []grpcweb.Option

	RegisterInterval time.Duration
	RegisterTTL      time.Duration
}

func defaultOptions() Config {
	return Config{
		ID:               uuid.New().String(),
		Name:             defaultName,
		Port:             defaultPort,
		RegisterInterval: defaultRegisterInterval,
		RegisterTTL:      defaultRegisterTTL,
	}
}
