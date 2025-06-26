package main

import (
	"context"
	"fmt"
	"os"

	"github.com/azer/crud/v2"
	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	Id        int `sql:"auto-increment primary-key"`
	FirstName string
	LastName  string
}

func main() {
	DB, err := crud.Connect("mysql", os.Getenv("DATABASE_URL"), nil)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	if err := DB.CreateTables(ctx, User{}); err != nil {
		panic(err)
	}

	azer := User{
		FirstName: "Azer",
		LastName:  "Koculu",
	}

	if err := DB.Create(ctx, &azer); err != nil {
		panic(err)
	}

	copy := User{}
	if err := DB.Read(ctx, &copy, "SELECT * FROM users WHERE first_name='Azer'"); err != nil {
		panic(err)
	}

	fmt.Println(copy.Id)
}
