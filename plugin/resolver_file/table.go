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

package resolver_file

import (
	"encoding/json"
	rt "github.com/lastbackend/toolkit/pkg/client/grpc/resolver/route"
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
	"os"
	"sync"
	"time"
)

type table struct {
	sync.RWMutex
	file    string
	log     logger.Logger
	x       int
	updated time.Time                    `json:"Updated"`
	Routes  map[string]map[string]*Route `json:"routes"`
}

type Route struct {
	Route   rt.Route  `json:"route"`
	Updated time.Time `json:"updated"`
}

func newTable(file string, log logger.Logger) (*table, error) {

	log.Info("create new resolver table")

	t := &table{
		file:   file,
		log:    log,
		Routes: make(map[string]map[string]*Route, 0),
	}

	if err := t.load(); err != nil {
		return t, err
	}

	go func() {
		for {
			t.watchFile()
			t.load()
		}
	}()

	go func() {
		for {
			t.save()
			time.Sleep(1 * time.Second)
		}
	}()

	return t, nil
}

func (t *table) load() error {

	if _, err := os.Stat(t.file); err != nil {
		if os.IsNotExist(err) {
			t.log.Infof("can not load file: %s, create a new one", err.Error())
			f, err := os.Create(t.file)
			if err != nil {
				t.log.Errorf("can not create file: %s", err.Error())
				return err
			}
			f.Close()
			return nil
		}

	}

	file, err := os.ReadFile(t.file)
	if err != nil {
		t.log.Errorf("can not load table: %s", err.Error())
		return err
	}

	err = json.Unmarshal([]byte(file), t.Routes)
	if err != nil {
		t.log.Errorf("can not decode table from file: %s: %s > %s", t.file, string(file), err.Error())
		return err
	}

	return nil
}

func (t *table) save() error {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		t.log.Errorf("can not create table: %s", err.Error())
		return err
	}

	if err = os.WriteFile(t.file, data, 0644); err != nil {
		t.log.Errorf("can not create table file: %s", err.Error())
		return err
	}

	return nil
}

func (t *table) watchFile() error {

	initialStat, err := os.Stat(t.file)
	if err != nil {
		if os.IsNotExist(err) {
			time.Sleep(1 * time.Second)
			return nil
		}
	}

	for {
		stat, err := os.Stat(t.file)
		if err != nil {
			return err
		}

		if stat.Size() != initialStat.Size() || stat.ModTime() != initialStat.ModTime() {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}

func (t *table) Find(service string) ([]rt.Route, error) {
	var routes []rt.Route

	t.RLock()
	defer t.RUnlock()

	if len(service) > 0 {
		routeMap, ok := t.Routes[service]
		if !ok {
			return nil, rt.ErrRouteNotFound
		}
		for _, rm := range routeMap {
			routes = append(routes, rm.Route)
		}
		return routes, nil
	}
	for _, serviceRoutes := range t.Routes {
		for _, sr := range serviceRoutes {
			routes = append(routes, sr.Route)
		}
	}
	return routes, nil
}

func (t *table) Create(r rt.Route) error {

	t.log.Infof("create Route record in table: %s with addr: %s", r.Service, r.Address)

	service := r.Service
	sum := r.Hash()
	t.x = 100
	t.Lock()
	defer t.Unlock()
	defer t.save()

	if _, ok := t.Routes[service]; !ok {
		t.Routes[service] = make(map[string]*Route)
	}

	t.Routes[service][sum] = &Route{r, time.Now()}
	t.updated = time.Now()

	return nil
}

func (t *table) Delete(r rt.Route) error {

	service := r.Service
	sum := r.Hash()

	t.Lock()
	defer t.Unlock()

	if _, ok := t.Routes[service]; !ok {
		return rt.ErrRouteNotFound
	}

	if _, ok := t.Routes[service][sum]; !ok {
		return rt.ErrRouteNotFound
	}

	delete(t.Routes[service], sum)

	if len(t.Routes[service]) == 0 {
		delete(t.Routes, service)
	}
	return nil
}

func (t *table) Update(r rt.Route) error {

	service := r.Service
	sum := r.Hash()

	t.Lock()
	defer t.Unlock()

	if _, ok := t.Routes[service]; !ok {
		t.Routes[service] = make(map[string]*Route)
	}

	if _, ok := t.Routes[service][sum]; !ok {
		t.Routes[service][sum] = &Route{r, time.Now()}
		return nil
	}
	t.Routes[service][sum] = &Route{r, time.Now()}
	return nil
}
