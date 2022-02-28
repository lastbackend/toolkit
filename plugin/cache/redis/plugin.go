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
	"crypto/tls"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/lastbackend/engine"
	"time"

	"context"
	"fmt"
	"strings"
)

const (
	defaultPrefix   = "redis"
	defaultName     = "server"
	defaultEndpoint = ":6379"
)

type Plugin interface {
	engine.Plugin

	DB() *cache.Cache
	Register(app engine.Service, opts *Options) error
}

type Options struct {
	Name string
}

type options struct {
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

type plugin struct {
	prefix string
	opts   options

	db *cache.Cache
}

func NewPlugin(app engine.Service, opts *Options) Plugin {
	p := new(plugin)
	err := p.Register(app, opts)
	if err != nil {
		return nil
	}
	return p
}

// Register - registers the plug implements storage using Postgres as a database storage
func (p *plugin) Register(app engine.Service, opts *Options) error {
	p.prefix = opts.Name
	if p.prefix == "" {
		p.prefix = defaultPrefix
	}

	p.addFlags(app)

	if err := app.PluginRegister(p); err != nil {
		return err
	}

	return nil
}

func (p *plugin) DB() *cache.Cache {
	return p.db
}

func (p *plugin) Start(ctx context.Context) (err error) {
	ring := redis.NewRing(p.prepareOptions(p.opts))
	p.db = cache.New(&cache.Options{
		Redis:      ring,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})
	return nil
}

func (p *plugin) Stop() error {
	return nil
}

func (p *plugin) withPrefix(name string) string {
	return fmt.Sprintf("%s-%s", p.prefix, name)
}

func (p *plugin) withEnvPrefix(name string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s", p.prefix, name))
}

func (p *plugin) addFlags(app engine.Service) {
	app.CLI().AddStringFlag(p.withPrefix("endpoints"), &p.opts.Endpoints).
		Env(p.withEnvPrefix("ENDPOINTS")).
		Usage("Set endpoints for connecting to the server as <host(optional)>:<port>,<host(optional)>:<port>,<etc.> string. (Default: :6379)")

	app.CLI().AddIntFlag(p.withPrefix("db"), &p.opts.DB).
		Env(p.withEnvPrefix("DB")).
		Usage("Set database for connecting to the server. (Default: 0)")

	app.CLI().AddStringFlag(p.withPrefix("username"), &p.opts.Username).
		Env(p.withEnvPrefix("USERNAME")).
		Usage("Set username to authenticate the current connection")

	app.CLI().AddStringFlag(p.withPrefix("password"), &p.opts.Password).
		Env(p.withEnvPrefix("PASSWORD")).
		Usage("Set password to authenticate the current connection")

	app.CLI().AddDurationFlag(p.withPrefix("heartbeat-frequency"), &p.opts.HeartbeatFrequency).
		Env(p.withEnvPrefix("HEARNEAT_FREQUENCY")).
		Usage("Set frequency of PING commands sent to check shards availability.")

	app.CLI().AddIntFlag(p.withPrefix("max-retries"), &p.opts.MaxRetries).
		Env(p.withEnvPrefix("MAX_RETRIES")).
		Usage("Set maximum number of retries before giving up. (Default is 3 retries)")

	app.CLI().AddDurationFlag(p.withPrefix("min-retry-backoff"), &p.opts.MinRetryBackoff).
		Env(p.withEnvPrefix("MIN_RETRY_BACKOFF")).
		Usage("Set minimum backoff between each retry. (Default is 512 milliseconds)")

	app.CLI().AddDurationFlag(p.withPrefix("max-retry-backoff"), &p.opts.MaxRetryBackoff).
		Env(p.withEnvPrefix("MAX_RETRY_BACKOFF")).
		Usage("Set maximum backoff between each retry.")

	app.CLI().AddDurationFlag(p.withPrefix("dial-timeout"), &p.opts.DialTimeout).
		Env(p.withEnvPrefix("DIAL_TIMEOUT")).
		Usage("Set dial timeout for establishing new connections. (Default is 5 seconds)")

	app.CLI().AddDurationFlag(p.withPrefix("read-timeout"), &p.opts.ReadTimeout).
		Env(p.withEnvPrefix("READ_TIMEOUT")).
		Usage("Set timeout for socket reads. If reached, commands will fail. (Default is 3 seconds)")

	app.CLI().AddDurationFlag(p.withPrefix("write-timeout"), &p.opts.WriteTimeout).
		Env(p.withEnvPrefix("WRITE_TIMEOUT")).
		Usage("Set timeout for socket writes. If reached, commands will fail with a timeout instead of blocking. (Default is ReadTimeout)")

	app.CLI().AddIntFlag(p.withPrefix("pool-size"), &p.opts.PoolSize).
		Env(p.withEnvPrefix("POOL_SIZE")).
		Usage("Set maximum number of socket connections. (Default is 10 connections per every CPU)")

	app.CLI().AddIntFlag(p.withPrefix("min-idle-conns"), &p.opts.MinIdleConns).
		Env(p.withEnvPrefix("MIN_IDLE_CONNS")).
		Usage("Set connection age at which client retires (closes) the connection.")

	app.CLI().AddDurationFlag(p.withPrefix("max-conn-age"), &p.opts.MaxConnAge).
		Env(p.withEnvPrefix("MAX_CONN_AGE")).
		Usage("Set connection age at which client retires (closes) the connection. (Default is to not close aged connections.)")

	app.CLI().AddDurationFlag(p.withPrefix("pool-timeout"), &p.opts.PoolTimeout).
		Env(p.withEnvPrefix("POOL_TIMEOUT")).
		Usage("Set amount of time client waits for connection if all connections are busy before returning an error. (Default is ReadTimeout + 1 second.)")

	app.CLI().AddDurationFlag(p.withPrefix("idle-timeout"), &p.opts.IdleTimeout).
		Env(p.withEnvPrefix("IDLE_TIMEOUT")).
		Usage("Set amount of time after which client closes idle connections. Should be less than server's timeout. (Default is 5 minutes.)")

	app.CLI().AddDurationFlag(p.withPrefix("idle-check-frequency"), &p.opts.IdleCheckFrequency).
		Env(p.withEnvPrefix("IDLE_CHECK_FREQUENCY")).
		Usage("Set frequency of idle checks made by idle connections reaper. (Default is 1 minute. -1)")
}

func (p *plugin) prepareOptions(opts options) *redis.RingOptions {

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
