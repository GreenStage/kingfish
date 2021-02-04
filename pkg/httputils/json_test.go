package httputils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GreenStage/kingfish/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"io/ioutil"
	"math"
	"net/http/httptest"
	"strings"
	"testing"
)

type intransigentRecorder struct {
	*httptest.ResponseRecorder
	throwError error
}

func (rw *intransigentRecorder) Write(buf []byte) (int, error) {
	if rw.throwError != nil {
		return 0, rw.throwError
	}
	return rw.ResponseRecorder.Write(buf)
}

func mockedRecorder(err error) intransigentRecorder {
	return intransigentRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		throwError:       err,
	}
}

func TestWriteJson(t *testing.T) {
	tests := []struct {
		name     string
		dataIn   interface{}
		bufferIn intransigentRecorder
		wantCode int
		wantBody string
		wantLog  string
	}{
		{
			name:     "invalid json logs and writes 500",
			bufferIn: mockedRecorder(nil),
			dataIn:   math.Inf(1),
			wantCode: 500,
			wantLog:  (&json.UnsupportedValueError{}).Error(),
		},
		{
			name:     "logs error writing body",
			bufferIn: mockedRecorder(errors.New("nil body")),
			dataIn:   "dummy",
			wantLog:  "nil body",
		},
		{
			name:     "ok",
			bufferIn: mockedRecorder(nil),
			dataIn: struct {
				X string `json:"tagkey"`
			}{
				X: "hi",
			},
			wantCode: 200,
			wantBody: "{\"tagkey\":\"hi\"}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, recorded := observer.New(zapcore.InfoLevel)
			ctx := logger.ToContext(context.Background(), zap.New(core))

			WriteJson(ctx, &tt.bufferIn, tt.dataIn)

			if tt.wantCode != 0 {
				assert.Equal(t, tt.wantCode, tt.bufferIn.Code)
			}

			if tt.wantBody != "" {
				bytes, err := ioutil.ReadAll(tt.bufferIn.Body)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantBody, string(bytes))
			}

			found := false

			for _, log := range recorded.All() {
				found = found || strings.Contains(log.Message, tt.wantLog)
				for _, field := range log.Context {
					found = found || strings.Contains(field.String, tt.wantLog)
					found = found || strings.Contains(fmt.Sprintf("%v", field.Interface), tt.wantLog)
				}
			}

			if found == (tt.wantLog == "") {
				t.Fatalf("log found %v, want log %s", found, tt.wantLog)
			}
		})
	}
}
