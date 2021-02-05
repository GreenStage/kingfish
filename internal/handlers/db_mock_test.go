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

func (m *MockDbConnection) Close() error {
	args := m.MethodCalled("Close")
	return args.Error(0)
}
