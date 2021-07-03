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
	"fmt"
	"github.com/lastbackend/engine/cmd"
	"strings"
)

type rpcClient struct {
	client *client

	prefix string
	opts   Options
}

func newRpc(prefix string) *rpcClient {
	s := new(rpcClient)
	if len(prefix) == 0 {
		s.prefix = defaultPrefix
	}
	s.prefix = prefix
	s.opts = defaultOptions()
	s.client = newClient(prefix)
	return s
}

func (s *rpcClient) Name() string {
	return PluginName
}

func (s *rpcClient) Flags() []cmd.Flag {
	return []cmd.Flag{
		&cmd.StringSliceFlag{
			Name:        s.withPrefix("endpoints"),
			EnvVar:      s.withEnvPrefix("ENDPOINTS"),
			Usage:       "Set gRPC addresses (Ex: <service>:<endpoint>)",
			Required:    true,
			Destination: &s.opts.Addresses,
		},
		&cmd.StringFlag{
			Name:        s.withPrefix("resolver"),
			EnvVar:      s.withEnvPrefix("RESOLVER"),
			Usage:       "Set resolver service (default: local)",
			Required:    false,
			Value:       defaultResolverService,
			Destination: &s.opts.ResolverService,
		},
		&cmd.IntFlag{
			Name:        s.withPrefix("pool-size"),
			EnvVar:      s.withEnvPrefix("POOL_SIZE"),
			Usage:       "Set pool size",
			Required:    false,
			Value:       defaultPoolSize,
			Destination: s.opts.Pool.Size,
		},
		&cmd.DurationFlag{
			Name:        s.withPrefix("pool-ttl"),
			EnvVar:      s.withEnvPrefix("POOL_TTL"),
			Usage:       "Set pool ttl",
			Required:    false,
			Value:       defaultPoolTTL,
			Destination: s.opts.Pool.Ttl,
		},
		&cmd.IntFlag{
			Name:        s.withPrefix("max-recv-msg-size"),
			EnvVar:      s.withEnvPrefix("MAX-RECV-MSG-SIZE"),
			Usage:       "Sets the maximum message size in bytes the client can receive (default 16 MB)",
			Required:    false,
			Value:       defaultMaxRecvMsgSize,
			Destination: &s.opts.MaxRecvMsgSize,
		},
		&cmd.IntFlag{
			Name:        s.withPrefix("max-send-msg-size"),
			EnvVar:      s.withEnvPrefix("MAX-SEND-MSG-SIZE"),
			Usage:       "Sets the maximum message size in bytes the client can send (default 16 MB)",
			Required:    false,
			Value:       defaultMaxSendMsgSize,
			Destination: &s.opts.MaxSendMsgSize,
		},
	}
}

func (s *rpcClient) Commands() []cmd.Command {
	return make([]cmd.Command, 0)
}

func (s *rpcClient) Start() error {
	return s.client.Init(s.opts)
}

func (s *rpcClient) Stop() error {
	return s.client.Close()
}

func (s *rpcClient) getClient() *client {
	return s.client
}

func (s *rpcClient) withPrefix(name string) string {
	return fmt.Sprintf("%s-%s", s.prefix, name)
}

func (s *rpcClient) withEnvPrefix(name string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s", s.prefix, name))
}
