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

package consul

import (
	"github.com/hashicorp/consul/api"
	resolver2 "github.com/lastbackend/toolkit/pkg/client/network/resolver"
	"github.com/lastbackend/toolkit/pkg/client/network/resolver/route"
	logger2 "github.com/lastbackend/toolkit/pkg/runtime/logger"
	"time"

	"fmt"
)

type Resolver struct {
	address string

	table   *table
	options resolver2.Options

	watchers map[string]uint64
}

type Options struct {
	Address string
}

func NewResolver(opts ...resolver2.Option) resolver2.Resolver {
	options := resolver2.DefaultOptions()
	for _, o := range opts {
		o(&options)
	}
	r := &Resolver{
		options:  options,
		table:    newTable(),
		watchers: make(map[string]uint64, 0),
	}
	return r
}

func (c *Resolver) Lookup(service string, opts ...resolver2.LookupOption) (route.List, error) {

	routes, err := c.table.Find(service)
	if err != nil && err != route.ErrRouteNotFound {
		return nil, err
	}
	if err == nil && routes != nil {
		return routes, nil
	}

	config := api.DefaultConfig()
	config.Address = c.address

	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	routes, err = c.getRoutes(client, service)
	if err != nil {
		return nil, err
	}

	for _, r := range routes {
		if err := c.table.Update(route.Route{Service: r.Service, Address: r.Address}); err != nil {
			if logger2.V(logger2.ErrorLevel, logger2.DefaultLogger) {
				logger2.Errorf("update state error: %v", err)
			}
		}
	}

	go c.watcher(client, service)

	return routes, nil
}

func (c *Resolver) Table() resolver2.Table {

	return c.table
}

func (c *Resolver) watcher(client *api.Client, name string) {

	if logger2.V(logger2.DebugLevel, logger2.DefaultLogger) {
		logger2.Debugf("calling consul watcher")
	}

	for {

		routes, err := c.getRoutes(client, name)
		if err != nil {
			if logger2.V(logger2.ErrorLevel, logger2.DefaultLogger) {
				logger2.Errorf("error retrieving instances from Consul: %v", err)
			}

			<-time.After(10 * time.Second)
		}

		for _, r := range routes {
			if err := c.table.Update(route.Route{Service: r.Service, Address: r.Address}); err != nil {
				if logger2.V(logger2.ErrorLevel, logger2.DefaultLogger) {
					logger2.Errorf("update state error: %v", err)
				}
			}
		}

		<-time.After(5 * time.Second)
	}

}

func (c *Resolver) getRoutes(client *api.Client, service string) (route.List, error) {

	routes := make(route.List, 0)

	services, _, err := client.Health().Service(service, "", true, &api.QueryOptions{})
	if err != nil {
		if logger2.V(logger2.ErrorLevel, logger2.DefaultLogger) {
			logger2.Errorf("error retrieving instances from Consul: %v", err)
		}
		return nil, err
	}

	for _, service := range services {
		addr := fmt.Sprintf("%v:%v", service.Service.Address, service.Service.Port)
		if logger2.V(logger2.DebugLevel, logger2.DefaultLogger) {
			logger2.Debugf("adding service addrs: %v", addr)
		}
		routes = append(routes, route.Route{Service: service.Service.Service, Address: addr})
	}

	return routes, nil
}

func (c *Resolver) Scheme() string {
	return "consul"
}
