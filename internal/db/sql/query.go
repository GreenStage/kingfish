package sql

import (
	"database/sql"
	"github.com/GreenStage/kingfish/internal/db"
)

type TypeToPtrConverter func(columnType *sql.ColumnType) interface{}

func ParseRows(rows *sql.Rows, converter TypeToPtrConverter) (db.QueryResult, error) {
	var (
		err error
		res db.QueryResult
	)

	res.Header, err = rows.Columns()
	if err != nil {
		return db.QueryResult{}, err
	}

	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return db.QueryResult{}, err
	}

	for rows.Next() {
		rowPtrs := make([]interface{}, len(res.Header))
		for i, _ := range colTypes {
			rowPtrs[i] = converter(colTypes[i])
		}

		if err := rows.Scan(rowPtrs...); err != nil {
			return db.QueryResult{}, err
		}

		res.Rows = append(res.Rows, rowPtrs)
	}

	return res, nil
}
