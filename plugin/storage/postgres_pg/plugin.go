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
	"github.com/go-pg/pg/v10"
	"github.com/lastbackend/engine/plugin"
	"io"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"context"
)

const (
	PluginName    = "postgres-pg"
	defaultPrefix = "psql"
)

// Register - registers the plugin implements storage using Postgres as a database storage
func Register(f plugin.RegisterFunc) plugin.CreatorFunc {
	return func(o plugin.Option) interface{} {
		p := newStorage(o.Prefix)
		f(p)
		return p.getClient()
	}
}

type Storage interface {
	Begin() (ClientTx, error)
	BeginContext(ctx context.Context) (ClientTx, error)
	Param(param string) interface{}
	Exec(query interface{}, params ...interface{}) (res pg.Result, err error)
	ExecContext(ctx context.Context, query interface{}, params ...interface{}) (pg.Result, error)
	ExecOne(query interface{}, params ...interface{}) (pg.Result, error)
	ExecOneContext(ctx context.Context, query interface{}, params ...interface{}) (pg.Result, error)
	Query(model, query interface{}, params ...interface{}) (res pg.Result, err error)
	QueryContext(ctx context.Context, model, query interface{}, params ...interface{}) (pg.Result, error)
	QueryOne(model, query interface{}, params ...interface{}) (pg.Result, error)
	QueryOneContext(ctx context.Context, model, query interface{}, params ...interface{}) (pg.Result, error)
	CopyFrom(r io.Reader, query interface{}, params ...interface{}) (res pg.Result, err error)
	CopyTo(w io.Writer, query interface{}, params ...interface{}) (res pg.Result, err error)
	Model(model ...interface{}) *pg.Query
	ModelContext(ctx context.Context, model ...interface{}) *pg.Query
	Prepare(q string) (*pg.Stmt, error)
}

type ClientTx interface {
	Commit() error
	CommitContext(ctx context.Context) error
	Rollback() error
	RollbackContext(ctx context.Context) error
	RunInTransaction(ctx context.Context, fn func(ClientTx) error) error
	Stmt(stmt *pg.Stmt) *pg.Stmt
	Prepare(q string) (*pg.Stmt, error)
	Exec(query interface{}, params ...interface{}) (pg.Result, error)
	ExecContext(c context.Context, query interface{}, params ...interface{}) (pg.Result, error)
	ExecOne(query interface{}, params ...interface{}) (pg.Result, error)
	ExecOneContext(c context.Context, query interface{}, params ...interface{}) (pg.Result, error)
	Query(model interface{}, query interface{}, params ...interface{}) (pg.Result, error)
	QueryContext(c context.Context, model interface{}, query interface{}, params ...interface{}, ) (pg.Result, error)
	QueryOne(model interface{}, query interface{}, params ...interface{}) (pg.Result, error)
	QueryOneContext(c context.Context, model interface{}, query interface{}, params ...interface{}, ) (pg.Result, error)
	Model(model ...interface{}) *pg.Query
	ModelContext(c context.Context, model ...interface{}) *pg.Query
	CopyFrom(r io.Reader, query interface{}, params ...interface{}) (res pg.Result, err error)
	CopyTo(w io.Writer, query interface{}, params ...interface{}) (res pg.Result, err error)
}
