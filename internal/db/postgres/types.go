package postgres

import (
	"database/sql"
	kfSql "github.com/GreenStage/kingfish/internal/db/sql"
	"github.com/GreenStage/kingfish/internal/types"
)

func TypeInstantiator(colType *sql.ColumnType) interface{} {
	switch colType.DatabaseTypeName() {
	case "NUMERIC":
		var v types.Numeric
		return &v
	default:
		return kfSql.Instantiate(colType)
	}
}
