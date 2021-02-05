package logger

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestLoggerContext_NotfoundReturnsDefaultLogger(t *testing.T) {
	ctx := context.Background()

	log := FromContext(ctx)

	assert.Equal(t, Default, log)
	assert.NotNil(t, log)
}

func TestLoggerContext_ReturnsInjectedLogger(t *testing.T) {
	injected, err := zap.NewProduction()
	assert.NoError(t, err)

	ctx := ToContext(context.Background(), injected)

	log := FromContext(ctx)

	assert.Equal(t, injected, log)
	assert.NotNil(t, log)
}
