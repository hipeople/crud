package crud

import (
	"iter"

	stdsql "database/sql"
)

type Client interface {
	Exec(string, ...interface{}) (stdsql.Result, error)
	Query(string, ...interface{}) (*stdsql.Rows, error)
	Create(interface{}) error
	CreateAndRead(interface{}) error
	Read(interface{}, ...interface{}) error
	ReadIter(interface{}, ...interface{}) (iter.Seq2[interface{}, error], error)
	Update(interface{}) error
	Delete(interface{}) error
}
