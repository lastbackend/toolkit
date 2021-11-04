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
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"io"
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
	// Enable the query logger
	Logger bool

	// Sets the connection string for connecting to the database
	Connection string

	// Sets the dial timeout for establishing new connections.
	// Default is 5 seconds.
	DialTimeout *time.Duration

	// Sets the timeout for socket reads. If reached, commands will fail
	// with a timeout instead of blocking.
	ReadTimeout *time.Duration
	// Sets the timeout for socket writes. If reached, commands will fail
	// with a timeout instead of blocking.
	WriteTimeout *time.Duration

	// Sets the maximum number of retries before giving up.
	// Default is to not retry failed queries.
	MaxRetries *int
	// Sets the whether to retry queries cancelled because of statement_timeout.
	RetryStatementTimeout *bool
	// Sets the minimum backoff between each retry.
	// Default is 250 milliseconds; -1 disables backoff.
	MinRetryBackoff *time.Duration
	// Sets the maximum backoff between each retry.
	// Default is 4 seconds; -1 disables backoff.
	MaxRetryBackoff *time.Duration

	// Sets the maximum number of socket connections.
	// Default is 10 connections per every CPU as reported by runtime.NumCPU.
	PoolSize *int
	// Sets the minimum number of idle connections which is useful when establishing
	// new connection is slow.
	MinIdleConns *int
	// Sets the connection age at which client retires (closes) the connection.
	// It is useful with proxies like PgBouncer and HAProxy.
	// Default is to not close aged connections.
	MaxConnAge *time.Duration
	// Sets the time for which client waits for free connection if all
	// connections are busy before returning an error.
	// Default is 30 seconds if ReadTimeOut is not defined, otherwise,
	// ReadTimeout + 1 second.
	PoolTimeout *time.Duration
	// Sets the amount of time after which client closes idle connections.
	// Should be less than server's timeout.
	// Default is 5 minutes. -1 disables idle timeout check.
	IdleTimeout *time.Duration
	// Sets the frequency of idle checks made by idle connections reaper.
	// Default is 1 minute. -1 disables idle connections reaper,
	// but idle connections are still discarded by the client
	// if IdleTimeout is set.
	IdleCheckFrequency *time.Duration
}

type client struct {
	c *pg.DB

	connection string
}

func newClient() *client {
	return new(client)
}

func (c *client) open(opts clientOptions) error {

	if len(opts.Connection) == 0 {
		return errors.New(errMissingConnectionString)
	}

	opt, err := pg.ParseURL(opts.Connection)
	if err != nil {
		panic(err)
	}

	if opts.DialTimeout != nil {
		opt.DialTimeout = *opts.DialTimeout
	}
	if opts.ReadTimeout != nil {
		opt.ReadTimeout = *opts.ReadTimeout
	}
	if opts.WriteTimeout != nil {
		opt.WriteTimeout = *opts.WriteTimeout
	}
	if opts.MaxRetries != nil {
		opt.MaxRetries = *opts.MaxRetries
	}
	if opts.RetryStatementTimeout != nil {
		opt.RetryStatementTimeout = *opts.RetryStatementTimeout
	}
	if opts.MinRetryBackoff != nil {
		opt.MinRetryBackoff = *opts.MinRetryBackoff
	}
	if opts.MaxRetryBackoff != nil {
		opt.MaxRetryBackoff = *opts.MaxRetryBackoff
	}
	if opts.PoolSize != nil {
		opt.PoolSize = *opts.PoolSize
	}
	if opts.MinIdleConns != nil {
		opt.MinIdleConns = *opts.MinIdleConns
	}
	if opts.MaxConnAge != nil {
		opt.MaxConnAge = *opts.MaxConnAge
	}
	if opts.PoolTimeout != nil {
		opt.PoolTimeout = *opts.PoolTimeout
	}
	if opts.IdleTimeout != nil {
		opt.IdleTimeout = *opts.IdleTimeout
	}
	if opts.IdleCheckFrequency != nil {
		opt.IdleCheckFrequency = *opts.IdleCheckFrequency
	}

	c.connection = opts.Connection
	db := pg.Connect(opt)

	if opts.Logger {
		db.AddQueryHook(dbLogger{})
	}

	c.c = db

	return nil
}

// ====================================================================================
// Client =============================================================================
// ====================================================================================
func (c *client) Begin() (ClientTx, error) {
	tx, err := c.c.Begin()
	if err != nil {
		return nil, err
	}
	return &clientTx{c: tx}, nil
}

func (c *client) BeginContext(ctx context.Context) (ClientTx, error) {
	tx, err := c.c.BeginContext(ctx)
	if err != nil {
		return nil, err
	}
	return &clientTx{c: tx}, nil
}

func (c *client) RunInTransaction(ctx context.Context, fn func(ClientTx) error) error {
	return c.c.RunInTransaction(ctx, func(tx *pg.Tx) error {
		return fn(&clientTx{c: tx})
	})
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

func (c *client) Publish(ctx context.Context, channel string, data json.RawMessage) (pg.Result, error) {

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

func (c *client) Model(model ...interface{}) *pg.Query {
	return c.c.Model(model...)
}

func (c *client) ModelContext(ctx context.Context, model ...interface{}) *pg.Query {
	return c.c.ModelContext(ctx, model...)
}

func (c *client) Exec(query interface{}, params ...interface{}) (res pg.Result, err error) {
	return c.c.Exec(query, params...)
}

func (c *client) ExecContext(ctx context.Context, query interface{}, params ...interface{}) (pg.Result, error) {
	return c.c.ExecContext(ctx, query, params...)
}

func (c *client) ExecOne(query interface{}, params ...interface{}) (pg.Result, error) {
	return c.c.ExecOne(query, params...)
}

func (c *client) ExecOneContext(ctx context.Context, query interface{}, params ...interface{}) (pg.Result, error) {
	return c.c.ExecOneContext(ctx, query, params...)
}

func (c *client) Query(model, query interface{}, params ...interface{}) (res pg.Result, err error) {
	return c.c.Query(model, query, params...)
}

func (c *client) QueryContext(ctx context.Context, model, query interface{}, params ...interface{}) (pg.Result, error) {
	return c.c.QueryContext(ctx, model, query, params...)
}

func (c *client) QueryOne(model, query interface{}, params ...interface{}) (pg.Result, error) {
	return c.c.QueryOne(model, query, params...)
}

func (c *client) QueryOneContext(ctx context.Context, model, query interface{}, params ...interface{}) (pg.Result, error) {
	return c.c.QueryOneContext(ctx, model, query, params...)
}

func (c *client) CopyFrom(r io.Reader, query interface{}, params ...interface{}) (res pg.Result, err error) {
	return c.c.CopyFrom(r, query, params...)
}

func (c *client) CopyTo(w io.Writer, query interface{}, params ...interface{}) (res pg.Result, err error) {
	return c.c.CopyTo(w, query, params...)
}

func (c *client) Prepare(q string) (*pg.Stmt, error) {
	return c.c.Prepare(q)
}

func (c *client) Context() context.Context {
	return c.c.Context()
}

func (c *client) Formatter() orm.QueryFormatter {
	return c.c.Formatter()
}

// ====================================================================================
// Transaction client =================================================================
// ====================================================================================
type clientTx struct {
	c *pg.Tx
}

func (c *clientTx) Commit() error {
	return c.c.Commit()
}

func (c *clientTx) CommitContext(ctx context.Context) error {
	return c.c.CommitContext(ctx)
}

func (c *clientTx) Rollback() error {
	return c.c.Rollback()
}

func (c *clientTx) RollbackContext(ctx context.Context) error {
	return c.c.RollbackContext(ctx)
}

func (c *clientTx) RunInTransaction(ctx context.Context, fn func(ClientTx) error) error {
	return c.c.RunInTransaction(ctx, func(tx *pg.Tx) error {
		return fn(&clientTx{c: tx})
	})
}

func (c *clientTx) Model(model ...interface{}) *pg.Query {
	return c.c.Model(model...)
}

func (c *clientTx) ModelContext(ctx context.Context, model ...interface{}) *pg.Query {
	return c.c.ModelContext(ctx, model...)
}

func (c *clientTx) Stmt(stmt *pg.Stmt) *pg.Stmt {
	return c.c.Stmt(stmt)
}

func (c *clientTx) Prepare(q string) (*pg.Stmt, error) {
	return c.c.Prepare(q)
}

func (c *clientTx) Exec(query interface{}, params ...interface{}) (pg.Result, error) {
	return c.c.Exec(query, params...)
}

func (c *clientTx) ExecContext(ctx context.Context, query interface{}, params ...interface{}) (pg.Result, error) {
	return c.c.ExecContext(ctx, query, params...)
}

func (c *clientTx) ExecOne(query interface{}, params ...interface{}) (pg.Result, error) {
	return c.c.ExecOne(query, params...)
}

func (c *clientTx) ExecOneContext(ctx context.Context, query interface{}, params ...interface{}) (pg.Result, error) {
	return c.c.ExecOneContext(ctx, query, params...)
}

func (c *clientTx) Query(model interface{}, query interface{}, params ...interface{}) (pg.Result, error) {
	return c.c.Query(model, query, params...)
}

func (c *clientTx) QueryContext(ctx context.Context, model interface{}, query interface{}, params ...interface{}) (pg.Result, error) {
	return c.c.QueryContext(ctx, model, query, params...)
}

func (c *clientTx) QueryOne(model interface{}, query interface{}, params ...interface{}) (pg.Result, error) {
	return c.c.QueryOne(model, query, params...)
}

func (c *clientTx) QueryOneContext(ctx context.Context, model interface{}, query interface{}, params ...interface{}) (pg.Result, error) {
	return c.c.QueryOneContext(ctx, model, query, params...)
}

func (c *clientTx) CopyFrom(r io.Reader, query interface{}, params ...interface{}) (res pg.Result, err error) {
	return c.c.CopyFrom(r, query, params...)
}

func (c *clientTx) CopyTo(w io.Writer, query interface{}, params ...interface{}) (res pg.Result, err error) {
	return c.c.CopyTo(w, query, params...)
}

func (c *clientTx) Context() context.Context {
	return c.c.Context()
}

func (c *clientTx) Formatter() orm.QueryFormatter {
	return c.c.Formatter()
}
