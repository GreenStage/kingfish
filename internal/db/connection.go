package db

import "context"

type ConnectionConfig struct {
	Username string
	Password string
	Dbname   string
	Host     string
	Persist  bool
}

type Connection interface {
	GetTables(ctx context.Context) ([]Table, error)
	Close() error
}
