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

package postgres_gorm

import (
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
	"github.com/lastbackend/engine"
	"github.com/lastbackend/engine/cmd"
	"github.com/pkg/errors"
	psql "gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/golang-migrate/migrate/v4/source/file" // nolint
	_ "github.com/lib/pq"

	"context"
	"database/sql"
	"fmt"
	"strings"
)

const (
	defaultPrefix = "psql"
	driverName    = "postgres"
)

type Plugin interface {
	engine.Plugin

	DB() *gorm.DB
	Register(app engine.Service, opts *Options) error
}

type Options struct {
	Name string
}

type options struct {
	Connection    string
	MigrationsDir *string
}

type plugin struct {
	prefix string
	opts   options

	db *gorm.DB
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
	p.addCommands(app)

	if err := app.PluginRegister(p); err != nil {
		return err
	}

	return nil
}

func (p *plugin) DB() *gorm.DB {
	return p.db
}

func (p *plugin) Start(ctx context.Context) (err error) {
	sqlDB, err := sql.Open("postgres", p.opts.Connection)
	if err != nil {
		return err
	}
	db, err := gorm.Open(psql.New(psql.Config{
		Conn: sqlDB,
	}))
	if err != nil {
		return err
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
	app.CLI().AddStringFlag(p.withPrefix("connection"), &p.opts.Connection).
		Env(p.withEnvPrefix("CONNECTION")).
		Usage("PostgreSQL connection string (Ex: postgres://user:pass@localhost:5432/db_name)").
		Required()
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
			if err != nil {
				return err
			}

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

	migrateCmd.AddStringFlag(p.withPrefix("connection"), nil).
		Env(p.withEnvPrefix("CONNECTION")).
		Usage("PostgreSQL connection string (Ex: host=localhost port=5432 user=<db_user> password=<db_pass> dbname=<db_name>)").
		Required()

	app.CLI().AddCommand(migrateCmd)
}
