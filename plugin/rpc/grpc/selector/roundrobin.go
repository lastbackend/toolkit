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

package selector

import (
	"math/rand"
)

func newRRSelector() Selector {
	return new(roundRobin)
}

type roundRobin struct{}

func (s *roundRobin) Select(routes []string) (Next, error) {
	if len(routes) == 0 {
		return nil, ErrNotAvailable
	}
	i := rand.Intn(len(routes))
	return func() string {
		route := routes[i%len(routes)]
		i++
		return route
	}, nil
}
