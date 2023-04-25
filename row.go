package crud

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
	values, err := GetRowValuesOf(st)
	if err != nil {
		return nil, err
	}

	tableName := SQLTableNameOf(st)

	if customTableName, ok := LookupCustomTableName(st); ok {
		tableName = customTableName
	}

	return &Row{
		SQLTableName: tableName,
		Values:       values,
	}, nil
}

// Scans given struct record and returns a list of crud.Row instances for each
// struct field. It's useful for extracting values and corresponding SQL meta information
// from structs representing database tables.
func GetRowValuesOf(st interface{}) ([]*RowValue, error) {
	fields, err := CollectRows(st, []*RowValue{})
	if err != nil {
		return nil, err
	}

	return fields, nil
}

func CollectRows(st interface{}, rows []*RowValue) ([]*RowValue, error) {
	iter := NewFieldIteration(st)
	for iter.Next() {
		sqlOptions, err := iter.SQLOptions()

		if err != nil {
			return nil, err
		}

		if sqlOptions.Ignore {
			continue
		}

		if sqlOptions.IsInline {
			rows, err = CollectRows(iter.Value(), rows)
			if err != nil {
				return nil, err
			}
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
