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

package route

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/pkg/errors"
)

var (
	ErrRouteNotFound  = errors.New("route not found")
	ErrDuplicateRoute = errors.New("duplicate route")
)

type Route struct {
	Service string
	Address string
}

func (r *Route) Hash() string {
	hash := sha256.Sum256([]byte(r.Service + r.Address))
	return hex.EncodeToString(hash[:])
}

type List []Route

func (r *List) Addresses() []string {
	list := make([]string, 0)
	for _, item := range *r {
		list = append(list, item.Address)
	}
	return list
}
