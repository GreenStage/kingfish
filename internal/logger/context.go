package logger

import (
	"context"
	"go.uber.org/zap"
)

var (
	loggerCtxKey = struct{}{}
)

func FromContext(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(loggerCtxKey).(*zap.Logger); ok {
		return logger
	}
	return Default
}

func ToContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey, logger)
}
