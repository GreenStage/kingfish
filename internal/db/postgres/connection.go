package postgres

import (
	sqlDriver "database/sql"

	//pq driver
	_ "github.com/lib/pq"
)

// Connection represents a postgresql connection
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

// Close closes a postgresql connection
func (s *Connection) Close() error {
	return s.db.Close()
}
