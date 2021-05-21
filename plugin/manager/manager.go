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

package manager

import (
	"github.com/pkg/errors"
	"gitlab.com/lastbackend/engine/cmd"
	"gitlab.com/lastbackend/engine/plugin"
	"gitlab.com/lastbackend/engine/plugin/storage/postgres"

	"fmt"
	"reflect"
)

type manager struct {
	plugins map[string]plugin.Plugin
}

func NewManager() plugin.Manager {
	mng := new(manager)
	mng.plugins = make(map[string]plugin.Plugin, 0)
	return mng
}

func (pm *manager) Register(in interface{}) error {
	val := reflect.ValueOf(in).Elem()

	for i := 0; i < val.NumField(); i++ {

		valueField := val.Field(i)
		typeField := val.Type().Field(i)

		pluginName := typeField.Tag.Get("plugin")
		prefixName := typeField.Tag.Get("prefix")

		pluginType, ok := plugins[pluginName]
		if !ok {
			return fmt.Errorf("plugin %s not registered", pluginName)
		}

		switch pluginType {
		case reflect.TypeOf((*postgres.Plugin)(nil)).Elem():
			p := postgres.New()
			p.SetPrefix(prefixName)

			name := fmt.Sprintf("%s-%s", prefixName, p.Name())

			pl, exists := pm.plugins[name]
			if exists {
				p = pl.(plugin.StoragePlugin)
			}
			pm.plugins[name] = p
			valueField.Set(reflect.ValueOf(p.Client()))
		default:
			return fmt.Errorf("plugin %s not foud", typeField.Type)
		}
	}

	return nil
}

func (pm *manager) ExtendСLI(cli cmd.CLI) {
	for _, p := range pm.plugins {
		cli.AddFlags(p.Flags()...)
	}

	for _, p := range pm.plugins {
		cli.AddCommands(p.Commands()...)
	}
}

func (pm *manager) Start() error {
	for name, plugin := range pm.plugins {
		if err := plugin.Start(); err != nil {
			return errors.Wrap(err, fmt.Sprintf("Start plugin %s failed", name))
		}
	}
	return nil
}

func (pm *manager) Stop() {
	for name, plugin := range pm.plugins {
		if err := plugin.Stop(); err != nil {
			fmt.Println(errors.Wrap(err, fmt.Sprintf("Stop plugin %s failed", name)))
		}
	}
}
