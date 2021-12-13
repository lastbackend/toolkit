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
	"github.com/lastbackend/engine"
	"github.com/lastbackend/engine/cmd"
	"github.com/pkg/errors"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"context"
	"fmt"
	"strings"
	"time"
)

const (
	defaultPrefix = "psql"
	driverName    = "postgres"
)

const (
	errMissingConnectionString = "Missing connection string"
)

type Plugin interface {
	engine.Plugin

	DB() *pg.DB
	Register(app engine.Service, opts Options) error
}

type Options struct {
	Name string
}

type options struct {
	// Enable the query logger
	Logger bool

	// Sets the connection string for connecting to the database
	Connection string

	// Sets the dial timeout for establishing new connections.
	// Default is 5 seconds.
	DialTimeout *time.Duration

	// Sets the timeout for socket reads. If reached, commands will fail
	// with a timeout instead of blocking.
	ReadTimeout *time.Duration
	// Sets the timeout for socket writes. If reached, commands will fail
	// with a timeout instead of blocking.
	WriteTimeout *time.Duration

	// Sets the maximum number of retries before giving up.
	// Default is to not retry failed queries.
	MaxRetries *int
	// Sets the whether to retry queries cancelled because of statement_timeout.
	RetryStatementTimeout *bool
	// Sets the minimum backoff between each retry.
	// Default is 250 milliseconds; -1 disables backoff.
	MinRetryBackoff *time.Duration
	// Sets the maximum backoff between each retry.
	// Default is 4 seconds; -1 disables backoff.
	MaxRetryBackoff *time.Duration

	// Sets the maximum number of socket connections.
	// Default is 10 connections per every CPU as reported by runtime.NumCPU.
	PoolSize *int
	// Sets the minimum number of idle connections which is useful when establishing
	// new connection is slow.
	MinIdleConns *int
	// Sets the connection age at which client retires (closes) the connection.
	// It is useful with proxies like PgBouncer and HAProxy.
	// Default is to not close aged connections.
	MaxConnAge *time.Duration
	// Sets the time for which client waits for free connection if all
	// connections are busy before returning an error.
	// Default is 30 seconds if ReadTimeOut is not defined, otherwise,
	// ReadTimeout + 1 second.
	PoolTimeout *time.Duration
	// Sets the amount of time after which client closes idle connections.
	// Should be less than server's timeout.
	// Default is 5 minutes. -1 disables idle timeout check.
	IdleTimeout *time.Duration
	// Sets the frequency of idle checks made by idle connections reaper.
	// Default is 1 minute. -1 disables idle connections reaper,
	// but idle connections are still discarded by the client
	// if IdleTimeout is set.
	IdleCheckFrequency *time.Duration

	MigrationsDir *string
}

type plugin struct {
	prefix     string
	connection string
	opts       options

	db *pg.DB
}

func Register(app engine.Service, opts Options) Plugin {
	db := new(plugin)
	db.Register(app, opts)
	return db
}

// Register - registers the plug implements storage using Postgres as a database storage
func (p *plugin) Register(app engine.Service, opts Options) error {

	p.prefix = opts.Name
	if p.prefix == "" {
		p.prefix = defaultPrefix
	}

	if err := app.PluginRegister(p); err != nil {
		return err
	}

	p.addFlags(app)
	p.addCommands(app)

	return nil
}

func (p *plugin) DB() *pg.DB {
	return p.db
}

func (p *plugin) Start(ctx context.Context) (err error) {
	if len(p.opts.Connection) == 0 {
		return errors.New(errMissingConnectionString)
	}

	opt, err := pg.ParseURL(p.opts.Connection)
	if err != nil {
		panic(err)
	}

	if p.opts.DialTimeout != nil {
		opt.DialTimeout = *p.opts.DialTimeout
	}
	if p.opts.ReadTimeout != nil {
		opt.ReadTimeout = *p.opts.ReadTimeout
	}
	if p.opts.WriteTimeout != nil {
		opt.WriteTimeout = *p.opts.WriteTimeout
	}
	if p.opts.MaxRetries != nil {
		opt.MaxRetries = *p.opts.MaxRetries
	}
	if p.opts.RetryStatementTimeout != nil {
		opt.RetryStatementTimeout = *p.opts.RetryStatementTimeout
	}
	if p.opts.MinRetryBackoff != nil {
		opt.MinRetryBackoff = *p.opts.MinRetryBackoff
	}
	if p.opts.MaxRetryBackoff != nil {
		opt.MaxRetryBackoff = *p.opts.MaxRetryBackoff
	}
	if p.opts.PoolSize != nil {
		opt.PoolSize = *p.opts.PoolSize
	}
	if p.opts.MinIdleConns != nil {
		opt.MinIdleConns = *p.opts.MinIdleConns
	}
	if p.opts.MaxConnAge != nil {
		opt.MaxConnAge = *p.opts.MaxConnAge
	}
	if p.opts.PoolTimeout != nil {
		opt.PoolTimeout = *p.opts.PoolTimeout
	}
	if p.opts.IdleTimeout != nil {
		opt.IdleTimeout = *p.opts.IdleTimeout
	}
	if p.opts.IdleCheckFrequency != nil {
		opt.IdleCheckFrequency = *p.opts.IdleCheckFrequency
	}

	p.connection = p.opts.Connection
	db := pg.Connect(opt)

	if p.opts.Logger {
		db.AddQueryHook(dbLogger{})
	}

	p.db = db

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
	app.CLI().AddStringFlag(p.withPrefix("connection"), "", "", &p.opts.Connection, p.withEnvPrefix("CONNECTION"), true, "PostgreSQL connection string (Ex: postgres://user:pass@localhost:5432/db_name)")
	app.CLI().AddBoolFlag(p.withPrefix("logger"), "", false, &p.opts.Logger, p.withEnvPrefix("LOGGER"), false, "Enable the query logger")
	app.CLI().AddDurationFlag(p.withPrefix("dial-timeout"), "", 0, p.opts.DialTimeout, p.withEnvPrefix("DIAL_TIMEOUT"), false, "Enable the query logger")
	app.CLI().AddDurationFlag(p.withPrefix("read-timeout"), "", 0, p.opts.ReadTimeout, p.withEnvPrefix("READ_TIMEOUT"), false, "Sets the timeout for socket reads.\nIf reached, commands will fail with a timeout instead of blocking.")
	app.CLI().AddDurationFlag(p.withPrefix("write-timeout"), "", 0, p.opts.WriteTimeout, p.withEnvPrefix("WRITE_TIMEOUT"), false, "Sets the timeout for socket writes.\nIf reached, commands will fail with a timeout instead of blocking.")
	app.CLI().AddBoolFlag(p.withPrefix("retry-statement-timeout"), "", false, p.opts.RetryStatementTimeout, p.withEnvPrefix("MAX_RETRIES"), false, "Sets the whether to retry queries cancelled because of statement_timeout.")
	app.CLI().AddDurationFlag(p.withPrefix("min-retry-backoff"), "", 0, p.opts.MinRetryBackoff, p.withEnvPrefix("MIN_RETRY_BACKOFF"), false, "Sets the minimum backoff between each retry.\nDefault is 250 milliseconds; -1 disables backoff.")
	app.CLI().AddDurationFlag(p.withPrefix("max-retry-backoff"), "", 0, p.opts.MaxRetryBackoff, p.withEnvPrefix("MAX_RETRY_BACKOFF"), false, "Sets the maximum backoff between each retry.\nDefault is 4 seconds; -1 disables backoff.")
	app.CLI().AddIntFlag(p.withPrefix("pool-size"), "", 0, p.opts.PoolSize, p.withEnvPrefix("POOL_SIZE"), false, "Sets the maximum number of socket connections.\nDefault is 10 connections per every CPU.")
	app.CLI().AddIntFlag(p.withPrefix("min-idle-conns"), "", 0, p.opts.MinIdleConns, p.withEnvPrefix("MIN_IDLE_CONNS"), false, "Minimum number of idle connections which is useful when establishing new connection is slow.")
	app.CLI().AddDurationFlag(p.withPrefix("max-conn-age"), "", 0, p.opts.MaxConnAge, p.withEnvPrefix("MAX_CONN_AGE"), false, "Sets the connection age at which client retires (closes) the connection.\nDefault is to not close aged connections.")
	app.CLI().AddDurationFlag(p.withPrefix("pool-timeout"), "", 0, p.opts.PoolTimeout, p.withEnvPrefix("POOL_TIMEOUT"), false, "Sets the time for which client waits for free connection if all connections are busy before returning an error.\nDefault is 30 seconds if ReadTimeOut is not defined, otherwise, ReadTimeout + 1 second.")
	app.CLI().AddDurationFlag(p.withPrefix("idle-timeout"), "", 0, p.opts.IdleTimeout, p.withEnvPrefix("IDLE_TIMEOUT"), false, "Sets the amount of time after which client closes idle connections.\nShould be less than server's timeout.\nDefault is 5 minutes. -1 disables idle timeout check.")
	app.CLI().AddDurationFlag(p.withPrefix("idle-check-frequency"), "", 0, p.opts.IdleCheckFrequency, p.withEnvPrefix("IDLE_CHECK_FREQUENCY"), false, "Sets the frequency of idle checks made by idle connections reaper.\nDefault is 1 minute. -1 disables idle connections reaper, but idle connections are still discarded by the client if IdleTimeout is set.")
}

func (p *plugin) addCommands(app engine.Service) {
	migrateCmd := &cmd.Command{
		Use:       "migrate [SOURCE_PATH]",
		ShortDesc: "Database migrations",
		Run: func(cmd *cmd.Command, args []string) error {

			if len(args) == 0 {
				return fmt.Errorf("argument \"source path\" is not set, programmer error, please correct")
			}

			connection, err := cmd.Flags().GetString(p.withPrefix("connection"))
			if err != nil {
				return errors.Wrapf(err, "\"%s\" flag is non-string, programmer error, please correct", p.withPrefix("connection"))
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

	migrateCmd.AddStringFlag(p.withPrefix("connection"), "", "", nil, p.withEnvPrefix("CONNECTION"), true, "PostgreSQL connection string (Ex: postgres://user:pass@localhost:5432/db_name)")

	app.CLI().AddCommand(migrateCmd)
}
