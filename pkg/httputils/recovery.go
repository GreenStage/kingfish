package httputils

import (
	"github.com/GreenStage/kingfish/pkg/logger"
	"go.uber.org/zap"
	"net/http"
)

func RecoveryHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				logger.FromContext(request.Context()).Error("panic recovered", zap.Any("panic", r))
			}
		}()
		next.ServeHTTP(writer, request)
	})
}
