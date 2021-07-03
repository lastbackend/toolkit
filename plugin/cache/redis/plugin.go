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

package redis

import (
	"context"
	"github.com/lastbackend/engine/plugin"
	"time"
)

const (
	PluginName    = "redis"
	defaultPrefix = "redis"
)

// Register - registers the plugin implements a cache using Redis as a key/value storage
func Register(f plugin.RegisterFunc) plugin.CreatorFunc {
	return func(o plugin.Option) interface{} {
		p := newCache(o.Prefix)
		f(p)
		return p.getClient()
	}
}

type Cache interface {
	Exists(ctx context.Context, key string) bool
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string, value interface{}) error
	Del(ctx context.Context, key string) error
}
