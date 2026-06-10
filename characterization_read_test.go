package crud_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These lock in how the scan path maps result columns to struct fields, before
// caching the column->field-index resolution. They cover the cases that would
// break a wrong cache: reordered columns, a subset of columns, custom column
// names (email, modified_col), and an extra column with no matching field.

func TestCharacterization_ReadReorderedSubset(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, CreateUserProfiles(ctx))

	var rows []*UserProfile
	// Columns out of struct order, a subset, including a custom-named column.
	err := DB.Read(ctx, &rows,
		"SELECT modified_col, email, name FROM user_profiles WHERE name = 'Azer'")
	require.NoError(t, err)
	require.Len(t, rows, 1)

	assert.Equal(t, "Azer", rows[0].Name)
	assert.Equal(t, "azer@roadbeats.com", rows[0].Email)
	// Unselected columns stay zero-valued.
	assert.Equal(t, "", rows[0].Bio)
	assert.Nil(t, rows[0].Attachment)
}

func TestCharacterization_ReadExtraColumn(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, CreateUserProfiles(ctx))

	var rows []*UserProfile
	// A computed column with no matching struct field must be ignored, and the
	// real columns must still populate correctly.
	err := DB.Read(ctx, &rows,
		"SELECT *, 99 AS bogus_column FROM user_profiles WHERE name = 'Nova'")
	require.NoError(t, err)
	require.Len(t, rows, 1)

	assert.Equal(t, "Nova", rows[0].Name)
	assert.Equal(t, "Photographer", rows[0].Bio)
	assert.Equal(t, "nova@roadbeats.com", rows[0].Email)
}

func TestCharacterization_ReadValueAndPointerSlices(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, CreateUserProfiles(ctx))

	var ptrs []*UserProfile
	require.NoError(t, DB.Read(ctx, &ptrs, "SELECT * FROM user_profiles WHERE name='Azer'"))
	require.Len(t, ptrs, 1)

	var vals []UserProfile
	require.NoError(t, DB.Read(ctx, &vals, "SELECT * FROM user_profiles WHERE name='Azer'"))
	require.Len(t, vals, 1)

	// Pointer-slice and value-slice scans must agree field-for-field.
	assert.Equal(t, ptrs[0].Name, vals[0].Name)
	assert.Equal(t, ptrs[0].Email, vals[0].Email)
	assert.Equal(t, ptrs[0].Modified, vals[0].Modified)
}
