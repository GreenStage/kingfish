package handlers

import (
	"context"
	"errors"
	"github.com/GreenStage/kingfish/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_GetTables_HandlesNoSessionGracefully(t *testing.T) {
	router := NewRouter(Config{})
	w := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		router.GetTables(w, &http.Request{})
	})
	assert.Equal(t, 500, w.Code)
}

func Test_GetTables_LogsAndHandlesDBErrorGracefully(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	ctx := logger.ToContext(context.Background(), zap.New(core))

	conn := &MockDbConnection{}
	conn.On("GetTables", mock.Anything).Return(nil, errors.New("boom!"))

	ctx = context.WithValue(ctx, sessionCtxKey, sessionData{
		Conn: conn,
	})
	r := (&http.Request{}).WithContext(ctx)

	router := NewRouter(Config{})
	w := httptest.NewRecorder()
	assert.NotPanics(t, func() {
		router.GetTables(w, r)
	})

	assert.Equal(t, 500, w.Code)

	logFound := logger.IsStringLogged(recorded, "boom!")
	assert.True(t, logFound)
}
