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

	ConnectionString string
	MigrationsDir    *string
}

type postgresStorage struct {
	client *client

	prefix string
	opts   Options
}

func newPostgresStorage(prefix string) *postgresStorage {
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
			EnvVars:     []string{s.withEnvPrefix("CONNECTION")},
			Usage:       "PostgreSQL connection string",
			Required:    true,
			Destination: &s.opts.ConnectionString,
		},
	}
}

func (s *postgresStorage) Commands() []cmd.Command {

	migrateCmd := &cmd.Cmd{
		Use:       "migrate [SOURCE_PATH]",
		ShortDesc: "Database migrations",
		Run: func(cmd cmd.Command, args []string) error {

			c := newClient()

			psqlConnection, err := cmd.Flags().GetString(s.withPrefix("connection"))
			if err != nil {
				return errors.Wrapf(err, "\"%s\" flag is non-string, programmer error, please correct", s.withPrefix("connection"))
			}

			if len(args) == 0 {
				return errors.Wrapf(err, "argument \"source path\" is not set, programmer error, please correct")
			}

			if err := c.open(psqlConnection); err != nil {
				return err
			}

			// Parse connection string and get database name
			items := strings.Split(psqlConnection, " ")

			dbName := ""
			for _, item := range items {
				if strings.HasPrefix(item, "dbname") {
					dbName = strings.Split(item, "=")[1]
					break
				}
			}

			driver, err := postgres.WithInstance(c.conn.DB, &postgres.Config{})
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

	migrateCmd.AddStringFlag(s.withPrefix("connection"), "", "", nil, []string{s.withEnvPrefix("CONNECTION")}, true, "PostgreSQL connection string")

	return []cmd.Command{migrateCmd}
}

func (s *postgresStorage) Start() error {

	if err := s.client.open(s.opts.ConnectionString); err != nil {
		return err
	}

	if err := s.client.conn.Ping(); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-time.After(defaultPingTimeout):
				if err := s.client.conn.Ping(); err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}()

	return nil
}

func (s *postgresStorage) Stop() error {
	return s.client.Close()
}

func (s *postgresStorage) getClient() Postgres {
	return s.client
}

func (s *postgresStorage) withPrefix(name string) string {
	return fmt.Sprintf("%s-%s", s.prefix, name)
}

func (s *postgresStorage) withEnvPrefix(name string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s", s.prefix, name))
}
