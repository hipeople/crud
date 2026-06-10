package crud_test

import (
	"testing"

	"github.com/hipeople/crud/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests lock in the current behavior of the write/update metadata paths
// (CollectRows via GetRowValuesOf, and the Table update column/value sets)
// before any performance refactor, so a behavior-preserving change is provable.

func TestCharacterization_UpdateColumnSet(t *testing.T) {
	table, err := crud.NewTable(UserProfile{})
	require.NoError(t, err)

	// Auto-increment PK (Id) is excluded from the update column set.
	assert.Equal(t,
		[]string{"name", "bio", "email", "attachment", "modified_col"},
		table.SQLUpdateColumnSet(),
	)
}

func TestCharacterization_UpdateValueSet(t *testing.T) {
	table, err := crud.NewTable(UserProfile{})
	require.NoError(t, err)

	values := table.SQLUpdateValueSet(UserProfile{
		Id:         7,
		Name:       "Azer",
		Bio:        "Engineer",
		Email:      "azer@roadbeats.com",
		Attachment: []byte("{}"),
		Modified:   42,
	})

	// Updatable columns in field order, followed by the PK (for the WHERE clause).
	require.Len(t, values, 6)
	assert.Equal(t, "Azer", values[0])
	assert.Equal(t, "Engineer", values[1])
	assert.Equal(t, "azer@roadbeats.com", values[2])
	assert.Equal(t, []byte("{}"), values[3])
	assert.Equal(t, int64(42), values[4])
	assert.Equal(t, 7, values[5]) // trailing primary key
}

func TestCharacterization_UpdateValueSetFromPointer(t *testing.T) {
	table, err := crud.NewTable(UserProfile{})
	require.NoError(t, err)

	// Pointer record must yield identical values to a value record.
	values := table.SQLUpdateValueSet(&UserProfile{Id: 9, Name: "Nova", Modified: 1})
	require.Len(t, values, 6)
	assert.Equal(t, "Nova", values[0])
	assert.Equal(t, int64(1), values[4])
	assert.Equal(t, 9, values[5])
}

func TestCharacterization_RowValuesFromPointer(t *testing.T) {
	// CollectRows must behave identically for value and pointer inputs.
	want, err := crud.GetRowValuesOf(UserProfile{Name: "Azer", Email: "a@b.com", Modified: 5})
	require.NoError(t, err)

	got, err := crud.GetRowValuesOf(&UserProfile{Name: "Azer", Email: "a@b.com", Modified: 5})
	require.NoError(t, err)

	require.Equal(t, len(want), len(got))
	for i := range want {
		assert.Equal(t, want[i].SQLColumn, got[i].SQLColumn)
		assert.Equal(t, want[i].Value, got[i].Value)
	}
}

func TestCharacterization_RowValuesWideStruct(t *testing.T) {
	// Locks the column->value mapping for the wide benchmark struct, including
	// the auto-increment PK skip (Id is omitted when zero).
	rows, err := crud.GetRowValuesOf(BenchCandidate{
		Field01: "alpha", IntField1: 7, Big1: 7000, Bool1: true, OrgId: 3,
	})
	require.NoError(t, err)

	byCol := map[string]interface{}{}
	for _, r := range rows {
		byCol[r.SQLColumn] = r.Value
	}

	_, hasID := byCol["id"]
	assert.False(t, hasID, "auto-increment id should be skipped when zero")
	assert.Equal(t, "alpha", byCol["field01"])
	assert.Equal(t, 7, byCol["int_field1"])
	assert.Equal(t, int64(7000), byCol["big1"])
	assert.Equal(t, true, byCol["bool1"])
	assert.Equal(t, 3, byCol["org_id"])
}
