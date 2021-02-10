package postgres

import (
	"context"
	"github.com/GreenStage/kingfish/internal/db"
	"github.com/GreenStage/kingfish/internal/db/sql"
	"github.com/GreenStage/kingfish/pkg/logger"
	"go.uber.org/zap"

	_ "github.com/lib/pq"
)

func (s *Connection) GetTables(_ context.Context) ([]db.Table, error) {
	rows, err := s.db.Query(`
SELECT tsc.tablename, cl.reltuples::bigint
	FROM pg_catalog.pg_tables tsc
		LEFT JOIN pg_catalog.pg_class cl ON tsc.tablename = cl.relname
	WHERE tsc.schemaname != 'pg_catalog' AND  tsc.schemaname != 'information_schema';
`,
	)
	if err != nil {
		return nil, err
	}

	var out []db.Table
	for rows.Next() {
		var (
			name  string
			count int64
		)
		if err := rows.Scan(&name, &count); err != nil {
			return nil, err
		}
		out = append(out, db.Table{
			Name:          name,
			AproxRowCount: count,
		})
	}
	return out, nil
}

func (s *Connection) Query(ctx context.Context, queryStr string) (db.QueryResult, error) {
	rows, err := s.db.Query(queryStr)
	if err != nil {
		return db.QueryResult{}, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			logger.FromContext(ctx).Error("error closing postgres rows", zap.Error(err))
		}
	}()

	return sql.ParseRows(rows, TypeInstantiator)
}
