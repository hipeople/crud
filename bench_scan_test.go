package crud_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// BenchCandidate is a wide struct (~45 columns) modelled on the production
// Candidate / OpenJobRole entities, so the scan path pays a realistic per-row
// cost: one []interface{} alloc and one FieldByName lookup per column, per row.
type BenchCandidate struct {
	Id        int    `sql:"name=id table-name=bench_candidates primary-key auto-increment unsigned"`
	Field01   string `sql:"name=field01"`
	Field02   string `sql:"name=field02"`
	Field03   string `sql:"name=field03"`
	Field04   string `sql:"name=field04"`
	Field05   string `sql:"name=field05"`
	Field06   string `sql:"name=field06"`
	Field07   string `sql:"name=field07"`
	Field08   string `sql:"name=field08"`
	Field09   string `sql:"name=field09"`
	Field10   string `sql:"name=field10"`
	IntField1 int    `sql:"name=int_field1"`
	IntField2 int    `sql:"name=int_field2"`
	IntField3 int    `sql:"name=int_field3"`
	IntField4 int    `sql:"name=int_field4"`
	IntField5 int    `sql:"name=int_field5"`
	IntField6 int    `sql:"name=int_field6"`
	IntField7 int    `sql:"name=int_field7"`
	IntField8 int    `sql:"name=int_field8"`
	Big1      int64  `sql:"name=big1"`
	Big2      int64  `sql:"name=big2"`
	Big3      int64  `sql:"name=big3"`
	Big4      int64  `sql:"name=big4"`
	Big5      int64  `sql:"name=big5"`
	Big6      int64  `sql:"name=big6"`
	Bool1     bool   `sql:"name=bool1"`
	Bool2     bool   `sql:"name=bool2"`
	Bool3     bool   `sql:"name=bool3"`
	Bool4     bool   `sql:"name=bool4"`
	Bool5     bool   `sql:"name=bool5"`
	Str11     string `sql:"name=str11"`
	Str12     string `sql:"name=str12"`
	Str13     string `sql:"name=str13"`
	Str14     string `sql:"name=str14"`
	Str15     string `sql:"name=str15"`
	Str16     string `sql:"name=str16"`
	Str17     string `sql:"name=str17"`
	Str18     string `sql:"name=str18"`
	Str19     string `sql:"name=str19"`
	Str20     string `sql:"name=str20"`
	CreatedAt int64  `sql:"name=created_at"`
	UpdatedAt int64  `sql:"name=updated_at"`
	DeletedAt int64  `sql:"name=deleted_at"`
	OrgId     int    `sql:"name=org_id"`
	RoleId    int    `sql:"name=role_id"`
}

// seedBenchCandidates creates the table once and fills it with n rows. Safe to
// call repeatedly; it only seeds when the table is short of n rows.
func seedBenchCandidates(ctx context.Context, n int) error {
	if err := DB.CreateTables(ctx, BenchCandidate{}); err != nil {
		return err
	}

	var existing []*BenchCandidate
	if err := DB.Read(ctx, &existing, "SELECT id FROM bench_candidates LIMIT ?", n); err != nil {
		return err
	}
	if len(existing) >= n {
		return nil
	}

	rows := make([]BenchCandidate, 0, n)
	for i := range make([]struct{}, n) {
		rows = append(rows, BenchCandidate{
			Field01: "alpha", Field05: "bravo", Field10: "charlie",
			IntField1: i, Big1: int64(i) * 1000,
			Bool1: i%2 == 0, Str11: "delta", Str20: "echo",
			OrgId: i % 50, RoleId: i % 200,
		})
	}
	return DB.BulkCreate(ctx, rows)
}

// BenchmarkReadList measures the hot list path: SELECT N wide rows scanned into
// a []*BenchCandidate. Run with -benchmem; allocs/op is the GC-pressure signal.
// The DB round-trip cost is constant across crud changes, so the allocs/op and
// ns/op deltas between runs attribute to the scan path.
func BenchmarkReadList(b *testing.B) {
	ctx := context.Background()
	require.NoError(b, seedBenchCandidates(ctx, 1000))

	for _, size := range []int{1, 100, 1000} {
		b.Run(fmt.Sprintf("rows=%d", size), func(b *testing.B) {
			query := fmt.Sprintf("SELECT * FROM bench_candidates LIMIT %d", size)
			b.ReportAllocs()

			for b.Loop() {
				var result []*BenchCandidate
				if err := DB.Read(ctx, &result, query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkReadOne measures the single-row path (candidates.First etc.).
func BenchmarkReadOne(b *testing.B) {
	ctx := context.Background()
	require.NoError(b, seedBenchCandidates(ctx, 1000))

	b.ReportAllocs()

	for b.Loop() {
		var result BenchCandidate
		if err := DB.Read(ctx, &result, "SELECT * FROM bench_candidates LIMIT 1"); err != nil {
			b.Fatal(err)
		}
	}
}
