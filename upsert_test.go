package crud_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpsert(t *testing.T) {
	ctx := context.Background()

	err := DB.ResetTables(ctx, UserProfile{})
	assert.Nil(t, err)

	azer := UserProfile{
		Name:  "Azer",
		Bio:   "I like photography",
		Email: "azer@roadbeats.com",
	}

	err = DB.Upsert(ctx, azer)
	assert.Nil(t, err)

	DB.DropTables(ctx, UserProfile{})
}

func TestUpsertAndRead(t *testing.T) {
	ctx := context.Background()

	DB.ResetTables(ctx, UserProfile{})

	azer := UserProfile{
		Name:  "Azer",
		Bio:   "I like photography",
		Email: "azer@roadbeats.com",
	}

	assert.Equal(t, azer.Id, 0)
	err := DB.UpsertAndRead(ctx, &azer)
	assert.Nil(t, err)
	assert.Equal(t, azer.Id, 1)
	assert.Equal(t, azer.Name, "Azer")
	assert.Equal(t, azer.Bio, "I like photography")

	DB.DropTables(ctx, UserProfile{})
}

func TestUpsertUpdate(t *testing.T) {
	ctx := context.Background()

	DB.ResetTables(ctx, UserProfile{})

	// First insert
	azer := UserProfile{
		Name:  "Azer",
		Bio:   "I like photography",
		Email: "azer@roadbeats.com",
	}

	err := DB.UpsertAndRead(ctx, &azer)
	assert.Nil(t, err)
	assert.Equal(t, azer.Id, 1)

	// Update using upsert with same email (unique constraint)
	azer2 := UserProfile{
		Name:  "Azer Updated",
		Bio:   "I like music",
		Email: "azer@roadbeats.com",
	}

	err = DB.UpsertAndRead(ctx, &azer2)
	assert.Nil(t, err)

	// Read back and verify update
	var result UserProfile
	err = DB.Read(ctx, &result, "SELECT * FROM user_profiles WHERE email = ?", "azer@roadbeats.com")
	assert.Nil(t, err)
	assert.Equal(t, result.Name, "Azer Updated")
	assert.Equal(t, result.Bio, "I like music")
	assert.Equal(t, result.Email, "azer@roadbeats.com")

	DB.DropTables(ctx, UserProfile{})
}

func TestUpsertWithMixed(t *testing.T) {
	ctx := context.Background()

	DB.ResetTables(ctx, Mixed{})

	// First insert
	m1 := Mixed{
		UserId:    1,
		Secret:    "secret1",
		CreatedAt: 100,
		UpdatedAt: 100,
	}

	err := DB.UpsertAndRead(ctx, &m1)
	assert.Nil(t, err)
	assert.Equal(t, m1.Id, 1)
	assert.Equal(t, m1.CreatedAt, int64(100))

	// Update using upsert with same id (primary key)
	m2 := Mixed{
		Id:        1,
		UserId:    2,
		Secret:    "secret2",
		CreatedAt: 999, // Should NOT change because created_at has no-update tag
		UpdatedAt: 200, // Should update
	}

	err = DB.Upsert(ctx, &m2)
	assert.Nil(t, err)

	// Read back and verify
	var result Mixed
	err = DB.Read(ctx, &result, "SELECT * FROM __mixed__ WHERE id = ?", 1)
	assert.Nil(t, err)
	assert.Equal(t, result.Id, 1)
	assert.Equal(t, result.UserId, 2)
	assert.Equal(t, result.Secret, "secret2")
	assert.Equal(t, result.CreatedAt, int64(100)) // created_at preserved (no-update tag)
	assert.Equal(t, result.UpdatedAt, int64(200))

	DB.DropTables(ctx, Mixed{})
}

func TestUpsertMultipleRows(t *testing.T) {
	ctx := context.Background()

	DB.ResetTables(ctx, UserProfile{})

	// Insert multiple users
	users := []UserProfile{
		{
			Name:  "User1",
			Bio:   "Bio1",
			Email: "user1@test.com",
		},
		{
			Name:  "User2",
			Bio:   "Bio2",
			Email: "user2@test.com",
		},
		{
			Name:  "User3",
			Bio:   "Bio3",
			Email: "user3@test.com",
		},
	}

	for i := range users {
		err := DB.UpsertAndRead(ctx, &users[i])
		assert.Nil(t, err)
		assert.NotEqual(t, users[i].Id, 0)
	}

	// Update one of them
	updateUser := UserProfile{
		Name:  "User2 Updated",
		Bio:   "Bio2 Updated",
		Email: "user2@test.com",
	}

	err := DB.Upsert(ctx, &updateUser)
	assert.Nil(t, err)

	// Verify update
	var result UserProfile
	err = DB.Read(ctx, &result, "SELECT * FROM user_profiles WHERE email = ?", "user2@test.com")
	assert.Nil(t, err)
	assert.Equal(t, result.Name, "User2 Updated")
	assert.Equal(t, result.Bio, "Bio2 Updated")

	// Verify others are unchanged
	var user1 UserProfile
	err = DB.Read(ctx, &user1, "SELECT * FROM user_profiles WHERE email = ?", "user1@test.com")
	assert.Nil(t, err)
	assert.Equal(t, user1.Name, "User1")

	DB.DropTables(ctx, UserProfile{})
}

func TestUpsertTransaction(t *testing.T) {
	ctx := context.Background()

	DB.ResetTables(ctx, UserProfile{})

	tx, err := DB.Begin(ctx, false)
	assert.Nil(t, err)

	azer := UserProfile{
		Name:  "Azer",
		Bio:   "I like photography",
		Email: "azer@roadbeats.com",
	}

	err = tx.UpsertAndRead(ctx, &azer)
	assert.Nil(t, err)
	assert.Equal(t, azer.Id, 1)

	// Update in same transaction
	azer2 := UserProfile{
		Name:  "Azer Updated",
		Bio:   "Updated bio",
		Email: "azer@roadbeats.com",
	}

	err = tx.Upsert(ctx, &azer2)
	assert.Nil(t, err)

	err = tx.Commit(ctx)
	assert.Nil(t, err)

	// Verify the update committed
	var result UserProfile
	err = DB.Read(ctx, &result, "SELECT * FROM user_profiles WHERE email = ?", "azer@roadbeats.com")
	assert.Nil(t, err)
	assert.Equal(t, result.Name, "Azer Updated")
	assert.Equal(t, result.Bio, "Updated bio")

	DB.DropTables(ctx, UserProfile{})
}
