package postgres

import (
	"context"
	sqlDriver "database/sql"
	"errors"
	"github.com/GreenStage/kingfish/internal/db"
	"github.com/GreenStage/kingfish/internal/db/sql"
	"github.com/GreenStage/kingfish/internal/logger"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type Connection struct {
	db *sqlDriver.DB
}

func connect(psqlInfo string) (*Connection, error) {
	db, err := sqlDriver.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	return &Connection{
		db: db,
	}, db.Ping()
}

func (s *Connection) Close() error {
	return s.db.Close()
}

func (s *Connection) GetTables(ctx context.Context) ([]db.Table, error) {
	result, err := s.Query(ctx, `
SELECT tsc.tablename, cl.reltuples::bigint
	FROM pg_catalog.pg_tables tsc
		LEFT JOIN pg_catalog.pg_class cl ON tsc.tablename = cl.relname
	WHERE tsc.schemaname != 'pg_catalog' AND  tsc.schemaname != 'information_schema';
`,
	)
	if err != nil {
		return nil, err
	}

	out := make([]db.Table, 0, len(result.Rows))
	for _, row := range result.Rows {
		name, _ := row[0].(*sqlDriver.NullString)
		if name == nil {
			return nil, errors.New("could not cast table name pointer")
		}
		out = append(out, db.Table{
			Name: name.String,
		})
	}
	return out, nil
}

func (s *Connection) Query(ctx context.Context, querystr string) (db.QueryResult, error) {
	rows, err := s.db.Query(querystr)
	if err != nil {
		return db.QueryResult{}, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			logger.FromContext(ctx).Error("error closing postgres rows", zap.Error(err))
		}
	}()

	return sql.ParseRows(rows, convertType)
}
