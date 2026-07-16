package crud_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// These benchmarks model the entity ID validator hot path. Every Save issues
// an existence check on each referenced ID (e.g. organizations.IsValidId calls
// Get → SELECT <all columns> FROM table WHERE id = ?, scanned into the full
// wide struct). The question: how much does swapping that full-row Get for a
// COUNT(id) save? Both are a single round-trip on a primary-key lookup, so the
// DB-side cost is near-identical; the delta is column transfer + the scan into
// a 45-field struct vs a single int.
//
// Run with:
//   go test -run=^$ -bench=BenchmarkValidateExists -benchmem ./...

// BenchmarkValidateExists_Get is the current validator: read the entire row
// into the wide entity struct just to check rec != nil.
func BenchmarkValidateExists_Get(b *testing.B) {
	ctx := context.Background()
	require.NoError(b, seedBenchCandidates(ctx, 1000))

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		var result BenchCandidate
		if err := DB.Read(ctx, &result, "SELECT * FROM bench_candidates WHERE id = ?", 1); err != nil {
			b.Fatal(err)
		}
		if result.Id == 0 {
			b.Fatal("expected row")
		}
	}
}

// BenchmarkValidateExists_Count is the proposed validator: ask the DB for the
// count and scan a single int. Same PK lookup, no wide-row materialization.
func BenchmarkValidateExists_Count(b *testing.B) {
	ctx := context.Background()
	require.NoError(b, seedBenchCandidates(ctx, 1000))

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		var count int
		if err := DB.Read(ctx, &count, "SELECT COUNT(id) FROM bench_candidates WHERE id = ?", 1); err != nil {
			b.Fatal(err)
		}
		if count == 0 {
			b.Fatal("expected row")
		}
	}
}
