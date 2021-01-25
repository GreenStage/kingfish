package handlers

import (
	"github.com/GreenStage/kingfish/internal/db"
	"github.com/GreenStage/kingfish/internal/logger"
	"go.uber.org/zap"
	"net/http"
)

var LoginConfig = struct {
	DriverKey, HostNameKey, UsernameKey, PasswordKey, DatabaseKey string
}{
	DriverKey:   "driver",
	HostNameKey: "hostname",
	UsernameKey: "username",
	PasswordKey: "password",
	DatabaseKey: "db",
}

func (r *router) GetLoginPage(w http.ResponseWriter, req *http.Request) {
	err := r.render.HTML(w, http.StatusOK, "loginPage", map[string]interface{}{
		"config":          LoginConfig,
		sessionExpiredKey: req.URL.Query().Get(sessionExpiredKey),
	})
	if err != nil {
		logger.FromContext(req.Context()).Error("error rendering home template", zap.Error(err))
	}
}

func (r *router) PostLoginForm(w http.ResponseWriter, req *http.Request) {
	log := logger.FromContext(req.Context())

	if contentType := req.Header.Get("Content-Type"); contentType != "application/x-www-form-urlencoded" {
		log.Error("invalid content type: " + contentType)
		w.WriteHeader(http.StatusUnsupportedMediaType)
	}

	if err := req.ParseForm(); err != nil {
		log.Error("could not parse form", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data := sessionData{
		Driver: req.FormValue(LoginConfig.DriverKey),
		ConnConfig: db.ConnectionConfig{
			Username: req.FormValue(LoginConfig.UsernameKey),
			Password: req.FormValue(LoginConfig.PasswordKey),
			Dbname:   req.FormValue(LoginConfig.DatabaseKey),
			Host:     req.FormValue(LoginConfig.HostNameKey),
		},
	}

	driver, ok := r.config.Drivers[data.Driver]
	if !ok {
		log.Error("driver not found", zap.String("driver", LoginConfig.DriverKey))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	testSession, err := driver.NewConnection(data.ConnConfig)
	if err != nil {
		log.Error("could not connect to db", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer testSession.Close()

	if err := r.writeSession(w, data); err != nil {
		log.Error("could not write session config", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", "home")
	w.WriteHeader(http.StatusFound)
}
