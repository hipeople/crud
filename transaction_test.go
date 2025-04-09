package crud_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSuccessfulCommit(t *testing.T) {
	ctx := context.Background()

	assert.Nil(t, CreateUserProfiles(ctx))

	tx, err := DB.Begin(context.Background(), false)
	assert.Nil(t, err)

	n := UserProfile{}
	err = tx.Read(ctx, &n, "SELECT * from user_profiles WHERE id = ?", 2)
	assert.Nil(t, err)

	n.Bio = "let's go somewhere"

	assert.Nil(t, tx.Update(ctx, &n))

	azer := UserProfile{}
	err = DB.Read(ctx, &azer, "SELECT * from user_profiles WHERE id = ?", 2)
	assert.Nil(t, err)
	assert.Equal(t, azer.Bio, "Engineer")

	assert.Nil(t, tx.Commit(ctx))

	time.Sleep(time.Second * 1)

	azerc := UserProfile{}
	err = DB.Read(ctx, &azerc, "SELECT * from user_profiles WHERE id = ?", 2)
	assert.Nil(t, err)
	assert.Equal(t, "let's go somewhere", azerc.Bio)

	DB.DropTables(ctx, UserProfile{})
}

func TestRollback(t *testing.T) {
	ctx := context.Background()

	assert.Nil(t, CreateUserProfiles(ctx))

	tx, err := DB.Begin(context.Background(), false)
	assert.Nil(t, err)

	err = tx.Create(ctx, &UserProfile{
		Email: "row1@rows.com",
		Name:  "Row1",
		Bio:   "testing transactions",
	})

	assert.Nil(t, err)

	err = tx.Create(ctx, &UserProfile{
		Email: "row2@rows.com",
		Name:  "Row2",
		Bio:   "testing transactions",
	})

	assert.Nil(t, err)

	err = tx.Create(ctx, &UserProfile{
		Email: "row1@rows.com",
		Name:  "Row3",
		Bio:   "testing transactions, should fail",
	})

	assert.Error(t, err)
	assert.Nil(t, tx.Rollback(ctx))

	shouldNotExist := UserProfile{}
	err = DB.Read(ctx, &shouldNotExist, "SELECT * from user_profiles WHERE email = ?", "row1@rows.com")
	assert.Error(t, err)
	assert.True(t, err == sql.ErrNoRows)

	DB.DropTables(ctx, UserProfile{})
}
