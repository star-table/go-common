// Copyright 2019-present Facebook Inc. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package dialect

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"

	entDialect "entgo.io/ent/dialect"

	"github.com/google/uuid"
)

// Dialect names for external usage.
const (
	MySQL    = "mysql"
	SQLite   = "sqlite3"
	Postgres = "postgres"
	Gremlin  = "gremlin"
)

// ExecQuerier wraps the 2 database operations.
type ExecQuerier interface {
	// Exec executes a query that does not return records. For example, in SQL, INSERT or UPDATE.
	// It scans the result into the pointer v. For SQL drivers, it is dialect/sql.Result.
	Exec(ctx context.Context, query string, args, v interface{}) error
	// Query executes a query that returns rows, typically a SELECT in SQL.
	// It scans the result into the pointer v. For SQL drivers, it is *dialect/sql.Rows.
	Query(ctx context.Context, query string, args, v interface{}) error
}

// Driver is the interface that wraps all necessary operations for ent clients.
type Driver interface {
	ExecQuerier
	// Tx starts and returns a new transaction.
	// The provided context is used until the transaction is committed or rolled back.
	Tx(context.Context) (Tx, error)
	// Close closes the underlying connection.
	Close() error
	// Dialect returns the dialect name of the driver.
	Dialect() string
}

// Tx wraps the Exec and Query operations in transaction.
type Tx interface {
	ExecQuerier
	driver.Tx
}

type nopTx struct {
	Driver
}

func (nopTx) Commit() error   { return nil }
func (nopTx) Rollback() error { return nil }

// NopTx returns a Tx with a no-op Commit / Rollback methods wrapping
// the provided Driver d.
func NopTx(d Driver) Tx {
	return nopTx{d}
}

// DebugDriver is a driver that logs all driver operations.
type DebugDriver struct {
	entDialect.Driver                                       // underlying driver.
	hookStart         func(context.Context, ...interface{}) // log function. defaults to log.Println.
	hookEnd           func(context.Context, error, ...interface{})
}

// DebugWithContext gets a driver and a logging function, and returns
// a new debugged-driver that prints all outgoing operations with context.
func DebugWithContext(d entDialect.Driver, hookStart func(context.Context, ...interface{}), hookEnd func(context.Context, error, ...interface{})) entDialect.Driver {
	drv := &DebugDriver{d, hookStart, hookEnd}
	return drv
}

// Exec logs its params and calls the underlying driver Exec method.
func (d *DebugDriver) Exec(ctx context.Context, query string, args, v interface{}) error {
	d.hookStart(ctx, query, args)
	err := d.Driver.Exec(ctx, query, args, v)
	d.hookEnd(ctx, err, query, args)
	return err
}

// Query logs its params and calls the underlying driver Query method.
func (d *DebugDriver) Query(ctx context.Context, query string, args, v interface{}) error {
	d.hookStart(ctx, query, args)
	err := d.Driver.Query(ctx, query, args, v)
	d.hookEnd(ctx, err, query, args)
	return err
}

// Tx adds an log-id for the transaction and calls the underlying driver Tx command.
func (d *DebugDriver) Tx(ctx context.Context) (entDialect.Tx, error) {
	tx, err := d.Driver.Tx(ctx)
	if err != nil {
		return nil, err
	}
	id := uuid.New().String()
	return &DebugTx{tx, id, d.hookStart, d.hookEnd, ctx}, nil
}

// BeginTx adds an log-id for the transaction and calls the underlying driver BeginTx command if it's supported.
func (d *DebugDriver) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	drv, ok := d.Driver.(interface {
		BeginTx(context.Context, *sql.TxOptions) (Tx, error)
	})
	if !ok {
		return nil, fmt.Errorf("Driver.BeginTx is not supported")
	}
	tx, err := drv.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	id := uuid.New().String()
	return &DebugTx{tx, id, d.hookStart, d.hookEnd, ctx}, nil
}

// DebugTx is a transaction implementation that logs all transaction operations.
type DebugTx struct {
	Tx                                              // underlying transaction.
	id        string                                // transaction logging id.
	hookStart func(context.Context, ...interface{}) // log function. defaults to log.Println.
	hookEnd   func(context.Context, error, ...interface{})
	ctx       context.Context // underlying transaction context.
}

// Exec logs its params and calls the underlying transaction Exec method.
func (d *DebugTx) Exec(ctx context.Context, query string, args, v interface{}) error {
	d.hookStart(ctx, query, args)
	err := d.Tx.Exec(ctx, query, args, v)
	d.hookEnd(ctx, err, query, args)
	return err
}

// Query logs its params and calls the underlying transaction Query method.
func (d *DebugTx) Query(ctx context.Context, query string, args, v interface{}) error {
	d.hookStart(ctx, query, args)
	err := d.Tx.Query(ctx, query, args, v)
	d.hookEnd(ctx, err, query, args)
	return err
}

// Commit logs this step and calls the underlying transaction Commit method.
func (d *DebugTx) Commit() error {
	return d.Tx.Commit()
}

// Rollback logs this step and calls the underlying transaction Rollback method.
func (d *DebugTx) Rollback() error {
	return d.Tx.Rollback()
}
