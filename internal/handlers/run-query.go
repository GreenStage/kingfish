package handlers

import (
	"github.com/GreenStage/kingfish/pkg/httputils"
	"github.com/GreenStage/kingfish/pkg/logger"
	"go.uber.org/zap"
	"io/ioutil"
	"mime"
	"net/http"
)

type responseColumn struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type queryResponse struct {
	Columns []responseColumn `json:"columns,omitempty"`
	Rows    [][]interface{}  `json:"rows,omitempty"`
	Error   string           `json:"error,omitempty"`
}

func (r *router) RunQuery(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	contentType, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		logger.FromContext(ctx).Debug("invalid Content-Type header", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.FromContext(ctx).Debug("could not read request body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var queryStr string
	switch contentType {
	case "text/plain":
		queryStr = string(bodyBytes)
	default:
		// todo: support other content types?
		w.WriteHeader(http.StatusUnsupportedMediaType)
	}

	sess, ok := ctx.Value(sessionCtxKey{}).(sessionData)
	if !ok {
		logger.FromContext(ctx).Debug("no session found in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	result, err := sess.Conn.Query(ctx, queryStr)
	response := queryResponse{
		Columns: make([]responseColumn, len(result.Columns)),
		Rows:    result.Rows,
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Error = err.Error()
	}

	for i, col := range result.Columns {
		response.Columns[i] = responseColumn{
			Name: col.Name,
			Type: col.Type,
		}
	}
	httputils.WriteJson(ctx, w, response)
}
