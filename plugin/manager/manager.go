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

	registered map[plugin.Plugin]bool
}

func NewManager() Manager {
	mng := new(manager)
	mng.registered = make(map[plugin.Plugin]bool, 0)
	return mng
}

func (pm *manager) RegisterPlugin(p plugin.Plugin) {
	pm.Lock()
	defer pm.Unlock()
	if _, ok := pm.registered[p]; !ok {
		pm.registered[p] = true
	}
}

func (pm *manager) Register(i interface{}, f func(f plugin.RegisterFunc) plugin.CreatorFunc, o plugin.Option) error {

	valIface := reflect.ValueOf(i).Elem()
	funcType := reflect.TypeOf(plugin.RegisterFunc(nil))
	funcRegister := reflect.MakeFunc(funcType, func(args []reflect.Value) []reflect.Value {
		pm.RegisterPlugin(args[0].Interface().(plugin.Plugin))
		return nil
	})
	funcCreator := reflect.ValueOf(f).Call([]reflect.Value{funcRegister})
	valueOptions := reflect.ValueOf(o)
	valueInstance := funcCreator[0].Call([]reflect.Value{valueOptions})

	fmt.Println("+++ ", valIface.Field(0).Type(), valIface.Field(0).Kind(), valIface.Field(0).Type().PkgPath())

	valIface.Field(0).Set(valueInstance[0].Elem().Convert(valIface.Field(0).Type()))

	//for i := 0; i < valIface.NumField(); i++ {
	//
	//	valueField := valIface.Field(i)
	//	typeField := valIface.Type().Field(i)
	//	tagPrefix := valIface.Type().Field(i).Tag.Get("prefix")
	//
	//	pm.Lock()
	//	valueFunc, ok := registered[typeField.Type]
	//	if !ok {
	//		return fmt.Errorf("plugin %s not registered", typeField.Type)
	//	}
	//	pm.Unlock()
	//
	//	funcType := reflect.TypeOf(plugin.RegisterFunc(nil))
	//	funcRegister := reflect.MakeFunc(funcType, func(args []reflect.Value) []reflect.Value {
	//		pm.RegisterPlugin(args[0].Interface().(plugin.Plugin))
	//		return nil
	//	})
	//
	//	funcCreator := valueFunc.Call([]reflect.Value{funcRegister})
	//
	//	valueOptions := reflect.ValueOf(plugin.Option{Prefix: tagPrefix})
	//	valueInstance := funcCreator[0].Call([]reflect.Value{valueOptions})
	//
	//	if len(valueInstance) > 0 {
	//		valueField.Set(valueInstance[0].Elem().Convert(valueField.Type()))
	//	}
	//}

	return nil
}

func (pm *manager) Flags() []cmd.Flag {
	if pm.registered == nil {
		return make([]cmd.Flag, 0)
	}

	pm.Lock()
	defer pm.Unlock()

	flags := make([]cmd.Flag, 0)

	for p := range pm.registered {
		flags = append(flags, p.Flags()...)
	}

	return flags
}

func (pm *manager) Commands() []cmd.Command {
	if pm.registered == nil {
		return make([]cmd.Command, 0)
	}

	pm.Lock()
	defer pm.Unlock()

	commands := make([]cmd.Command, 0)

	for p := range pm.registered {
		commands = append(commands, p.Commands()...)
	}

	return commands
}

func (pm *manager) Start() error {
	for p := range pm.registered {
		if err := p.Start(); err != nil {
			return errors.Wrap(err, fmt.Sprintf("Start plugin %s failed", p.Name()))
		}
	}
	return nil
}

func (pm *manager) Stop() {
	for p := range pm.registered {
		if err := p.Stop(); err != nil {
			fmt.Println(errors.Wrap(err, fmt.Sprintf("Stop plugin %s failed", p.Name())))
		}
		delete(pm.registered, p)
	}
}
