package crud

import (
	"context"
	stdsql "database/sql"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"

	"github.com/go-sql-driver/mysql"
	"github.com/hipeople/crud/v2/sql"
)

const (
	InvalidOperationErrNumber = 1062
)

var ErrInvalidOperation = errors.New("failed to run query, operation is not valid")

func createAndGetResult(ctx context.Context, exec ExecFn, record interface{}) (stdsql.Result, error) {
	row, columns, values, err := valuesForRecord(record)
	if err != nil {
		return nil, err
	}

	res, err := exec(ctx, sql.InsertQuery(row.SQLTableName, columns), values...)
	return res, checkMysqlError(err)
}

func create(ctx context.Context, exec ExecFn, record interface{}) error {
	_, err := createAndGetResult(ctx, exec, record)
	return checkMysqlError(err)
}

func createAndRead(ctx context.Context, exec ExecFn, query QueryFn, record interface{}) error {
	result, err := createAndGetResult(ctx, exec, record)
	if err != nil {
		return err
	}

	if err := checkMysqlError(err); err != nil {
		return err
	}

	return readLastInsert(ctx, query, record, result)
}

func checkMysqlError(err error) error {
	if err != nil {
		if mysqlErr, ok := errors.AsType[*mysql.MySQLError](err); ok {
			if mysqlErr.Number == InvalidOperationErrNumber {
				return ErrInvalidOperation
			}
		}
	}

	return err
}

func bulkCreate(ctx context.Context, exec ExecFn, value any) error {
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

func upsertAndGetResult(ctx context.Context, exec ExecFn, record interface{}) (stdsql.Result, error) {
	row, columns, values, err := valuesForRecord(record)
	if err != nil {
		return nil, err
	}

	table, err := NewTable(record)
	if err != nil {
		return nil, err
	}

	// Build INSERT ... ON DUPLICATE KEY UPDATE query
	query := fmt.Sprintf("INSERT INTO `%s` ", row.SQLTableName)

	// Add column names
	query += "("
	for i, col := range columns {
		if i > 0 {
			query += ", "
		}
		query += "`" + col + "`"
	}
	query += ") VALUES ("

	// Add placeholders
	for i := range columns {
		if i > 0 {
			query += ", "
		}
		query += "?"
	}
	query += ") ON DUPLICATE KEY UPDATE "

	// Build map of fields to exclude from updates (primary key and fields with no-update tag)
	excludeFromUpdate := make(map[string]struct{})
	if pkField := table.PrimaryKeyField(); pkField != nil {
		excludeFromUpdate[pkField.SQL.Name] = struct{}{}
	}
	for _, field := range table.Fields {
		if field.SQL.NoUpdate {
			excludeFromUpdate[field.SQL.Name] = struct{}{}
		}
	}

	// MySQL returns 0 affected rows and doesn't update LastInsertId when all column values match the existing row exactly.
	// Touch the pk column to force a change so LastInsertId is the affected row's ID, and readLastInsert can re-read.
	var updateParts []string
	if pk := table.PrimaryKeyField(); pk != nil {
		updateParts = append(updateParts, fmt.Sprintf("`%s` = LAST_INSERT_ID(`%s`)", pk.SQL.Name, pk.SQL.Name))
	}
	for _, col := range columns {
		if _, ok := excludeFromUpdate[col]; !ok {
			updateParts = append(updateParts, fmt.Sprintf("`%s` = VALUES(`%s`)", col, col))
		}
	}
	query += updateParts[0]
	for i := 1; i < len(updateParts); i++ {
		query += ", " + updateParts[i]
	}

	return exec(ctx, query, values...)
}

func upsert(ctx context.Context, exec ExecFn, record interface{}) error {
	_, err := upsertAndGetResult(ctx, exec, record)
	return err
}

func upsertAndRead(ctx context.Context, exec ExecFn, query QueryFn, record interface{}) error {
	result, err := upsertAndGetResult(ctx, exec, record)
	if err != nil {
		return err
	}

	return readLastInsert(ctx, query, record, result)
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
