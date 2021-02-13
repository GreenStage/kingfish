package db

import "context"

// ConnectionConfig defines the input configuration of a Connection
type ConnectionConfig struct {
	Username string
	Password string
	Dbname   string
	Host     string
	Persist  bool
}

// Connection defines the db interface each driver should implement
type Connection interface {
	GetTables(ctx context.Context) ([]Table, error)
	Query(ctx context.Context, queryStr string) (QueryResult, error)
	Close() error
}
