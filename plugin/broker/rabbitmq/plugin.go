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

package rabbitmq

import (
	"fmt"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"gitlab.com/lastbackend/engine/cmd"
	"gitlab.com/lastbackend/engine/plugin"
	"gitlab.com/lastbackend/engine/plugin/broker"
	"strings"
)

const (
	PluginName    = "rabbitmq"
	defaultPrefix = "rbmq"
)

type Plugin plugin.BrokerPlugin

type Options struct {
}

type rabbitMQPlugin struct {
	client *client

	prefix string
	opts   Options
}

func New() plugin.BrokerPlugin {
	s := new(rabbitMQPlugin)
	s.prefix = defaultPrefix
	s.client = newClient()
	return s
}

func (s *rabbitMQPlugin) Name() string {
	return PluginName
}

func (s *rabbitMQPlugin) Type() plugin.PluginType {
	return plugin.PluginBroker
}

func (s *rabbitMQPlugin) SetPrefix(prefix string) {
	s.prefix = prefix
}

func (s *rabbitMQPlugin) Flags() []cmd.Flag {
	return []cmd.Flag{}
}

func (s *rabbitMQPlugin) Commands() []cmd.Command {
	return []cmd.Command{}
}

func (s *rabbitMQPlugin) Client() broker.Client {
	return s.client
}

func (s *rabbitMQPlugin) Start() error {
	return nil
}

func (s *rabbitMQPlugin) Stop() error {
	return s.client.Close()
}

func (s *rabbitMQPlugin) withPrefix(name string) string {
	return fmt.Sprintf("%s-%s", s.prefix, name)
}

func (s *rabbitMQPlugin) withEnvPrefix(name string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s", s.prefix, name))
}
