package crud_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hipeople/crud/v2"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	ctx := context.Background()

	err := DB.ResetTables(ctx, UserProfile{})
	assert.Nil(t, err)

	azer := UserProfile{
		Name:  "Azer",
		Bio:   "I like photography",
		Email: "azer@roadbeats.com",
	}

	err = DB.Create(ctx, azer)
	assert.Nil(t, err)

	DB.DropTables(ctx, UserProfile{})
}

func TestCreateAndRead(t *testing.T) {
	ctx := context.Background()

	DB.ResetTables(ctx, UserProfile{})

	azer := UserProfile{
		Name:  "Azer",
		Bio:   "I like photography",
		Email: "azer@roadbeats.com",
	}

	assert.Equal(t, azer.Id, 0)
	err := DB.CreateAndRead(ctx, &azer)
	assert.Nil(t, err)
	assert.Equal(t, azer.Id, 1)

	DB.DropTables(ctx, UserProfile{})
}

func TestCreateEmpty(t *testing.T) {
	ctx := context.Background()

	DB.ResetTables(ctx, UserProfile{})

	azer := UserProfile{
		Name: "Azer",
	}

	err := DB.Create(ctx, azer)
	assert.Nil(t, err)

	DB.DropTables(ctx, UserProfile{})
}

func TestCreatingRenamedTableRow(t *testing.T) {
	t.Skip("timestamp not working with MariaDB")

	ctx := context.Background()

	DB.ResetTables(ctx, Post{})

	p := Post{
		Title:     "Foo",
		Text:      "bar",
		Category:  PostCategoryDrink,
		CreatedAt: time.Now(),
	}

	assert.Equal(t, p.Id, 0)
	err := DB.CreateAndRead(ctx, &p)
	assert.Nil(t, err)
	assert.Equal(t, p.Id, 1)

	DB.DropTables(ctx, Post{})
}

func TestCreateDuplicateEntry(t *testing.T) {
	ctx := context.Background()

	err := DB.ResetTables(ctx, UserProfile{})
	assert.Nil(t, err)

	azer := UserProfile{
		Name:  "Azer",
		Bio:   "I like photography",
		Email: "azer@roadbeats.com",
	}

	err = DB.Create(ctx, azer)
	assert.Nil(t, err)

	duplicate := UserProfile{
		Name:  "Azer Clone",
		Bio:   "Same email, different person",
		Email: "azer@roadbeats.com",
	}

	err = DB.Create(ctx, duplicate)
	assert.True(t, errors.Is(err, crud.ErrDuplicateEntry))

	DB.DropTables(ctx, UserProfile{})
}

func TestCreateBulk(t *testing.T) {
	ctx := context.Background()

	DB.ResetTables(ctx, UserProfile{})

	users := []UserProfile{
		{
			Name:  "Azer",
			Bio:   "I like photography",
			Email: "azer@roadbeats.com",
		},
		{
			Name:  "Azer2",
			Bio:   "I like photography2",
			Email: "azer2@roadbeats.com",
		},
		{
			Name:  "Azer3",
			Bio:   "I like photography3",
			Email: "azer3@roadbeats.com",
		},
	}

	err := DB.BulkCreate(ctx, users)
	assert.Nil(t, err)

	var users2 []UserProfile
	err = DB.Read(ctx, &users2, "SELECT * FROM user_profiles")
	assert.Nil(t, err)
	assert.Equal(t, len(users), 3)

	for i, user := range users2 {
		assert.Equal(t, user.Name, users[i].Name)
		assert.Equal(t, user.Bio, users[i].Bio)
		assert.Equal(t, user.Email, users[i].Email)
	}

	DB.DropTables(ctx, UserProfile{})
}
