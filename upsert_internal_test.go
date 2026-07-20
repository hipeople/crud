package crud

import (
	"context"
	stdsql "database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These internal tests lock in the exact SQL produced by upsertAndGetResult —
// the INSERT ... ON DUPLICATE KEY UPDATE builder — before switching its string
// concatenation to a strings.Builder, so the refactor is provably byte-identical.

type wNoUpd struct {
	Id        int   `sql:"name=id primary-key auto-increment"`
	CreatedAt int64 `sql:"name=created_at no-update"`
	Val       int   `sql:"name=val"`
}

// captureUpsertQuery runs upsertAndGetResult with a no-op exec that records the
// generated query string.
func captureUpsertQuery(t *testing.T, rec interface{}) string {
	t.Helper()

	var got string
	exec := func(_ context.Context, q string, _ ...interface{}) (stdsql.Result, error) {
		got = q
		return upsertNopResult{}, nil
	}

	_, err := upsertAndGetResult(context.Background(), exec, rec)
	require.NoError(t, err)
	return got
}

type upsertNopResult struct{}

func (upsertNopResult) LastInsertId() (int64, error) { return 1, nil }
func (upsertNopResult) RowsAffected() (int64, error) { return 1, nil }

func TestUpsertQuery_IncludesNonZeroPK(t *testing.T) {
	assert.Equal(t,
		"INSERT INTO `w_recs` (`age`, `id`, `name`) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE `id` = LAST_INSERT_ID(`id`), `age` = VALUES(`age`), `name` = VALUES(`name`)",
		captureUpsertQuery(t, wRec{Id: 5, Name: "x", Age: 7}),
	)
}

func TestUpsertQuery_SkipsZeroAutoIncPK(t *testing.T) {
	// id is dropped from the INSERT column list, but the UPDATE clause still
	// touches it via LAST_INSERT_ID so readLastInsert can re-read the row.
	assert.Equal(t,
		"INSERT INTO `w_recs` (`age`, `name`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `id` = LAST_INSERT_ID(`id`), `age` = VALUES(`age`), `name` = VALUES(`name`)",
		captureUpsertQuery(t, wRec{Name: "x", Age: 7}),
	)
}

func TestUpsertQuery_ExcludesNoUpdateColumns(t *testing.T) {
	// created_at carries no-update, so it appears in the INSERT but not the
	// UPDATE set.
	assert.Equal(t,
		"INSERT INTO `w_no_upds` (`created_at`, `id`, `val`) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE `id` = LAST_INSERT_ID(`id`), `val` = VALUES(`val`)",
		captureUpsertQuery(t, wNoUpd{Id: 3, CreatedAt: 100, Val: 9}),
	)
}

// BenchmarkUpsertQueryBuild measures the upsert query-building path (no DB
// round-trip) with a no-op exec, on the wide struct.
func BenchmarkUpsertQueryBuild(b *testing.B) {
	ctx := context.Background()
	exec := func(context.Context, string, ...interface{}) (stdsql.Result, error) {
		return upsertNopResult{}, nil
	}
	rec := wWide{Id: 1, C01: "a", C05: "b", I01: 7, B01: 7000, Bo01: true, S01: "c", Org: 3, Role: 42}

	b.ReportAllocs()
	for b.Loop() {
		if _, err := upsertAndGetResult(ctx, exec, rec); err != nil {
			b.Fatal(err)
		}
	}
}
