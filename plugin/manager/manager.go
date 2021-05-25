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
	"github.com/lastbackend/engine/cmd"
	"github.com/lastbackend/engine/plugin"
	"github.com/pkg/errors"

	"fmt"
	"reflect"
	"sync"
)

type manager struct {
	sync.RWMutex

	plugins map[plugin.Plugin]bool
}

func NewManager() Manager {
	mng := new(manager)
	mng.plugins = make(map[plugin.Plugin]bool, 0)
	return mng
}

func (pm *manager) RegisterPlugin(p plugin.Plugin) {
	pm.Lock()
	defer pm.Unlock()
	if _, ok := pm.plugins[p]; !ok {
		pm.plugins[p] = true
	}
}

func (pm *manager) Register(in interface{}) error {

	val := reflect.ValueOf(in).Elem()

	for i := 0; i < val.NumField(); i++ {

		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tagPrefix := val.Type().Field(i).Tag.Get("prefix")

		pm.Lock()
		valueFunc, ok := plugins[typeField.Type]
		if !ok {
			return fmt.Errorf("plugin %s not registered", typeField.Type)
		}
		pm.Unlock()

		funcType := reflect.TypeOf(plugin.RegisterFunc(nil))
		funcRegister := reflect.MakeFunc(funcType, func(args []reflect.Value) []reflect.Value {
			pm.RegisterPlugin(args[0].Interface().(plugin.Plugin))
			return nil
		})

		funcCreator := valueFunc.Call([]reflect.Value{funcRegister})

		valueOptions := reflect.ValueOf(plugin.Option{Prefix: tagPrefix})
		valueInstance := funcCreator[0].Call([]reflect.Value{valueOptions})

		if len(valueInstance) > 0 {
			valueField.Set(valueInstance[0].Elem().Convert(valueField.Type()))
		}
	}

	return nil
}

func (pm *manager) Flags() []cmd.Flag {
	if pm.plugins == nil {
		return make([]cmd.Flag, 0)
	}

	pm.Lock()
	defer pm.Unlock()

	flags := make([]cmd.Flag, 0)

	for p := range pm.plugins {
		flags = append(flags, p.Flags()...)
	}

	return flags
}

func (pm *manager) Commands() []cmd.Command {
	if pm.plugins == nil {
		return make([]cmd.Command, 0)
	}

	pm.Lock()
	defer pm.Unlock()

	commands := make([]cmd.Command, 0)

	for p := range pm.plugins {
		commands = append(commands, p.Commands()...)
	}

	return commands
}

func (pm *manager) Start() error {
	for p := range pm.plugins {
		if err := p.Start(); err != nil {
			return errors.Wrap(err, fmt.Sprintf("Start plugin %s failed", p.Name()))
		}
	}
	return nil
}

func (pm *manager) Stop() {
	for p := range pm.plugins {
		if err := p.Stop(); err != nil {
			fmt.Println(errors.Wrap(err, fmt.Sprintf("Stop plugin %s failed", p.Name())))
		}
	}
}
