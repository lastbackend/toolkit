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

package consul

import (
	"github.com/hashicorp/consul/api"
	"github.com/lastbackend/engine/logger"
	"github.com/lastbackend/engine/network/resolver"
	"github.com/lastbackend/engine/network/resolver/route"
	"time"

	"errors"
	"fmt"
	"regexp"
)

const (
	defaultPort = "8500"
)

var (
	errMissingAddr   = errors.New("consul resolver: missing address")
	errAddrMisMatch  = errors.New("consul resolver: invalied uri")
	errEndsWithColon = errors.New("consul resolver: missing port after port-separator colon")
)

var (
	regexConsul, _ = regexp.Compile("^([A-z0-9.]+)(:[0-9]{1,5})?/([A-z_]+)$")
)

type Resolver struct {
	address string
	cache   bool

	table   *table
	options resolver.Options

	watchers map[string]uint64
}

type Options struct {
	Address string
}

func NewResolver(opts ...resolver.Option) resolver.Resolver {
	options := resolver.DefaultOptions()
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

func (c *Resolver) Lookup(service string, opts ...resolver.LookupOption) (route.RouteList, error) {

	routes := make(route.RouteList, 0)

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
			if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
				logger.Errorf("update state error: %v", err)
			}
		}
	}

	go c.watcher(client, service)

	return routes, nil
}

func (c *Resolver) Table() resolver.Table {

	return c.table
}

func (c *Resolver) watcher(client *api.Client, name string) {

	if logger.V(logger.DebugLevel, logger.DefaultLogger) {
		logger.Debugf("calling consul watcher")
	}

	for {

		routes, err := c.getRoutes(client, name)
		if err != nil {
			if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
				logger.Errorf("error retrieving instances from Consul: %v", err)
			}

			<-time.After(10 * time.Second)
		}

		for _, r := range routes {
			if err := c.table.Update(route.Route{Service: r.Service, Address: r.Address}); err != nil {
				if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
					logger.Errorf("update state error: %v", err)
				}
			}
		}

		<-time.After(5 * time.Second)
	}

}

func (c *Resolver) getRoutes(client *api.Client, service string) (route.RouteList, error) {

	routes := make(route.RouteList, 0)

	services, _, err := client.Health().Service(service, "", true, &api.QueryOptions{})
	if err != nil {
		if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
			logger.Errorf("error retrieving instances from Consul: %v", err)
		}
		return nil, err
	}

	for _, service := range services {
		addr := fmt.Sprintf("%v:%v", service.Service.Address, service.Service.Port)
		if logger.V(logger.DebugLevel, logger.DefaultLogger) {
			logger.Debugf("adding service addrs: %v", addr)
		}
		routes = append(routes, route.Route{Service: service.Service.Service, Address: addr})
	}

	return routes, nil
}

func (c *Resolver) Scheme() string {
	return "consul"
}
