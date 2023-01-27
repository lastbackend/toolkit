/*
Copyright [2014] - [2022] The Last.Backend authors.

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

package postgres

import (
	"database/sql"
	"github.com/go-pg/pg/v10"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/pkg/cmd"
	"github.com/lastbackend/toolkit/pkg/probe"
	"github.com/pkg/errors"
	"strconv"

	_ "github.com/golang-migrate/migrate/v4/source/file" // nolint
	_ "github.com/lib/pq"

	"context"
	"fmt"
	"os"
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

	DB() *sqlx.DB
	Register(app toolkit.Service, opts *Options) error
}

type Options struct {
	Name string
}

type options struct {
	// Sets the connection string for connecting to the database
	Connection string

	// Sets the maximum number of connections in the idle
	// connection pool.
	//
	// If MaxOpenConns is greater than 0 but less than the new MaxIdleConns,
	// then the new MaxIdleConns will be reduced to match the MaxOpenConns limit.
	//
	// If n <= 0, no idle connections are retained.
	//
	// The default max idle connections is currently 2. This may change in
	// a future release.
	MaxIdleConns *int

	// Sets the maximum number of open connections to the database.
	//
	// If MaxIdleConns is greater than 0 and the new MaxOpenConns is less than
	// MaxIdleConns, then MaxIdleConns will be reduced to match the new
	// MaxOpenConns limit.
	//
	// If n <= 0, then there is no limit on the number of open connections.
	// The default is 0 (unlimited).
	MaxOpenConns *int

	// Sets the maximum amount of time a connection may be reused.
	//
	// Expired connections may be closed lazily before reuse.
	//
	// If d <= 0, connections are not closed due to a connection's age.
	ConnMaxLifetime *time.Duration

	// Sets the maximum amount of time a connection may be reused.
	//
	// Expired connections may be closed lazily before reuse.
	//
	// If d <= 0, connections are not closed due to a connection's age.
	ConnMaxIdleTime *time.Duration

	MigrationsDir string
}

type plugin struct {
	prefix     string
	envPrefix  string
	connection string
	opts       options

	db *sqlx.DB

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

func (p *plugin) DB() *sqlx.DB {
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

	conn, err := sqlx.Open(driverName, p.opts.Connection)
	if err != nil {
		return err
	}

	if p.opts.MaxIdleConns != nil {
		conn.SetMaxIdleConns(*p.opts.MaxIdleConns)
	}
	if p.opts.MaxOpenConns != nil {
		conn.SetMaxOpenConns(*p.opts.MaxOpenConns)
	}
	if p.opts.ConnMaxLifetime != nil {
		conn.SetConnMaxLifetime(*p.opts.ConnMaxLifetime)
	}
	if p.opts.ConnMaxIdleTime != nil {
		conn.SetConnMaxIdleTime(*p.opts.ConnMaxIdleTime)
	}

	if p.opts.MigrationsDir != "" {
		fmt.Printf("\nRun migration from dir: %s", p.opts.MigrationsDir)
		if err = p.migration(conn.DB, p.opts.MigrationsDir, p.opts.Connection); err != nil {
			return err
		}
		fmt.Printf("\nMigration completed!\n")
	}

	p.probe.AddReadinessFunc(p.prefix, probe.PostgresPingChecker(conn.DB, 1*time.Second))

	p.connection = p.opts.Connection
	p.db = conn

	return nil
}

func (p *plugin) Stop() error {
	return nil
}

func (p *plugin) genUsage() string {
	return fmt.Sprintf(`PostgreSQL connection string (Ex: host=localhost port=5432 user=<db_user> password=<db_pass> dbname=<db_name>) 
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

	// define plugin connection
	app.CLI().AddStringFlag(p.withPrefix("connection"), &p.opts.Connection).
		Env(p.generateEnvName("CONNECTION")).
		Usage(p.genUsage())

	// define connection max lifetime flag
	app.CLI().AddDurationFlag(p.withPrefix("conn-max-lifetime"), p.opts.ConnMaxLifetime).
		Env(p.generateEnvName("CONN_MAX_LIFETIME")).
		Usage("Sets the maximum amount of time a connection may be reused.\nIf <= 0, connections are not closed due to a connection's age").
		Default(0)

	// define connection max idle flag
	app.CLI().AddDurationFlag(p.withPrefix("conn-max-idle-time"), p.opts.ConnMaxIdleTime).
		Env(p.generateEnvName("CONN_MAX_IDLE_TIME")).
		Usage("Sets the maximum amount of time a connection may be idle.\nIf <= 0, connections are not closed due to a connection's idle time").
		Default(0)

	// define max idle connections flag
	app.CLI().AddIntFlag(p.withPrefix("max-idle-conns"), p.opts.MaxIdleConns).
		Env(p.generateEnvName("MAX_IDLE_CONNS")).
		Usage("Sets the maximum number of connections in the idle connection pool.\nIf <= 0, no idle connections are retained.\n(The default max idle connections is currently 2)").
		Default(0)

	// define max idle connections flag
	app.CLI().AddIntFlag(p.withPrefix("max-open-conns"), p.opts.MaxOpenConns).
		Env(p.generateEnvName("MAX_OPEN_CONNS")).
		Usage("Sets the maximum number of open connections to the database.\nIf <= 0, then there is no limit on the number of open connections.\n(default unlimited)").
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

			conn, err := sqlx.Open(driverName, connection)
			if err != nil {
				return fmt.Errorf("failed to db open: %w", err)
			}
			defer conn.Close()

			fmt.Println("Start migration")

			if err = p.migration(conn.DB, args[0], connection); err != nil {
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
