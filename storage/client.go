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
	"database/sql"
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"context"
	"fmt"
	"time"
)

const (
	receivedSubscribeDelay = 60 * time.Second
)

const (
	errMissingConnectionString = "Missing connection string"
)

type Client interface {
	Begin() (ClientTX, error)
	Beginx() (ClientTX, error)
	Subscribe(ctx context.Context, channel string, listener chan string) error
	Publish(ctx context.Context, channel string, data json.RawMessage) error
	MustBegin() ClientTX
	MapperFunc(mf func(string) string)
	Rebind(query string) string
	BindNamed(query string, arg interface{}) (string, []interface{}, error)
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
	NamedExec(query string, arg interface{}) (sql.Result, error)
	Select(dest interface{}, query string, args ...interface{}) error
	Get(dest interface{}, query string, args ...interface{}) error
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Prepare(query string) (*sql.Stmt, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryRow(query string, args ...interface{}) *sql.Row
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	QueryRowx(query string, args ...interface{}) *sqlx.Row
	MustExec(query string, args ...interface{}) sql.Result
	Preparex(query string) (*sqlx.Stmt, error)
	PrepareNamed(query string) (*sqlx.NamedStmt, error)
}

type ClientTX interface {
	Commit() error
	Rollback() error
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Prepare(query string) (*sql.Stmt, error)
	StmtContext(ctx context.Context, stmt *sql.Stmt) *sql.Stmt
	Stmt(stmt *sql.Stmt) *sql.Stmt
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryRow(query string, args ...interface{}) *sql.Row
	Rebind(query string) string
	BindNamed(query string, arg interface{}) (string, []interface{}, error)
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
	NamedExec(query string, arg interface{}) (sql.Result, error)
	Select(dest interface{}, query string, args ...interface{}) error
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	QueryRowx(query string, args ...interface{}) *sqlx.Row
	Get(dest interface{}, query string, args ...interface{}) error
	MustExec(query string, args ...interface{}) sql.Result
	Preparex(query string) (*sqlx.Stmt, error)
	Stmtx(stmt interface{}) *sqlx.Stmt
	NamedStmt(stmt *sqlx.NamedStmt) *sqlx.NamedStmt
	PrepareNamed(query string) (*sqlx.NamedStmt, error)
}

type clientOptions struct {
	Connection      string
	MaxIdleConns    *int
	MaxOpenConns    *int
	ConnMaxLifetime *time.Duration
	ConnMaxIdleTime *time.Duration
}

type client struct {
	opts clientOptions
	conn *sqlx.DB
}

func newClient(opts clientOptions) (*client, error) {
	c := new(client)

	if len(opts.Connection) == 0 {
		return nil, errors.New(errMissingConnectionString)
	}

	conn, err := sqlx.Open(driverName, opts.Connection)
	if err != nil {
		return nil, err
	}

	if opts.ConnMaxLifetime != nil {
		conn.SetConnMaxLifetime(*opts.ConnMaxLifetime)
	}
	if opts.ConnMaxIdleTime != nil {
		conn.SetConnMaxIdleTime(*opts.ConnMaxIdleTime)
	}
	if opts.MaxIdleConns != nil {
		conn.SetMaxIdleConns(*opts.MaxIdleConns)
	}
	if opts.MaxOpenConns != nil {
		conn.SetMaxOpenConns(*opts.MaxOpenConns)
	}

	c.opts = opts
	c.conn = conn

	return c, nil
}

func (c *client) Begin() (*clientTx, error) {
	tx := c.conn.MustBegin()
	return &clientTx{conn: tx}, nil
}

func (c *client) Beginx() (*clientTx, error) {
	tx, err := c.conn.Beginx()
	if err != nil {
		return nil, err
	}
	return &clientTx{conn: tx}, nil
}

func (c *client) MustBegin() *clientTx {
	tx := c.conn.MustBegin()
	return &clientTx{conn: tx}
}

func (c *client) Subscribe(ctx context.Context, channel string, listener chan string) error {

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Println("ListenerEventType: ", ev, err)
		}
	}

	l := pq.NewListener(c.opts.Connection, 10*time.Second, time.Minute, reportProblem)

	if err := l.Listen(channel); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case n := <-l.Notify:
			if n == nil {
				continue
			}
			listener <- n.Extra
		case <-time.After(receivedSubscribeDelay):
			go func() {
				if err := l.Ping(); err != nil {
					fmt.Println(err)
				}
			}()
		}
	}
}

func (c *client) Publish(ctx context.Context, channel string, data json.RawMessage) (sql.Result, error) {

	publishQuery := `
		CREATE OR REPLACE FUNCTION Publish(channel TEXT, payload TEXT)
		  RETURNS VOID AS
		$BODY$
		BEGIN
		  PERFORM pg_notify(channel, payload);
		END;
		$BODY$
		  LANGUAGE 'plpgsql'
		  VOLATILE
		  COST 100;`

	_, err := c.conn.Exec(publishQuery)
	if err != nil {
		return nil, err
	}

	query := `SELECT TRUE FROM Publish($1, $2)`

	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return c.conn.ExecContext(ctx, query, channel, string(payload))
}

func (c *client) MapperFunc(mf func(string) string) {
	c.conn.MapperFunc(mf)
}

func (c *client) Rebind(query string) string {
	return c.conn.Rebind(query)
}

func (c *client) BindNamed(query string, arg interface{}) (string, []interface{}, error) {
	return c.conn.BindNamed(query, arg)
}

func (c *client) NamedQuery(query string, arg interface{}) (*sqlx.Rows, error) {
	return c.conn.NamedQuery(query, arg)
}

func (c *client) NamedExec(query string, arg interface{}) (sql.Result, error) {
	return c.conn.NamedExec(query, arg)
}

func (c *client) Select(dest interface{}, query string, args ...interface{}) error {
	return c.conn.Select(dest, query, args)
}

func (c *client) Get(dest interface{}, query string, args ...interface{}) error {
	return c.conn.Get(dest, query, args)
}

func (c *client) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return c.conn.PrepareContext(ctx, query)
}

func (c *client) Prepare(query string) (*sql.Stmt, error) {
	return c.conn.Prepare(query)
}

func (c *client) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return c.conn.ExecContext(ctx, query, args)
}

func (c *client) Exec(query string, args ...interface{}) (sql.Result, error) {
	return c.conn.Exec(query, args)
}

func (c *client) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return c.conn.QueryContext(ctx, query, args)
}

func (c *client) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return c.conn.Query(query, args)
}

func (c *client) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return c.conn.QueryRowContext(ctx, query, args)
}

func (c *client) QueryRow(query string, args ...interface{}) *sql.Row {
	return c.conn.QueryRow(query, args)
}

func (c *client) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	return c.conn.Queryx(query, args)
}

func (c *client) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	return c.conn.QueryRowx(query, args)
}

func (c *client) MustExec(query string, args ...interface{}) sql.Result {
	return c.conn.MustExec(query, args)
}

func (c *client) Preparex(query string) (*sqlx.Stmt, error) {
	return c.conn.Preparex(query)
}

func (c *client) PrepareNamed(query string) (*sqlx.NamedStmt, error) {
	return c.conn.PrepareNamed(query)
}

// ====================================================================================
// Transaction client // ==============================================================
// ====================================================================================
type clientTx struct {
	conn *sqlx.Tx
}

func (c *clientTx) Rollback() error {
	return c.conn.Rollback()
}

func (c *clientTx) Commit() error {
	return c.conn.Commit()
}

func (c *clientTx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return c.conn.PrepareContext(ctx, query)
}

func (c *clientTx) Prepare(query string) (*sql.Stmt, error) {
	return c.conn.Prepare(query)
}

func (c *clientTx) StmtContext(ctx context.Context, stmt *sql.Stmt) *sql.Stmt {
	return c.conn.StmtContext(ctx, stmt)
}

func (c *clientTx) Stmt(stmt *sql.Stmt) *sql.Stmt {
	return c.conn.Stmt(stmt)
}

func (c *clientTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return c.conn.ExecContext(ctx, query, args)
}

func (c *clientTx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return c.conn.Exec(query, args)
}

func (c *clientTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return c.conn.QueryContext(ctx, query, args)
}

func (c *clientTx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return c.conn.Query(query, args)
}

func (c *clientTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return c.conn.QueryRowContext(ctx, query, args)
}

func (c *clientTx) QueryRow(query string, args ...interface{}) *sql.Row {
	return c.conn.QueryRow(query, args)
}

func (c *clientTx) Rebind(query string) string {
	return c.conn.Rebind(query)
}

func (c *clientTx) BindNamed(query string, arg interface{}) (string, []interface{}, error) {
	return c.conn.BindNamed(query, arg)
}

func (c *clientTx) NamedQuery(query string, arg interface{}) (*sqlx.Rows, error) {
	return c.conn.NamedQuery(query, arg)
}

func (c *clientTx) NamedExec(query string, arg interface{}) (sql.Result, error) {
	return c.conn.NamedExec(query, arg)
}

func (c *clientTx) Select(dest interface{}, query string, args ...interface{}) error {
	return c.conn.Select(dest, query, args)
}

func (c *clientTx) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	return c.conn.Queryx(query, args)
}

func (c *clientTx) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	return c.conn.QueryRowx(query, args)
}

func (c *clientTx) Get(dest interface{}, query string, args ...interface{}) error {
	return c.conn.Get(dest, query, args)
}

func (c *clientTx) MustExec(query string, args ...interface{}) sql.Result {
	return c.conn.MustExec(query, args)
}

func (c *clientTx) Preparex(query string) (*sqlx.Stmt, error) {
	return c.conn.Preparex(query)
}

func (c *clientTx) Stmtx(stmt interface{}) *sqlx.Stmt {
	return c.conn.Stmtx(stmt)
}

func (c *clientTx) NamedStmt(stmt *sqlx.NamedStmt) *sqlx.NamedStmt {
	return c.conn.NamedStmt(stmt)
}

func (c *clientTx) PrepareNamed(query string) (*sqlx.NamedStmt, error) {
	return c.conn.PrepareNamed(query)
}
