package handlers

import (
	"github.com/GreenStage/kingfish/pkg/httputils"
	"github.com/GreenStage/kingfish/pkg/logger"
	"go.uber.org/zap"
	"net/http"
)

type getTableResponse struct {
	Name          string `json:"name"`
	AproxRowCount int64  `json:"aprox_row_count"`
}

func (r *router) GetTables(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	sess, ok := ctx.Value(sessionCtxKey{}).(sessionData)
	if !ok {
		logger.FromContext(ctx).Debug("no session found in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tables, err := sess.Conn.GetTables(ctx)
	if err != nil {
		logger.FromContext(ctx).Error("error running query", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tableResp := make([]getTableResponse, 0, len(tables))
	for _, t := range tables {
		tableResp = append(tableResp, getTableResponse{
			Name:          t.Name,
			AproxRowCount: t.AproxRowCount,
		})
	}

	httputils.WriteJson(ctx, w, tableResp)
}
