## CRUD

A minimalistic relational database library for Go, with simple and familiar interface.

* [Install](#install)
* [Initialize](#initialize)
* [Define](#define)
  * [Create & Drop Tables](#create-drop--tables)
* CRUD:
  * [Create](#create)
  * [Read](#read)
    * [Reading a single row](#reading-a-single-row)
    * [Reading multiple rows](#reading-multiple-rows)
    * [Scanning to custom values](#scanning-to-custom-values)
  * [Update](#update)
  * [MustUpdate](#mustupdate)
  * [Delete](#delete)
  * [MustDelete](#mustdelete)
  * [Transactions](#transactions)
* [Logs](#logs)
* [Transactions](#transactions)
* [Custom Queries](#custom-queries)
* [Running Tests](#running-tests)
* [Why another ORMish library for Go?](#why-another-ormish-library-for-go)

## Install

```bash
$ go get github.com/azer/crud
```

## Initialize

```go
import (
  "github.com/azer/crud"
  _ "github.com/go-sql-driver/mysql"
)

var DB *crud.DB

func init () {
  DB, err := crud.Connect("mysql", os.Getenv("DATABASE_URL"))
  err := DB.Ping()
}
```

## Define

```go
type User struct {
  Id int `sql:"auto-increment primary-key"`
  FirstName string
  LastName string
  ProfileId int
}

type Profile struct {
  Id int `sql:"auto-increment primary-key"`
  Bio string `sql:"text"`
}
```

You don't need to specify SQL field names manually. CRUD will automatically convert column names like "FirstName" to "first_name" for you. You can still choose custom names though;

```go
type Post {
  Slug string `sql:"name=slug_id varchar(255) primary-key required"`
}
```

##### Create & Drop Tables

`CreateTables` takes list of structs and makes sure they exist in the database.

```go
err := DB.CreateTables(User{}, Profile{})

err := DB.DropTables(User{}, Profile{})
```

## Create

```go
user := &User{1, "Foo", "Bar", 1}
err := DB.Create(user)
```

## Read

You can read single/multiple rows, or custom values, with the `Read` method.

##### Reading a single row:

```go
user := User{}
err := DB.Read(user, "WHERE id = ?", 1) // You can type the full query if preferred.
// => SELECT * FROM users WHERE id = 1

fmt.Println(user.Name)
// => Foo
```

##### Reading multiple rows:

```go
users := []*User{}

err := DB.Read(&users)
// => SELECT * FROM users

fmt.Println(len(users))
// => 10
```

##### Scanning to custom values:

```go
names := []string{}
err := DB.Read(&names, "SELECT name FROM users")
```

```
name := ""
err := DB.Read(&name, "SELECT name FROM users WHERE id=1")
```

```go
totalUsers := 0
err := DB.Read(&totalUsers, "SELECT COUNT(id) FROM users"
```

## Update

```go
user := &User{}
err := DB.Read(user, "WHERE id = ?", 1)

user.Name = "Yolo"
err := DB.Update(user)
```

## MustUpdate

Same as [Update](#update), returns error when there is no matching row.

```go
err := DB.MustUpdate(&User{
  Id: 123,
  Name: "Foo"
})
```

## Delete

```go
err := DB.Delete(&User{
  Id: 1
})
```

## MustDelete

Same as [Delete](#delete), returns error when there is no matching row.

```
err := DB.MustDelete(&User{
  Id: 1
})
```

## Transactions

Use `Begin` method of a `crud.DB` instance to create a new transaction. Each transaction will provide you following methods;

* Commit
* Rollback
* Exec
* Query
* Create
* Read
* Update
* MustUpdate
* Delete
* MustDelete

```go
tx, err := DB.Begin()

err := tx.Create(&User{
  Name: "yolo"
})

err := tx.Delete(&User{
  Id: 123
})

err := tx.Commit()
```

## Logs

If you want to see crud's internal logs, specify `crud` in the `LOG` environment variable when you run your app. For example;

```
$ LOG=crud go run myapp.go
```

[(More info about how crud's logging work)](http://github.com/azer/logger)

## Custom Queries

````go
result, err := DB.Query("DROP DATABASE yolo") // or .Exec
````

## Running Tests

```bash
DATABASE_URL="?" go test ./...
```

## Why another ORMish library for Go?

* Simplicity, taking more advantage of `reflect` library to keep the API simple.
* Building less things with more complete abstractions
* Handling errors in an idiomatic way
* Good test coverage
* Modular & reusable code
* Making less unsafe assumptions. e.g: not mapping structs to SQL rows by column index.
