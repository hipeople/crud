package crud

import (
	"fmt"

	"github.com/azer/crud/v2/meta"
	"github.com/azer/snakecase"
	"github.com/jinzhu/inflection"
)

// Find out what given interface should be called in the database. It first looks up
// if a table name was explicitly specified (see "table-name" option), or automatically
// generates a plural name from the name of the struct type.
func SQLTableNameOf(any interface{}) string {
	if customTableName, ok := LookupCustomTableName(any); ok {
		return customTableName
	}

	return escapeTableName(snakecase.SnakeCase(inflection.Plural(meta.TypeNameOf(any))))
}

func LookupCustomTableName(any interface{}) (string, bool) {
	if meta.IsSlice(any) {
		any = meta.CreateElement(any).Interface()
	}

	fields, err := GetFieldsOf(any)
	if err != nil {
		return "", false
	}

	for _, f := range fields {
		if len(f.SQL.TableName) > 0 {
			return escapeTableName(f.SQL.TableName), true
		}
	}

	return "", false
}

func escapeTableName(tableName string) string {
	return fmt.Sprintf("`%s`", tableName)
}
