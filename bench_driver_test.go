package crud_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// BenchCandidateListItem is a narrow read model (6 of BenchCandidate's ~45
// columns) over the same table — what a list view actually needs.
type BenchCandidateListItem struct {
	Id        int    `sql:"name=id table-name=bench_candidates primary-key auto-increment unsigned"`
	Field01   string `sql:"name=field01"`
	IntField1 int    `sql:"name=int_field1"`
	Big1      int64  `sql:"name=big1"`
	Bool1     bool   `sql:"name=bool1"`
	OrgId     int    `sql:"name=org_id"`
}

const sixColumns = "id, field01, int_field1, big1, bool1, org_id"

// BenchmarkProjection compares a realistic "list view needs 6 of 47 columns"
// read three ways, over 100 rows.
func BenchmarkProjection(b *testing.B) {
	ctx := context.Background()
	require.NoError(b, seedBenchCandidates(ctx, 1000))

	// Baseline: full SELECT * into the wide 45-field entity.
	b.Run("full_47col_into_wide", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			var res []*BenchCandidate
			if err := DB.Read(ctx, &res, "SELECT * FROM bench_candidates LIMIT 100"); err != nil {
				b.Fatal(err)
			}
		}
	})

	// Project 6 columns, but still scan into the wide entity: cuts driver
	// per-column allocations, but the 45-field struct is still allocated per row.
	b.Run("project_6col_into_wide", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			var res []*BenchCandidate
			if err := DB.Read(ctx, &res, "SELECT "+sixColumns+" FROM bench_candidates LIMIT 100"); err != nil {
				b.Fatal(err)
			}
		}
	})

	// Project 6 columns into a 6-field read model: also drops the wide-struct
	// allocation — the full win.
	b.Run("project_6col_into_narrow", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			var res []*BenchCandidateListItem
			if err := DB.Read(ctx, &res, "SELECT "+sixColumns+" FROM bench_candidates LIMIT 100"); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// These benchmarks isolate where the go-sql-driver/mysql allocations come from,
// to evaluate driver-level levers: text vs binary protocol, and column count.
// All read ~100 wide rows so the per-row driver cost dominates.

func BenchmarkDriver(b *testing.B) {
	ctx := context.Background()
	require.NoError(b, seedBenchCandidates(ctx, 1000))

	// Text protocol: a parameter-less query goes through textRows (string/[]byte
	// per column per row + database/sql string->type conversion).
	b.Run("text_protocol_wide", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			var res []*BenchCandidate
			if err := DB.Read(ctx, &res, "SELECT * FROM bench_candidates LIMIT 100"); err != nil {
				b.Fatal(err)
			}
		}
	})

	// Binary protocol: a parameterized query goes through a prepared statement
	// and binaryRows (numerics parsed directly from bytes, no strconv).
	b.Run("binary_protocol_wide", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			var res []*BenchCandidate
			if err := DB.Read(ctx, &res,
				"SELECT * FROM bench_candidates WHERE id < ? LIMIT 100", 100000000); err != nil {
				b.Fatal(err)
			}
		}
	})

	// Narrow column set: same rows, only 5 of 45 columns selected.
	b.Run("text_protocol_narrow", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			var res []*BenchCandidate
			if err := DB.Read(ctx, &res,
				"SELECT id, field01, int_field1, big1, bool1 FROM bench_candidates LIMIT 100"); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("binary_protocol_narrow", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			var res []*BenchCandidate
			if err := DB.Read(ctx, &res,
				"SELECT id, field01, int_field1, big1, bool1 FROM bench_candidates WHERE id < ? LIMIT 100", 100000000); err != nil {
				b.Fatal(err)
			}
		}
	})
}
