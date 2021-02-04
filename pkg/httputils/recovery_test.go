package httputils

import (
	"context"
	"fmt"
	"github.com/GreenStage/kingfish/pkg/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecoveryHandler_HandlesPanicByLogging(t *testing.T) {
	mockHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		a := []int{}
		fmt.Print(a[2])
	})
	recovery := RecoveryHandler(mockHandler)

	core, recorded := observer.New(zapcore.InfoLevel)

	contextInjectHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		ctx := logger.ToContext(context.Background(), zap.New(core))
		recovery.ServeHTTP(writer, request.WithContext(ctx))
	})

	contextInjectHandler.ServeHTTP(httptest.NewRecorder(), &http.Request{})

	found := false
	searchFor := "panic"
	for _, log := range recorded.All() {
		found = found || strings.Contains(log.Message, searchFor)
		for _, field := range log.Context {
			found = found || strings.Contains(field.String, searchFor)
			found = found || strings.Contains(fmt.Sprintf("%v", field.Interface), searchFor)
		}
	}
}
