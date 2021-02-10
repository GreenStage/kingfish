package sql

import (
	"database/sql"
	"fmt"
	"github.com/GreenStage/kingfish/internal/db"
)

func ParseRows(rows *sql.Rows, instantiate TypeInstantiator) (db.QueryResult, error) {
	var (
		retval db.QueryResult
		err    error
	)

	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return retval, fmt.Errorf("could not parse rows' types: %v", err)
	}

	retval.Columns = make([]db.ColInfo, 0, len(colTypes))
	for _, colType := range colTypes {
		retval.Columns = append(retval.Columns, db.ColInfo{
			Name: colType.Name(),
			Type: colType.DatabaseTypeName(),
		})
	}

	for rows.Next() {
		row := make([]interface{}, len(retval.Columns))

		for i := range colTypes {
			row[i] = instantiate(colTypes[i])
		}

		if err := rows.Scan(row...); err != nil {
			return retval, fmt.Errorf("error reading row: %v", err)
		}

		retval.Rows = append(retval.Rows, row)
	}

	return retval, nil
}
