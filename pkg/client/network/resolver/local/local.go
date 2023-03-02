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
	resolver2 "github.com/lastbackend/toolkit/pkg/client/network/resolver"
	"github.com/lastbackend/toolkit/pkg/client/network/resolver/route"
)

type Resolver struct {
	table   *table
	options resolver2.Options
}

func NewResolver(opts ...resolver2.Option) resolver2.Resolver {
	options := resolver2.DefaultOptions()
	for _, o := range opts {
		o(&options)
	}
	r := &Resolver{
		options: options,
		table:   newTable(),
	}
	return r
}

func (c *Resolver) Lookup(service string, opts ...resolver2.LookupOption) (route.List, error) {
	q := resolver2.NewLookup(opts...)
	routes, err := c.table.Find(service)
	if err != nil {
		return nil, err
	}
	routes = resolver2.Filter(routes, q)
	if len(routes) == 0 {
		return nil, route.ErrRouteNotFound
	}
	return routes, nil
}

func (c *Resolver) Table() resolver2.Table {
	return c.table
}
