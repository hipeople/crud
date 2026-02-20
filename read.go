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

// readIter is similar to ReadIter but works with interface{} for method compatibility.
// It returns an iterator that yields interface{} values which must be type-asserted by the caller.
// Any errors during setup or iteration are returned as the error in the first yielded value.
func readIter(ctx context.Context, query QueryFn, typ any, allparams []any) iter.Seq2[any, error] {
	return func(yield func(any, error) bool) {
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
		scanner, err := NewScan(typ)
		if err != nil {
			yield(nil, err)
			return
		}

		for rows.Next() {
			record := meta.CreateElement(typ)

			if err := scanner.Scan(rows, record); err != nil {
				if !yield(nil, err) {
					return
				}
				continue
			}

			// Extract the value from reflect.Value
			var value any
			if record.Kind() == reflect.Pointer {
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
