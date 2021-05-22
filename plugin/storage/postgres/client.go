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
	"context"
	"database/sql"
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"fmt"
	"time"
)

const (
	driverName             = "postgres"
	receivedSubscribeDelay = 60 * time.Second
)

const (
	errMissingConnectionString = "Missing connection string"
)

type clientOptions struct {
	MaxIdleConns    *int
	MaxOpenConns    *int
	ConnMaxLifetime *time.Duration
	ConnMaxIdleTime *time.Duration
}

type client struct {
	conn *sqlx.DB

	connection string
}

func newClient() *client {
	return new(client)
}

func (c *client) open(connection string, opts ...clientOptions) error {

	if len(connection) == 0 {
		return errors.New(errMissingConnectionString)
	}

	conn, err := sqlx.Open(driverName, connection)
	if err != nil {
		return err
	}

	if len(opts) > 0 {
		o := opts[0]
		if o.ConnMaxLifetime != nil {
			conn.SetConnMaxLifetime(*o.ConnMaxLifetime)
		}
		if o.ConnMaxIdleTime != nil {
			conn.SetConnMaxIdleTime(*o.ConnMaxIdleTime)
		}
		if o.MaxIdleConns != nil {
			conn.SetMaxIdleConns(*o.MaxIdleConns)
		}
		if o.MaxOpenConns != nil {
			conn.SetMaxOpenConns(*o.MaxOpenConns)
		}

	}

	c.connection = connection
	c.conn = conn

	return nil
}

func (c *client) Begin() (ClientTx, error) {
	tx := c.conn.MustBegin()
	return &clientTx{conn: tx}, nil
}

func (c *client) Beginx() (ClientTx, error) {
	tx, err := c.conn.Beginx()
	if err != nil {
		return nil, err
	}
	return &clientTx{conn: tx}, nil
}

func (c *client) MustBegin() ClientTx {
	tx := c.conn.MustBegin()
	return &clientTx{conn: tx}
}

func (c *client) Subscribe(ctx context.Context, channel string, listener chan string) error {

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Println("ListenerEventType: ", ev, err)
		}
	}

	l := pq.NewListener(c.connection, 10*time.Second, time.Minute, reportProblem)

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

func (c *client) Close() error {
	return c.conn.Close()
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
