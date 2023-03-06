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

package file

import (
	"context"
	"fmt"
	"github.com/lastbackend/toolkit/pkg/client/grpc/resolver"
	"github.com/lastbackend/toolkit/pkg/client/grpc/resolver/route"
	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/util/addr"
	"os"
	"path"
)

const (
	prefix          string = "resolver"
	defaultFileName string = ".lb.resolver"
)

type Plugin interface {
	Print()
}

type Resolver struct {
	prefix  string
	runtime runtime.Runtime
	table   *table
	opts    Config
}

type Config struct {
	Cache bool
	File  string `env:"FILEPATH" comment:"Filepath to resolver storage file. default(.lb.resolver in home dir)"`
}

type Options struct {
	Name string
}

func NewResolver(runtime runtime.Runtime) resolver.Resolver {

	var err error

	r := &Resolver{
		prefix:  prefix,
		runtime: runtime,
	}

	if err := runtime.Config().Parse(&r.opts, prefix); err != nil {
		runtime.Log().Errorf("Can not parse config %s: %s", prefix, err.Error())
	}

	if r.opts.File == "" {
		dirname, err := os.UserHomeDir()
		if err != nil {
			runtime.Log().Errorf("Can not obtain default home dir: %s", prefix, err.Error())
			return nil
		}
		r.opts.File = path.Join(dirname, defaultFileName)
	}

	if r.table, err = newTable(r.opts.File, runtime.Log()); err != nil {
		return nil
	}

	return r
}

func (c *Resolver) OnStart(ctx context.Context) error {
	c.runtime.Log().Info("resolver file on-start call")

	ip, err := addr.DetectIP()
	if err != nil {
		c.runtime.Log().Errorf("Can not detect local ip: %s", err.Error())
		return err
	}

	srvs := c.runtime.Server().GRPCList()
	for name, srv := range srvs {
		c.table.Create(route.Route{
			Service: name,
			Address: fmt.Sprintf("%s:%d", ip, srv.Info().Port),
		})
	}
	return nil
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

func (c *Resolver) Print() {
	return
}
