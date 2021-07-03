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
	"github.com/pkg/errors"
)

type Type int

const (
	Random Type = iota
	RoundRobin
)

var (
	ErrSelectorNotDetected = errors.New("selector not detected")
	ErrNotAvailable        = errors.New("not available")
)

type Selector interface {
	Select([]string) (Next, error)
}

type Next func() string

func New(t Type) (selector Selector, err error) {
	switch t {
	case Random:
		selector = newRandomSelector()
	case RoundRobin:
		selector = newRRSelector()
	default:
		err = ErrSelectorNotDetected
	}
	return selector, err
}
