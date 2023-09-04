package crud

import (
	"context"
	stdsql "database/sql"
	"fmt"
	"log/slog"
	"time"
)

type Tx struct {
	Context context.Context
	Client  *stdsql.Tx
}

// Execute any SQL query on the transaction client. Returns sql.Result.
func (tx *Tx) Exec(sql string, params ...interface{}) (stdsql.Result, error) {
	start := time.Now()
	result, err := tx.Client.ExecContext(tx.Context, sql, params...)
	slog.InfoContext(tx.Context, "Executed SQL query", "sql", sql, "took", time.Since(start))
	return result, err
}

// Execute any SQL query on the transaction client. Returns sql.Rows.
func (tx *Tx) Query(sql string, params ...interface{}) (*stdsql.Rows, error) {
	start := time.Now()
	result, err := tx.Client.QueryContext(tx.Context, sql, params...)
	slog.InfoContext(tx.Context, "Ran SQL query", "sql", sql, "took", time.Since(start))
	return result, err
}

// Commit the transaction.
func (tx *Tx) Commit() error {
	slog.InfoContext(tx.Context, "Committing")
	return tx.Client.Commit()
}

// Rollback the transaction.
func (tx *Tx) Rollback() error {
	slog.InfoContext(tx.Context, "Rolling back")
	return tx.Client.Rollback()
}

// Insert given record to the database.
func (tx *Tx) Create(record interface{}) error {
	return create(tx.Exec, record)
}

// Inserts given record and scans the inserted row back to the given row.
func (tx *Tx) CreateAndRead(record interface{}) error {
	return createAndRead(tx.Exec, tx.Query, record)
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
func (tx *Tx) Read(scanTo interface{}, params ...interface{}) error {
	return read(tx.Query, scanTo, params)
}

// Run an update query on the transaction, finding out the primary-key field of the given row.
func (tx *Tx) Update(record interface{}) error {
	return mustUpdate(tx.Exec, record)
}

// Executes a DELETE query on the transaction for given struct record. It matches
// the database row by finding out the primary key field defined in the table schema.
func (tx *Tx) Delete(record interface{}) error {
	return mustDelete(tx.Exec, record)
}

func (tx *Tx) Begin(ctx context.Context) (*Tx, error) {
	return nil, fmt.Errorf("can't created nested transactions")
}
