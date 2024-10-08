package crud

import (
	"github.com/azer/crud/v2/sql"
)

type Field struct {
	Name string
	SQL  *sql.Options
}

// Get DB fields of any valid struct given
func GetFieldsOf(st interface{}) ([]*Field, error) {
	fields, err := CollectFields(st, []*Field{})
	if err != nil {
		return nil, err
	}

	return fields, nil
}

func CollectFields(st interface{}, fields []*Field) ([]*Field, error) {
	iter := NewFieldIteration(st)
	for iter.Next() {
		sqlOptions, err := iter.SQLOptions()
		if err != nil {
			return nil, err
		}

		if sqlOptions.Ignore {
			continue
		}

		fields = append(fields, &Field{
			Name: iter.Name(),
			SQL:  sqlOptions,
		})
	}

	return fields, nil
}

// If no PK is specified, then set `id` to be PK.
func SetDefaultPK(fields []*Field) {
	if HasPK(fields) {
		return
	}

	for i, f := range fields {
		if !f.SQL.IsPrimaryKey && f.SQL.Name == "id" && f.SQL.Type == "int" {
			fields[i].SQL.IsAutoIncrementing = true
			fields[i].SQL.AutoIncrement = 1
			fields[i].SQL.IsRequired = true
			fields[i].SQL.IsPrimaryKey = true
			return
		}
	}
}

func HasPK(fields []*Field) bool {
	for _, f := range fields {
		if f.SQL.IsPrimaryKey {
			return true
		}
	}

	return false
}
