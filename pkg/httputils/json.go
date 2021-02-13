package httputils

import (
	"context"
	"encoding/json"
	"github.com/GreenStage/kingfish/pkg/logger"
	"go.uber.org/zap"
	"net/http"
)

type JsonWriteOption func(w http.ResponseWriter)

func WithCode(code int) JsonWriteOption {
	return func(w http.ResponseWriter) {
		w.WriteHeader(code)
	}
}

func WriteJson(ctx context.Context, w http.ResponseWriter, data interface{}, options ...JsonWriteOption) {
	out, err := json.Marshal(data)
	if err != nil {
		logger.FromContext(ctx).Error("could not marshal json response", zap.Error(err))
		w.WriteHeader(500)
		return
	}

	for _, opt := range options {
		opt(w)
	}

	if _, err := w.Write(out); err != nil {
		logger.FromContext(ctx).Error("could not write json response", zap.Error(err))
	}
}
