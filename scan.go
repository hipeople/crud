package crud

import (
	"database/sql"
	"iter"
	"reflect"

	"github.com/azer/crud/v2/meta"
)

// Create a scanner for any given interface. This function will be called for
// every target interface passed to DB methods that scans results.
func NewScan(to interface{}) (*Scan, error) {
	scan := &Scan{
		To:         to,
		ToPointers: meta.HasPointers(to),
		ToStructs:  meta.HasAnyStruct(to),
	}

	if scan.ToStructs {
		table, err := NewTable(to)
		if err != nil {
			return nil, err
		}

		scan.SQLColumnDict = table.SQLColumnDict()
	}

	return scan, nil
}

type Scan struct {
	To            interface{}
	ToPointers    bool
	ToStructs     bool
	SQLColumnDict map[string]string
}

func (scan *Scan) All(rows *sql.Rows) error {
	writeTo := meta.Addressable(scan.To)

	for rows.Next() {
		record := meta.CreateElement(scan.To)

		if err := scan.Scan(rows, record); err != nil {
			return err
		}

		meta.Push(writeTo, record)
	}

	return nil
}

func (scan *Scan) One(rows *sql.Rows) error {
	for rows.Next() {
		return scan.Scan(rows, meta.DirectValueOf(scan.To))
	}

	return sql.ErrNoRows
}

func (scan *Scan) Scan(rows *sql.Rows, record reflect.Value) error {
	if scan.ToStructs {
		return scan.ScanToStruct(rows, record)
	}

	if record.Kind() != reflect.Ptr {
		return rows.Scan(record.Addr().Interface())
	} else {
		return rows.Scan(record.Interface())
	}
}

func (scan *Scan) ScanToStruct(rows *sql.Rows, record reflect.Value) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	values := make([]interface{}, len(columns))

	for i, column := range columns {
		var field reflect.Value

		fieldName := scan.SQLColumnDict[column]

		if scan.ToPointers {
			field = record.Elem().FieldByName(fieldName)
		} else {
			field = record.FieldByName(fieldName)
		}

		if field.IsValid() {
			values[i] = field.Addr().Interface()
		} else {
			values[i] = &values[i]
		}
	}

	return rows.Scan(values...)
}

// Yield returns an iterator that yields each row from sql.Rows as a value of type T.
// The iterator handles row scanning automatically and ensures rows are properly closed.
// It yields (value, error) pairs - on success error is nil, on failure value is zero.
//
// Example usage:
//
//	type User struct {
//		ID   int    `sql:"id"`
//		Name string `sql:"name"`
//	}
//
//	rows, _ := db.Query(ctx, "SELECT id, name FROM users")
//	for user, err := range Yield[User](rows) {
//		if err != nil {
//			return err
//		}
//		fmt.Println(user.Name)
//	}
func Yield[T any](rows *sql.Rows) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		defer rows.Close()

		// Create a scanner for type T
		var target []T
		scanner, err := NewScan(&target)
		if err != nil {
			var zero T
			yield(zero, err)
			return
		}

		// Iterate through all rows
		for rows.Next() {
			record := meta.CreateElement(&target)

			if err := scanner.Scan(rows, record); err != nil {
				var zero T
				if !yield(zero, err) {
					return
				}
				continue
			}

			// Extract the value from reflect.Value
			var value T
			if record.Kind() == reflect.Ptr {
				value = record.Elem().Interface().(T)
			} else {
				value = record.Interface().(T)
			}

			if !yield(value, nil) {
				return
			}
		}

		// Check for errors from iteration
		if err := rows.Err(); err != nil {
			var zero T
			yield(zero, err)
		}
	}
}
