package httputils

import (
	"context"
	"encoding/json"
	"github.com/GreenStage/kingfish/pkg/logger"
	"go.uber.org/zap"
	"net/http"
)

func WriteJson(ctx context.Context, w http.ResponseWriter, data interface{}) {
	out, err := json.Marshal(data)
	if err != nil {
		logger.FromContext(ctx).Error("could not marshal json response", zap.Error(err))
		w.WriteHeader(500)
		return
	}
	if _, err := w.Write(out); err != nil {
		logger.FromContext(ctx).Error("could not write json response", zap.Error(err))
	}
}
