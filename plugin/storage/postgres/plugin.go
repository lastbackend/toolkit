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

package postgres

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

	"os"
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
	connection string
	opts       options

	db *sqlx.DB
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

func (p *plugin) DB() *sqlx.DB {
	return p.db
}

func (p *plugin) Start(ctx context.Context) (err error) {
	if len(p.opts.Connection) == 0 {
		return errors.New(errMissingConnectionString)
	}

	if len(p.opts.Connection) == 0 {
		return errors.New(errMissingConnectionString)
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

	p.connection = p.opts.Connection
	p.db = conn

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

	// define plugin connection
	app.CLI().AddStringFlag(p.withPrefix("connection"), &p.opts.Connection).
		Env(p.withEnvPrefix("CONNECTION")).
		Usage("PostgreSQL connection string (Ex: host=localhost port=5432 user=<db_user> password=<db_pass> dbname=<db_name>)").
		Required()

	// define connection max lifetime flag
	app.CLI().AddDurationFlag(p.withPrefix("conn-max-lifetime"), p.opts.ConnMaxLifetime).
		Env(p.withEnvPrefix("CONN_MAX_LIFETIME")).
		Usage("Sets the maximum amount of time a connection may be reused.\nIf <= 0, connections are not closed due to a connection's age").
		Default(0)

	// define connection max idle flag
	app.CLI().AddDurationFlag(p.withPrefix("conn-max-idle-time"), p.opts.ConnMaxIdleTime).
		Env(p.withEnvPrefix("CONN_MAX_IDLE_TIME")).
		Usage("Sets the maximum amount of time a connection may be idle.\nIf <= 0, connections are not closed due to a connection's idle time").
		Default(0)

	// define max idle connections flag
	app.CLI().AddIntFlag(p.withPrefix("max-idle-conns"), p.opts.MaxIdleConns).
		Env(p.withEnvPrefix("MAX_IDLE_CONNS")).
		Usage("Sets the maximum number of connections in the idle connection pool.\nIf <= 0, no idle connections are retained.\n(The default max idle connections is currently 2)").
		Default(0)

	// define max idle connections flag
	app.CLI().AddIntFlag(p.withPrefix("max-open-conns"), p.opts.MaxOpenConns).
		Env(p.withEnvPrefix("MAX_OPEN_CONNS")).
		Usage("Sets the maximum number of open connections to the database.\nIf <= 0, then there is no limit on the number of open connections.\n(default unlimited)").
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
			defer c.Close()

			if err = p.migration(c.DB, args[0], connection); err != nil {
				return err
			}

			os.Exit(0)

			return nil
		},
	}

	migrateCmd.AddStringFlag(p.withPrefix("connection"), nil).
		Env(p.withEnvPrefix("CONNECTION")).
		Usage("PostgreSQL connection string (Ex: host=localhost port=5432 user=<db_user> password=<db_pass> dbname=<db_name>)").
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
