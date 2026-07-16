package crud

import (
	"reflect"

	"github.com/hipeople/crud/v2/meta"
)

type RowValue struct {
	SQLColumn string
	Value     interface{}
}

type Row struct {
	SQLTableName string
	Values       []*RowValue
}

func (row *Row) SQLValues() map[string]interface{} {
	result := map[string]interface{}{}

	for _, v := range row.Values {
		result[v.SQLColumn] = v.Value
	}

	return result
}

func NewRow(st interface{}) (*Row, error) {
	table, err := NewTable(st)
	if err != nil {
		return nil, err
	}

	// SQLName already resolves any custom table-name option (SQLTableNameOf
	// checks it first), so no separate LookupCustomTableName pass is needed.
	return &Row{
		SQLTableName: table.SQLName,
		Values:       rowValuesFromPlan(table.rowPlan, meta.ValueOf(st)),
	}, nil
}

// Scans given struct record and returns a list of crud.Row instances for each
// struct field. It's useful for extracting values and corresponding SQL meta information
// from structs representing database tables.
func GetRowValuesOf(st interface{}) ([]*RowValue, error) {
	table, err := NewTable(st)
	if err != nil {
		return nil, err
	}

	return rowValuesFromPlan(table.rowPlan, meta.ValueOf(st)), nil
}

// rowValuesFromPlan reads the insertable field values off record in the plan's
// (struct-field) order, skipping a zero explicit-auto-increment int PK so the
// database assigns it. rv must be the indirected struct value.
func rowValuesFromPlan(plan []rowPlanEntry, rv reflect.Value) []*RowValue {
	// One backing array for the RowValue structs instead of a heap allocation
	// per field; values holds pointers into it.
	backing := make([]RowValue, len(plan))
	values := make([]*RowValue, 0, len(plan))

	i := 0
	for _, entry := range plan {
		value := rv.Field(entry.index).Interface()

		if entry.autoIncSkip {
			if n, ok := value.(int); ok && n == 0 {
				continue
			}
		}

		backing[i] = RowValue{
			SQLColumn: entry.column,
			Value:     value,
		}
		values = append(values, &backing[i])
		i++
	}

	return values
}

func CollectRows(st interface{}, rows []*RowValue) ([]*RowValue, error) {
	iter := NewFieldIteration(st)
	for iter.Next() {
		sqlOptions, err := iter.SQLOptions()
		if err != nil {
			return nil, err
		}

		if sqlOptions.Ignore || sqlOptions.Generated {
			continue
		}

		value := iter.Value()

		if n, ok := value.(int); ok && sqlOptions.AutoIncrement > 0 && n == 0 {
			continue
		}

		rows = append(rows, &RowValue{
			SQLColumn: sqlOptions.Name,
			Value:     value,
		})
	}

	return rows, nil
}
