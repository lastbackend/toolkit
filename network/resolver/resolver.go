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

package resolver

import (
	"github.com/lastbackend/engine/network/resolver/route"
	"github.com/pkg/errors"
)

const (
	LocalResolver = "local"
)

var (
	ErrResolverNotDetected = errors.New("resolver not detected")
	ErrNotAvailable        = errors.New("not available")
)

type Option func(*Options)

type Options struct {
	Cache bool
}

type LookupOption func(*LookupOptions)

type Resolver interface {
	Table() Table
	Lookup(service string, opts ...LookupOption) (route.RouteList, error)
}

type Table interface {
	Create(route.Route) error
	Delete(route.Route) error
	Update(route.Route) error
	Find(service string) ([]route.Route, error)
}

type LookupOptions struct {
	Address string
}

func NewLookup(opts ...LookupOption) LookupOptions {
	qopts := LookupOptions{
		Address: "*",
	}
	for _, o := range opts {
		o(&qopts)
	}
	return qopts
}
