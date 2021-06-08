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
	"crypto/tls"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"strings"

	"fmt"
	"time"
)

const (
	defaultName     = "server"
	defaultEndpoint = ":6379"
)

type clientOptions struct {
	// Endpoints = host:port,host:port addresses of ring shards.
	Endpoints string

	// Frequency of PING commands sent to check shards availability.
	// Shard is considered down after 3 subsequent failed checks.
	HeartbeatFrequency time.Duration

	// Use the specified Username to authenticate the current connection
	// with one of the connections defined in the ACL list when connecting
	// to a Redis 6.0 instance, or greater, that is using the Redis ACL system.
	Username string
	// Optional password. Must match the password specified in the
	// requirepass server configuration option (if connecting to a Redis 5.0 instance, or lower),
	// or the User Password when connecting to a Redis 6.0 instance, or greater,
	// that is using the Redis ACL system.
	Password string
	// Database to be selected after connecting to the server.
	DB int

	// Maximum number of retries before giving up.
	// Default is 3 retries; -1 (not 0) disables retries.
	MaxRetries int
	// Minimum backoff between each retry.
	// Default is 8 milliseconds; -1 disables backoff.
	MinRetryBackoff time.Duration
	// Maximum backoff between each retry.
	// Default is 512 milliseconds; -1 disables backoff.
	MaxRetryBackoff time.Duration

	// Dial timeout for establishing new connections.
	// Default is 5 seconds.
	DialTimeout time.Duration
	// Timeout for socket reads. If reached, commands will fail
	// with a timeout instead of blocking. Use value -1 for no timeout and 0 for default.
	// Default is 3 seconds.
	ReadTimeout time.Duration
	// Timeout for socket writes. If reached, commands will fail
	// with a timeout instead of blocking.
	// Default is ReadTimeout.
	WriteTimeout time.Duration

	// Maximum number of socket connections.
	// Default is 10 connections per every CPU as reported by runtime.NumCPU.
	PoolSize int
	// Minimum number of idle connections which is useful when establishing
	// new connection is slow.
	MinIdleConns int
	// Connection age at which client retires (closes) the connection.
	// Default is to not close aged connections.
	MaxConnAge time.Duration
	// Amount of time client waits for connection if all connections
	// are busy before returning an error.
	// Default is ReadTimeout + 1 second.
	PoolTimeout time.Duration
	// Amount of time after which client closes idle connections.
	// Should be less than server's timeout.
	// Default is 5 minutes. -1 disables idle timeout check.
	IdleTimeout time.Duration
	// Frequency of idle checks made by idle connections reaper.
	// Default is 1 minute. -1 disables idle connections reaper,
	// but idle connections are still discarded by the client
	// if IdleTimeout is set.
	IdleCheckFrequency time.Duration

	// TODO: need to implement the ability to install tls
	// TLS Config to use. When set TLS will be negotiated.
	TLSConfig *tls.Config
}

type client struct {
	c *cache.Cache
}

func newClient() *client {
	return new(client)
}

func (c *client) open(opts clientOptions) error {
	ring := redis.NewRing(c.prepareOptions(opts))
	c.c = cache.New(&cache.Options{
		Redis:      ring,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})
	return nil
}

func (c *client) prepareOptions(opts clientOptions) *redis.RingOptions {

	addrs := map[string]string{defaultName: defaultEndpoint}
	if len(opts.Endpoints) > 0 {
		addrs = make(map[string]string, 0)

		opts.Endpoints = strings.Replace(opts.Endpoints, " ", "", -1)
		addrsList := strings.Split(opts.Endpoints, ",")
		var index = 1
		for _, v := range addrsList {
			addrs[fmt.Sprintf("%s-%d", defaultName, index)] = v
		}
	}

	return &redis.RingOptions{
		Addrs:              addrs,
		Username:           opts.Username,
		Password:           opts.Password,
		MaxRetries:         opts.MaxRetries,
		MinRetryBackoff:    opts.MinRetryBackoff,
		MaxRetryBackoff:    opts.MaxRetryBackoff,
		DialTimeout:        opts.DialTimeout,
		ReadTimeout:        opts.ReadTimeout,
		WriteTimeout:       opts.WriteTimeout,
		PoolSize:           opts.PoolSize,
		MinIdleConns:       opts.MinIdleConns,
		MaxConnAge:         opts.MaxConnAge,
		PoolTimeout:        opts.PoolTimeout,
		IdleTimeout:        opts.IdleTimeout,
		IdleCheckFrequency: opts.IdleCheckFrequency,
		TLSConfig:          opts.TLSConfig,
	}
}

func (c *client) Exists(ctx context.Context, key string) bool {
	return c.c.Exists(ctx, key)
}

func (c *client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if ttl == time.Duration(0) {
		ttl = 1 * time.Hour
	}
	return c.c.Set(&cache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: value,
		TTL:   ttl,
	})
}

func (c *client) Get(ctx context.Context, key string, value interface{}) error {
	return c.c.Get(ctx, key, value)
}

func (c *client) Del(ctx context.Context, key string) error {
	return c.c.Delete(ctx, key)
}
