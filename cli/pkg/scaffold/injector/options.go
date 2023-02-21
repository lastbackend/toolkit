/*
Copyright [2014] - [2023] The Last.Backend authors.

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

package injector

type Options struct {
	Workdir string
	Methods []MethodOption
	Plugins []PluginOption
}

type Option func(o *Options)

func Workdir(d string) Option {
	return func(o *Options) {
		o.Workdir = d
	}
}

type MethodOption struct {
	Name           string
	ExposeHTTP     *ExposeHTTP
	ExposeWS       *ExposeWS
	RPCProxy       *RpcProxy
	SubscriptionWS bool
}

func Method(m MethodOption) Option {
	return func(o *Options) {
		if o.Methods == nil {
			o.Methods = make([]MethodOption, 0)
		}
		o.Methods = append(o.Methods, m)
	}
}

type PluginOption struct {
	Name   string
	Prefix string
}

func Plugin(p PluginOption) Option {
	return func(o *Options) {
		if o.Plugins == nil {
			o.Plugins = make([]PluginOption, 0)
		}
		o.Plugins = append(o.Plugins, p)
	}
}

type ImportOption struct {
	Path string
}

type MessageOption struct {
	Name string
}
