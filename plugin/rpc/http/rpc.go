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

package http

import (
	"github.com/lastbackend/engine/cmd"

	"fmt"
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
		&cmd.StringFlag{
			Name:        s.withPrefix("endpoint"),
			EnvVar:      s.withEnvPrefix("ENDPOINT"),
			Usage:       "Set HTTP endpoint (Ex: <endpoint>)",
			Required:    true,
			Destination: &s.opts.Endpoint,
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
