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

package generator

type Options struct {
	Service string
	Vendor  string
	Workdir string

	RedisPlugin        bool
	RabbitMQPlugin     bool
	PostgresPGPlugin   bool
	PostgresGORMPlugin bool
	PostgresPlugin     bool
	CentrifugePlugin   bool
	annotationImport   bool
}

func (o Options) AnnotationImport() bool {
	return o.annotationImport
}

type Option func(o *Options)

func Service(s string) Option {
	return func(o *Options) {
		o.Service = s
	}
}
func Vendor(s string) Option {
	return func(o *Options) {
		o.Vendor = s
	}
}

func Workdir(d string) Option {
	return func(o *Options) {
		o.Workdir = d
	}
}

func RedisPlugin(b bool) Option {
	return func(o *Options) {
		if b {
			o.annotationImport = true
		}
		o.RedisPlugin = b
	}
}

func RabbitMQPlugin(b bool) Option {
	return func(o *Options) {
		if b {
			o.annotationImport = true
		}
		o.RabbitMQPlugin = b
	}
}

func PostgresPGPlugin(b bool) Option {
	return func(o *Options) {
		if b {
			o.annotationImport = true
		}
		o.PostgresPGPlugin = b
	}
}

func PostgresGORMPlugin(b bool) Option {
	return func(o *Options) {
		if b {
			o.annotationImport = true
		}
		o.PostgresGORMPlugin = b
	}
}

func PostgresPlugin(b bool) Option {
	return func(o *Options) {
		if b {
			o.annotationImport = true
		}
		o.PostgresPlugin = b
	}
}

func CentrifugePlugin(b bool) Option {
	return func(o *Options) {
		if b {
			o.annotationImport = true
		}
		o.CentrifugePlugin = b
	}
}
