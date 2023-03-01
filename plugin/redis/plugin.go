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

package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/pkg/config"
	"strings"
	"time"

	"github.com/lastbackend/toolkit/pkg/probes"
	"github.com/redis/go-redis/v9"
)

const (
	defaultPrefix   = "REDIS"
	defaultEndpoint = ":6379"
)

type Config struct {
	Endpoint string `env:"ENDPOINT" envDefault:":6379" comment:"Endpoint = host:port,host:port addresses of ring shards."`

	Cluster bool `env:"CLUSTER" comment:"Cluster = enable cluster mode"`

	Database int `env:"DATABASE" required:"true" comment:"Database to be selected after connecting to the server."`

	Username string `env:"USERNAME" comment:"Use the specified Username to authenticate the current connection with one of the connections defined in the ACL list when connecting to a Redis 6.0 instance, or greater, that is using the Redis ACL system."`

	Password string `env:"PASSWORD" comment:"Optional password. Must match the password specified in the requirepass server configuration option (if connecting to a Redis 5.0 instance, or lower), or the User Password when connecting to a Redis 6.0 instance, or greater, that is using the Redis ACL system."`

	HeartbeatFrequency time.Duration `env:"HEARNEAT_FREQUENCY" comment:"Frequency of PING commands sent to check shards availability. Shard is considered down after 3 subsequent failed checks."`

	MaxRetries int `env:"MAX_RETRIES" comment:"Maximum number of retries before giving up. Default is 3 retries; -1 (not 0) disables retries."`

	MinRetryBackoff time.Duration `env:"MIN_RETRY_BACKOFF" comment:"Minimum backoff between each retry. Default is 8 milliseconds; -1 disables backoff."`

	MaxRetryBackoff time.Duration `env:"MAX_RETRY_BACKOFF" comment:"Maximum backoff between each retry. Default is 512 milliseconds; -1 disables backoff."`

	DialTimeout time.Duration `env:"DIAL_TIMEOUT" comment:"Dial timeout for establishing new connections. Default is 5 seconds."`

	ReadTimeout time.Duration `env:"READ_TIMEOUT" comment:"Timeout for socket reads. If reached, commands will fail with a timeout instead of blocking. Use value -1 for no timeout and 0 for default. Default is 3 seconds."`

	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" comment:"Timeout for socket writes. If reached, commands will fail with a timeout instead of blocking. Default is Read Timeout."`

	PoolSize int `env:"POOL_SIZE" comment:"Maximum number of socket connections. Default is 10 connections per every CPU as reported by runtime.NumCPU."`

	MinIdleConns int `env:"MIN_IDLE_CONNS" comment:"Minimum number of idle connections which is useful when establishing new connection is slow."`

	PoolTimeout time.Duration `env:"POOL_TIMEOUT" comment:"Amount of time client waits for connection if all connections are busy before returning an error. Default is ReadTimeout + 1 second."`

	// TODO: need to implement the ability to install tls
	// TLS Config to use. When set TLS will be negotiated.
	TLSConfig *tls.Config
}

type Plugin interface {
	toolkit.Plugin

	DB() *redis.Client
	ClusterDB() *redis.ClusterClient

	Register(app toolkit.Service, opts *Options) error

	Print()
}

type Options struct {
	Name string
}

type plugin struct {
	prefix string

	opts Config
	db   *redis.Client
	cdb  *redis.ClusterClient

	//probe toolkit.Probe
}

func NewPlugin(service toolkit.Service, opts *Options) Plugin {

	p := new(plugin)

	p.prefix = opts.Name
	if p.prefix == "" {
		p.prefix = defaultPrefix
	}

	if err := config.Parse(&p.opts, p.prefix); err != nil {
		return nil
	}

	//p.probe = service.Probe()
	err := p.Register(service, opts)

	if err != nil {
		return nil
	}

	return p
}

// Register - registers the plugin implements storage using Postgres as a database storage
func (p *plugin) Register(app toolkit.Service, opts *Options) error {

	//if err := app.PluginRegister(p); err != nil {
	//	return err
	//}

	return nil
}

func (p *plugin) DB() *redis.Client {
	return p.db
}

func (p *plugin) ClusterDB() *redis.ClusterClient {
	return p.cdb
}

func (p *plugin) Start(ctx context.Context) (err error) {

	if p.opts.Cluster {
		client := redis.NewClusterClient(p.prepareClusterOptions(p.opts))
		//p.probes.AddReadinessFunc(p.prefix, redisClusterPingChecker(client, 1*time.Second))
		p.cdb = client
	} else {
		client := redis.NewClient(p.prepareOptions(p.opts))
		//p.probes.AddReadinessFunc(p.prefix, redisPingChecker(client, 1*time.Second))
		p.db = client
	}
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

func (p *plugin) addFlags(app toolkit.Service) {

}

func (p *plugin) prepareOptions(opts Config) *redis.Options {

	addr := defaultEndpoint
	if len(opts.Endpoint) > 0 {
		opts.Endpoint = strings.Replace(opts.Endpoint, " ", "", -1)
		addr = strings.Split(opts.Endpoint, ",")[0]
	}

	return &redis.Options{
		Addr:            addr,
		Username:        opts.Username,
		Password:        opts.Password,
		MaxRetries:      opts.MaxRetries,
		MinRetryBackoff: opts.MinRetryBackoff,
		MaxRetryBackoff: opts.MaxRetryBackoff,
		DialTimeout:     opts.DialTimeout,
		ReadTimeout:     opts.ReadTimeout,
		WriteTimeout:    opts.WriteTimeout,
		PoolSize:        opts.PoolSize,
		MinIdleConns:    opts.MinIdleConns,
		PoolTimeout:     opts.PoolTimeout,
		TLSConfig:       opts.TLSConfig,
	}
}

func (p *plugin) prepareClusterOptions(opts Config) *redis.ClusterOptions {

	addrs := []string{defaultEndpoint}
	if len(opts.Endpoint) > 0 {
		opts.Endpoint = strings.Replace(opts.Endpoint, " ", "", -1)
		addrs = strings.Split(opts.Endpoint, ",")
	}

	return &redis.ClusterOptions{
		Addrs:           addrs,
		Username:        opts.Username,
		Password:        opts.Password,
		MaxRetries:      opts.MaxRetries,
		MinRetryBackoff: opts.MinRetryBackoff,
		MaxRetryBackoff: opts.MaxRetryBackoff,
		DialTimeout:     opts.DialTimeout,
		ReadTimeout:     opts.ReadTimeout,
		WriteTimeout:    opts.WriteTimeout,
		PoolSize:        opts.PoolSize,
		MinIdleConns:    opts.MinIdleConns,
		PoolTimeout:     opts.PoolTimeout,
		TLSConfig:       opts.TLSConfig,
	}
}

func (p *plugin) Print() {
	config.Print(p.opts, p.prefix)
}

func redisClusterPingChecker(client *redis.ClusterClient, timeout time.Duration) probes.HandleFunc {
	return func() error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if client == nil {
			return fmt.Errorf("connection is nil")
		}
		_, err := client.Ping(ctx).Result()
		if err != nil {
			return err
		}
		return nil
	}
}

func redisPingChecker(client *redis.Client, timeout time.Duration) probes.HandleFunc {
	return func() error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if client == nil {
			return fmt.Errorf("connection is nil")
		}
		_, err := client.Ping(ctx).Result()
		if err != nil {
			return err
		}
		return nil
	}
}
