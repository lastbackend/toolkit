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
	rt "github.com/lastbackend/engine/network/resolver/route"

	"sync"
	"time"
)

type table struct {
	sync.RWMutex
	x int
	routes map[string]map[string]*resolverRoute
}

type resolverRoute struct {
	route   rt.Route
	updated time.Time
}

func newTable() *table {
	return &table{
		routes: make(map[string]map[string]*resolverRoute, 0),
	}
}

func (t *table) Find(service string) ([]rt.Route, error) {
	var routes []rt.Route

	t.RLock()
	defer t.RUnlock()

	if len(service) > 0 {
		routeMap, ok := t.routes[service]
		if !ok {
			return nil, rt.ErrRouteNotFound
		}
		for _, rm := range routeMap {
			routes = append(routes, rm.route)
		}
		return routes, nil
	}
	for _, serviceRoutes := range t.routes {
		for _, sr := range serviceRoutes {
			routes = append(routes, sr.route)
		}
	}
	return routes, nil
}

func (t *table) Create(r rt.Route) error {
	service := r.Service
	sum := r.Hash()
	t.x = 100
	t.Lock()
	defer t.Unlock()

	if _, ok := t.routes[service]; !ok {
		t.routes[service] = make(map[string]*resolverRoute)
	}

	if _, ok := t.routes[service][sum]; ok {
		return rt.ErrDuplicateRoute
	}

	t.routes[service][sum] = &resolverRoute{r, time.Now()}

	return nil
}

func (t *table) Delete(r rt.Route) error {

	service := r.Service
	sum := r.Hash()

	t.Lock()
	defer t.Unlock()

	if _, ok := t.routes[service]; !ok {
		return rt.ErrRouteNotFound
	}

	if _, ok := t.routes[service][sum]; !ok {
		return rt.ErrRouteNotFound
	}

	delete(t.routes[service], sum)

	if len(t.routes[service]) == 0 {
		delete(t.routes, service)
	}
	return nil
}

func (t *table) Update(r rt.Route) error {

	service := r.Service
	sum := r.Hash()

	t.Lock()
	defer t.Unlock()

	if _, ok := t.routes[service]; !ok {
		t.routes[service] = make(map[string]*resolverRoute)
	}

	if _, ok := t.routes[service][sum]; !ok {
		t.routes[service][sum] = &resolverRoute{r, time.Now()}
		return nil
	}
	t.routes[service][sum] = &resolverRoute{r, time.Now()}
	return nil
}


