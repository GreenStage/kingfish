package sql

import (
	"database/sql"
	"reflect"
)

// TypeInstantiator defines a function that each driver should implement in order
// to instantiate a typed variable pointer according to a DB column type.
type TypeInstantiator func(colType *sql.ColumnType) interface{}

func Instantiate(colType *sql.ColumnType) interface{} {
	return reflect.New(colType.ScanType()).Interface()
}
