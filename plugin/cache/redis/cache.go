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
	"github.com/lastbackend/engine/cmd"

	"fmt"
	"strings"
)

type Options struct {
	clientOptions
}

type redisCache struct {
	client *client

	prefix string
	opts   Options
}

func newCache(prefix string) *redisCache {
	s := new(redisCache)
	if len(prefix) == 0 {
		s.prefix = defaultPrefix
	}
	s.prefix = prefix
	s.client = newClient()
	return s
}

func (c *redisCache) Name() string {
	return PluginName
}

func (c *redisCache) Flags() []cmd.Flag {
	return []cmd.Flag{
		&cmd.StringFlag{
			Name:        c.withPrefix("endpoints"),
			EnvVars:     []string{c.withEnvPrefix("ENDPOINTS")},
			Usage:       "Set endpoints for connecting to the server as <host(optional)>:<port>,<host(optional)>:<port>,<etc.> string. (Default: :6379)",
			Required:    false,
			Destination: &c.opts.Endpoints,
		},
		&cmd.IntFlag{
			Name:        c.withPrefix("db"),
			EnvVars:     []string{c.withEnvPrefix("DB")},
			Usage:       "Set database for connecting to the server. (Default: 0)",
			Required:    false,
			Destination: &c.opts.DB,
		},
		&cmd.StringFlag{
			Name:        c.withPrefix("username"),
			EnvVars:     []string{c.withEnvPrefix("USERNAME")},
			Usage:       "Set username to authenticate the current connection",
			Required:    false,
			Destination: &c.opts.Username,
		},
		&cmd.StringFlag{
			Name:        c.withPrefix("password"),
			EnvVars:     []string{c.withEnvPrefix("PASSWORD")},
			Usage:       "Set password to authenticate the current connection",
			Required:    false,
			Destination: &c.opts.Password,
		},
		&cmd.DurationFlag{
			Name:        c.withPrefix("heartbeat-frequency"),
			EnvVars:     []string{c.withEnvPrefix("HEARNEAT_FREQUENCY")},
			Usage:       "Set frequency of PING commands sent to check shards availability.",
			Required:    false,
			Destination: &c.opts.HeartbeatFrequency,
		},
		&cmd.IntFlag{
			Name:        c.withPrefix("max-retries"),
			EnvVars:     []string{c.withEnvPrefix("MAX_RETRIES")},
			Usage:       "Set maximum number of retries before giving up. (Default is 3 retries)",
			Required:    false,
			Destination: &c.opts.MaxRetries,
		},
		&cmd.DurationFlag{
			Name:        c.withPrefix("min-retry-backoff"),
			EnvVars:     []string{c.withEnvPrefix("MIN_RETRY_BACKOFF")},
			Usage:       "Set minimum backoff between each retry. (Default is 512 milliseconds)",
			Required:    false,
			Destination: &c.opts.MinRetryBackoff,
		},
		&cmd.DurationFlag{
			Name:        c.withPrefix("max-retry-backoff"),
			EnvVars:     []string{c.withEnvPrefix("MAX_RETRY_BACKOFF")},
			Usage:       "Set maximum backoff between each retry.",
			Required:    false,
			Destination: &c.opts.MaxRetryBackoff,
		},
		&cmd.DurationFlag{
			Name:        c.withPrefix("dial-timeout"),
			EnvVars:     []string{c.withEnvPrefix("DIAL_TIMEOUT")},
			Usage:       "Set dial timeout for establishing new connections. (Default is 5 seconds)",
			Required:    false,
			Destination: &c.opts.DialTimeout,
		},
		&cmd.DurationFlag{
			Name:        c.withPrefix("read-timeout"),
			EnvVars:     []string{c.withEnvPrefix("READ_TIMEOUT")},
			Usage:       "Set timeout for socket reads. If reached, commands will fail. (Default is 3 seconds)",
			Required:    false,
			Destination: &c.opts.ReadTimeout,
		},
		&cmd.DurationFlag{
			Name:        c.withPrefix("write-timeout"),
			EnvVars:     []string{c.withEnvPrefix("WRITE_TIMEOUT")},
			Usage:       "Set timeout for socket writes. If reached, commands will fail with a timeout instead of blocking. (Default is ReadTimeout)",
			Required:    false,
			Destination: &c.opts.WriteTimeout,
		},
		&cmd.IntFlag{
			Name:        c.withPrefix("pool-size"),
			EnvVars:     []string{c.withEnvPrefix("POOL_SIZE")},
			Usage:       "Set maximum number of socket connections. (Default is 10 connections per every CPU)",
			Required:    false,
			Destination: &c.opts.PoolSize,
		},
		&cmd.IntFlag{
			Name:        c.withPrefix("min-idle-conns"),
			EnvVars:     []string{c.withEnvPrefix("MIN_IDLE_CONNS")},
			Usage:       "Set connection age at which client retires (closes) the connection.",
			Required:    false,
			Destination: &c.opts.MinIdleConns,
		},
		&cmd.DurationFlag{
			Name:        c.withPrefix("max-conn-age"),
			EnvVars:     []string{c.withEnvPrefix("MAX_CONN_AGE")},
			Usage:       "Set connection age at which client retires (closes) the connection. (Default is to not close aged connections.)",
			Required:    false,
			Destination: &c.opts.MaxConnAge,
		},
		&cmd.DurationFlag{
			Name:        c.withPrefix("pool-timeout"),
			EnvVars:     []string{c.withEnvPrefix("POOL_TIMEOUT")},
			Usage:       "Set amount of time client waits for connection if all connections are busy before returning an error. (Default is ReadTimeout + 1 second.)",
			Required:    false,
			Destination: &c.opts.PoolTimeout,
		},
		&cmd.DurationFlag{
			Name:        c.withPrefix("idle-timeout"),
			EnvVars:     []string{c.withEnvPrefix("IDLE_TIMEOUT")},
			Usage:       "Set amount of time after which client closes idle connections. Should be less than server's timeout. (Default is 5 minutes.)",
			Required:    false,
			Destination: &c.opts.IdleTimeout,
		},
		&cmd.DurationFlag{
			Name:        c.withPrefix("idle-check-frequency"),
			EnvVars:     []string{c.withEnvPrefix("IDLE_CHECK_FREQUENCY")},
			Usage:       "Set frequency of idle checks made by idle connections reaper. (Default is 1 minute. -1)",
			Required:    false,
			Destination: &c.opts.IdleCheckFrequency,
		},
	}
}

func (c *redisCache) Commands() []cmd.Command {
	return []cmd.Command{}
}

func (c *redisCache) Start() error {
	return c.client.open(c.opts.clientOptions)
}

func (c *redisCache) Stop() error {
	return nil
}

func (c *redisCache) getClient() *client {
	return c.client
}

func (c *redisCache) withPrefix(name string) string {
	return fmt.Sprintf("%s-%s", c.prefix, name)
}

func (c *redisCache) withEnvPrefix(name string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s", c.prefix, name))
}
