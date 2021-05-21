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

package redis

import (
	"fmt"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"gitlab.com/lastbackend/engine/cmd"
	"gitlab.com/lastbackend/engine/plugin"
	"gitlab.com/lastbackend/engine/plugin/cache"
	"strings"
)

const (
	PluginName    = "redis"
	defaultPrefix = "redis"
)

type Plugin plugin.CachePlugin

type Options struct {
}

type redisPlugin struct {
	client *client

	prefix string
	opts   Options
}

func New() plugin.CachePlugin {
	s := new(redisPlugin)
	s.prefix = defaultPrefix
	s.client = newClient()
	return s
}

func (s *redisPlugin) Name() string {
	return PluginName
}

func (s *redisPlugin) Type() plugin.PluginType {
	return plugin.PluginCache
}

func (s *redisPlugin) SetPrefix(prefix string) {
	s.prefix = prefix
}

func (s *redisPlugin) Flags() []cmd.Flag {
	return []cmd.Flag{}
}

func (s *redisPlugin) Commands() []cmd.Command {
	return []cmd.Command{}
}

func (s *redisPlugin) Client() cache.Client {
	return s.client
}

func (s *redisPlugin) Start() error {
	return nil
}

func (s *redisPlugin) Stop() error {
	return s.client.Close()
}

func (s *redisPlugin) withPrefix(name string) string {
	return fmt.Sprintf("%s-%s", s.prefix, name)
}

func (s *redisPlugin) withEnvPrefix(name string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s", s.prefix, name))
}
