package handlers

import (
	"github.com/GreenStage/kingfish/internal/db"
	"github.com/go-chi/chi"
	"github.com/thedevsaddam/renderer"
	"net/http"
	"time"
)

type Driver interface {
	NewConnection(config db.ConnectionConfig) (session db.Connection, err error)
}

type Config struct {
	Drivers              map[string]Driver
	TemplatesDir         string
	AssetsDir            string
	SessionEncryptionKey []byte
	MaxSessionLifetime   time.Duration
}

type router struct {
	config Config
	render *renderer.Render
	chiMux *chi.Mux
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.chiMux.ServeHTTP(w, req)
}

func NewRouter(conf Config) *router {
	mux := chi.NewMux()
	render := renderer.New(
		renderer.Options{
			ParseGlobPattern: conf.TemplatesDir + "/*.gohtml",
		})

	r := router{
		config: conf,
		chiMux: mux,
		render: render,
	}

	staticFileDirectory := http.Dir(conf.AssetsDir)
	staticFileHandler := http.StripPrefix("/assets/", http.FileServer(staticFileDirectory))

	mux.Get("/assets/*", staticFileHandler.ServeHTTP)

	mux.Get("/login", r.GetLoginPage)
	mux.Post("/login", r.PostLoginForm)

	mux.Group(func(subMux chi.Router) {
		subMux.Use(r.sessionReader)

		subMux.Get("/home", r.GetHome)
	})

	return &r
}
