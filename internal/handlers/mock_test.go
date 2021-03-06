package handlers

import (
	"context"
	"github.com/GreenStage/kingfish/internal/db"
	"github.com/stretchr/testify/mock"
)

type MockDbConnection struct {
	mock.Mock
}

func (m *MockDbConnection) GetTables(ctx context.Context) ([]db.Table, error) {
	args := m.MethodCalled("GetTables", ctx)
	if args.Get(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.Table), nil
}

func (m *MockDbConnection) Query(ctx context.Context, queryStr string) (db.QueryResult, error) {
	args := m.MethodCalled("Query", ctx, queryStr)
	if args.Get(1) != nil {
		return args.Get(0).(db.QueryResult), args.Error(1)
	}
	return args.Get(0).(db.QueryResult), nil
}

func (m *MockDbConnection) Close() error {
	args := m.MethodCalled("Close")
	return args.Error(0)
}

type MockDriver struct {
	mock.Mock
}

func (m *MockDriver) NewConnection(config db.ConnectionConfig) (db.Connection, error) {
	args := m.MethodCalled("NewConnection", config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(db.Connection), args.Error(1)
}
