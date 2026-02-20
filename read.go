package crud

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"reflect"

	"github.com/azer/crud/v2/meta"
)

func read(ctx context.Context, query QueryFn, scanTo interface{}, allparams []interface{}) error {
	sql, params, err := ResolveReadParams(allparams)
	if err != nil {
		return err
	}

	if !meta.IsPointer(scanTo) {
		return errors.New("A pointer has to be passed for scanning rows to.")
	}

	if meta.IsSlice(scanTo) {
		return readAll(ctx, query, scanTo, sql, params)
	}

	return readOne(ctx, query, scanTo, sql, params)
}

func readOne(ctx context.Context, query QueryFn, scanTo interface{}, sql string, params []interface{}) error {
	scanner, err := NewScan(scanTo)
	if err != nil {
		return err
	}

	rows, err := query(ctx, sql, params...)
	if err != nil {
		return err
	}

	defer rows.Close()

	if err := scanner.One(rows); err != nil {
		return err
	}

	return rows.Err()
}

func readAll(ctx context.Context, query QueryFn, scanTo interface{}, sql string, params []interface{}) error {
	scanner, err := NewScan(scanTo)
	if err != nil {
		return err
	}

	rows, err := query(ctx, sql, params...)
	if err != nil {
		return err
	}

	defer rows.Close()

	if err := scanner.All(rows); err != nil {
		return err
	}

	return rows.Err()
}

// Arguments of the common CRUD functions can start with a query in string type, followed
// by parameters for the query itself. ResolveReadParams takes a list of any type,
// returns a query as a string type, parameters as a slice of interface, and potentially an
// error the last parameter.
func ResolveReadParams(params []interface{}) (string, []interface{}, error) {
	if len(params) == 0 {
		return "", []interface{}{}, nil
	}

	var (
		query string
		ok    bool
	)

	if query, ok = params[0].(string); !ok {
		return "", nil, fmt.Errorf("Invalid query: %v", params[0])
	}

	if len(params) == 1 {
		return query, []interface{}{}, nil
	}

	return query, params[1:], nil
}

// ReadIter executes the given query and returns an iterator that yields each row as type T.
// The iterator automatically handles scanning and ensures proper resource cleanup.
// Unlike read/readAll, this doesn't load all rows into memory at once, making it ideal for large result sets.
//
// Usage Example with DB:
//
//	type User struct {
//		ID   int    `sql:"id"`
//		Name string `sql:"name"`
//	}
//
//	iter, err := crud.ReadIter[User](ctx, db.Query, "SELECT * FROM users WHERE active = ?", true)
//	if err != nil {
//		return err
//	}
//
//	for user, err := range iter {
//		if err != nil {
//			return err
//		}
//		fmt.Println(user.Name)
//	}
//
// Usage Example with Transaction:
//
//	tx, _ := db.Begin(ctx, false)
//	iter, err := crud.ReadIter[User](ctx, tx.Query, "SELECT * FROM users")
//	if err != nil {
//		return err
//	}
//
//	for user, err := range iter {
//		if err != nil {
//			return err
//		}
//		fmt.Println(user.Name)
//	}
func ReadIter[T any](ctx context.Context, query QueryFn, params ...interface{}) (iter.Seq2[T, error], error) {
	sql, queryParams, err := ResolveReadParams(params)
	if err != nil {
		return nil, err
	}

	rows, err := query(ctx, sql, queryParams...)
	if err != nil {
		return nil, err
	}

	return Yield[T](rows), nil
}

// readIter is similar to ReadIter but works with interface{} for method compatibility.
// It returns an iterator that yields interface{} values which must be type-asserted by the caller.
// Any errors during setup or iteration are returned as the error in the first yielded value.
func readIter(ctx context.Context, query QueryFn, result interface{}, allparams []interface{}) iter.Seq2[interface{}, error] {
	return func(yield func(interface{}, error) bool) {
		sql, params, err := ResolveReadParams(allparams)
		if err != nil {
			yield(nil, err)
			return
		}

		rows, err := query(ctx, sql, params...)
		if err != nil {
			yield(nil, err)
			return
		}
		defer rows.Close()

		// Create scanner based on result type to determine what we're scanning to
		scanner, err := NewScan(result)
		if err != nil {
			yield(nil, err)
			return
		}

		for rows.Next() {
			record := meta.CreateElement(result)

			if err := scanner.Scan(rows, record); err != nil {
				if !yield(nil, err) {
					return
				}
				continue
			}

			// Extract the value from reflect.Value
			var value interface{}
			if record.Kind() == reflect.Ptr {
				value = record.Elem().Interface()
			} else {
				value = record.Interface()
			}

			if !yield(value, nil) {
				return
			}
		}

		// Check for errors from iteration
		if err := rows.Err(); err != nil {
			yield(nil, err)
		}
	}
}
