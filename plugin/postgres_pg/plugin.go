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

package postgres_pg

import (
	"database/sql"
	"os"
	"strconv"

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
	defaultPort   = 5432
)

const (
	envHostName     = "HOST"
	envPortName     = "PORT"
	envUserName     = "USERNAME"
	envPasswordName = "PASSWORD"
	envDatabaseName = "DATABASE"
	envSslmodeName  = "SSL_MODE"
	envTimezoneName = "TIMEZONE"
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
	envPrefix  string
	connection string
	opts       options

	db *pg.DB

	probe toolkit.Probe
}

func NewPlugin(service toolkit.Service, opts *Options) Plugin {
	p := new(plugin)
	p.envPrefix = service.Meta().GetEnvPrefix()
	p.probe = service.Probe()
	err := p.Register(service, opts)
	if err != nil {
		return nil
	}
	return p
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
	if p.opts.Connection == "" {
		config := p.getDBConfig()
		if config.Host == "" {
			return fmt.Errorf("%s flag or %s environment variable required but not set",
				p.withPrefix("connection"), p.generateWithEnvPrefix(envHostName))
		}
		p.opts.Connection = config.getConnectionString()
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

	p.probe.AddReadinessFunc(p.prefix, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if db == nil {
			return fmt.Errorf("database is nil")
		}
		return db.Ping(ctx)
	})

	if p.opts.Logger {
		db.AddQueryHook(dbLogger{})
	}

	if p.opts.MigrationsDir != "" {
		fmt.Printf("\nRun migration from dir: %s", p.opts.MigrationsDir)
		conn, err := sql.Open(driverName, p.opts.Connection)
		if err != nil {
			return fmt.Errorf("failed to dbection open: %w", err)
		}
		if err = p.migration(conn, p.opts.MigrationsDir, p.opts.Connection); err != nil {
			return err
		}
		if err = conn.Close(); err != nil {
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

func (p *plugin) genUsage() string {
	return fmt.Sprintf(`PostgreSQL connection string (Ex: postgres://user:pass@localhost:5432/db_name) 
or use environment variables: 
	%s - The host to connect to (required), 
	%s - The port to bind to (default: 5432), 
	%s - The username to connect with. Not required if using IntegratedSecurity, 
	%s - The password to connect with. Not required if using IntegratedSecurity, 
	%s - The database to connect to, 
	%s - Whether or not to use SSL, 
	%s - Sets the session timezone`,
		p.generateWithEnvPrefix(envHostName), p.generateWithEnvPrefix(envPortName), p.generateWithEnvPrefix(envUserName),
		p.generateWithEnvPrefix(envPasswordName), p.generateWithEnvPrefix(envDatabaseName), p.generateWithEnvPrefix(envSslmodeName),
		p.generateWithEnvPrefix(envTimezoneName))
}

func (p *plugin) addFlags(app toolkit.Service) {

	app.CLI().AddStringFlag(p.withPrefix("connection"), &p.opts.Connection).
		Env(p.generateEnvName("CONNECTION")).
		Usage(p.genUsage())

	app.CLI().AddBoolFlag(p.withPrefix("logger"), &p.opts.Logger).
		Env(p.generateEnvName("LOGGER")).
		Usage("Enable the query logger").
		Default(false)

	app.CLI().AddDurationFlag(p.withPrefix("dial-timeout"), p.opts.DialTimeout).
		Env(p.generateEnvName("DIAL_TIMEOUT")).
		Usage("Enable the query logger").
		Default(0)

	app.CLI().AddDurationFlag(p.withPrefix("read-timeout"), p.opts.ReadTimeout).
		Env(p.generateEnvName("READ_TIMEOUT")).
		Usage("Sets the timeout for socket reads.\nIf reached, commands will fail with a timeout instead of blocking.").
		Default(0)

	app.CLI().AddDurationFlag(p.withPrefix("write-timeout"), p.opts.WriteTimeout).
		Env(p.generateEnvName("WRITE_TIMEOUT")).
		Usage("Sets the timeout for socket writes.\nIf reached, commands will fail with a timeout instead of blocking.").
		Default(0)

	app.CLI().AddBoolFlag(p.withPrefix("retry-statement-timeout"), p.opts.RetryStatementTimeout).
		Env(p.generateEnvName("MAX_RETRIES")).
		Usage("Sets the whether to retry queries cancelled because of statement_timeout.").
		Default(false)

	app.CLI().AddDurationFlag(p.withPrefix("min-retry-backoff"), p.opts.MinRetryBackoff).
		Env(p.generateEnvName("MIN_RETRY_BACKOFF")).
		Usage("Sets the minimum backoff between each retry.\nDefault is 250 milliseconds; -1 disables backoff.").
		Default(0)

	app.CLI().AddDurationFlag(p.withPrefix("max-retry-backoff"), p.opts.MaxRetryBackoff).
		Env(p.generateEnvName("MAX_RETRY_BACKOFF")).
		Usage("Sets the maximum backoff between each retry.\nDefault is 4 seconds; -1 disables backoff.").
		Default(0)

	app.CLI().AddIntFlag(p.withPrefix("pool-size"), p.opts.PoolSize).
		Env(p.generateEnvName("POOL_SIZE")).
		Usage("Sets the maximum number of socket connections.\nDefault is 10 connections per every CPU.").
		Default(0)

	app.CLI().AddIntFlag(p.withPrefix("min-idle-conns"), p.opts.MinIdleConns).
		Env(p.generateEnvName("MIN_IDLE_CONNS")).
		Usage("Minimum number of idle connections which is useful when establishing new connection is slow.").
		Default(0)

	app.CLI().AddDurationFlag(p.withPrefix("max-conn-age"), p.opts.MaxConnAge).
		Env(p.generateEnvName("MAX_CONN_AGE")).
		Usage("Sets the connection age at which client retires (closes) the connection.\nDefault is to not close aged connections.").
		Default(0)

	app.CLI().AddDurationFlag(p.withPrefix("pool-timeout"), p.opts.PoolTimeout).
		Env(p.generateEnvName("POOL_TIMEOUT")).
		Usage("Sets the time for which client waits for free connection if all connections are busy before returning an error.\nDefault is 30 seconds if ReadTimeOut is not defined, otherwise, ReadTimeout + 1 second.").
		Default(0)

	app.CLI().AddDurationFlag(p.withPrefix("idle-timeout"), p.opts.IdleTimeout).
		Env(p.generateEnvName("IDLE_TIMEOUT")).
		Usage("Sets the amount of time after which client closes idle connections.\nShould be less than server's timeout.\nDefault is 5 minutes. -1 disables idle timeout check.").
		Default(0)

	app.CLI().AddDurationFlag(p.withPrefix("idle-check-frequency"), p.opts.IdleCheckFrequency).
		Env(p.generateEnvName("IDLE_CHECK_FREQUENCY")).
		Usage("Sets the frequency of idle checks made by idle connections reaper.\nDefault is 1 minute. -1 disables idle connections reaper, but idle connections are still discarded by the client if IdleTimeout is set.").
		Default(0)

	app.CLI().AddStringFlag(p.withPrefix("migration-dir"), &p.opts.MigrationsDir).
		Env(p.generateEnvName("MIGRATION_DIR")).
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
				return errors.Wrapf(err, "\"%s\" flag is non-string, programmer error, please correct",
					p.withPrefix("connection"))
			}

			if connection == "" {
				config := p.getDBConfig()
				if config.Host == "" {
					return fmt.Errorf("%s flag or %s environment variable required but not set",
						p.withPrefix("connection"), p.generateWithEnvPrefix(envHostName))
				}
				connection = config.getConnectionString()
			}

			c, err := sqlx.Open(driverName, connection)
			if err != nil {
				return fmt.Errorf("failed to db open: %w", err)
			}
			defer c.Close()

			fmt.Println("Start migration")

			if err = p.migration(c.DB, args[0], connection); err != nil {
				return err
			}

			fmt.Println("Migration is completed successfully!")

			os.Exit(0)

			return nil
		},
	}

	migrateCmd.AddStringFlag(p.withPrefix("connection"), nil).
		Env(p.generateEnvName("CONNECTION")).
		Usage(p.genUsage())

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

type dbConfig struct {
	Host     string
	Port     int32
	Database string
	Username string
	Password string
	SSLMode  string
	TimeZone string
}

func (c *dbConfig) getConnectionString() string {
	var connection = fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		c.Username, c.Password, c.Host, c.Port, c.Database)

	var qs = make([]string, 0)

	if c.TimeZone != "" {
		qs = append(qs, fmt.Sprintf("TimeZone=%s", c.TimeZone))
	}
	if c.SSLMode != "" {
		qs = append(qs, fmt.Sprintf("sslmode=%s", c.SSLMode))
	}
	if len(qs) > 0 {
		connection += "?" + strings.Join(qs, "&")
	}

	return connection
}

func (p *plugin) getDBConfig() dbConfig {
	config := dbConfig{Port: defaultPort}

	if host, ok := os.LookupEnv(p.generateWithEnvPrefix(envHostName)); ok {
		config.Host = host
	}
	if port, ok := os.LookupEnv(p.generateWithEnvPrefix(envPortName)); ok {
		if value, err := strconv.ParseInt(port, 10, 32); err == nil {
			config.Port = int32(value)
		}
	}
	if user, ok := os.LookupEnv(p.generateWithEnvPrefix(envUserName)); ok {
		config.Username = user
	}
	if password, ok := os.LookupEnv(p.generateWithEnvPrefix(envPasswordName)); ok {
		config.Password = password
	}
	if name, ok := os.LookupEnv(p.generateWithEnvPrefix(envDatabaseName)); ok {
		config.Database = name
	}
	if sslMode, ok := os.LookupEnv(p.generateWithEnvPrefix(envSslmodeName)); ok {
		config.SSLMode = sslMode
	}
	if timeZone, ok := os.LookupEnv(p.generateWithEnvPrefix(envTimezoneName)); ok {
		config.TimeZone = timeZone
	}

	return config
}

func (p *plugin) withPrefix(name string) string {
	return fmt.Sprintf("%s-%s", p.prefix, name)
}

func (p *plugin) generateEnvName(name string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s", p.prefix, strings.Replace(name, "-", "_", -1)))
}

func (p *plugin) generateWithEnvPrefix(name string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s", p.envPrefix, p.generateEnvName(name)))
}
