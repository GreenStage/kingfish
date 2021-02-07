package logger

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"testing"
)

func TestIsStringLogged_IsLogged(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	log := zap.New(core)

	log.Info("test", zap.String("testKey", "testMessage"))
	log.Error("test", zap.Error(errors.New("testError")))

	assert.True(t, IsStringLogged(recorded, "testMessage"))
	assert.True(t, IsStringLogged(recorded, "testError"))
	assert.False(t, IsStringLogged(recorded, "NotLogged"))
}
