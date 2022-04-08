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
	"database/sql"
	"github.com/go-pg/pg/v10"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/cmd"
	"github.com/pkg/errors"

	_ "github.com/golang-migrate/migrate/v4/source/file" // nolint
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
	toolkit.Plugin

	DB() *pg.DB
	Register(app toolkit.Service, opts *Options) error
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

	MigrationsDir string
}

type plugin struct {
	prefix     string
	connection string
	opts       options

	db *pg.DB
}

func NewPlugin(app toolkit.Service, opts *Options) Plugin {
	db := new(plugin)
	err := db.Register(app, opts)
	if err != nil {
		return nil
	}
	return db
}

// Register - registers the plugin implements storage using Postgres as a database storage
func (p *plugin) Register(app toolkit.Service, opts *Options) error {

	p.prefix = opts.Name
	if p.prefix == "" {
		p.prefix = defaultPrefix
	}

	p.addFlags(app)
	p.addCommands(app)

	if err := app.PluginRegister(p); err != nil {
		return err
	}

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
		return errors.New(errMissingConnectionString)
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

	if p.opts.MigrationsDir != "" {
		fmt.Printf("\nRun migration from dir: %s", p.opts.MigrationsDir)
		sqlDB, err := sql.Open(driverName, p.opts.Connection)
		if err != nil {
			return fmt.Errorf("failed to db open: %w", err)
		}
		if err = p.migration(sqlDB, p.opts.MigrationsDir, p.opts.Connection); err != nil {
			return err
		}
		if err = sqlDB.Close(); err != nil {
			return err
		}
		fmt.Printf("\nMigration completed!\n")
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

func (p *plugin) addFlags(app toolkit.Service) {

	app.CLI().AddStringFlag(p.withPrefix("connection"), &p.opts.Connection).
		Env(p.withEnvPrefix("CONNECTION")).
		Usage("PostgreSQL connection string (Ex: postgres://user:pass@localhost:5432/db_name)").
		Required()

	app.CLI().AddBoolFlag(p.withPrefix("logger"), &p.opts.Logger).
		Env(p.withEnvPrefix("LOGGER")).
		Usage("Enable the query logger").
		Default(false)

	app.CLI().AddDurationFlag(p.withPrefix("dial-timeout"), p.opts.DialTimeout).
		Env(p.withEnvPrefix("DIAL_TIMEOUT")).
		Usage("Enable the query logger").
		Default(0)

	app.CLI().AddDurationFlag(p.withPrefix("read-timeout"), p.opts.ReadTimeout).
		Env(p.withEnvPrefix("READ_TIMEOUT")).
		Usage("Sets the timeout for socket reads.\nIf reached, commands will fail with a timeout instead of blocking.").
		Default(0)

	app.CLI().AddDurationFlag(p.withPrefix("write-timeout"), p.opts.WriteTimeout).
		Env(p.withEnvPrefix("WRITE_TIMEOUT")).
		Usage("Sets the timeout for socket writes.\nIf reached, commands will fail with a timeout instead of blocking.").
		Default(0)

	app.CLI().AddBoolFlag(p.withPrefix("retry-statement-timeout"), p.opts.RetryStatementTimeout).
		Env(p.withEnvPrefix("MAX_RETRIES")).
		Usage("Sets the whether to retry queries cancelled because of statement_timeout.").
		Default(false)

	app.CLI().AddDurationFlag(p.withPrefix("min-retry-backoff"), p.opts.MinRetryBackoff).
		Env(p.withEnvPrefix("MIN_RETRY_BACKOFF")).
		Usage("Sets the minimum backoff between each retry.\nDefault is 250 milliseconds; -1 disables backoff.").
		Default(0)

	app.CLI().AddDurationFlag(p.withPrefix("max-retry-backoff"), p.opts.MaxRetryBackoff).
		Env(p.withEnvPrefix("MAX_RETRY_BACKOFF")).
		Usage("Sets the maximum backoff between each retry.\nDefault is 4 seconds; -1 disables backoff.").
		Default(0)

	app.CLI().AddIntFlag(p.withPrefix("pool-size"), p.opts.PoolSize).
		Env(p.withEnvPrefix("POOL_SIZE")).
		Usage("Sets the maximum number of socket connections.\nDefault is 10 connections per every CPU.").
		Default(0)

	app.CLI().AddIntFlag(p.withPrefix("min-idle-conns"), p.opts.MinIdleConns).
		Env(p.withEnvPrefix("MIN_IDLE_CONNS")).
		Usage("Minimum number of idle connections which is useful when establishing new connection is slow.").
		Default(0)

	app.CLI().AddDurationFlag(p.withPrefix("max-conn-age"), p.opts.MaxConnAge).
		Env(p.withEnvPrefix("MAX_CONN_AGE")).
		Usage("Sets the connection age at which client retires (closes) the connection.\nDefault is to not close aged connections.").
		Default(0)

	app.CLI().AddDurationFlag(p.withPrefix("pool-timeout"), p.opts.PoolTimeout).
		Env(p.withEnvPrefix("POOL_TIMEOUT")).
		Usage("Sets the time for which client waits for free connection if all connections are busy before returning an error.\nDefault is 30 seconds if ReadTimeOut is not defined, otherwise, ReadTimeout + 1 second.").
		Default(0)

	app.CLI().AddDurationFlag(p.withPrefix("idle-timeout"), p.opts.IdleTimeout).
		Env(p.withEnvPrefix("IDLE_TIMEOUT")).
		Usage("Sets the amount of time after which client closes idle connections.\nShould be less than server's timeout.\nDefault is 5 minutes. -1 disables idle timeout check.").
		Default(0)

	app.CLI().AddDurationFlag(p.withPrefix("idle-check-frequency"), p.opts.IdleCheckFrequency).
		Env(p.withEnvPrefix("IDLE_CHECK_FREQUENCY")).
		Usage("Sets the frequency of idle checks made by idle connections reaper.\nDefault is 1 minute. -1 disables idle connections reaper, but idle connections are still discarded by the client if IdleTimeout is set.").
		Default(0)

	app.CLI().AddStringFlag(p.withPrefix("migration-dir"), &p.opts.MigrationsDir).
		Env(p.withEnvPrefix("MIGRATION_DIR")).
		Usage("PostgreSQL migration dir path")
}

func (p *plugin) addCommands(app toolkit.Service) {
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

			c, err := sqlx.Open(driverName, connection)
			if err != nil {
				return fmt.Errorf("failed to db open: %w", err)
			}

			if err = p.migration(c.DB, args[0], connection); err != nil {
				return err
			}

			return c.Close()
		},
	}

	migrateCmd.AddStringFlag(p.withPrefix("connection"), nil).
		Env(p.withEnvPrefix("CONNECTION")).
		Usage("PostgreSQL connection string (Ex: postgres://user:pass@localhost:5432/db_name)").
		Required()

	app.CLI().AddCommand(migrateCmd)
}

func (p *plugin) migration(db *sql.DB, migrateDir, connectionString string) error {

	opts, err := pg.ParseURL(connectionString)
	if err != nil {
		return errors.New(errMissingConnectionString)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", migrateDir), opts.Database, driver)
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

	return nil
}
