package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/GreenStage/kingfish/internal/db"
	"github.com/GreenStage/kingfish/internal/logger"
	"github.com/GreenStage/kingfish/pkg/aes"
	"go.uber.org/zap"
	"net/http"
)

const (
	clientKeyLen       = 32
	clientKeyCookieKey = "CLIENT_KEY"
	sessionCookieKey   = "SESSION_CFG"
	sessionCtxKey      = "SESSION_CFG"
	sessionExpiredKey  = "session_expired"
)

type sessionData struct {
	Driver     string
	ConnConfig db.ConnectionConfig
	Conn       db.Connection
}

func (r *router) writeSession(w http.ResponseWriter, data sessionData) error {
	confBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("could not marshal session config into valid json: %v", err)
	}

	clientKey, err := aes.GenerateKey(clientKeyLen)
	if err != nil {
		return fmt.Errorf("could generate client key: %v", err)
	}

	// encrypt first with the generated client key
	clientEncryptedSess, err := aes.Encrypt(confBytes, clientKey)
	if err != nil {
		return fmt.Errorf("error encrypting session config with server key: %v", err)
	}

	// encrypt with the server key
	serverEncryptedSess, err := aes.Encrypt(clientEncryptedSess, r.config.SessionEncryptionKey)
	if err != nil {
		return fmt.Errorf("error encrypting session config with server key: %v", err)
	}

	base64ClientKey := make([]byte, base64.StdEncoding.EncodedLen(clientKeyLen))
	base64.StdEncoding.Encode(base64ClientKey, clientKey)

	// set the client key as cookie
	http.SetCookie(w, &http.Cookie{
		Name:     clientKeyCookieKey,
		Value:    string(base64ClientKey),
		MaxAge:   int(r.config.MaxSessionLifetime.Seconds()),
		HttpOnly: true, // prevent XSS
	})

	base64Session := make([]byte, base64.StdEncoding.EncodedLen(len(serverEncryptedSess)))
	base64.StdEncoding.Encode(base64Session, serverEncryptedSess)

	// set the encrypted session as a key
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieKey,
		Value:    string(base64Session),
		MaxAge:   int(r.config.MaxSessionLifetime.Seconds()),
		HttpOnly: true, // prevent XSS
	})

	return nil
}

func (r *router) sessionReader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var err error
		defer func() {
			if err != nil {
				dropSession(w)
			}
		}()

		ctx := req.Context()

		sessionCookie, err := req.Cookie(sessionCookieKey)
		if err != nil {
			logger.FromContext(ctx).Debug("cookie " + sessionCookieKey + " not found in request")
			return
		}

		clientKeyCookie, err := req.Cookie(clientKeyCookieKey)
		if err != nil {
			logger.FromContext(ctx).Debug("cookie " + clientKeyCookieKey + " not found in request")
			return
		}

		clientKey, err := base64.StdEncoding.DecodeString(clientKeyCookie.Value)
		if err != nil {
			logger.FromContext(ctx).Debug("could not decode client key from base64", zap.Error(err))
			return
		}

		serverEncryptedSession, err := base64.StdEncoding.DecodeString(sessionCookie.Value)
		if err != nil {
			logger.FromContext(ctx).Debug("could not decode session config from base64", zap.Error(err))
			return
		}

		clientEncryptedSession, err := aes.Decrypt(serverEncryptedSession, r.config.SessionEncryptionKey)
		if err != nil {
			logger.FromContext(ctx).Debug("could not decrypt session config with server key", zap.Error(err))
			return
		}

		jsonSession, err := aes.Decrypt(clientEncryptedSession, clientKey)
		if err != nil {
			logger.FromContext(ctx).Debug("could not decrypt session config with client key", zap.Error(err))
			return
		}

		var data sessionData
		if err := json.Unmarshal(jsonSession, &data); err != nil {
			logger.FromContext(ctx).Debug("could not unmarshal session config from json", zap.Error(err))
			return
		}

		driver, ok := r.config.Drivers[data.Driver]
		if !ok {
			logger.FromContext(ctx).Error("no session found in context")
			return
		}

		conn, err := driver.NewConnection(data.ConnConfig)
		if err != nil {
			logger.FromContext(ctx).Error("error establishing connection", zap.Error(err))
			return
		}

		defer func() {
			if e := conn.Close(); e != nil {
				logger.FromContext(ctx).Error("error closing db connection", zap.Error(e))
			}
		}()

		data.Conn = conn
		req = req.WithContext(context.WithValue(ctx, sessionCtxKey, data))
		next.ServeHTTP(w, req)
	})
}

func dropSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   sessionCookieKey,
		MaxAge: -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:   clientKeyCookieKey,
		MaxAge: -1,
	})
	w.Header().Set("Location", "/login?"+sessionExpiredKey+"=1")
	w.WriteHeader(http.StatusFound)
}
