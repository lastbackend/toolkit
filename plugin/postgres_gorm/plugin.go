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

package postgres_gorm

import (
	"github.com/go-pg/pg/v10"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // nolint
	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/pkg/logger"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	psql "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"context"
	"database/sql"
	"fmt"
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

	DB() *gorm.DB
	Info()
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
	SSLMode       string `env:"SSLMODE"  comment:" Whether or not to use SSL mode (disable, allow, prefer, require, verify-ca, verify-full)"`
	TimeZone      string `env:"TIMEZONE" comment:"Sets the session timezone"`
	MigrationsDir string `env:"MIGRATIONS_DIR" comment:"Migrations directory to run migration when plugin is started"`
}

type plugin struct {
	log     logger.Logger
	runtime toolkit.Runtime

	prefix    string
	envPrefix string

	opts Config

	db *gorm.DB

	//probe toolkit.Probe
}

// Register - registers the plugin implements storage using Postgres as a database storage
func (p *plugin) Register(app toolkit.Runtime, _ *Options) error {
	app.Manager().Plugin().Register(p)
	return nil
}

func (p *plugin) DB() *gorm.DB {
	return p.db
}

func (p *plugin) Info() {
	p.runtime.Manager().Config().Print(p.opts, p.prefix)
}

func (p *plugin) PreStart(ctx context.Context) (err error) {

	p.log.Debug("-- postgresql:plugin: pre start --")

	if p.opts.DSN == "" {
		if p.opts.Host == "" {
			return fmt.Errorf("%s_DSN or %s_Host environment variable required but not set",
				p.prefix, p.prefix)
		}
		p.opts.DSN = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			p.opts.Username, p.opts.Password, p.opts.Host, p.opts.Port, p.opts.Database, p.opts.SSLMode)
	}

	conn, err := sql.Open(driverName, p.opts.DSN)
	if err != nil {
		return err
	}
	db, err := gorm.Open(psql.New(psql.Config{
		Conn: conn,
	}))
	if err != nil {
		return err
	}

	//p.probe.AddReadinessFunc(p.prefix, probes.PostgresPingChecker(conn, 1*time.Second))

	p.db = db

	return nil
}

func (p *plugin) OnStop() error {
	return nil
}

func (p *plugin) RunMigration() error {

	if p.opts.MigrationsDir == "" {
		return fmt.Errorf("can not run migration: dir is not set: %s", p.opts.MigrationsDir)
	}

	fmt.Printf("\nRun migration from dir: %s", p.opts.MigrationsDir)

	opts, err := pg.ParseURL(p.opts.DSN)
	if err != nil {
		return errors.New(errMissingConnectionString)
	}

	db, err := p.db.DB()
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
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

func NewPlugin(service toolkit.Runtime, opts *Options) Plugin {
	p := new(plugin)
	p.runtime = service

	p.log = service.Log()

	p.prefix = opts.Name
	if p.prefix == "" {
		p.prefix = defaultPrefix
	}

	if err := service.Manager().Config().Parse(&p.opts, p.prefix); err != nil {
		return nil
	}

	//p.probe = service.Probe()
	err := p.Register(service, opts)
	if err != nil {
		return nil
	}
	return p
}
