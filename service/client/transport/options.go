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

package transport

import (
	"github.com/lastbackend/engine/service/codec"

	"context"
	"crypto/tls"
	"time"
)

type Options struct {
	Addrs     []string
	Secure    bool
	Codec     codec.Marshaler
	TLSConfig *tls.Config
	Timeout   time.Duration
	Context   context.Context
}

type DialOptions struct {
	Stream  bool
	Timeout time.Duration
	Context context.Context
}

type ListenOptions struct {
	Context context.Context
}
