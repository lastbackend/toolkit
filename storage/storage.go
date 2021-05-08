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

package storage

import (
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"fmt"
	"strings"
	"time"
)

const (
	defaultPingTimeout = 60 * time.Second
)

type Storage interface {
	Register(h ...Handler)
}

type storage struct {
	*client

	opts Options
}

type Option func(*Options)

func (s *storage) Register(h ...Handler) {
	for _, fn := range h {
		fn.setClient(s.client)
	}
}

func newStorage(opts ...Option) (Storage, error) {
	options := newOptions(opts...)

	cli, err := newClient(options.clientOptions)
	if err != nil {
		return nil, err
	}

	if err := cli.conn.Ping(); err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case <-time.After(defaultPingTimeout):
				if err := cli.conn.Ping(); err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}()

	if options.MigrationsDir != nil {
		// Parse connection string and get database name
		items := strings.Split(options.Connection, " ")

		dbName := ""
		for _, item := range items {
			if strings.HasPrefix(item, "dbname") {
				dbName = strings.Split(item, "=")[1]
				break
			}
		}

		driver, err := postgres.WithInstance(cli.conn.DB, &postgres.Config{})
		m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", *options.MigrationsDir), dbName, driver)
		if err != nil {
			return nil, err
		}

		version, dirty, err := m.Version()
		if err != nil && err != migrate.ErrNilVersion {
			return nil, err
		}
		if dirty {
			if err := m.Force(int(version)); err != nil {
				return nil, err
			}
			if err := m.Down(); err != nil {
				return nil, err
			}
		}

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return nil, err
		}
	}

	return &storage{
		client: cli,
		opts:   options,
	}, nil
}
