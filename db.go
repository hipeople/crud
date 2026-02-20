package crud

import (
	"context"
	stdsql "database/sql"
	"iter"
	"log/slog"
	"time"

	"github.com/azer/crud/v2/sql"
)

type ExecFn func(context.Context, string, ...interface{}) (stdsql.Result, error)
type QueryFn func(context.Context, string, ...interface{}) (*stdsql.Rows, error)

type DB struct {
	Client *stdsql.DB
	Logger *slog.Logger
}

func (db *DB) Ping() error {
	return db.Client.Ping()
}

// Run any query on the database client, passing parameters optionally. Returns sql.Result.
func (db *DB) Exec(ctx context.Context, sql string, params ...interface{}) (stdsql.Result, error) {
	start := time.Now()
	result, err := db.Client.ExecContext(ctx, sql, params...)

	if db.Logger != nil {
		db.Logger.DebugContext(ctx, "Executed SQL query", "sql", sql, "took", time.Since(start), "error", err)
	}

	return result, err
}

// Run any query on the database client, passing parameters optionally. Its difference with
// `Exec` method is returning `sql.Rows` instead of `sql.Result`.
func (db *DB) Query(ctx context.Context, sql string, params ...interface{}) (*stdsql.Rows, error) {
	start := time.Now()
	result, err := db.Client.QueryContext(ctx, sql, params...)

	if db.Logger != nil {
		db.Logger.DebugContext(ctx, "Ran SQL query", "sql", sql, "took", time.Since(start), "error", err)
	}

	return result, err
}

// Takes any valid struct and creates a SQL table from it.
func (db *DB) CreateTable(ctx context.Context, st interface{}, ifexists bool) error {
	t, err := NewTable(st)
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, sql.NewTableQuery(t.SQLName, t.SQLOptions(), ifexists))
	return err
}

// Takes any valid struct, finds out its corresponding SQL table and drops it.
func (db *DB) DropTable(ctx context.Context, st interface{}, ifexists bool) error {
	t, err := NewTable(st)
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, sql.DropTableQuery(t.SQLName, true))
	return err
}

// Creates multiple tables from given any amount of structs. Calls `CreateTable` internally.
func (db *DB) CreateTables(ctx context.Context, structs ...interface{}) error {
	for _, st := range structs {
		if err := db.CreateTable(ctx, st, true); err != nil {
			return err
		}
	}

	return nil
}

// Drops correspoinding SQL tables of the given structs.
func (db *DB) DropTables(ctx context.Context, structs ...interface{}) error {
	for _, st := range structs {
		if err := db.DropTable(ctx, st, true); err != nil {
			return err
		}
	}

	return nil
}

// Drops (if they exist) and re-creates corresponding SQL tables for the given structs.
func (db *DB) ResetTables(ctx context.Context, structs ...interface{}) error {
	if err := db.DropTables(ctx, structs...); err != nil {
		return err
	}

	if err := db.CreateTables(ctx, structs...); err != nil {
		return err
	}

	return nil
}

// Runs a query to check if the given table exists and returns bool
func (db *DB) CheckIfTableExists(ctx context.Context, name string) bool {
	var result string
	err := db.Client.QueryRow(sql.ShowTablesLikeQuery(name)).Scan(&result)
	return err == nil && result == name
}

// Inserts given record into the database, generating an insert query for it.
func (db *DB) Create(ctx context.Context, record interface{}) error {
	return create(ctx, db.Exec, record)
}

func (db *DB) CreateAndGetResult(ctx context.Context, record interface{}) (stdsql.Result, error) {
	return createAndGetResult(ctx, db.Exec, record)
}

// Inserts given record and scans the inserted row back to the given row.
func (db *DB) CreateAndRead(ctx context.Context, record interface{}) error {
	return createAndRead(ctx, db.Exec, db.Query, record)
}

// Replaces given record into the database, generating a replace query for it.
func (db *DB) Replace(ctx context.Context, record interface{}) error {
	return replace(ctx, db.Exec, record)
}

func (db *DB) ReplaceAndGetResult(ctx context.Context, record interface{}) (stdsql.Result, error) {
	return replaceAndGetResult(ctx, db.Exec, record)
}

// Replaces given record and scans the replaceed row back to the given row.
func (db *DB) ReplaceAndRead(ctx context.Context, record interface{}) error {
	return replaceAndRead(ctx, db.Exec, db.Query, record)
}

// Upserts given record into the database using INSERT ... ON DUPLICATE KEY UPDATE.
// This is safer than Replace as it uses a single atomic operation and is less prone to deadlocks.
func (db *DB) Upsert(ctx context.Context, record interface{}) error {
	return upsert(ctx, db.Exec, record)
}

func (db *DB) UpsertAndGetResult(ctx context.Context, record interface{}) (stdsql.Result, error) {
	return upsertAndGetResult(ctx, db.Exec, record)
}

// Upserts given record and scans the upserted row back to the given row.
func (db *DB) UpsertAndRead(ctx context.Context, record interface{}) error {
	return upsertAndRead(ctx, db.Exec, db.Query, record)
}

// Runs given SQL query and scans the result rows into the given target interface. The target
// interface could be both a single record or a slice of records.
//
// Usage Example:
//
// user := &User{}
// err := tx.Read(user, "SELECT * FROM users WHERE id = ?", 1)
//
// users := &[]*User{}
// err := tx.Read(users, "SELECT * FROM users", 1)
func (db *DB) Read(ctx context.Context, scanTo interface{}, params ...interface{}) error {
	return read(ctx, db.Query, scanTo, params)
}

// ReadIter executes the given query and returns an iterator that yields each row.
// Unlike Read, this doesn't load all rows into memory at once, making it ideal for large result sets.
// Any errors during query execution are returned on the first iteration.
//
// Usage Example:
//
//	var users []UserProfile
//	iter := db.ReadIter(ctx, &users, "SELECT * FROM users WHERE active = ?", true)
//
//	for val, err := range iter {
//		if err != nil {
//			return err
//		}
//		user := val.(UserProfile)
//		fmt.Println(user.Name)
//	}
func (db *DB) ReadIter(ctx context.Context, result interface{}, params ...interface{}) iter.Seq2[interface{}, error] {
	return readIter(ctx, db.Query, result, params)
}

// Finding out the primary-key field of the given row, updates the corresponding record on the table
// with the values in the given record.
func (db *DB) Update(ctx context.Context, record interface{}) error {
	return mustUpdate(ctx, db.Exec, record)
}

// Generates and executes a DELETE query for given struct record. It matches the database row by finding
// out the primary key field defined in the table schema.
func (db *DB) Delete(ctx context.Context, record interface{}) error {
	return mustDelete(ctx, db.Exec, record)
}

func (db *DB) BulkDelete(ctx context.Context, records interface{}) error {
	return mustBulkDelete(ctx, db.Exec, records)
}

// Start a DB transaction. It returns an interface w/ most of the methods DB provides.
func (db *DB) Begin(ctx context.Context, readOnly bool) (*Tx, error) {
	client, err := db.Client.BeginTx(ctx, &stdsql.TxOptions{
		ReadOnly: readOnly,
	})
	if err != nil {
		return nil, err
	}

	return &Tx{
		Client: client,
		Logger: db.Logger,
	}, nil
}

// Establish DB connection and return a crud.DB instance w/ methods needed for accessing / writing the database.
// Example call: Connect("mysql", "root:123456@tcp(localhost:3306)/database_name?parseTime=true")
func Connect(driver, url string, logger *slog.Logger) (*DB, error) {
	client, err := stdsql.Open(driver, url)
	if err != nil {
		return nil, err
	}

	return &DB{
		Client: client,
		Logger: logger,
	}, nil
}

func (db *DB) BulkCreate(ctx context.Context, records any) error {
	return bulkCreate(ctx, db.Exec, records)
}
