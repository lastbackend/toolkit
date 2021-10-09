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
	driverName             = "postgres"
	receivedSubscribeDelay = 60 * time.Second
)

const (
	errMissingConnectionString = "Missing connection string"
)

type clientOptions struct {
	// Sets the connection string for connecting to the database
	Connection string

	// Sets the maximum number of connections in the idle
	// connection pool.
	//
	// If MaxOpenConns is greater than 0 but less than the new MaxIdleConns,
	// then the new MaxIdleConns will be reduced to match the MaxOpenConns limit.
	//
	// If n <= 0, no idle connections are retained.
	//
	// The default max idle connections is currently 2. This may change in
	// a future release.
	MaxIdleConns *int

	// Sets the maximum number of open connections to the database.
	//
	// If MaxIdleConns is greater than 0 and the new MaxOpenConns is less than
	// MaxIdleConns, then MaxIdleConns will be reduced to match the new
	// MaxOpenConns limit.
	//
	// If n <= 0, then there is no limit on the number of open connections.
	// The default is 0 (unlimited).
	MaxOpenConns *int

	// Sets the maximum amount of time a connection may be reused.
	//
	// Expired connections may be closed lazily before reuse.
	//
	// If d <= 0, connections are not closed due to a connection's age.
	ConnMaxLifetime *time.Duration

	// Sets the maximum amount of time a connection may be reused.
	//
	// Expired connections may be closed lazily before reuse.
	//
	// If d <= 0, connections are not closed due to a connection's age.
	ConnMaxIdleTime *time.Duration
}

type client struct {
	c *sqlx.DB

	connection string
}

func newClient() *client {
	return new(client)
}

func (c *client) open(opts clientOptions) error {

	if len(opts.Connection) == 0 {
		return errors.New(errMissingConnectionString)
	}

	conn, err := sqlx.Open(driverName, opts.Connection)
	if err != nil {
		return err
	}

	if opts.MaxIdleConns != nil {
		conn.SetMaxIdleConns(*opts.MaxIdleConns)
	}
	if opts.MaxOpenConns != nil {
		conn.SetMaxOpenConns(*opts.MaxOpenConns)
	}
	if opts.ConnMaxLifetime != nil {
		conn.SetConnMaxLifetime(*opts.ConnMaxLifetime)
	}
	if opts.ConnMaxIdleTime != nil {
		conn.SetConnMaxIdleTime(*opts.ConnMaxIdleTime)
	}
	
	c.connection = opts.Connection
	c.c = conn

	return nil
}

func (c *client) Begin() (ClientTx, error) {
	tx, err := c.c.Begin()
	if err != nil {
		return nil, err
	}
	return &clientTx{c: tx}, nil
}

func (c *client) Beginx() (ClientTxx, error) {
	tx, err := c.c.Beginx()
	if err != nil {
		return nil, err
	}
	return &clientTxx{c: tx}, nil
}

func (c *client) MustBegin() ClientTxx {
	tx := c.c.MustBegin()
	return &clientTxx{c: tx}
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

	_, err := c.c.Exec(publishQuery)
	if err != nil {
		return nil, err
	}

	sqlStatement := `SELECT TRUE FROM Publish($1, $2)`

	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return c.c.ExecContext(ctx, sqlStatement, channel, string(payload))
}

func (c *client) MapperFunc(mf func(string) string) {
	c.c.MapperFunc(mf)
}

func (c *client) Rebind(sqlStatement string) string {
	return c.c.Rebind(sqlStatement)
}

func (c *client) BindNamed(sqlStatement string, arg interface{}) (string, []interface{}, error) {
	return c.c.BindNamed(sqlStatement, arg)
}

func (c *client) NamedQuery(sqlStatement string, arg interface{}) (*sqlx.Rows, error) {
	return c.c.NamedQuery(sqlStatement, arg)
}

func (c *client) NamedExec(sqlStatement string, arg interface{}) (sql.Result, error) {
	return c.c.NamedExec(sqlStatement, arg)
}

func (c *client) Select(dest interface{}, sqlStatement string, args ...interface{}) error {
	return c.c.Select(dest, sqlStatement, args...)
}

func (c *client) Get(dest interface{}, sqlStatement string, args ...interface{}) error {
	return c.c.Get(dest, sqlStatement, args...)
}

func (c *client) PrepareContext(ctx context.Context, sqlStatement string) (*sql.Stmt, error) {
	return c.c.PrepareContext(ctx, sqlStatement)
}

func (c *client) Prepare(sqlStatement string) (*sql.Stmt, error) {
	return c.c.Prepare(sqlStatement)
}

func (c *client) ExecContext(ctx context.Context, sqlStatement string, args ...interface{}) (sql.Result, error) {
	return c.c.ExecContext(ctx, sqlStatement, args...)
}

func (c *client) Exec(sqlStatement string, args ...interface{}) (sql.Result, error) {
	return c.c.Exec(sqlStatement, args...)
}

func (c *client) QueryContext(ctx context.Context, sqlStatement string, args ...interface{}) (*sql.Rows, error) {
	return c.c.QueryContext(ctx, sqlStatement, args...)
}

func (c *client) Query(sqlStatement string, args ...interface{}) (*sql.Rows, error) {
	return c.c.Query(sqlStatement, args...)
}

func (c *client) QueryRowContext(ctx context.Context, sqlStatement string, args ...interface{}) *sql.Row {
	return c.c.QueryRowContext(ctx, sqlStatement, args...)
}

func (c *client) QueryRow(sqlStatement string, args ...interface{}) *sql.Row {
	return c.c.QueryRow(sqlStatement, args...)
}

func (c *client) Queryx(sqlStatement string, args ...interface{}) (*sqlx.Rows, error) {
	return c.c.Queryx(sqlStatement, args...)
}

func (c *client) QueryRowx(sqlStatement string, args ...interface{}) *sqlx.Row {
	return c.c.QueryRowx(sqlStatement, args...)
}

func (c *client) MustExec(sqlStatement string, args ...interface{}) sql.Result {
	return c.c.MustExec(sqlStatement, args...)
}

func (c *client) Preparex(sqlStatement string) (*sqlx.Stmt, error) {
	return c.c.Preparex(sqlStatement)
}

func (c *client) PrepareNamed(sqlStatement string) (*sqlx.NamedStmt, error) {
	return c.c.PrepareNamed(sqlStatement)
}

func (c *client) Close() error {
	return c.c.Close()
}

// ====================================================================================
// Transaction client // ==============================================================
// ====================================================================================
type clientTx struct {
	c *sql.Tx
}

func (c *clientTx) Commit() error {
	return c.c.Commit()
}

func (c *clientTx) Rollback() error {
	return c.c.Rollback()
}

func (c *clientTx) PrepareContext(ctx context.Context, sqlStatement string) (*sql.Stmt, error) {
	return c.c.PrepareContext(ctx, sqlStatement)
}

func (c *clientTx) Prepare(sqlStatement string) (*sql.Stmt, error) {
	return c.c.Prepare(sqlStatement)
}

func (c *clientTx) Exec(sqlStatement string, args ...interface{}) (sql.Result, error) {
	return c.c.Exec(sqlStatement, args...)
}

func (c *clientTx) ExecContext(ctx context.Context, sqlStatement string, args ...interface{}) (sql.Result, error) {
	return c.c.ExecContext(ctx, sqlStatement, args...)
}

func (c *clientTx) QueryContext(ctx context.Context, sqlStatement string, args ...interface{}) (*sql.Rows, error) {
	return c.c.QueryContext(ctx, sqlStatement, args...)
}

func (c *clientTx) Query(sqlStatement string, args ...interface{}) (*sql.Rows, error) {
	return c.c.Query(sqlStatement, args...)
}

func (c *clientTx) QueryRowContext(ctx context.Context, sqlStatement string, args ...interface{}) *sql.Row {
	return c.c.QueryRowContext(ctx, sqlStatement, args...)
}

func (c *clientTx) QueryRow(sqlStatement string, args ...interface{}) *sql.Row {
	return c.c.QueryRow(sqlStatement, args...)
}

type clientTxx struct {
	c *sqlx.Tx
}

func (c *clientTxx) Rollback() error {
	return c.c.Rollback()
}

func (c *clientTxx) Commit() error {
	return c.c.Commit()
}

func (c *clientTxx) PrepareContext(ctx context.Context, sqlStatement string) (*sql.Stmt, error) {
	return c.c.PrepareContext(ctx, sqlStatement)
}

func (c *clientTxx) Prepare(sqlStatement string) (*sql.Stmt, error) {
	return c.c.Prepare(sqlStatement)
}

func (c *clientTxx) StmtContext(ctx context.Context, stmt *sql.Stmt) *sql.Stmt {
	return c.c.StmtContext(ctx, stmt)
}

func (c *clientTxx) Stmt(stmt *sql.Stmt) *sql.Stmt {
	return c.c.Stmt(stmt)
}

func (c *clientTxx) ExecContext(ctx context.Context, sqlStatement string, args ...interface{}) (sql.Result, error) {
	return c.c.ExecContext(ctx, sqlStatement, args...)
}

func (c *clientTxx) Exec(sqlStatement string, args ...interface{}) (sql.Result, error) {
	return c.c.Exec(sqlStatement, args...)
}

func (c *clientTxx) QueryContext(ctx context.Context, sqlStatement string, args ...interface{}) (*sql.Rows, error) {
	return c.c.QueryContext(ctx, sqlStatement, args...)
}

func (c *clientTxx) Query(sqlStatement string, args ...interface{}) (*sql.Rows, error) {
	return c.c.Query(sqlStatement, args...)
}

func (c *clientTxx) QueryRowContext(ctx context.Context, sqlStatement string, args ...interface{}) *sql.Row {
	return c.c.QueryRowContext(ctx, sqlStatement, args...)
}

func (c *clientTxx) QueryRow(sqlStatement string, args ...interface{}) *sql.Row {
	return c.c.QueryRow(sqlStatement, args...)
}

func (c *clientTxx) Rebind(sqlStatement string) string {
	return c.c.Rebind(sqlStatement)
}

func (c *clientTxx) BindNamed(sqlStatement string, arg interface{}) (string, []interface{}, error) {
	return c.c.BindNamed(sqlStatement, arg)
}

func (c *clientTxx) NamedQuery(sqlStatement string, arg interface{}) (*sqlx.Rows, error) {
	return c.c.NamedQuery(sqlStatement, arg)
}

func (c *clientTxx) NamedExec(sqlStatement string, arg interface{}) (sql.Result, error) {
	return c.c.NamedExec(sqlStatement, arg)
}

func (c *clientTxx) Select(dest interface{}, sqlStatement string, args ...interface{}) error {
	return c.c.Select(dest, sqlStatement, args...)
}

func (c *clientTxx) Queryx(sqlStatement string, args ...interface{}) (*sqlx.Rows, error) {
	return c.c.Queryx(sqlStatement, args...)
}

func (c *clientTxx) QueryRowx(sqlStatement string, args ...interface{}) *sqlx.Row {
	return c.c.QueryRowx(sqlStatement, args...)
}

func (c *clientTxx) Get(dest interface{}, sqlStatement string, args ...interface{}) error {
	return c.c.Get(dest, sqlStatement, args...)
}

func (c *clientTxx) MustExec(sqlStatement string, args ...interface{}) sql.Result {
	return c.c.MustExec(sqlStatement, args...)
}

func (c *clientTxx) Preparex(sqlStatement string) (*sqlx.Stmt, error) {
	return c.c.Preparex(sqlStatement)
}

func (c *clientTxx) Stmtx(stmt interface{}) *sqlx.Stmt {
	return c.c.Stmtx(stmt)
}

func (c *clientTxx) NamedStmt(stmt *sqlx.NamedStmt) *sqlx.NamedStmt {
	return c.c.NamedStmt(stmt)
}

func (c *clientTxx) PrepareNamed(sqlStatement string) (*sqlx.NamedStmt, error) {
	return c.c.PrepareNamed(sqlStatement)
}
