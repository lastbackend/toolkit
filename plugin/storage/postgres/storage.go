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
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/lastbackend/engine/cmd"
	"github.com/pkg/errors"

	"fmt"
	"strings"
	"time"
)

const (
	defaultPingTimeout = 60 * time.Second
)

type Options struct {
	clientOptions
	MigrationsDir *string
}

type postgresStorage struct {
	client *client

	prefix string
	opts   Options
}

func newStorage(prefix string) *postgresStorage {
	s := new(postgresStorage)
	if len(prefix) == 0 {
		s.prefix = defaultPrefix
	}
	s.prefix = prefix
	s.client = newClient()
	return s
}

func (s *postgresStorage) Name() string {
	return PluginName
}

func (s *postgresStorage) Flags() []cmd.Flag {
	return []cmd.Flag{
		&cmd.StringFlag{
			Name:        s.withPrefix("connection"),
			EnvVar:      s.withEnvPrefix("CONNECTION"),
			Usage:       "PostgreSQL connection string (Ex: host=localhost port=5432 user=<db_user> password=<db_pass> dbname=<db_name>)",
			Required:    true,
			Destination: &s.opts.Connection,
		},
		&cmd.DurationFlag{
			Name:        s.withPrefix("conn-max-lifetime"),
			EnvVar:      s.withEnvPrefix("CONN_MAX_LIFETIME"),
			Usage:       "Sets the maximum amount of time a connection may be reused.\nIf <= 0, connections are not closed due to a connection's age",
			Required:    false,
			Destination: s.opts.ConnMaxLifetime,
		},
		&cmd.DurationFlag{
			Name:        s.withPrefix("conn-max-idle-time"),
			EnvVar:      s.withEnvPrefix("CONN_MAX_IDLE_TIME"),
			Usage:       "Sets the maximum amount of time a connection may be idle.\nIf <= 0, connections are not closed due to a connection's idle time",
			Required:    false,
			Destination: s.opts.ConnMaxIdleTime,
		},
		&cmd.IntFlag{
			Name:        s.withPrefix("max-idle-conns"),
			EnvVar:      s.withEnvPrefix("MAX_IDLE_CONNS"),
			Usage:       "Sets the maximum number of connections in the idle connection pool.\nIf <= 0, no idle connections are retained.\n(The default max idle connections is currently 2)",
			Required:    false,
			Destination: s.opts.MaxIdleConns,
		},
		&cmd.IntFlag{
			Name:        s.withPrefix("max-open-conns"),
			EnvVar:      s.withEnvPrefix("MAX_OPEN_CONNS"),
			Usage:       "Sets the maximum number of open connections to the database.\nIf <= 0, then there is no limit on the number of open connections.\n(default unlimited)",
			Required:    false,
			Destination: s.opts.MaxOpenConns,
		},
	}
}

func (s *postgresStorage) Commands() []cmd.Command {

	migrateCmd := &cmd.Cmd{
		Use:       "migrate [SOURCE_PATH]",
		ShortDesc: "Database migrations",
		Run: func(cmd cmd.Command, args []string) error {

			if len(args) == 0 {
				return fmt.Errorf( "argument \"source path\" is not set, programmer error, please correct")
			}

			c := newClient()

			connection, err := cmd.Flags().GetString(s.withPrefix("connection"))
			if err != nil {
				return errors.Wrapf(err, "\"%s\" flag is non-string, programmer error, please correct", s.withPrefix("connection"))
			}

			if err := c.open(clientOptions{Connection: connection}); err != nil {
				return err
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

			driver, err := postgres.WithInstance(c.c.DB, &postgres.Config{})
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

			return c.c.Close()
		},
	}

	migrateCmd.AddStringFlag(s.withPrefix("connection"), "", "", nil, s.withEnvPrefix("CONNECTION"), true, "PostgreSQL connection string (Ex: host=localhost port=5432 user=<db_user> password=<db_pass> dbname=<db_name>)")

	return []cmd.Command{migrateCmd}
}

func (s *postgresStorage) Start() error {

	if err := s.client.open(s.opts.clientOptions); err != nil {
		return err
	}

	if err := s.client.c.Ping(); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-time.After(defaultPingTimeout):
				if err := s.client.c.Ping(); err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}()

	return nil
}

func (s *postgresStorage) Stop() error {
	return s.client.c.Close()
}

func (s *postgresStorage) getClient() *client {
	return s.client
}

func (s *postgresStorage) withPrefix(name string) string {
	return fmt.Sprintf("%s-%s", s.prefix, name)
}

func (s *postgresStorage) withEnvPrefix(name string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s", s.prefix, name))
}
