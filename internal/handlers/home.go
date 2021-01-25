package handlers

import (
	"github.com/GreenStage/kingfish/internal/db"
	"github.com/GreenStage/kingfish/internal/logger"
	"go.uber.org/zap"
	"net/http"
)

type HomeTmplData struct {
	Username, DbName, Hostname string
	Tables                     []db.Table
}

func (r *router) GetHome(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	sess, ok := ctx.Value(sessionCtxKey).(sessionData)
	if !ok {
		logger.FromContext(ctx).Debug("no session found in context")
		return
	}

	tables, err := sess.Conn.GetTables(req.Context())
	if err != nil {
		logger.FromContext(ctx).Error("error running query", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = r.render.HTML(w, http.StatusOK, "homePage", HomeTmplData{
		Username: sess.ConnConfig.Username,
		DbName:   sess.ConnConfig.Dbname,
		Hostname: sess.ConnConfig.Host,
		Tables:   tables,
	})
	if err != nil {
		logger.FromContext(ctx).Error("error rendering home template", zap.Error(err))
	}
}
