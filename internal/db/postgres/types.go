package postgres

import (
	sqlDriver "database/sql"
	"github.com/GreenStage/kingfish/internal/db/sql"
)

func convertType(columnType *sqlDriver.ColumnType) interface{} {
	switch columnType.DatabaseTypeName() {
	case "NAME":
		var a sqlDriver.NullString
		return &a
	}

	return sql.NewPtrForType(columnType)
}
