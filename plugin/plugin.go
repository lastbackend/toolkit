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

package plugin

import (
	"gitlab.com/lastbackend/engine/cmd"
	"gitlab.com/lastbackend/engine/plugin/broker"
	"gitlab.com/lastbackend/engine/plugin/cache"
	"gitlab.com/lastbackend/engine/plugin/server"
	"gitlab.com/lastbackend/engine/plugin/storage"
)

type PluginType int

const (
	PluginStorage PluginType = iota
	PluginBroker
	PluginCache
	PluginServer
)

type Manager interface {
	Register(interface{}) error
	ExtendСLI(cmd.CLI)
	Start() error
	Stop()
}

type Plugin interface {
	Name() string
	Type() PluginType
	Flags() []cmd.Flag
	Commands() []cmd.Command
	SetPrefix(string)
	Start() error
	Stop() error
}

type StoragePlugin interface {
	Plugin
	Client() storage.Client
}

type CachePlugin interface {
	Plugin
	Client() cache.Client
}

type BrokerPlugin interface {
	Plugin
	Client() broker.Client
}

type ServerPlugin interface {
	Plugin
	Client() server.Client
}
