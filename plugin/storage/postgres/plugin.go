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

	DB() *sqlx.DB
	Register(app engine.Service, opts *Options) error
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

	MigrationsDir *string
}

type plugin struct {
	prefix     string
	connection string
	opts       options

	db *sqlx.DB
}

func NewPlugin(app engine.Service, opts *Options) Plugin {
	db := new(plugin)
	db.Register(app, opts)
	return db
}

// Register - registers the plug implements storage using Postgres as a database storage
func (p *plugin) Register(app engine.Service, opts *Options) error {

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

func (p *plugin) addFlags(app engine.Service) {
	app.CLI().AddStringFlag(p.withPrefix("connection"), "", "", &p.opts.Connection, p.withEnvPrefix("CONNECTION"), true, "PostgreSQL connection string (Ex: host=localhost port=5432 user=<db_user> password=<db_pass> dbname=<db_name>)")
	app.CLI().AddDurationFlag(p.withPrefix("conn-max-lifetime"), "", 0, p.opts.ConnMaxLifetime, p.withEnvPrefix("CONN_MAX_LIFETIME"), true, "Sets the maximum amount of time a connection may be reused.\nIf <= 0, connections are not closed due to a connection's age")
	app.CLI().AddDurationFlag(p.withPrefix("conn-max-idle-time"), "", 0, p.opts.ConnMaxIdleTime, p.withEnvPrefix("CONN_MAX_IDLE_TIME"), true, "Sets the maximum amount of time a connection may be idle.\nIf <= 0, connections are not closed due to a connection's idle time")
	app.CLI().AddIntFlag(p.withPrefix("max-idle-conns"), "", 0, p.opts.MaxIdleConns, p.withEnvPrefix("MAX_IDLE_CONNS"), true, "Sets the maximum number of connections in the idle connection pool.\nIf <= 0, no idle connections are retained.\n(The default max idle connections is currently 2)")
	app.CLI().AddIntFlag(p.withPrefix("max-open-conns"), "", 0, p.opts.MaxOpenConns, p.withEnvPrefix("MAX_OPEN_CONNS"), true, "Sets the maximum number of open connections to the database.\nIf <= 0, then there is no limit on the number of open connections.\n(default unlimited)")
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

			c, err := sqlx.Open(driverName, connection)
			if err != nil {
				return fmt.Errorf("failed to db open: %w", err)
			}

			// Parse connection string and get database name
			items := strings.Split(connection, " ")

			dbName := ""
			for _, item := range items {
				if strings.HasPrefix(item, "dbname") {
					dbName = strings.Split(item, "=")[1]
					break
				}
			}

			driver, err := postgres.WithInstance(c.DB, &postgres.Config{})
			m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", args[0]), dbName, driver)
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

	migrateCmd.AddStringFlag(p.withPrefix("connection"), "", "", nil, p.withEnvPrefix("CONNECTION"), true, "PostgreSQL connection string (Ex: host=localhost port=5432 user=<db_user> password=<db_pass> dbname=<db_name>)")

	app.CLI().AddCommand(migrateCmd)
}
