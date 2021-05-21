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
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"gitlab.com/lastbackend/engine/cmd"
	"gitlab.com/lastbackend/engine/plugin"
	"gitlab.com/lastbackend/engine/plugin/server"
	"strings"
)

const (
	PluginName    = "grpc"
	defaultPrefix = "grpc"
)

type Plugin plugin.ServerPlugin

type Options struct {
}

type grpcPlugin struct {
	client *client

	prefix string
	opts   Options
}

func New() plugin.ServerPlugin {
	s := new(grpcPlugin)
	s.prefix = defaultPrefix
	s.client = newClient()
	return s
}

func (s *grpcPlugin) Name() string {
	return PluginName
}

func (s *grpcPlugin) Type() plugin.PluginType {
	return plugin.PluginServer
}

func (s *grpcPlugin) SetPrefix(prefix string) {
	s.prefix = prefix
}

func (s *grpcPlugin) Flags() []cmd.Flag {
	return []cmd.Flag{}
}

func (s *grpcPlugin) Commands() []cmd.Command {
	return []cmd.Command{}
}

func (s *grpcPlugin) Client() server.Client {
	return s.client
}

func (s *grpcPlugin) Start() error {
	return nil
}

func (s *grpcPlugin) Stop() error {
	return s.client.Close()
}

func (s *grpcPlugin) withPrefix(name string) string {
	return fmt.Sprintf("%s-%s", s.prefix, name)
}

func (s *grpcPlugin) withEnvPrefix(name string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s", s.prefix, name))
}
