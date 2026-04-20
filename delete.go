package crud

import (
	"context"
	stdsql "database/sql"
	"errors"
	"fmt"
	"reflect"

	"github.com/hipeople/crud/v2/meta"
	"github.com/hipeople/crud/v2/sql"
)

func deleteRow(ctx context.Context, exec ExecFn, record any) (stdsql.Result, error) {
	table, err := NewTable(record)

	if err != nil {
		return nil, err
	}

	pk := table.PrimaryKeyField()
	if pk == nil {
		return nil, errors.New(fmt.Sprintf("Table '%s' (%s) doesn't have a primary-key field", table.Name, table.SQLName))
	}

	return exec(ctx, sql.DeleteQuery(table.SQLName, pk.SQL.Name), meta.StructFieldValue(record, pk.Name))
}

func deleteRows(ctx context.Context, exec ExecFn, records []any) (stdsql.Result, error) {
	if len(records) == 0 {
		return nil, errors.New("no records to delete")
	}

	table, err := NewTable(records[0])
	if err != nil {
		return nil, err
	}

	pk := table.PrimaryKeyField()
	if pk == nil {
		return nil, fmt.Errorf("Table '%s' (%s) doesn't have a primary-key field", table.Name, table.SQLName)
	}

	pkValues := make([]any, 0, len(records))
	for _, record := range records {
		pkValue := meta.StructFieldValue(record, pk.Name)
		pkValues = append(pkValues, pkValue)
	}

	return exec(ctx, sql.BulkDeleteQuery(table.SQLName, pk.SQL.Name, len(records)), pkValues...)
}

func mustDelete(ctx context.Context, exec ExecFn, record any) error {
	result, err := deleteRow(ctx, exec, record)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return stdsql.ErrNoRows
	}

	return nil
}

func mustBulkDelete(ctx context.Context, exec ExecFn, value any) error {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Slice {
		return errors.New("records must be a slice")
	}

	records := make([]any, 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		records = append(records, v.Index(i).Interface())
	}

	if len(records) == 0 {
		return errors.New("no records to delete")
	}

	result, err := deleteRows(ctx, exec, records)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return stdsql.ErrNoRows
	}

	return nil
}
