package crud_test

import (
	"context"
	"testing"

	"github.com/hipeople/crud/v2"
	"github.com/stretchr/testify/require"
)

func benchRecord() BenchCandidate {
	return BenchCandidate{
		Field01: "alpha", Field05: "bravo", Field10: "charlie",
		IntField1: 7, Big1: 7000, Bool1: true, Str11: "delta", Str20: "echo",
		OrgId: 3, RoleId: 42,
	}
}

// BenchmarkGetRowValues isolates the per-row write metadata extraction
// (CollectRows: tag parse + column name + value read) with no DB round-trip.
// This is the work BulkCreate repeats once per record.
func BenchmarkGetRowValues(b *testing.B) {
	rec := benchRecord()
	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		if _, err := crud.GetRowValuesOf(rec); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUpdateValueSet isolates the per-update value extraction
// (Table.SQLUpdateValueSet), currently a linear FieldByName per field.
func BenchmarkUpdateValueSet(b *testing.B) {
	table, err := crud.NewTable(BenchCandidate{})
	require.NoError(b, err)
	rec := benchRecord()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = table.SQLUpdateValueSet(rec)
	}
}

// BenchmarkBulkCreate measures an end-to-end 100-row bulk insert.
func BenchmarkBulkCreate(b *testing.B) {
	ctx := context.Background()
	require.NoError(b, DB.CreateTables(ctx, BenchCandidate{}))

	batch := make([]BenchCandidate, 100)
	for i := range batch {
		batch[i] = benchRecord()
	}

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		if err := DB.BulkCreate(ctx, batch); err != nil {
			b.Fatal(err)
		}
	}
}
