package handlers

import (
	"encoding/json"
	"github.com/GreenStage/kingfish/internal/db"
	"github.com/GreenStage/kingfish/pkg/httputils"
	"github.com/GreenStage/kingfish/pkg/logger"
	"go.uber.org/zap"
	"net/http"
)

type loginRequest struct {
	Driver   string `json:"driver"`
	Hostname string `json:"hostname"`
	Dbname   string `json:"dbname"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

func (r *router) Login(w http.ResponseWriter, req *http.Request) {
	log := logger.FromContext(req.Context())

	var loginRequest loginRequest
	if err := json.NewDecoder(req.Body).Decode(&loginRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data := sessionData{
		Driver: loginRequest.Driver,
		ConnConfig: db.ConnectionConfig{
			Username: loginRequest.Username,
			Password: loginRequest.Password,
			Dbname:   loginRequest.Dbname,
			Host:     loginRequest.Hostname,
		},

		// TODO : this should be configurable by request
		IdleLifeTime: r.config.SessionIdleLifetime,
	}

	driver, ok := r.config.Drivers[data.Driver]
	if !ok {
		log.Error("driver not supported", zap.String("driver", data.Driver))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	testSession, err := driver.NewConnection(data.ConnConfig)
	if err != nil {
		log.Error("could not connect to db", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest) // TODO: check what error code suits this scenario best
		return
	}
	defer testSession.Close()

	token, err := r.newSessionToken(data)
	if err != nil {
		log.Error("could not get session token", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	httputils.WriteJson(w, loginResponse{
		AccessToken: token,
		ExpiresIn:   int64(data.IdleLifeTime.Seconds()),
	})
}
