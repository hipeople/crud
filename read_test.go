package crud_test

import (
	"context"
	"testing"

	"github.com/azer/crud/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveReadParams(t *testing.T) {
	query, params, err := crud.ResolveReadParams([]interface{}{})
	assert.Equal(t, query, "")
	assert.Equal(t, len(params), 0)
	assert.Nil(t, err)

	query, params, err = crud.ResolveReadParams([]interface{}{123, 456})
	assert.Equal(t, query, "")
	assert.Nil(t, params)
	assert.NotNil(t, err)

	query, params, err = crud.ResolveReadParams([]interface{}{"yolo", 456, 123})
	assert.Equal(t, query, "yolo")
	assert.Equal(t, len(params), 2)
	assert.Nil(t, err)
}

func TestReadingMultipleRows(t *testing.T) {
	ctx := context.Background()

	assert.Nil(t, CreateUserProfiles(ctx))

	result := []UserProfile{}
	err := DB.Read(ctx, &result, "SELECT * FROM user_profiles")
	assert.Nil(t, err)
	assert.Equal(t, len(result), 3)
	assert.Equal(t, result[0].Name, "Nova")
	assert.Equal(t, result[0].Bio, "Photographer")
	assert.Equal(t, result[0].Email, "nova@roadbeats.com")
	assert.Equal(t, result[1].Name, "Azer")
	assert.Equal(t, result[1].Bio, "Engineer")
	assert.Equal(t, result[1].Email, "azer@roadbeats.com")
	assert.Equal(t, string(result[1].Attachment), "{ \"azer\": \"bar\" }")

	resultptr := []*UserProfile{}
	err = DB.Read(ctx, &resultptr, "SELECT * FROM user_profiles")
	assert.Nil(t, err)
	assert.Equal(t, len(resultptr), 3)
	assert.Equal(t, resultptr[0].Name, "Nova")
	assert.Equal(t, resultptr[0].Bio, "Photographer")
	assert.Equal(t, resultptr[0].Email, "nova@roadbeats.com")
	assert.Equal(t, resultptr[1].Name, "Azer")
	assert.Equal(t, resultptr[1].Bio, "Engineer")
	assert.Equal(t, resultptr[1].Email, "azer@roadbeats.com")

	var results []*UserProfile
	err = DB.Read(ctx, &results, "SELECT * FROM user_profiles")
	assert.Nil(t, err)
	assert.Equal(t, len(results), 3)
	assert.Equal(t, results[0].Name, "Nova")
	assert.Equal(t, results[0].Bio, "Photographer")
	assert.Equal(t, results[0].Email, "nova@roadbeats.com")
	assert.Equal(t, results[1].Name, "Azer")
	assert.Equal(t, results[1].Bio, "Engineer")
	assert.Equal(t, results[1].Email, "azer@roadbeats.com")

	var notmatching []*UserProfile
	err = DB.Read(ctx, &notmatching, "SELECT * FROM user_profiles WHERE name='not matching'")
	assert.Nil(t, err)
	assert.Equal(t, len(notmatching), 0)
}

func TestReadingSingleRow(t *testing.T) {
	ctx := context.Background()

	assert.Nil(t, CreateUserProfiles(ctx))

	nova := UserProfile{}
	err := DB.Read(ctx, &nova, "SELECT * FROM user_profiles WHERE name = ?", "Nova")
	assert.Nil(t, err)
	assert.Equal(t, nova.Id, 1)
	assert.Equal(t, nova.Name, "Nova")
	assert.Equal(t, nova.Bio, "Photographer")
	assert.Equal(t, nova.Email, "nova@roadbeats.com")

	var azer *UserProfile = &UserProfile{}
	err = DB.Read(ctx, azer, "SELECT * FROM user_profiles WHERE name = ?", "Azer")
	assert.Nil(t, err)
	assert.Equal(t, azer.Id, 2)
	assert.Equal(t, azer.Name, "Azer")
	assert.Equal(t, azer.Bio, "Engineer")
	assert.Equal(t, azer.Email, "azer@roadbeats.com")

	var az UserProfile
	err = DB.Read(ctx, &az, "SELECT * FROM user_profiles WHERE name = ?", "Azer")
	assert.Nil(t, err)
	assert.Equal(t, az.Id, 2)
	assert.Equal(t, az.Name, "Azer")
	assert.Equal(t, az.Bio, "Engineer")
	assert.Equal(t, az.Email, "azer@roadbeats.com")

	no := UserProfile{}
	err = DB.Read(ctx, &no, "SELECT * FROM user_profiles WHERE name = ?", "Not matching")
	assert.NotNil(t, err)

	DB.DropTables(ctx, UserProfile{})
}

func TestGeneratingQueries(t *testing.T) {
	ctx := context.Background()

	assert.Nil(t, CreateUserProfiles(ctx))

	result := []UserProfile{}
	err := DB.Read(ctx, &result, "SELECT * FROM user_profiles")
	assert.Nil(t, err)
	assert.Equal(t, len(result), 3)
	assert.Equal(t, result[0].Name, "Nova")
	assert.Equal(t, result[0].Bio, "Photographer")
	assert.Equal(t, result[0].Email, "nova@roadbeats.com")
	assert.Equal(t, result[1].Name, "Azer")
	assert.Equal(t, result[1].Bio, "Engineer")
	assert.Equal(t, result[1].Email, "azer@roadbeats.com")
	assert.Equal(t, result[2].Name, "Hola")
	assert.Equal(t, result[2].Bio, "")
	assert.Equal(t, result[2].Email, "hola@roadbeats.com")

	nova := UserProfile{}
	err = DB.Read(ctx, &nova, "SELECT * FROM user_profiles WHERE name=?", "Nova")
	assert.Nil(t, err)
	assert.Equal(t, nova.Name, "Nova")
	assert.Equal(t, nova.Bio, "Photographer")
	assert.Equal(t, nova.Email, "nova@roadbeats.com")

	DB.DropTables(ctx, UserProfile{})
}

func TestScanningToCustomValues(t *testing.T) {
	ctx := context.Background()

	assert.Nil(t, CreateUserProfiles(ctx))

	names := []string{}
	err := DB.Read(ctx, &names, "SELECT name FROM user_profiles ORDER BY id ASC")
	assert.Nil(t, err)
	assert.Equal(t, len(names), 3)
	assert.Equal(t, names[0], "Nova")
	assert.Equal(t, names[1], "Azer")

	name := ""
	err = DB.Read(ctx, &name, "SELECT name FROM user_profiles WHERE id=1")
	assert.Nil(t, err)
	assert.Equal(t, name, "Nova")

	count := 0
	err = DB.Read(ctx, &count, "SELECT COUNT(id) FROM user_profiles")
	assert.Nil(t, err)
	assert.Equal(t, count, 3)

	DB.DropTables(ctx, UserProfile{})
}

func TestScanningToNullTypes(t *testing.T) {
	ctx := context.Background()

	assert.Nil(t, CreateUserProfiles(ctx))

	nova := UserProfileNull{}
	err := DB.Read(ctx, &nova, "SELECT * FROM user_profiles WHERE name = ?", "Nova")
	assert.Nil(t, err)

	assert.Equal(t, nova.Id.Int64, int64(1))
	assert.Equal(t, nova.Name.String, "Nova")
	assert.Equal(t, nova.Bio.String, "Photographer")
	assert.Equal(t, nova.Email.String, "nova@roadbeats.com")

	DB.DropTables(ctx, UserProfile{})
}

func TestUnexistingFields(t *testing.T) {
	ctx := context.Background()

	assert.Nil(t, CreateUserProfiles(ctx))

	nova := UserProfile{}
	err := DB.Read(ctx, &nova, "SELECT u.*, COUNT(u.id) as ucount FROM user_profiles u WHERE name=? GROUP BY u.id", "Nova")
	assert.Nil(t, err)
	assert.Equal(t, nova.Name, "Nova")
	assert.Equal(t, nova.Bio, "Photographer")
	assert.Equal(t, nova.Email, "nova@roadbeats.com")

	DB.DropTables(ctx, UserProfile{})
}

func CreateUserProfiles(ctx context.Context) error {
	if err := DB.ResetTables(ctx, UserProfile{}); err != nil {
		return err
	}

	if err := DB.Create(ctx, UserProfile{
		Name:       "Nova",
		Bio:        "Photographer",
		Email:      "nova@roadbeats.com",
		Attachment: []byte("{ \"nova\": \"bar\" }"),
	}); err != nil {
		return err
	}

	if err := DB.Create(ctx, UserProfile{
		Name:       "Azer",
		Bio:        "Engineer",
		Email:      "azer@roadbeats.com",
		Attachment: []byte("{ \"azer\": \"bar\" }"),
	}); err != nil {
		return err
	}

	if err := DB.Create(ctx, UserProfile{
		Name:       "Hola",
		Email:      "hola@roadbeats.com",
		Attachment: []byte("{ \"hola\": \"bar\" }"),
	}); err != nil {
		return err
	}

	return nil
}

func BenchmarkRead(b *testing.B) {
	ctx := context.Background()

	require.NoError(b, CreateUserProfiles(ctx))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var user UserProfile
		err := DB.Read(ctx, &user, "SELECT * FROM user_profiles LIMIT 1")
		require.NoError(b, err)
		assert.NotZero(b, user.Id)
		assert.NotZero(b, user.Name)
	}
}
