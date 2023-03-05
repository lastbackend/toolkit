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

package local

import (
	"github.com/lastbackend/toolkit/pkg/client/grpc/resolver"
	"github.com/lastbackend/toolkit/pkg/client/grpc/resolver/route"
	"github.com/lastbackend/toolkit/pkg/runtime"
	"strings"
)

const (
	prefix    string = "resolver"
	separator string = "="
)

type Resolver struct {
	runtime runtime.Runtime
	table   *table
	options *Options
}

type Options struct {
	Cache     bool
	Endpoints []string `env:"ENDPOINTS" envSeparator:":"`
}

func NewResolver(runtime runtime.Runtime) resolver.Resolver {

	opts := new(Options)
	if err := runtime.Config().Parse(opts, prefix); err != nil {
		runtime.Log().Errorf("Can not parse config %s: $s", prefix, err.Error())
	}

	table := newTable()
	r := &Resolver{
		table:   table,
		options: opts,
	}

	for _, s := range opts.Endpoints {

		if s != "" {

			parts := strings.Split(s, separator)
			if len(parts) < 2 {
				continue
			}

			table.Create(route.Route{
				Service: parts[0],
				Address: parts[1],
			})
		}
	}

	return r
}

func (c *Resolver) Lookup(service string, opts ...resolver.LookupOption) (route.List, error) {
	q := resolver.NewLookup(opts...)
	routes, err := c.table.Find(service)
	if err != nil {
		return nil, err
	}
	routes = resolver.Filter(routes, q)
	if len(routes) == 0 {
		return nil, route.ErrRouteNotFound
	}
	return routes, nil
}

func (c *Resolver) Table() resolver.Table {
	return c.table
}
