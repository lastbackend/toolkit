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
	"os"

	"github.com/go-pg/pg/v10"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/cmd"
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

const (
	dbhost     = "HOST"
	dbport     = "PORT"
	dbuser     = "USERNAME"
	dbpass     = "PASSWORD"
	dbname     = "DATABASE"
	dbsslmode  = "SSL_MODE"
	dbtimezone = "TIMEZONE"
)

const (
	errMissingConnectionString = "Missing connection string"
)

type Plugin interface {
	toolkit.Plugin

	DB() *gorm.DB
	Register(app toolkit.Service, opts *Options) error
}

type Options struct {
	Name string
}

type options struct {
	Connection    string
	MigrationsDir string
}

type plugin struct {
	prefix string
	opts   options

	db *gorm.DB
}

func NewPlugin(app toolkit.Service, opts *Options) Plugin {
	p := new(plugin)
	err := p.Register(app, opts)
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

func (p *plugin) DB() *gorm.DB {
	return p.db
}

func (p *plugin) Start(ctx context.Context) (err error) {

	if p.opts.Connection == "" {
		config := p.dbConfig()
		if _, ok := config[dbhost]; !ok {
			return fmt.Errorf("%s flag or %s environment variable required but not set", p.withPrefix("connection"), p.withEnvPrefix(dbhost))
		}
		p.opts.Connection = config.getConnectionString()
	}

	sqlDB, err := sql.Open(driverName, p.opts.Connection)
	if err != nil {
		return err
	}
	db, err := gorm.Open(psql.New(psql.Config{
		Conn: sqlDB,
	}))
	if err != nil {
		return err
	}

	if p.opts.MigrationsDir != "" {
		fmt.Printf("\nRun migration from dir: %s", p.opts.MigrationsDir)
		if err = p.migration(sqlDB, p.opts.MigrationsDir, p.opts.Connection); err != nil {
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
		Usage(fmt.Sprintf(`PostgreSQL connection string (Ex: postgres://user:pass@localhost:5432/db_name) 
			or use environment variables: 
			%s - The host to connect to (required), 
			%s - The port to bind to (default: 5432), 
			%s - The username to connect with. Not required if using IntegratedSecurity., 
			%s - The password to connect with. Not required if using IntegratedSecurity., 
			%s - The database to connect to, 
			%s - Whether or not to use SSL, 
			%s - Sets the session timezone`,
			p.withEnvPrefix(dbhost), p.withEnvPrefix(dbport), p.withEnvPrefix(dbuser),
			p.withEnvPrefix(dbpass), p.withEnvPrefix(dbname), p.withEnvPrefix(dbsslmode), p.withEnvPrefix(dbtimezone)))

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

			if connection == "" {
				config := p.dbConfig()
				if _, ok := config[dbhost]; !ok {
					return fmt.Errorf("%s flag or %s environment variable required but not set", p.withPrefix("connection"), p.withEnvPrefix(dbhost))
				}
				p.opts.Connection = config.getConnectionString()
			}

			sqlDB, err := sql.Open(driverName, connection)
			if err != nil {
				return fmt.Errorf("failed to db open: %w", err)
			}
			defer sqlDB.Close()

			if err = p.migration(sqlDB, args[0], connection); err != nil {
				return err
			}

			os.Exit(0)

			return nil
		},
	}

	migrateCmd.AddStringFlag(p.withPrefix("connection"), nil).
		Env(p.withEnvPrefix("CONNECTION")).
		Usage(fmt.Sprintf(`PostgreSQL connection string (Ex: postgres://user:pass@localhost:5432/db_name) 
		or use environment variables: %s, %s, %s, %s, %s, %s, %s`,
			p.withEnvPrefix(dbhost), p.withEnvPrefix(dbport), p.withEnvPrefix(dbuser),
			p.withEnvPrefix(dbpass), p.withEnvPrefix(dbname), p.withEnvPrefix(dbsslmode),
			p.withEnvPrefix(dbtimezone)))

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

type dbConfig map[string]string

func (d dbConfig) getConnectionString() string {
	var connection = ""
	if v, ok := d[dbhost]; ok {
		connection += fmt.Sprintf("host=%s", v)
	}
	if v, ok := d[dbport]; ok {
		connection += fmt.Sprintf(" port=%s", v)
	}
	if v, ok := d[dbuser]; ok {
		connection += fmt.Sprintf(" user=%s", v)
	}
	if v, ok := d[dbpass]; ok {
		connection += fmt.Sprintf(" password=%s", v)
	}
	if v, ok := d[dbname]; ok {
		connection += fmt.Sprintf(" dbname=%s", v)
	}
	if v, ok := d[dbsslmode]; ok {
		connection += fmt.Sprintf(" sslmode=%s", v)
	}
	if v, ok := d[dbtimezone]; ok {
		connection += fmt.Sprintf(" TimeZone=%s", v)
	}
	return connection
}

func (p *plugin) dbConfig() dbConfig {
	conf := make(map[string]string)
	if host, ok := os.LookupEnv(p.withEnvPrefix(dbhost)); ok {
		conf[dbhost] = host
	}
	if port, ok := os.LookupEnv(p.withEnvPrefix(dbport)); ok {
		conf[dbport] = port
	}
	if user, ok := os.LookupEnv(p.withEnvPrefix(dbuser)); ok {
		conf[dbuser] = user
	}
	if password, ok := os.LookupEnv(p.withEnvPrefix(dbpass)); ok {
		conf[dbpass] = password
	}
	if name, ok := os.LookupEnv(p.withEnvPrefix(dbname)); ok {
		conf[dbname] = name
	}
	if sslMode, ok := os.LookupEnv(p.withEnvPrefix(dbsslmode)); ok {
		conf[dbsslmode] = sslMode
	}
	if timeZone, ok := os.LookupEnv(p.withEnvPrefix(dbtimezone)); ok {
		conf[dbtimezone] = timeZone
	}
	return dbConfig(conf)
}
