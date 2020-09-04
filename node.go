package nestedset

import (
	"database/sql"
	"database/sql/driver"
)

type ID interface {
	driver.Valuer
	sql.Scanner
	MustValue() driver.Value
	Raw() interface{}
}

type Id struct {
	Valuer  func() (driver.Value, error)
	Scanner func(v interface{}) error
	Rawer   func() interface{}
}

func (this Id) Value() (driver.Value, error) {
	return this.Valuer()
}

func (this Id) Scan(src interface{}) error {
	return this.Scanner(src)
}

func (this Id) Raw() interface{} {
	return this.Rawer()
}

func (this Id) MustValue() driver.Value {
	return Dv(this)
}

type Node struct {
	ID, ParentID ID
	Lft, Rgt     int64
	Depth        int32
	Path         string
	DenyChindren bool
}
