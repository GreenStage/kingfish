package logger

import (
	"fmt"
	"go.uber.org/zap/zaptest/observer"
	"strings"
)

func IsStringLogged(recorded *observer.ObservedLogs, str string) bool {
	found := false
	for _, log := range recorded.All() {
		found = found || strings.Contains(log.Message, str)
		for _, field := range log.Context {
			found = found || strings.Contains(field.String, str)
			found = found || strings.Contains(fmt.Sprintf("%v", field.Interface), str)
		}
	}
	return found
}
