package crud

import (
	"context"
	stdsql "database/sql"
	"fmt"
	"log/slog"
	"time"
)

type Tx struct {
	Client *stdsql.Tx
	Logger *slog.Logger
}

// Execute any SQL query on the transaction client. Returns sql.Result.
func (tx *Tx) Exec(ctx context.Context, sql string, params ...interface{}) (stdsql.Result, error) {
	start := time.Now()
	result, err := tx.Client.ExecContext(ctx, sql, params...)
	if tx.Logger != nil {
		tx.Logger.DebugContext(ctx, "Executed SQL query", "sql", sql, "took", time.Since(start))
	}
	return result, err
}

// Execute any SQL query on the transaction client. Returns sql.Rows.
func (tx *Tx) Query(ctx context.Context, sql string, params ...interface{}) (*stdsql.Rows, error) {
	start := time.Now()
	result, err := tx.Client.QueryContext(ctx, sql, params...)
	if tx.Logger != nil {
		tx.Logger.DebugContext(ctx, "Ran SQL query", "sql", sql, "took", time.Since(start))
	}
	return result, err
}

// Commit the transaction.
func (tx *Tx) Commit(ctx context.Context) error {
	return tx.Client.Commit()
}

// Rollback the transaction.
func (tx *Tx) Rollback(ctx context.Context) error {
	return tx.Client.Rollback()
}

// Insert given record to the database.
func (tx *Tx) Create(ctx context.Context, record interface{}) error {
	return create(ctx, tx.Exec, record)
}

// Inserts given record and scans the inserted row back to the given row.
func (tx *Tx) CreateAndRead(ctx context.Context, record interface{}) error {
	return createAndRead(ctx, tx.Exec, tx.Query, record)
}

// Replace given record to the database.
func (tx *Tx) Replace(ctx context.Context, record interface{}) error {
	return replace(ctx, tx.Exec, record)
}

// Replaces given record and scans the replaceed row back to the given row.
func (tx *Tx) ReplaceAndRead(ctx context.Context, record interface{}) error {
	return replaceAndRead(ctx, tx.Exec, tx.Query, record)
}

// Run a select query on the databaase (w/ given parameters optionally) and scan the result(s) to the
// target interface specified as the first parameter.
//
// Usage Example:
//
// user := &User{}
// err := tx.Read(user, "SELECT * FROM users WHERE id = ?", 1)
//
// users := &[]*User{}
// err := tx.Read(users, "SELECT * FROM users", 1)
func (tx *Tx) Read(ctx context.Context, scanTo interface{}, params ...interface{}) error {
	return read(ctx, tx.Query, scanTo, params)
}

// Run an update query on the transaction, finding out the primary-key field of the given row.
func (tx *Tx) Update(ctx context.Context, record interface{}) error {
	return mustUpdate(ctx, tx.Exec, record)
}

// Executes a DELETE query on the transaction for given struct record. It matches
// the database row by finding out the primary key field defined in the table schema.
func (tx *Tx) Delete(ctx context.Context, record interface{}) error {
	return mustDelete(ctx, tx.Exec, record)
}

func (tx *Tx) Begin(ctx context.Context, readOnly bool) (*Tx, error) {
	return nil, fmt.Errorf("can't created nested transactions")
}
