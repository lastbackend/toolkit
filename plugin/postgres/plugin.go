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

package postgres

import (
	"github.com/go-pg/pg/v10"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // nolint
	"github.com/jmoiron/sqlx"
	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/pkg/config"
	"github.com/lastbackend/toolkit/pkg/runtime/types"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"

	"context"
	"fmt"
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
	types.Plugin

	DB() *sqlx.DB
	Register(app toolkit.Service, opts *Options) error
}

type Options struct {
	Name string
}

type Config struct {
	DSN           string `env:"DSN"  envDefault:"" comment:"DSN = postgresql://[user[:password]@][netloc][:port][/dbname][?param1=value1&...] complete connection string"`
	Host          string `env:"HOST" envDefault:"127.0.0.1"  comment:"The host to connect to (required)"`
	Port          int32  `env:"PORT" envDefault:"5432" comment:"The port to bind to (default: 5432)"`
	Database      string `env:"DATABASE" comment:"Database to be selected after connecting to the server."`
	Username      string `env:"USERNAME" comment:"The username to connect with. Not required if using IntegratedSecurity"`
	Password      string `env:"PASSWORD" comment:"The password to connect with. Not required if using IntegratedSecurity"`
	SSLMode       string `env:"SSLMODE" comment:" Whether or not to use SSL mode (disable, allow, prefer, require, verify-ca, verify-full)"`
	TimeZone      string `env:"TIMEZONE" comment:"Sets the session timezone"`
	MigrationsDir string `env:"MIGRATIONS_DIR" comment:"Migrations directory to run migration when plugin is started"`
}

type plugin struct {
	prefix     string
	envPrefix  string
	connection string
	opts       Config

	db *sqlx.DB

	//probe toolkit.Probe
}

func NewPlugin(service toolkit.Service, opts *Options) Plugin {
	p := new(plugin)

	p.prefix = opts.Name
	if p.prefix == "" {
		p.prefix = defaultPrefix
	}

	if err := config.Parse(&p.opts, p.prefix); err != nil {
		return nil
	}

	//p.probe = service.ProbesServer().AddReadinessFunc("")
	err := p.Register(service, opts)
	if err != nil {
		return nil
	}
	return p
}

// Register - registers the plugin implements storage using Postgres as a database storage
func (p *plugin) Register(app toolkit.Service, _ *Options) error {

	//if err := app.PluginRegister(p); err != nil {
	//	return err
	//}

	return nil
}

func (p *plugin) DB() *sqlx.DB {
	return p.db
}

func (p *plugin) Start(ctx context.Context) (err error) {

	if p.opts.DSN == "" {
		if p.opts.Host == "" {
			return fmt.Errorf("%s_DSN or %s_Host environment variable required but not set",
				p.prefix, p.prefix)
		}
		p.opts.DSN = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			p.opts.Username, p.opts.Password, p.opts.Host, p.opts.Port, p.opts.Database, p.opts.SSLMode)
	}

	conn, err := sqlx.Open(driverName, p.opts.DSN)
	if err != nil {
		return err
	}

	//p.probe.AddReadinessFunc(p.prefix, PostgresPingChecker(conn.DB, 1*time.Second))

	p.connection = p.opts.DSN
	p.db = conn

	return nil
}

func (p *plugin) Stop() error {
	return nil
}

//func PostgresPingChecker(database *sql.DB, timeout time.Duration) types.HandleFunc {
//	return func() error {
//		ctx, cancel := context.WithTimeout(context.Background(), timeout)
//		defer cancel()
//		if database == nil {
//			return fmt.Errorf("connection is nil")
//		}
//		return database.PingContext(ctx)
//	}
//}

func (p *plugin) RunMigration() error {

	if p.opts.MigrationsDir == "" {
		return fmt.Errorf("can not run migration: dir is not set: %s", p.opts.MigrationsDir)
	}

	fmt.Printf("\nRun migration from dir: %s", p.opts.MigrationsDir)

	opts, err := pg.ParseURL(p.opts.DSN)
	if err != nil {
		return errors.New(errMissingConnectionString)
	}

	driver, err := postgres.WithInstance(p.DB().DB, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", p.opts.MigrationsDir), opts.Database, driver)
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

	fmt.Printf("\nMigration completed!\n")
	return nil
}
