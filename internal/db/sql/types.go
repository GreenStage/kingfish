package sql

import (
	"database/sql"
	"reflect"
)

func NewPtrForType(columnType *sql.ColumnType) interface{} {
	r := reflect.New(columnType.ScanType()).Interface()
	return &r
}
