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
	"context"
	"gitlab.com/lastbackend/engine/codec"

	"crypto/tls"
)

type Options struct {
	Codecs    map[string]codec.NewCodec
	Metadata  map[string]string
	Name      string
	Address   string
	Advertise string
	Id        string
	Version   string
	TLSConfig *tls.Config
	Context   context.Context
}

func newOptions(opt ...Option) Options {
	opts := Options{
		Codecs:   make(map[string]codec.NewCodec),
		Metadata: map[string]string{},
	}

	for _, o := range opt {
		o(&opts)
	}

	if len(opts.Address) == 0 {
		opts.Address = DefaultAddress
	}

	if len(opts.Name) == 0 {
		opts.Name = DefaultName
	}

	if len(opts.Id) == 0 {
		opts.Id = DefaultId
	}

	if len(opts.Version) == 0 {
		opts.Version = DefaultVersion
	}

	return opts
}

func Name(n string) Option {
	return func(o *Options) {
		o.Name = n
	}
}

func Version(v string) Option {
	return func(o *Options) {
		o.Version = v
	}
}
