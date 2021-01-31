package postgres

import (
	sqlDriver "database/sql"
	_ "github.com/lib/pq"
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
