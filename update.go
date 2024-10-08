package crud

import (
	stdsql "database/sql"
	"errors"
	"fmt"

	"github.com/azer/crud/v2/sql"
)

func update(exec ExecFn, record interface{}) (stdsql.Result, error) {
	table, err := NewTable(record)
	if err != nil {
		return nil, err
	}

	pk := table.PrimaryKeyField()
	if pk == nil {
		return nil, errors.New(fmt.Sprintf("Table '%s' (%s) doesn't have a primary-key field", table.Name, table.SQLName))
	}

	return exec(sql.UpdateQuery(table.SQLName, pk.SQL.Name, table.SQLUpdateColumnSet()), table.SQLUpdateValueSet(record)...)
}

func mustUpdate(exec ExecFn, record interface{}) error {
	result, err := update(exec, record)
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
