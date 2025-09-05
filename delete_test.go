package crud_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*func TestDelete(t *testing.T) {
	assert.Nil(t, CreateUserProfiles())

	nova := UserProfile{}
	err := DB.Read(&nova, "SELECT * FROM user_profile WHERE name = 'Nova'")
	assert.Nil(t, err)

	assert.Nil(t, DB.Delete(nova))

	novac := UserProfile{}
	err = DB.Read(&novac, "SELECT * FROM user_profile WHERE name = 'Nova'")
	assert.NotNil(t, err)
}

func TestDeleteNotMatching(t *testing.T) {
	assert.Nil(t, DB.Delete(&UserProfile{
		Id:   123,
		Name: "Yolo",
	}))
}*/

func TestMustDelete(t *testing.T) {
	ctx := context.Background()

	require.NoError(t, CreateUserProfiles(ctx))

	nova := UserProfile{}
	err := DB.Read(ctx, &nova, "SELECT * FROM user_profiles WHERE name = 'Nova'")
	require.NoError(t, err)
	require.NotZero(t, nova.Id)

	require.NoError(t, DB.Delete(ctx, nova))

	novac := UserProfile{}
	err = DB.Read(ctx, &novac, "SELECT * FROM user_profiles WHERE name = 'Nova'")
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestMustDeleteNotMatching(t *testing.T) {
	ctx := context.Background()

	assert.NotNil(t, DB.Delete(ctx, &UserProfile{
		Id:   123,
		Name: "Yolo",
	}))
}

func TestBulkDelete(t *testing.T) {
	ctx := context.Background()

	require.NoError(t, CreateUserProfiles(ctx))

	nova := UserProfile{}
	err := DB.Read(ctx, &nova, "SELECT * FROM user_profiles WHERE name = 'Nova'")
	assert.Nil(t, err)

	assert.Nil(t, DB.BulkDelete(ctx, []UserProfile{nova}))

	novac := UserProfile{}
	err = DB.Read(ctx, &novac, "SELECT * FROM user_profile WHERE name = 'Nova'")
	assert.NotNil(t, err)
}
