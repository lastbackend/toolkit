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

package server

import (
	"fmt"
	"github.com/lastbackend/engine/cmd"
)

type rpcServer struct {
}

func New() *rpcServer {
	return new(rpcServer)
}

func (rpcServer) Start() error {
	return nil
}

func (rpcServer) Stop() error {
	return nil
}

func (r rpcServer) Register(i interface{}) error {

	type handler interface {
		register(Server) error
	}

	//val := reflect.ValueOf(i).Elem()
	//
	//for i := 0; i < val.NumField(); i++ {
	//
	//	valueField := val.Field(i)
	//	typeField := val.Type().Field(i)
	//
	//	fmt.Println("v >>>>", valueField.Type(), valueField.Kind())
	//	fmt.Println("t >>>>", typeField.Type, typeField.Name)
	//
	//	//v := i.(handler)
	//	//
	//	//if err := v.register(r); err != nil {
	//	//	return err
	//	//}
	//
	//}

	return nil
}

func (rpcServer) Handle(Handler) error {
	fmt.Println(">>>>>>>>>>>> Handle")
	return nil
}

func (rpcServer) NewHandler(interface{}, ...HandlerOption) Handler {
	fmt.Println(">>>>>>>>>>>> NewHandler")
	return nil
}

func (rpcServer) Flags() []cmd.Flag {
	return nil
}

func (rpcServer) Commands() []cmd.Command {
	return nil
}
