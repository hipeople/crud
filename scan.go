package crud

import (
	"database/sql"
	"reflect"

	"github.com/hipeople/crud/v2/meta"
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

	// plan maps each result column (by position) to the destination struct
	// field index path, computed once from the first row's columns and reused
	// for every subsequent row. A nil index means the column has no matching
	// field and is scanned into a throwaway. values is the reused scan buffer.
	plan   []columnPlan
	values []interface{}
}

type columnPlan struct {
	index []int // reflect field index path; nil = no matching struct field
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

	if record.Kind() != reflect.Pointer {
		return rows.Scan(record.Addr().Interface())
	} else {
		return rows.Scan(record.Interface())
	}
}

func (scan *Scan) ScanToStruct(rows *sql.Rows, record reflect.Value) error {
	target := record
	if scan.ToPointers {
		target = record.Elem()
	}

	if scan.plan == nil {
		if err := scan.buildPlan(rows, target.Type()); err != nil {
			return err
		}
	}

	for i, p := range scan.plan {
		if p.index == nil {
			scan.values[i] = &scan.values[i]
			continue
		}

		scan.values[i] = target.FieldByIndex(p.index).Addr().Interface()
	}

	return rows.Scan(scan.values...)
}

// buildPlan resolves each column to a struct field index once, replacing the
// per-row, per-column FieldByName linear search with a precomputed lookup.
func (scan *Scan) buildPlan(rows *sql.Rows, structType reflect.Type) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	plan := make([]columnPlan, len(columns))
	for i, column := range columns {
		if field, ok := structType.FieldByName(scan.SQLColumnDict[column]); ok {
			plan[i] = columnPlan{index: field.Index}
		}
	}

	scan.plan = plan
	scan.values = make([]interface{}, len(columns))

	return nil
}
