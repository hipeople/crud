package sql

import (
	"fmt"
	"slices"
	"strings"
)

func NewTableQuery(name string, fields []*Options, ifNotExists bool) string {
	ifNotExistsExt := ""
	if ifNotExists {
		ifNotExistsExt = " IF NOT EXISTS"
	}

	return fmt.Sprintf("CREATE TABLE%s `%s` (\n%s%s\n)%s;",
		ifNotExistsExt, name, NewFieldQueries(fields), NewPrimaryKeyQuery(fields), NewTableConfigQuery(fields))
}

func NewFieldQueries(fields []*Options) string {
	queries := []string{}

	for _, f := range fields {
		if f.Ignore {
			continue
		}

		queries = append(queries, NewFieldQuery(f))
	}

	return strings.Join(queries, ",\n")
}

func NewFieldQuery(field *Options) string {
	length := ""
	autoIncrement := ""
	required := ""
	defaultValue := ""
	unsigned := ""
	unique := ""

	if field.TypeArg != "" {
		length = fmt.Sprintf("(%s)", field.TypeArg)
	}

	if field.AutoIncrement > 0 {
		autoIncrement = " AUTO_INCREMENT"
	}

	if field.IsRequired {
		required = " NOT NULL"
	}

	if field.DefaultValue != "" {
		defaultValue = fmt.Sprintf(" DEFAULT %s", field.DefaultValue)
	}

	if field.IsUnsigned {
		unsigned = " UNSIGNED"
	}

	if field.IsUnique {
		unique = " UNIQUE"
	}

	query := fmt.Sprintf("%s%s%s%s%s%s",
		length, required, defaultValue, unsigned, unique, autoIncrement)

	return fmt.Sprintf("  `%s` %s%s", field.Name, field.Type, query)
}

func NewPrimaryKeyQuery(fields []*Options) string {
	keys := []string{}

	for _, f := range fields {
		if f.IsPrimaryKey {
			keys = append(keys, f.Name)
		}
	}

	if len(keys) == 0 {
		return ""
	}

	return fmt.Sprintf(",\n  PRIMARY KEY (`%s`)", strings.Join(keys, "`, `"))
}

func NewTableConfigQuery(fields []*Options) string {
	autoIncrement := ""
	for _, f := range fields {
		if f.AutoIncrement > 1 {
			autoIncrement = fmt.Sprintf(" AUTO_INCREMENT=%d", f.AutoIncrement)
		}
	}

	return autoIncrement
}

func DropTableQuery(name string, ifExists bool) string {
	ext := ""

	if ifExists {
		ext = " IF EXISTS"
	}

	return fmt.Sprintf("DROP TABLE%s `%s`", ext, name)
}

func ShowTablesLikeQuery(name string) string {
	return fmt.Sprintf("SHOW TABLES LIKE '%s'", name)
}

func InsertQuery(tableName string, columnNames []string) string {
	questionMarks := repeatComma(len(columnNames), "?")

	return fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)",
		tableName, strings.Join(quoteColumnNames(columnNames), ","), questionMarks)
}

func InsertBulkQuery(tableName string, columnNames []string, numRecords int) string {
	pattern := fmt.Sprintf("(%s)", repeatComma(len(columnNames), "?"))
	questionMarks := slices.Repeat([]string{pattern}, numRecords)

	return fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s",
		tableName, strings.Join(quoteColumnNames(columnNames), ","), strings.Join(questionMarks, ","))
}

func ReplaceQuery(tableName string, columnNames []string) string {
	questionMarks := repeatComma(len(columnNames), "?")

	return fmt.Sprintf("REPLACE INTO `%s` (%s) VALUES (%s)",
		tableName, strings.Join(quoteColumnNames(columnNames), ","), questionMarks)
}

func SelectQuery(tableName string, columnNames []string) string {
	columns := "*"
	if len(columnNames) > 0 {
		columns = strings.Join(quoteColumnNames(columnNames), ", ")
	}

	return fmt.Sprintf("SELECT %s FROM `%s`", columns, tableName)
}

func UpdateQuery(tableName, index string, columnNames []string) string {
	return fmt.Sprintf("%s WHERE `%s`=?", UpdateAllQuery(tableName, columnNames), index)
}

func UpdateAllQuery(tableName string, columnNames []string) string {
	return fmt.Sprintf("UPDATE `%s` SET %s=?", tableName, strings.Join(quoteColumnNames(columnNames), "=?, "))
}

func DeleteQuery(tableName, index string) string {
	return fmt.Sprintf("DELETE FROM `%s` WHERE `%s`=?", tableName, index)
}

func BulkDeleteQuery(tableName, index string, numRecords int) string {
	pattern := fmt.Sprintf("(%s)", repeatComma(1, "?"))
	questionMarks := slices.Repeat([]string{pattern}, numRecords)

	return fmt.Sprintf("DELETE FROM `%s` WHERE `%s` IN (%s)",
		tableName, index, strings.Join(questionMarks, ","))
}

func quoteColumnNames(columns []string) []string {
	var cols []string
	for _, c := range columns {
		cols = append(cols, "`"+c+"`")
	}

	return cols
}

func repeatComma(num int, char string) string {
	var out string

	if num > 0 {
		out = strings.Repeat(char+",", num)
		out = out[:len(out)-1]
	}

	return out
}
