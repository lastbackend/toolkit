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

package postgres_pg

import (
	"github.com/go-pg/pg/v10"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
	"github.com/lastbackend/engine/cmd"
	"github.com/pkg/errors"

	"context"
	"fmt"
	"strings"
	"time"
)

const (
	defaultPingTimeout = 60 * time.Second
)

type Options struct {
	clientOptions
	MigrationsDir *string
}

type postgresStorage struct {
	client *client

	prefix string
	opts   Options
}

func newStorage(prefix string) *postgresStorage {
	s := new(postgresStorage)
	if len(prefix) == 0 {
		s.prefix = defaultPrefix
	}
	s.prefix = prefix
	s.client = newClient()
	return s
}

func (s *postgresStorage) Name() string {
	return PluginName
}

func (s *postgresStorage) Flags() []cmd.Flag {
	return []cmd.Flag{
		&cmd.StringFlag{
			Name:        s.withPrefix("connection"),
			EnvVar:      s.withEnvPrefix("CONNECTION"),
			Usage:       "PostgreSQL connection string (Ex: postgres://user:pass@localhost:5432/db_name)",
			Required:    true,
			Destination: &s.opts.Connection,
		},
		&cmd.DurationFlag{
			Name:        s.withPrefix("dial-timeout"),
			EnvVar:      s.withEnvPrefix("DIAL_TIMEOUT"),
			Usage:       "Sets the dial timeout for establishing new connections.\nThe default is 5 seconds",
			Required:    false,
			Destination: s.opts.DialTimeout,
		},
		&cmd.DurationFlag{
			Name:        s.withPrefix("read_timeout"),
			EnvVar:      s.withEnvPrefix("READ_TIMEOUT"),
			Usage:       "Sets the timeout for socket reads.\nIf reached, commands will fail with a timeout instead of blocking.",
			Required:    false,
			Destination: s.opts.ReadTimeout,
		},
		&cmd.DurationFlag{
			Name:        s.withPrefix("write_timeout"),
			EnvVar:      s.withEnvPrefix("WRITE_TIMEOUT"),
			Usage:       "Sets the timeout for socket writes.\nIf reached, commands will fail with a timeout instead of blocking.",
			Required:    false,
			Destination: s.opts.WriteTimeout,
		},
		&cmd.IntFlag{
			Name:        s.withPrefix("max_retries"),
			EnvVar:      s.withEnvPrefix("MAX_RETRIES"),
			Usage:       "Sets the maximum number of retries before giving up.\nDefault is to not retry failed queries.",
			Required:    false,
			Destination: s.opts.MaxRetries,
		},
		&cmd.BoolFlag{
			Name:        s.withPrefix("retry_statement_timeout"),
			EnvVar:      s.withEnvPrefix("RETRY_STATEMENT_TIMEOUT"),
			Usage:       "Sets the whether to retry queries cancelled because of statement_timeout.",
			Required:    false,
			Destination: s.opts.RetryStatementTimeout,
		},
		&cmd.DurationFlag{
			Name:        s.withPrefix("min_retry_backoff"),
			EnvVar:      s.withEnvPrefix("MIN_RETRY_BACKOFF"),
			Usage:       "Sets the minimum backoff between each retry.\nDefault is 250 milliseconds; -1 disables backoff.",
			Required:    false,
			Destination: s.opts.MinRetryBackoff,
		},
		&cmd.DurationFlag{
			Name:        s.withPrefix("max_retry_backoff"),
			EnvVar:      s.withEnvPrefix("MAX_RETRY_BACKOFF"),
			Usage:       "Sets the maximum backoff between each retry.\nDefault is 4 seconds; -1 disables backoff.",
			Required:    false,
			Destination: s.opts.MaxRetryBackoff,
		},
		&cmd.IntFlag{
			Name:        s.withPrefix("pool_size"),
			EnvVar:      s.withEnvPrefix("POOL_SIZE"),
			Usage:       "Sets the maximum number of socket connections.\nDefault is 10 connections per every CPU.",
			Required:    false,
			Destination: s.opts.PoolSize,
		},
		&cmd.IntFlag{
			Name:        s.withPrefix("min_idle_conns"),
			EnvVar:      s.withEnvPrefix("MIN_IDLE_CONNS"),
			Usage:       "Minimum number of idle connections which is useful when establishing new connection is slow.",
			Required:    false,
			Destination: s.opts.MinIdleConns,
		},
		&cmd.DurationFlag{
			Name:        s.withPrefix("max_conn_age"),
			EnvVar:      s.withEnvPrefix("MAX_CONN_AGE"),
			Usage:       "Sets the connection age at which client retires (closes) the connection.\nDefault is to not close aged connections.",
			Required:    false,
			Destination: s.opts.MaxConnAge,
		},
		&cmd.DurationFlag{
			Name:        s.withPrefix("pool_timeout"),
			EnvVar:      s.withEnvPrefix("POOL_TIMEOUT"),
			Usage:       "Sets the time for which client waits for free connection if all connections are busy before returning an error.\nDefault is 30 seconds if ReadTimeOut is not defined, otherwise, ReadTimeout + 1 second.",
			Required:    false,
			Destination: s.opts.PoolTimeout,
		},
		&cmd.DurationFlag{
			Name:        s.withPrefix("idle_timeout"),
			EnvVar:      s.withEnvPrefix("IDLE_TIMEOUT"),
			Usage:       "Sets the amount of time after which client closes idle connections.\nShould be less than server's timeout.\nDefault is 5 minutes. -1 disables idle timeout check.",
			Required:    false,
			Destination: s.opts.IdleTimeout,
		},
		&cmd.DurationFlag{
			Name:        s.withPrefix("idle_check_frequency"),
			EnvVar:      s.withEnvPrefix("IDLE_CHECK_FREQUENCY"),
			Usage:       "Sets the frequency of idle checks made by idle connections reaper.\nDefault is 1 minute. -1 disables idle connections reaper, but idle connections are still discarded by the client if IdleTimeout is set.",
			Required:    false,
			Destination: s.opts.IdleCheckFrequency,
		},
	}
}


func (s *postgresStorage) Commands() []cmd.Command {

	migrateCmd := &cmd.Cmd{
		Use:       "migrate [SOURCE_PATH]",
		ShortDesc: "Database migrations",
		Run: func(cmd cmd.Command, args []string) error {

			if len(args) == 0 {
				return fmt.Errorf("argument \"source path\" is not set, programmer error, please correct")
			}

			connection, err := cmd.Flags().GetString(s.withPrefix("connection"))
			if err != nil {
				return errors.Wrapf(err, "\"%s\" flag is non-string, programmer error, please correct", s.withPrefix("connection"))
			}

			opt, err := pg.ParseURL(connection)
			if err != nil {
				return errors.New(errMissingConnectionString)
			}

			c, err := sqlx.Open(driverName, connection)
			if err != nil {
				return fmt.Errorf("failed to db open: %w", err)
			}

			driver, err := postgres.WithInstance(c.DB, &postgres.Config{})
			m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", args[0]), opt.Database, driver)
			if err != nil {
				return err
			}

			version, dirty, err := m.Version()
			if err != nil && err != migrate.ErrNilVersion {
				return err
			}
			if dirty {
				if err := m.Force(int(version)); err != nil {
					return err
				}
				if err := m.Down(); err != nil {
					return err
				}
			}

			if err := m.Up(); err != nil && err != migrate.ErrNoChange {
				return err
			}

			return c.Close()
		},
	}

	migrateCmd.AddStringFlag(s.withPrefix("connection"), "", "", nil, s.withEnvPrefix("CONNECTION"), true, "PostgreSQL connection string (Ex: postgres://user:pass@localhost:5432/db_name)")

	return []cmd.Command{migrateCmd}
}

func (s *postgresStorage) Start() error {

	if err := s.client.open(s.opts.clientOptions); err != nil {
		return err
	}

	if err := s.client.c.Ping(context.Background()); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-time.After(defaultPingTimeout):
				if err := s.client.c.Ping(context.Background()); err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}()

	return nil
}

func (s *postgresStorage) Stop() error {
	return s.client.c.Close()
}

func (s *postgresStorage) getClient() *client {
	return s.client
}

func (s *postgresStorage) withPrefix(name string) string {
	return fmt.Sprintf("%s-%s", s.prefix, name)
}

func (s *postgresStorage) withEnvPrefix(name string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s", s.prefix, name))
}
