package crud

import (
	"reflect"
	"slices"
	"strings"
	"sync"

	"github.com/hipeople/crud/v2/meta"
	"github.com/hipeople/crud/v2/sql"
)

var (
	tableCache   = map[reflect.Type]*Table{}
	tableCacheMu sync.RWMutex
)

// Create an internal representation of a database table, including its fields from given
// struct record
func NewTable(any interface{}) (*Table, error) {
	anyT := reflect.TypeOf(any)

	tableCacheMu.RLock()
	if t, ok := tableCache[anyT]; ok {
		tableCacheMu.RUnlock()
		return t, nil
	}
	tableCacheMu.RUnlock()

	if meta.IsSlice(any) {
		any = meta.CreateElement(any).Interface()
	}

	fields, err := GetFieldsOf(any)
	if err != nil {
		return nil, err
	}

	// Build the insert row plan from the raw field options, BEFORE SetDefaultPK
	// mutates them. This preserves the historical write-path behavior: the
	// auto-increment-zero skip only fires for fields with an explicit
	// auto-increment tag, not for an implicitly-defaulted `id` primary key.
	rowPlan := buildRowPlan(fields)

	SetDefaultPK(fields)

	name, sqlName := ReadTableName(any)

	t := &Table{
		Name:    name,
		SQLName: sqlName,
		Fields:  fields,
	}
	t.sqlColumnDict = buildSQLColumnDict(fields)
	t.columnIndex = buildColumnIndex(fields)
	t.rowPlan = rowPlan
	t.insertPlan = sortRowPlan(rowPlan)

	tableCacheMu.Lock()
	tableCache[anyT] = t
	tableCacheMu.Unlock()

	return t, nil
}

type Table struct {
	Name    string
	SQLName string
	Fields  []*Field

	// sqlColumnDict maps SQL column name -> struct field name. Built once in
	// NewTable (the Table is cached by reflect.Type), so reads are free.
	sqlColumnDict map[string]string

	// columnIndex maps SQL column name -> struct field position, so the scan
	// path resolves columns to fields with a map lookup instead of a linear
	// FieldByName search per column.
	columnIndex map[string]int

	// rowPlan is the precomputed set of insertable columns in struct-field
	// order (matching the historical CollectRows walk). insertPlan is the same
	// set sorted by column name — the deterministic order the INSERT builder
	// needs. Both are built once per type in NewTable, so the write path reads
	// values via a cached plan instead of re-parsing tags per record.
	rowPlan    []rowPlanEntry
	insertPlan []rowPlanEntry
}

// rowPlanEntry maps one insertable struct field to its SQL column. autoIncSkip
// records whether the field carries an explicit auto-increment tag, in which
// case a zero int value is omitted from the insert (letting the DB assign it).
type rowPlanEntry struct {
	column      string
	index       int
	autoIncSkip bool
}

func (table *Table) SQLOptions() []*sql.Options {
	result := []*sql.Options{}

	for _, f := range table.Fields {
		result = append(result, f.SQL)
	}

	return result
}

func (table *Table) SQLColumnDict() map[string]string {
	return table.sqlColumnDict
}

func buildSQLColumnDict(fields []*Field) map[string]string {
	result := make(map[string]string, len(fields))

	for _, field := range fields {
		result[field.SQL.Name] = field.Name
	}

	return result
}

func buildColumnIndex(fields []*Field) map[string]int {
	result := make(map[string]int, len(fields))

	for _, field := range fields {
		result[field.SQL.Name] = field.Index
	}

	return result
}

// buildRowPlan captures, in struct-field order, the columns that participate in
// an insert. It mirrors CollectRows: fields flagged Generated are excluded
// (Ignore fields are already filtered by GetFieldsOf), and a field's explicit
// auto-increment tag is recorded so a zero value can be skipped per record.
func buildRowPlan(fields []*Field) []rowPlanEntry {
	plan := make([]rowPlanEntry, 0, len(fields))

	for _, field := range fields {
		if field.SQL.Generated {
			continue
		}

		plan = append(plan, rowPlanEntry{
			column:      field.SQL.Name,
			index:       field.Index,
			autoIncSkip: field.SQL.AutoIncrement > 0,
		})
	}

	return plan
}

// sortRowPlan returns a copy of the plan ordered by SQL column name, matching
// the deterministic column order the INSERT builder historically produced by
// sorting the SQLValues map keys.
func sortRowPlan(plan []rowPlanEntry) []rowPlanEntry {
	sorted := make([]rowPlanEntry, len(plan))
	copy(sorted, plan)

	slices.SortFunc(sorted, func(a, b rowPlanEntry) int {
		return strings.Compare(a.column, b.column)
	})

	return sorted
}

func (table *Table) PrimaryKeyField() *Field {
	for _, f := range table.Fields {
		if f.SQL.IsPrimaryKey {
			return f
		}
	}

	return nil
}

func (table *Table) SQLUpdateColumnSet() []string {
	columns := []string{}

	for _, f := range table.Fields {
		if f.SQL.Ignore || f.SQL.IsAutoIncrementing || f.SQL.Generated {
			continue
		}

		columns = append(columns, f.SQL.Name)
	}

	return columns
}

func (table *Table) SQLUpdateValueSet(record interface{}) []interface{} {
	rv := meta.ValueOf(record)
	values := []interface{}{}

	for _, f := range table.Fields {
		if f.SQL.Ignore || f.SQL.IsAutoIncrementing || f.SQL.Generated {
			continue
		}

		values = append(values, rv.Field(f.Index).Interface())
	}

	pk := table.PrimaryKeyField()
	if pk != nil {
		values = append(values, rv.Field(pk.Index).Interface())
	}

	return values
}

// Return struct name and SQL table name
func ReadTableName(any interface{}) (string, string) {
	return meta.TypeNameOf(any), SQLTableNameOf(any)
}

// Return table columns for given struct, pointer to struct or slice of structs.
func ReadTableColumns(any interface{}) ([]string, error) {
	if meta.IsSlice(any) {
		any = meta.CreateElement(any).Interface()
	}

	fields, err := GetFieldsOf(any)
	if err != nil {
		return nil, err
	}

	columns := []string{}

	for _, col := range fields {
		columns = append(columns, col.SQL.Name)
	}

	return columns, nil
}
