package crud

import (
	"context"
	stdsql "database/sql"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"

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

func createBulk(ctx context.Context, exec ExecFn, value any) error {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Slice {
		return errors.New("records must be a slice")
	}

	records := make([]any, 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		records = append(records, v.Index(i).Interface())
	}

	row, columns, _, err := valuesForRecord(records[0])
	if err != nil {
		return err
	}

	query := sql.InsertBulkQuery(row.SQLTableName, columns, len(records))
	values := make([]interface{}, 0, len(records)*len(columns))

	for _, record := range records {
		_, _, v, err := valuesForRecord(record)
		if err != nil {
			return err
		}

		values = append(values, v...)
	}

	_, err = exec(ctx, query, values...)
	return err
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

	sqlValues := row.SQLValues()
	columns := make([]string, 0, len(sqlValues))
	values := make([]interface{}, 0, len(sqlValues))

	for _, c := range slices.Sorted(maps.Keys(sqlValues)) {
		v := sqlValues[c]
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
		fmt.Sprintf("SELECT * FROM `%s` WHERE `%s` = ?", table.SQLName, table.PrimaryKeyField().SQL.Name),
		id,
	}

	return read(ctx, query, record, params)
}
