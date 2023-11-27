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

package selector

import (
	"math/rand"
)

func newRandomSelector() Selector {
	return new(random)
}

type random struct{}

func (s *random) Select(routes []string) (Next, error) {
	if len(routes) == 0 {
		return nil, ErrNotAvailable
	}
	return func() string {
		if len(routes) == 1 {
			return routes[0]
		}
		return routes[rand.Intn(len(routes))]
	}, nil
}
