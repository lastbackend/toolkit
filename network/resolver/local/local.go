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

package local

import (
	"fmt"
	"github.com/lastbackend/engine/network/resolver"
	"github.com/lastbackend/engine/network/resolver/route"
)

type localResolver struct {
	table   *table
	options resolver.Options
}

func NewResolver(opts ...resolver.Option) resolver.Resolver {
	options := resolver.DefaultOptions()
	for _, o := range opts {
		o(&options)
	}
	r := &localResolver{
		options: options,
		table:   newTable(),
	}
	return r
}

func (c *localResolver) Lookup(service string, opts ...resolver.LookupOption) (route.RouteList, error) {
	q := resolver.NewLookup(opts...)
	fmt.Println("11 >>>>", service)
	routes, err := c.table.Find(service)
	if err != nil {
		return nil, err
	}
	fmt.Println("22 >>>>", routes)
	routes = resolver.Filter(routes, q)
	if len(routes) == 0 {
		return nil, route.ErrRouteNotFound
	}
	return routes, nil
}

func (c *localResolver) Table() resolver.Table {
	return c.table
}