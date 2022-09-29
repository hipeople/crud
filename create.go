package crud

import (
	stdsql "database/sql"
	"fmt"

	"github.com/azer/crud/v2/sql"
)

func createAndGetResult(exec ExecFn, record interface{}, isUpsert bool) (stdsql.Result, error) {
	row, err := NewRow(record)
	if err != nil {
		return nil, err
	}

	var columns []string
	var values []interface{}

	for c, v := range row.SQLValues() {
		columns = append(columns, c)
		values = append(values, v)
	}

	if isUpsert {
		vals := append(values, values...)
		return exec(sql.UpsertQuery(row.SQLTableName, columns), vals...)
	}
	return exec(sql.InsertQuery(row.SQLTableName, columns), values...)
}

func create(exec ExecFn, record interface{}) error {
	_, err := createAndGetResult(exec, record, false)
	return err
}

func createAndRead(exec ExecFn, query QueryFn, record interface{}, upsert bool) error {
	result, err := createAndGetResult(exec, record, upsert)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	table, err := NewTable(record)
	if err != nil {
		// this is a bad design choice made assuming that it'll never happen.
		return err
	}

	params := []interface{}{
		fmt.Sprintf("SELECT * FROM %s WHERE %s = ?", table.SQLName, table.PrimaryKeyField().SQL.Name),
		id,
	}

	return read(query, record, params)
}
