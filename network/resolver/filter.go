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
)

func isMatch(route route.Route, address string) bool {
	match := func(a, b string) bool {
		if a == "*" || b == "*" || a == b {
			return true
		}
		return false
	}
	type matcher struct {
		a string
		b string
	}
	values := []matcher{
		{address, route.Address},
	}
	for _, v := range values {
		if !match(v.a, v.b) {
			return false
		}
	}
	return true
}

func Filter(routes []route.Route, opts LookupOptions) []route.Route {
	address := opts.Address
	routeMap := make(map[string][]route.Route, 0)

	for _, r := range routes {
		if isMatch(r, address) {
			routeKey := r.Service
			routeMap[routeKey] = append(routeMap[routeKey], r)
		}
	}

	var results []route.Route

	for _, r := range routeMap {
		results = append(results, r...)
	}

	return results
}
