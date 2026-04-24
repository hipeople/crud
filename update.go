package crud

import (
	"context"
	stdsql "database/sql"
	"fmt"

	"github.com/hipeople/crud/v2/sql"
)

func update(ctx context.Context, exec ExecFn, record interface{}) (stdsql.Result, error) {
	table, err := NewTable(record)
	if err != nil {
		return nil, err
	}

	pk := table.PrimaryKeyField()
	if pk == nil {
		return nil, fmt.Errorf("Table '%s' (%s) doesn't have a primary-key field", table.Name, table.SQLName)
	}

	return exec(ctx, sql.UpdateQuery(table.SQLName, pk.SQL.Name, table.SQLUpdateColumnSet()), table.SQLUpdateValueSet(record)...)
}

func mustUpdate(ctx context.Context, exec ExecFn, record interface{}) error {
	result, err := update(ctx, exec, record)
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
