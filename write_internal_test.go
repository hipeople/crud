package crud

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These internal tests lock in the behavior of valuesForRecord — the create /
// insert metadata path (columns + values + table name) — before the perf
// refactor that routes it through the cached Table. valuesForRecord is
// unexported, so the characterization lives in package crud.

type wRec struct {
	Id   int    `sql:"name=id primary-key auto-increment"`
	Name string `sql:"name=name"`
	Age  int    `sql:"name=age"`
}

type wCustom struct {
	Foo int `sql:"table-name=yolo name=foo"`
}

type wGen struct {
	Id   int    `sql:"name=id primary-key auto-increment"`
	Comp string `sql:"name=comp generated"`
	Name string `sql:"name=name"`
}

// wWide is a wide struct (~24 columns) so the valuesForRecord benchmark pays a
// realistic per-record cost — mirroring the production write path.
type wWide struct {
	Id   int    `sql:"name=id primary-key auto-increment unsigned"`
	C01  string `sql:"name=c01"`
	C02  string `sql:"name=c02"`
	C03  string `sql:"name=c03"`
	C04  string `sql:"name=c04"`
	C05  string `sql:"name=c05"`
	C06  string `sql:"name=c06"`
	C07  string `sql:"name=c07"`
	C08  string `sql:"name=c08"`
	I01  int    `sql:"name=i01"`
	I02  int    `sql:"name=i02"`
	I03  int    `sql:"name=i03"`
	I04  int    `sql:"name=i04"`
	B01  int64  `sql:"name=b01"`
	B02  int64  `sql:"name=b02"`
	B03  int64  `sql:"name=b03"`
	Bo01 bool   `sql:"name=bo01"`
	Bo02 bool   `sql:"name=bo02"`
	S01  string `sql:"name=s01"`
	S02  string `sql:"name=s02"`
	S03  string `sql:"name=s03"`
	Org  int    `sql:"name=org_id"`
	Role int    `sql:"name=role_id"`
}

// BenchmarkValuesForRecord measures the create/insert metadata extraction —
// the columns+values a single INSERT is built from. This is finding #2's hot
// path (formerly a map build + column sort per record).
func BenchmarkValuesForRecord(b *testing.B) {
	rec := wWide{Id: 1, C01: "a", C05: "b", I01: 7, B01: 7000, Bo01: true, S01: "c", Org: 3, Role: 42}
	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		if _, _, _, err := valuesForRecord(rec); err != nil {
			b.Fatal(err)
		}
	}
}

// Columns are emitted sorted by SQL column name, with a zero explicit
// auto-increment PK skipped.
func TestValuesForRecord_SortedColumnsSkipZeroPK(t *testing.T) {
	row, columns, values, err := valuesForRecord(wRec{Name: "x", Age: 5})
	require.NoError(t, err)

	assert.Equal(t, "w_recs", row.SQLTableName)
	assert.Equal(t, []string{"age", "name"}, columns)
	assert.Equal(t, []interface{}{5, "x"}, values)
}

// A non-zero auto-increment PK is included, still in sorted column order.
func TestValuesForRecord_IncludesNonZeroPK(t *testing.T) {
	_, columns, values, err := valuesForRecord(wRec{Id: 7, Name: "x", Age: 5})
	require.NoError(t, err)

	assert.Equal(t, []string{"age", "id", "name"}, columns)
	assert.Equal(t, []interface{}{5, 7, "x"}, values)
}

// A pointer record yields identical columns/values to a value record.
func TestValuesForRecord_Pointer(t *testing.T) {
	_, columns, values, err := valuesForRecord(&wRec{Id: 7, Name: "x", Age: 5})
	require.NoError(t, err)

	assert.Equal(t, []string{"age", "id", "name"}, columns)
	assert.Equal(t, []interface{}{5, 7, "x"}, values)
}

// A custom table-name option is honored for the row's table name.
func TestValuesForRecord_CustomTableName(t *testing.T) {
	row, columns, _, err := valuesForRecord(wCustom{Foo: 1})
	require.NoError(t, err)

	assert.Equal(t, "yolo", row.SQLTableName)
	assert.Equal(t, []string{"foo"}, columns)
}

// Generated columns are excluded from the insert set.
func TestValuesForRecord_SkipsGenerated(t *testing.T) {
	_, columns, values, err := valuesForRecord(wGen{Name: "x"})
	require.NoError(t, err)

	assert.Equal(t, []string{"name"}, columns)
	assert.Equal(t, []interface{}{"x"}, values)
}
