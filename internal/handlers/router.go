package handlers

import (
	"github.com/GreenStage/kingfish/internal/db"
	"github.com/GreenStage/kingfish/pkg/httputils"
	"github.com/go-chi/chi"
	"net/http"
	"time"
)

type Driver interface {
	NewConnection(config db.ConnectionConfig) (session db.Connection, err error)
}

type Config struct {
	Drivers              map[string]Driver
	SessionEncryptionKey []byte
	SessionIdleLifetime  time.Duration
}

type sessionCtxKey struct{}

type router struct {
	config Config
	chiMux *chi.Mux
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.chiMux.ServeHTTP(w, req)
}

func NewRouter(conf Config) *router {
	mux := chi.NewMux()
	mux.Use(httputils.RecoveryHandler)

	r := router{
		config: conf,
		chiMux: mux,
	}

	mux.Post("/login", r.Login)

	mux.Group(func(subMux chi.Router) {
		subMux.Use(r.sessionReader)

		subMux.Get("/tables", r.GetTables)
		subMux.Post("/query", r.RunQuery)
	})

	return &r
}
