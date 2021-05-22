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
	"github.com/lastbackend/engine/cmd"
)

type Option struct {
	Prefix string
}

type CreatorFunc func(o Option) interface{}
type RegisterFunc func(Plugin)

type Plugin interface {
	Name() string
	Flags() []cmd.Flag
	Commands() []cmd.Command
	Start() error
	Stop() error
}
