package crud

import (
	"context"
	stdsql "database/sql"
	"fmt"

	"github.com/azer/crud/v2/sql"
)

func createAndGetResult(ctx context.Context, exec ExecFn, record interface{}) (stdsql.Result, error) {
	row, columns, values, err := valuesForRecord(record)
	if err != nil {
		return nil, err
	}

	return exec(ctx, sql.InsertQuery(row.SQLTableName, columns), values...)
}

func create(ctx context.Context, exec ExecFn, record interface{}) error {
	_, err := createAndGetResult(ctx, exec, record)
	return err
}

func createAndRead(ctx context.Context, exec ExecFn, query QueryFn, record interface{}) error {
	result, err := createAndGetResult(ctx, exec, record)
	if err != nil {
		return err
	}

	return readLastInsert(ctx, query, record, result)
}

func replaceAndGetResult(ctx context.Context, exec ExecFn, record interface{}) (stdsql.Result, error) {
	row, columns, values, err := valuesForRecord(record)
	if err != nil {
		return nil, err
	}

	return exec(ctx, sql.ReplaceQuery(row.SQLTableName, columns), values...)
}

func replace(ctx context.Context, exec ExecFn, record interface{}) error {
	_, err := replaceAndGetResult(ctx, exec, record)
	return err
}

func replaceAndRead(ctx context.Context, exec ExecFn, query QueryFn, record interface{}) error {
	result, err := replaceAndGetResult(ctx, exec, record)
	if err != nil {
		return err
	}

	return readLastInsert(ctx, query, record, result)
}

func valuesForRecord(record interface{}) (*Row, []string, []interface{}, error) {
	row, err := NewRow(record)
	if err != nil {
		return nil, nil, nil, err
	}

	columns := []string{}
	values := []interface{}{}

	for c, v := range row.SQLValues() {
		columns = append(columns, c)
		values = append(values, v)
	}

	return row, columns, values, nil
}

func readLastInsert(ctx context.Context, query QueryFn, record interface{}, result stdsql.Result) error {
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

	return read(ctx, query, record, params)
}
