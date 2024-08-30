package crud_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/azer/crud/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/mariadb"
)

var DB *crud.DB

func TestMain(m *testing.M) {
	var exit int
	defer func() { os.Exit(exit) }() // allow unroll defer stack with Exit

	errf := func(format string, a ...any) {
		exit = 1
		fmt.Fprintf(os.Stderr, format, a...)
	}

	cont, err := mariadb.Run(context.Background(), "mariadb")
	if err != nil {
		errf("run container: %v", err)
		return
	}
	defer cont.Terminate(context.Background())

	connStr, err := cont.ConnectionString(context.Background())
	if err != nil {
		errf("get connection string: %v", err)
		return
	}

	DB, err = crud.Connect("mysql", connStr)
	if err != nil {
		errf("connect: %v", err)
		return
	}

	exit = m.Run()
}

type UserProfile struct {
	Id         int    `json:"id" sql:"auto-increment primary-key required"`
	Name       string `json:"name" sql:"required"`
	Bio        string `json:"bio" sql:"type=text"`
	Email      string `json:"e-mail" sql:"name=email unique"`
	Attachment []byte `json:"attachment"`
	Modified   int64  `json:"modified" sql:"name=modified_col"`
}

type UserProfileNull struct {
	Id       sql.NullInt64  `json:"id" sql:"auto-increment primary-key required"`
	Name     sql.NullString `json:"name" sql:"required"`
	Bio      sql.NullString `json:"bio" sql:"type=text"`
	Email    sql.NullString `json:"e-mail" sql:"name=email"`
	Modified sql.NullInt64  `json:"modified" sql:"name=modified"`
}

type Mixed struct {
	Id        int    `json:"-" sql:"  primary-key auto-increment unsigned name=id table-name=__mixed__ "`
	UserId    int    `json:"-" valid:"User.Id~Specified user was not found" sql:" name=user_id"`
	Secret    string `json:"-" valid:"required" sql:" name=secret"`
	CreatedAt int64  `json:"-" sql:"default=0 name=created_at"`
	UpdatedAt int64  `json:"-" sql:"default=0 name=updated_at"`
}

type PostCategory string

const (
	PostCategoryFood       PostCategory = "food"
	PostCategoryDrink      PostCategory = "drink"
	PostCategoryWithHyphen PostCategory = "with-hyphen"
)

type Post struct {
	Id        int          `json:"id" sql:"auto-increment primary-key required table-name=renamed_posts"`
	Title     string       `json:"title"`
	Text      string       `json:"text"`
	Category  PostCategory `json:"post_category" sql:"name=post_category enum('food','drink','with-hyphen')"`
	CreatedAt time.Time    `json:"created_at"`
}

type Foo struct {
	Id     int
	APIKey string
	YOLO   bool
	Beast  string
}

type EmbeddedFoo struct {
	Foo
	Span int
	Eggs string
}

type FooSlice []Foo
type FooPTRSlice []*Foo

type CustomTableName struct {
	Foo int `sql:"table-name=yolo"`
}

func TestPing(t *testing.T) {
	assert.Nil(t, DB.Ping())
}

func TestExecuteSQL(t *testing.T) {
	result, err := DB.Client.Exec("SHOW TABLES LIKE 'shouldnotexist'")
	assert.Nil(t, err)

	l, err := result.LastInsertId()
	assert.Equal(t, err, nil)
	assert.Equal(t, l, int64(0))

	a, err := result.RowsAffected()
	assert.Equal(t, err, nil)
	assert.Equal(t, a, int64(0))
}

func TestCreateTables(t *testing.T) {
	err := DB.CreateTables(UserProfile{}, Post{})
	assert.Nil(t, err)
	assert.True(t, DB.CheckIfTableExists("user_profiles"))
	assert.True(t, DB.CheckIfTableExists("renamed_posts"))
}

func TestDropTables(t *testing.T) {
	err := DB.DropTables(UserProfile{}, Post{})
	assert.Nil(t, err)
	assert.False(t, DB.CheckIfTableExists("user_profiles"))
	assert.False(t, DB.CheckIfTableExists("posts"))
}
